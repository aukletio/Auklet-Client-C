// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a broker.Message.
type Handler func(data []byte) (broker.Message, error)

// Server provides a connection server for an Auklet agent.
type Server struct {
	// conn is the connection to be served.
	conn io.Reader

	// handlers is a collection of Handler functions keyed by message
	// type. When a message is received, the corresponding Handler is looked
	// up and called. The argument to the handler is the  message's
	// data. The message returned by a handler is sent via out.
	// Errors returned by a handler are logged, and do not shut down the
	// Server.
	handlers map[string]Handler
	out      chan broker.Message
}

// NewServer returns a new Server that reads from conn. Incoming messages are
// processed by the given handlers.
func NewServer(conn io.Reader, handlers map[string]Handler) Server {
	s := Server{
		conn:     conn,
		out:      make(chan broker.Message),
		handlers: handlers,
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
	dec := json.NewDecoder(s.conn)
	for {
		msg := &message{}
		if err := dec.Decode(msg); err == io.EOF {
			return
		} else if err != nil {
			// There was a problem decoding the stream into
			// message format.
			buf, _ := ioutil.ReadAll(dec.Buffered())
			errorlog.Print(err, string(buf))
			dec = json.NewDecoder(s.conn)
			continue
		}

		if handler, in := s.handlers[msg.Type]; in {
			pm, err := handler(msg.Data)
			switch err.(type) {
			case broker.ErrStorageFull:
				// Our persistent storage is full, so we drop
				// messages. This isn't an error; it's desired
				// behavior.
				log.Print(err)
				continue
			}
			s.out <- pm
		} else {
			log.Printf(`message of type "%v" not handled`, msg.Type)
		}
	}
}

// Output returns s's output stream.
func (s Server) Output() <-chan broker.Message {
	return s.out
}
