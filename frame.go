// Copyright 2014 The Go Authors.
// See https://code.google.com/p/go/source/browse/CONTRIBUTORS
// Licensed under the same terms as Go itself:
// https://code.google.com/p/go/source/browse/LICENSE

package http2

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const frameHeaderLen = 9

type FrameType uint8

// Defined in http://http2.github.io/http2-spec/#rfc.section.11.2
const (
	FrameData         FrameType = 0x0
	FrameHeaders      FrameType = 0x1
	FramePriority     FrameType = 0x2
	FrameRSTStream    FrameType = 0x3
	FrameSettings     FrameType = 0x4
	FramePushPromise  FrameType = 0x5
	FramePing         FrameType = 0x6
	FrameGoAway       FrameType = 0x7
	FrameWindowUpdate FrameType = 0x8
	FrameContinuation FrameType = 0x9
)

var frameName = map[FrameType]string{
	FrameData:         "DATA",
	FrameHeaders:      "HEADERS",
	FramePriority:     "PRIORITY",
	FrameRSTStream:    "RST_STREAM",
	FrameSettings:     "SETTINGS",
	FramePushPromise:  "PUSH_PROMISE",
	FramePing:         "PING",
	FrameGoAway:       "GOAWAY",
	FrameWindowUpdate: "WINDOW_UPDATE",
	FrameContinuation: "CONTINUATION",
}

func (t FrameType) String() string {
	if s, ok := frameName[t]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN_FRAME_TYPE_%d", uint8(t))
}

// Frame-specific FrameHeader flag bits.
const (
	// Settings Frame
	FlagSettingsAck Flags = 0x1

	// Headers Frame
	FlagHeadersEndStream  Flags = 0x1
	FlagHeadersEndSegment Flags = 0x2
	FlagHeadersEndHeaders Flags = 0x4
	FlagHeadersPadded     Flags = 0x8
	FlagHeadersPriority   Flags = 0x20
)

type SettingID uint16

const (
	SettingHeaderTableSize      SettingID = 0x1
	SettingEnablePush           SettingID = 0x2
	SettingMaxConcurrentStreams SettingID = 0x3
	SettingInitialWindowSize    SettingID = 0x4
	SettingMaxFrameSize         SettingID = 0x5
	SettingMaxHeaderListSize    SettingID = 0x6
)

var settingName = map[SettingID]string{
	SettingHeaderTableSize:      "HEADER_TABLE_SIZE",
	SettingEnablePush:           "ENABLE_PUSH",
	SettingMaxConcurrentStreams: "MAX_CONCURRENT_STREAMS",
	SettingInitialWindowSize:    "INITIAL_WINDOW_SIZE",
	SettingMaxFrameSize:         "MAX_FRAME_SIZE",
	SettingMaxHeaderListSize:    "MAX_HEADER_LIST_SIZE",
}

func (s SettingID) String() string {
	if v, ok := settingName[s]; ok {
		return v
	}
	return fmt.Sprintf("UNKNOWN_SETTING_%d", uint8(s))
}

// a frameParser parses a frame. The parser can assume that the Reader will
// not read past the length of a frame (e.g. it acts like an io.LimitReader
// bounded by the FrameHeader.Length)
type frameParser func(FrameHeader, []byte) (Frame, error)

var FrameParsers = map[FrameType]frameParser{
	FrameSettings:     parseSettingsFrame,
	FrameWindowUpdate: parseWindowUpdateFrame,
	FrameHeaders:      parseHeadersFrame,
}

func typeFrameParser(t FrameType) frameParser {
	if f, ok := FrameParsers[t]; ok {
		return f
	}
	return parseUnknownFrame
}

type Flags uint8

func (f Flags) Has(v Flags) bool {
	return (f & v) == v
}

