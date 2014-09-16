// Copyright 2014 The Go Authors.
// See https://code.google.com/p/go/source/browse/CONTRIBUTORS
// Licensed under the same terms as Go itself:
// https://code.google.com/p/go/source/browse/LICENSE

package http2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const frameHeaderLen = 9

// A FrameType is a registered frame type as defined in
// http://http2.github.io/http2-spec/#rfc.section.11.2
type FrameType uint8

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

// Flags is a bitmask of HTTP/2 flags.
// The meaning of flags varies depending on the frame type.
type Flags uint8

// Has reports whether f contains all (0 or more) flags in v.
func (f Flags) Has(v Flags) bool {
	return (f & v) == v
}

// Frame-specific FrameHeader flag bits.
const (
	// Data Frame
	FlagDataEndStream Flags = 0x1
	FlagDataPadded    Flags = 0x8

	// Headers Frame
	FlagHeadersEndStream  Flags = 0x1
	FlagHeadersEndHeaders Flags = 0x4
	FlagHeadersPadded     Flags = 0x8
	FlagHeadersPriority   Flags = 0x20

	// Settings Frame
	FlagSettingsAck Flags = 0x1

	// Ping Frame
	FlagPingAck Flags = 0x1

	// Continuation Frame
	FlagContinuationEndHeaders Flags = 0x4
)

var flagName = map[FrameType]map[Flags]string{
	FrameData: {
		FlagDataEndStream: "END_STREAM",
		FlagDataPadded:    "PADDED",
	},
	FrameHeaders: {
		FlagHeadersEndStream:  "END_STREAM",
		FlagHeadersEndHeaders: "END_HEADERS",
		FlagHeadersPadded:     "PADDED",
		FlagHeadersPriority:   "PRIORITY",
	},
	FrameSettings: {
		FlagSettingsAck: "ACK",
	},
	FramePing: {
		FlagPingAck: "ACK",
	},
	FrameContinuation: {
		FlagContinuationEndHeaders: "END_HEADERS",
	},
}

// A SettingID is an HTTP/2 setting as defined in
// http://http2.github.io/http2-spec/#iana-settings
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

// a frameParser parses a frame given its FrameHeader and payload
// bytes. The length of payload will always equal fh.Length (which
// might be 0).
type frameParser func(fh FrameHeader, payload []byte) (Frame, error)

var frameParsers = map[FrameType]frameParser{
	FrameData:         parseDataFrame,
	FrameHeaders:      parseHeadersFrame,
	FramePriority:     nil, // TODO
	FrameRSTStream:    parseRSTStreamFrame,
	FrameSettings:     parseSettingsFrame,
	FramePushPromise:  nil, // TODO
	FramePing:         parsePingFrame,
	FrameGoAway:       parseGoAwayFrame,
	FrameWindowUpdate: parseWindowUpdateFrame,
	FrameContinuation: parseContinuationFrame,
}

func typeFrameParser(t FrameType) frameParser {
	if f := frameParsers[t]; f != nil {
		return f
	}
	return parseUnknownFrame
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

// Header returns h. It exists so FrameHeaders can be embedded in other
// specific frame types and implement the Frame interface.
func (h FrameHeader) Header() FrameHeader { return h }

func (h FrameHeader) String() string {
	var buf bytes.Buffer
	buf.WriteString("[FrameHeader ")
	buf.WriteString(h.Type.String())
	if h.Flags != 0 {
		buf.WriteString(" flags=")
		set := 0
		for i := uint8(0); i < 8; i++ {
			if h.Flags&(1<<i) == 0 {
				continue
			}
			set++
			if set > 1 {
				buf.WriteByte('|')
			}
			name := flagName[h.Type][Flags(1<<i)]
			if name != "" {
				buf.WriteString(name)
			} else {
				fmt.Fprintf(&buf, "0x%x", 1<<i)
			}
		}
	}
	if h.StreamID != 0 {
		fmt.Fprintf(&buf, " stream=%d", h.StreamID)
	}
	fmt.Fprintf(&buf, " len=%d]", h.Length)
	return buf.String()
	return fmt.Sprintf("[FrameHeader type=%v flags=%v stream=%v len=%v]",
		h.Type, h.Flags, h.StreamID, h.Length)
}

func (h *FrameHeader) checkValid() {
	if !h.valid {
		panic("Frame accessor called on non-owned Frame")
	}
}

func (h *FrameHeader) invalidate() { h.valid = false }

// frame header bytes.
// Used only by ReadFrameHeader.
var fhBytes = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, frameHeaderLen)
		return &buf
	},
}

