// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// Message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type Message struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
	Error string
}

// Handler transforms a byte slice into a broker.Message.
type Handler func(data []byte) (broker.Message, error)

// Server provides a connection server for an Auklet agent.
type Server struct {
	in  io.Reader
	out chan Message
}

// NewServer returns a new Server that reads from conn. Incoming messages are
// processed by the given handlers.
func NewServer(in io.Reader) Server {
	s := Server{
		in:  in,
		out: make(chan Message),
	}
	go s.serve()
	return s
}

// serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s Server) serve() {
	defer close(s.out)
	log.Print("Server: accepted connection")
	defer log.Print("Server: connection closed")
	dec := json.NewDecoder(s.in)
	for {
		var msg Message
		if err := dec.Decode(&msg); err == io.EOF {
			return
		} else if err != nil {
			// There was a problem decoding the stream into
			// message format.
			buf, _ := ioutil.ReadAll(dec.Buffered())
			s.out <- Message{
				Error: fmt.Sprint(err.Error(), string(buf)),
			}
			dec = json.NewDecoder(s.in)
			continue
		}
		s.out <- msg
	}
}

// Output returns s's output stream.
func (s Server) Output() <-chan Message {
	return s.out
}