// A FrameHeader is the 9 byte header of all HTTP/2 frames.
//
// See http://http2.github.io/http2-spec/#FrameHeader
type FrameHeader struct {
	valid bool // caller can access []byte fields in the Frame

	Type     FrameType
	Flags    Flags
	Length   uint32 // actually a uint24 max; default is uint16 max
	StreamID uint32
}

func (h FrameHeader) Header() FrameHeader { return h }

func (h FrameHeader) String() string {
	return fmt.Sprintf("[FrameHeader type=%v flags=%v stream=%v len=%v]",
		h.Type, h.Flags, h.StreamID, h.Length)
}

func (h *FrameHeader) checkValid() {
	if !h.valid {
		panic("Frame accessor called on non-owned Frame")
	}
}

func (h *FrameHeader) invalidate() { h.valid = false }

// frame header bytes
var fhBytes = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, frameHeaderLen)
		return &buf
	},
}

func ReadFrameHeader(r io.Reader) (FrameHeader, error) {
	bufp := fhBytes.Get().(*[]byte)
	defer fhBytes.Put(bufp)
	return readFrameHeader(*bufp, r)
}

func readFrameHeader(buf []byte, r io.Reader) (FrameHeader, error) {
	_, err := io.ReadFull(r, buf[:frameHeaderLen])
	if err != nil {
		return FrameHeader{}, err
	}
	return FrameHeader{
		Length:   (uint32(buf[0])<<16 | uint32(buf[1])<<7 | uint32(buf[2])),
		Type:     FrameType(buf[3]),
		Flags:    Flags(buf[4]),
		StreamID: binary.BigEndian.Uint32(buf[5:]) & (1<<31 - 1),
		valid:    true,
	}, nil
}

type Frame interface {
	Header() FrameHeader
	invalidate()
}

// A Framer reads and writes Frames.
type Framer struct {
	r         io.Reader
	lr        io.LimitedReader
	lastFrame Frame
	readBuf   []byte

	w io.Writer
}

// NewFramer returns a Framer that writes frames to w and reads them from r.
func NewFramer(w io.Writer, r io.Reader) *Framer {
	return &Framer{
		w:       w,
		r:       r,
		readBuf: make([]byte, 1<<10),
	}
}

// ReadFrame reads a single frame. The returned Frame is only valid
// until the next call to ReadFrame.
func (fr *Framer) ReadFrame() (Frame, error) {
	if fr.lastFrame != nil {
		fr.lastFrame.invalidate()
	}
	fh, err := readFrameHeader(fr.readBuf, fr.r)
	if err != nil {
		return nil, err
	}
	if uint32(len(fr.readBuf)) < fh.Length {
		fr.readBuf = make([]byte, fh.Length)
	}
	payload := fr.readBuf[:fh.Length]
	if _, err := io.ReadFull(fr.r, payload); err != nil {
		return nil, err
	}
	f, err := typeFrameParser(fh.Type)(fh, payload)
	if err != nil {
		return nil, err
	}
	fr.lastFrame = f
	return f, nil
}

type SettingsFrame struct {
	FrameHeader
	p []byte
}