// ReadFrameHeader reads 9 bytes from r and returns a FrameHeader.
// Most users should use Framer.ReadFrame instead.
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

// A Frame is the base interface implemented by all frame types.
// Callers will generally type-assert the specific frame type:
// *HeadersFrame, *SettingsFrame, *WindowUpdateFrame, etc.
//
// Frames are only valid until the next call to Framer.ReadFrame.
type Frame interface {
	Header() FrameHeader

	// invalidate is called by Framer.ReadFrame to make this
	// frame's buffers as being invalid, since the subsequent
	// frame will reuse them.
	invalidate()
}

// A Framer reads and writes Frames.
type Framer struct {
	r         io.Reader
	lr        io.LimitedReader
	lastFrame Frame
	readBuf   []byte

	w    io.Writer
	wbuf []byte
}

func (f *Framer) startWrite(ftype FrameType, flags Flags, streamID uint32) {
	// Write the FrameHeader.
	f.wbuf = append(f.wbuf[:0],
		0, // 3 bytes of length, filled in in endWrite
		0,
		0,
		byte(ftype),
		byte(flags),
		byte(streamID>>24), // TODO: &127? Or do it in callers? Or allow for testing.
		byte(streamID>>16),
		byte(streamID>>8),
		byte(streamID))
}

func (f *Framer) endWrite() error {
	// Now that we know the final size, fill in the FrameHeader in
	// the space previously reserved for it. Abuse append.
	length := len(f.wbuf) - frameHeaderLen
	_ = append(f.wbuf[:0],
		byte(length>>16),
		byte(length>>8),
		byte(length))
	n, err := f.w.Write(f.wbuf)
	if err == nil && n != length {
		err = io.ErrShortWrite
	}
	return err
}

func (f *Framer) writeUint32(v uint32) {
	f.wbuf = append(f.wbuf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
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

// A DataFrame conveys arbitrary, variable-length sequences of octets
// associated with a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.1
type DataFrame struct {
	FrameHeader
	data []byte
}

// Data returns the frame's data octets, not including any padding
// size byte or padding suffix bytes.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *DataFrame) Data() []byte {
	f.checkValid()
	return f.data
}

func parseDataFrame(fh FrameHeader, payload []byte) (Frame, error) {
	if fh.StreamID == 0 {
		// DATA frames MUST be associated with a stream. If a
		// DATA frame is received whose stream identifier
		// field is 0x0, the recipient MUST respond with a
		// connection error (Section 5.4.1) of type
		// PROTOCOL_ERROR.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	f := &DataFrame{
		FrameHeader: fh,
	}
	var padSize byte
	if fh.Flags.Has(FlagDataPadded) {
		var err error
		payload, padSize, err = readByte(payload)
		if err != nil {
			return nil, err
		}
	}
	if int(padSize) > len(payload) {
		// If the length of the padding is greater than the
		// length of the frame payload, the recipient MUST
		// treat this as a connection error.
		// Filed: https://github.com/http2/http2-spec/issues/610
		return nil, ConnectionError(ErrCodeProtocol)
	}
	f.data = payload[:len(payload)-int(padSize)]
	return f, nil
}

// WriteData writes a DATA frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *Framer) WriteData(streamID uint32, endStream bool, data []byte) error {
	// TODO: ignoring padding for now. will add when somebody cares.
	if streamID == 0 {
		// TODO: return some error to tell the caller that
		// they're doing it wrong?  Or let them do as they'd
		// like? Might be useful for testing other people's
		// http2 implementations.  Maybe we have a
		// Framer.AllowStupid bool?
	}
	var flags Flags
	if endStream {
		flags |= FlagDataEndStream
	}
	f.startWrite(FrameData, flags, streamID)
	f.wbuf = append(f.wbuf, data...)
	return f.endWrite()
}

// A SettingsFrame conveys configuration parameters that affect how
// endpoints communicate, such as preferences and constraints on peer
// behavior.
//
// See http://http2.github.io/http2-spec/#SETTINGS
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

// A PingFrame is a mechanism for measuring a minimal round trip time
// from the sender, as well as determining whether an idle connection
// is still functional.
// See http://http2.github.io/http2-spec/#rfc.section.6.7
type PingFrame struct {
	FrameHeader
	Data [8]byte
}

