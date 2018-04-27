// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"

	msg "github.com/ESG-USA/Auklet-Client/message"
)

// message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a message.Message.
type Handler func(data []byte) (msg.Message, error)

// Server provides a Unix domain socket listener for an Auklet agent.
type Server struct {
	// in is the Unix domain socket on which the Server waits for
	// an incoming connection.
	in net.Listener

	// handlers is a collection of Handler functions keyed by message
	// type. When a message is received, the corresponding Handler is looked
	// up and called. The argument to the handler is the  message's
	// data. The message returned by a handler is sent via out.
	// Errors returned by a handler are logged, and do not shut down the
	// Server.
	handlers map[string]Handler
	out      chan msg.Message
}

// NewServer returns a new Server for the Unix domain socket at addr. Incoming
// messages are processed by the given handlers.
func NewServer(addr string, handlers map[string]Handler) Server {
	l, err := net.Listen("unixpacket", addr)
	if err != nil {
		log.Print(err)
	}
	p := Server{
		in:       l,
		out:      make(chan msg.Message),
		handlers: handlers,
	}
	return p
}

// Serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s Server) Serve() {
	defer close(s.out)
	defer s.in.Close()
	conn, err := s.in.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("accepted connection on %v", s.in.Addr())
	d := json.NewDecoder(conn)
	for {
		sm := &message{}
		if err := d.Decode(sm); err == io.EOF {
			log.Printf("connection on %v closed", s.in.Addr())
			break
		} else if err != nil {
			// There was a problem decoding the JSON into
			// message format.
			b, _ := ioutil.ReadAll(d.Buffered())
			log.Print(err, string(b))
			continue
		}

		if handler, in := s.handlers[sm.Type]; in {
			pm, err := handler(sm.Data)
			if err != nil {
				log.Print(err)
				continue
			}
			s.out <- pm
		} else {
			log.Printf(`message of type "%v" not handled`, sm.Type)
		}
	}
}

// Output returns s's output stream.
func (s Server) Output() <-chan msg.Message {
	return s.out
}