func parseSettingsFrame(fh FrameHeader, p []byte) (Frame, error) {
	if fh.Flags.Has(FlagSettingsAck) && fh.Length > 0 {
		// When this (ACK 0x1) bit is set, the payload of the
		// SETTINGS frame MUST be empty.  Receipt of a
		// SETTINGS frame with the ACK flag set and a length
		// field value other than 0 MUST be treated as a
		// connection error (Section 5.4.1) of type
		// FRAME_SIZE_ERROR.
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		// SETTINGS frames always apply to a connection,
		// never a single stream.  The stream identifier for a
		// SETTINGS frame MUST be zero (0x0).  If an endpoint
		// receives a SETTINGS frame whose stream identifier
		// field is anything other than 0x0, the endpoint MUST
		// respond with a connection error (Section 5.4.1) of
		// type PROTOCOL_ERROR.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	if len(p)%6 != 0 {
		// Expecting even number of 6 byte settings.
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	f := &SettingsFrame{FrameHeader: fh, p: p}
	if v, ok := f.Value(SettingInitialWindowSize); ok && v > (1<<31)-1 {
		// Values above the maximum flow control window size of 2^31 - 1 MUST
		// be treated as a connection error (Section 5.4.1) of type
		// FLOW_CONTROL_ERROR.
		return nil, ConnectionError(ErrCodeFlowControl)
	}
	return f, nil
}

func (f *SettingsFrame) Value(s SettingID) (v uint32, ok bool) {
	f.checkValid()
	buf := f.p
	for len(buf) > 0 {
		settingID := SettingID(binary.BigEndian.Uint16(buf[:2]))
		if settingID == s {
			return binary.BigEndian.Uint32(buf[2:4]), true
		}
		buf = buf[6:]
	}
	return 0, false
}

func (f *SettingsFrame) ForeachSetting(fn func(s SettingID, v uint32)) {
	f.checkValid()
	buf := f.p
	for len(buf) > 0 {
		fn(SettingID(binary.BigEndian.Uint16(buf[:2])), binary.BigEndian.Uint32(buf[2:4]))
		buf = buf[6:]
	}
}

type UnknownFrame struct {
	FrameHeader
	p []byte
}

// Payload returns the frame's payload (after the header).
// It is not valid to call this method after a subsequent
// call to Framer.ReadFrame.
func (f *UnknownFrame) Payload() []byte {
	f.checkValid()
	return f.p
}

func parseUnknownFrame(fh FrameHeader, p []byte) (Frame, error) {
	return &UnknownFrame{fh, p}, nil
}

type WindowUpdateFrame struct {
	FrameHeader
	Increment uint32
}

func parseWindowUpdateFrame(fh FrameHeader, p []byte) (Frame, error) {
	if len(p) < 4 {
		// Too short.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	f := &WindowUpdateFrame{
		FrameHeader: fh,
		Increment:   binary.BigEndian.Uint32(p[:4]) & 0x7fffffff, // mask off high reserved bit
	}
	return f, nil
}

type HeaderFrame struct {
	FrameHeader

	// If FlagHeadersPriority:
	ExclusiveDep bool
	StreamDep    uint32

	// Weight is [0,255]. Only valid if FrameHeader.Flags has the
	// FlagHeadersPriority bit set, in which case the caller must
	// also add 1 to get to spec-defined [1,256] range.
	Weight uint8

	headerFragBuf []byte // not owned
}

func (f *HeaderFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func parseHeadersFrame(fh FrameHeader, p []byte) (_ Frame, err error) {
	hf := &HeaderFrame{
		FrameHeader: fh,
	}
	if fh.StreamID == 0 {
		// HEADERS frames MUST be associated with a stream.  If a HEADERS frame
		// is received whose stream identifier field is 0x0, the recipient MUST
		// respond with a connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	var padLength uint8
	if fh.Flags.Has(FlagHeadersPadded) {
		if p, padLength, err = readByte(p); err != nil {
			return
		}
	}
	if fh.Flags.Has(FlagHeadersPriority) {
		var v uint32
		p, v, err = readUint32(p)
		if err != nil {
			return nil, err
		}
		hf.StreamDep = v & 0x7fffffff
		hf.ExclusiveDep = (v != hf.StreamDep) // high bit was set
		p, hf.Weight, err = readByte(p)
		if err != nil {
			return nil, err
		}
	}
	if len(p)-int(padLength) <= 0 {
		return nil, StreamError(fh.StreamID)
	}
	hf.headerFragBuf = p[:len(p)-int(padLength)]
	return hf, nil
}

func readByte(p []byte) (remain []byte, b byte, err error) {
	if len(p) == 0 {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return p[1:], p[0], nil
}

func readUint32(p []byte) (remain []byte, v uint32, err error) {
	if len(p) < 4 {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return p[4:], binary.BigEndian.Uint32(p[:4]), nil
}