func parsePingFrame(fh FrameHeader, payload []byte) (Frame, error) {
	if len(payload) != 8 {
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	if fh.StreamID != 0 {
		return nil, ConnectionError(ErrCodeProtocol)
	}
	f := &PingFrame{FrameHeader: fh}
	copy(f.Data[:], payload)
	return f, nil
}

// A GoAwayFrame informs the remote peer to stop creating streams on this connection.
// See http://http2.github.io/http2-spec/#rfc.section.6.8
type GoAwayFrame struct {
	FrameHeader
	LastStreamID uint32
	ErrCode      uint32
	debugData    []byte
}

// DebugData returns any debug data in the GOAWAY frame. Its contents
// are not defined.
// The caller must not retain the returned memory past the next
// call to ReadFrame.
func (f *GoAwayFrame) DebugData() []byte {
	f.checkValid()
	return f.debugData
}

func parseGoAwayFrame(fh FrameHeader, p []byte) (Frame, error) {
	if fh.StreamID != 0 {
		return nil, ConnectionError(ErrCodeProtocol)
	}
	if len(p) < 8 {
		return nil, ConnectionError(ErrCodeFrameSize)
	}
	return &GoAwayFrame{
		FrameHeader:  fh,
		LastStreamID: binary.BigEndian.Uint32(p[:4]) & (1<<31 - 1),
		ErrCode:      binary.BigEndian.Uint32(p[4:8]),
		debugData:    p[:8],
	}, nil
}

// An UnknownFrame is the frame type returned when the frame type is unknown
// or no specific frame type parser exists.
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

// A WindowUpdateFrame is used to implement flow control.
// See http://http2.github.io/http2-spec/#rfc.section.6.9
type WindowUpdateFrame struct {
	FrameHeader
	Increment uint32
}

func parseWindowUpdateFrame(fh FrameHeader, p []byte) (Frame, error) {
	if len(p) < 4 {
		// Too short.
		return nil, ConnectionError(ErrCodeProtocol)
	}
	return &WindowUpdateFrame{
		FrameHeader: fh,
		Increment:   binary.BigEndian.Uint32(p[:4]) & 0x7fffffff, // mask off high reserved bit
	}, nil
}

// A HeadersFrame is used to open a stream and additionally carries a
// header block fragment.
type HeadersFrame struct {
	FrameHeader

	// If FlagHeadersPriority:
	ExclusiveDep bool
	StreamDep    uint32

	// Weight is [0,255]. Only valid if FrameHeader.Flags has the
	// FlagHeadersPriority bit set, in which case the caller must
	// also add 1 to get to spec-defined [1,256] range.
	Weight        uint8
	headerFragBuf []byte // not owned
}

func (f *HeadersFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func (f *HeadersFrame) HeadersEnded() bool {
	return f.FrameHeader.Flags.Has(FlagHeadersEndHeaders)
}

func parseHeadersFrame(fh FrameHeader, p []byte) (_ Frame, err error) {
	hf := &HeadersFrame{
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

// A RSTStreamFrame allows for abnormal termination of a stream.
// See http://http2.github.io/http2-spec/#rfc.section.6.4
type RSTStreamFrame struct {
	FrameHeader
	ErrCode uint32
}

func parseRSTStreamFrame(fh FrameHeader, p []byte) (Frame, error) {
	if fh.StreamID == 0 || len(p) < 4 {
		return nil, ConnectionError(ErrCodeProtocol)
	}
	return &RSTStreamFrame{fh, binary.BigEndian.Uint32(p[:4])}, nil
}

// WriteRSTStream writes a RST_STREAM frame.
//
// It will perform exactly one Write to the underlying Writer.
// It is the caller's responsibility to not call other Write methods concurrently.
func (f *Framer) WriteRSTStream(streamID, errCode uint32) error {
	if streamID == 0 {
		// TODO: return some error to tell the caller that
		// they're doing it wrong?  Or let them do as they'd
		// like? Might be useful for testing other people's
		// http2 implementations.  Maybe we have a
		// Framer.AllowStupid bool?
	}
	f.startWrite(FrameRSTStream, 0, streamID)
	f.writeUint32(errCode)
	return f.endWrite()
}

// A ContinuationFrame is used to continue a sequence of header block fragments.
// See http://http2.github.io/http2-spec/#rfc.section.6.10
type ContinuationFrame struct {
	FrameHeader
	headerFragBuf []byte
}

func parseContinuationFrame(fh FrameHeader, p []byte) (Frame, error) {
	return &ContinuationFrame{fh, p}, nil
}

func (f *ContinuationFrame) HeaderBlockFragment() []byte {
	f.checkValid()
	return f.headerFragBuf
}

func (f *ContinuationFrame) HeadersEnded() bool {
	return f.FrameHeader.Flags.Has(FlagContinuationEndHeaders)
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
