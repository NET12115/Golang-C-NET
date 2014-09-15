// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// See https://code.google.com/p/go/source/browse/CONTRIBUTORS
// Licensed under the same terms as Go itself:
// https://code.google.com/p/go/source/browse/LICENSE

// Package http2 implements the HTTP/2 protocol.
//
// It currently targets draft-14. See http://http2.github.io/
package http2

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"sync"
)

const (
	// ClientPreface is the string that must be sent by new
	// connections from clients.
	ClientPreface = "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
)

var (
	clientPreface = []byte(ClientPreface)
)

const npnProto = "h2-14"

// Server is an HTTP2 server.
type Server struct {
	// MaxStreams optionally ...
	MaxStreams int

	mu sync.Mutex
}

func (srv *Server) handleClientConn(hs *http.Server, c *tls.Conn, h http.Handler) {
	cc := &clientConn{hs, c, h, NewFramer(c, c)}
	cc.serve()
}

type clientConn struct {
	hs *http.Server
	c  *tls.Conn
	h  http.Handler
	fr *Framer
}

func (cc *clientConn) logf(format string, args ...interface{}) {
	if lg := cc.hs.ErrorLog; lg != nil {
		lg.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (cc *clientConn) serve() {
	defer cc.c.Close()
	log.Printf("HTTP/2 connection from %v on %p", cc.c.RemoteAddr(), cc.hs)

	buf := make([]byte, len(ClientPreface))
	// TODO: timeout reading from the client
	if _, err := io.ReadFull(cc.c, buf); err != nil {
		cc.logf("error reading client preface: %v", err)
		return
	}
	if !bytes.Equal(buf, clientPreface) {
		cc.logf("bogus greeting from client: %q", buf)
		return
	}
	log.Printf("client %v said hello", cc.c.RemoteAddr())
	for {

		f, err := cc.fr.ReadFrame()
		if h2e, ok := err.(Error); ok {
			if h2e.IsConnectionError() {
				log.Printf("Disconnection; connection error: %v", err)
				return
			}
			// TODO: stream errors, etc
		}
		if err != nil {
			log.Printf("Disconnection due to other error: %v", err)
			return
		}
		log.Printf("got %v: %#v", f.Header(), f)
		if sf, ok := f.(*SettingsFrame); ok {
			log.Printf("Is a settings frame")
			sf.ForeachSetting(func(s SettingID, v uint32) {
				log.Printf("  setting %s = %v", s, v)
			})
		}
	}
}

// ConfigureServer adds HTTP2 support to s as configured by the HTTP/2
// server configuration in conf. The configuration may be nil.
//
// ConfigureServer must be called before s beings serving.
func ConfigureServer(s *http.Server, conf *Server) {
	if conf == nil {
		conf = new(Server)
	}
	if s.TLSConfig == nil {
		s.TLSConfig = new(tls.Config)
	}
	haveNPN := false
	for _, p := range s.TLSConfig.NextProtos {
		if p == npnProto {
			haveNPN = true
			break
		}
	}
	if !haveNPN {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, npnProto)
	}

	if s.TLSNextProto == nil {
		s.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
	}
	s.TLSNextProto[npnProto] = func(hs *http.Server, c *tls.Conn, h http.Handler) {
		if testHookOnConn != nil {
			testHookOnConn()
		}
		conf.handleClientConn(hs, c, h)
	}
}

var testHookOnConn func() // for testing
