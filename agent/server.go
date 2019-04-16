// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/aukletio/Auklet-Client-C/errorlog"
)

// Message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type Message struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
	Error string
}

// Server provides a connection server for an Auklet agent.
type Server struct {
	in  io.Reader
	dec *json.Decoder
	out chan Message
	// Done closes when the Server gets EOF.
	Done chan struct{}
	errd bool
	msg Message
	err error
}

// NewServer returns a new Server that reads from in. If dec is not nil, it is
// used directly.
func NewServer(in io.Reader, dec *json.Decoder) *Server {
	s := &Server{
		in:   in,
		dec:  dec,
		out:  make(chan Message),
		Done: make(chan struct{}),
	}
	go s.serve()
	return s
}

func (s *Server) scan() bool {
	var msg Message
	if err := s.dec.Decode(&msg); err == io.EOF {
		return false
	} else if err != nil {
		// There was a problem decoding the stream into
		// message format.
		buf, _ := ioutil.ReadAll(s.dec.Buffered())
		s.err = fmt.Errorf("%v in %v", err.Error(), string(buf))
		msg := Message{
			Type:  "log",
			Error: s.err.Error(),
		}
		s.msg = msg
		s.dec = json.NewDecoder(s.in)
		errorlog.Printf("Server.serve: %v in %q", err, string(buf))
		return true
	}
	if msg.Type == "event" {
		s.errd = true
	}
	s.msg = msg
	return true
}

// serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s *Server) serve() {
	defer close(s.Done)
	defer close(s.out)
	log.Print("Server: accepted connection")
	defer log.Print("Server: connection closed")
	if s.dec == nil {
		s.dec = json.NewDecoder(s.in)
	}
	for s.scan() {
		s.out <- s.msg
	}
	if !s.errd {
		s.out <- Message{Type: "cleanExit"}
	}
}

// Output returns s's output stream.
func (s *Server) Output() <-chan Message {
	return s.out
}
