// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/ESG-USA/Auklet-Client/errorlog"
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a kafka.Message.
type Handler func(data []byte) (kafka.Message, error)

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
	out      chan kafka.Message

	conf chan int
}

// NewServer returns a new Server for the Unix domain socket at addr. Incoming
// messages are processed by the given handlers.
func NewServer(addr string, handlers map[string]Handler) Server {
	l, err := net.Listen("unix", addr)
	if err != nil {
		errorlog.Print(err)
	}
	return Server{
		in:       l,
		out:      make(chan kafka.Message),
		handlers: handlers,
		conf:     make(chan int),
	}
}

// Serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s Server) Serve() {
	defer close(s.out)
	defer s.in.Close()
	conn, err := s.in.Accept()
	if err != nil {
		errorlog.Print(err)
		return
	}
	log.Printf("accepted connection on %v", s.in.Addr())
	go s.requestProfiles(conn)
	dec := json.NewDecoder(conn)
	for {
		msg := &message{}
		if err := dec.Decode(msg); err == io.EOF {
			log.Printf("connection on %v closed", s.in.Addr())
			break
		} else if err != nil {
			// There was a problem decoding the JSON into
			// message format.
			buf, _ := ioutil.ReadAll(dec.Buffered())
			errorlog.Print(err, string(buf))
			dec = json.NewDecoder(conn)
			continue
		}

		if handler, in := s.handlers[msg.Type]; in {
			pm, err := handler(msg.Data)
			switch err.(type) {
			case kafka.ErrStorageFull:
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

// Configure returns a channel on which the emission period in seconds can be
// set.
func (s Server) Configure() chan<- int {
	return s.conf
}

func (s Server) requestProfiles(out io.Writer) {
	emit := time.NewTicker(time.Second)
	for {
		select {
		case <-emit.C:
			if _, err := out.Write([]byte{0}); err != nil {
				errorlog.Print(err)
			}
		case dur := <-s.conf:
			emit.Stop()
			emit = time.NewTicker(time.Duration(dur) * time.Second)
		}
	}
}

// Output returns s's output stream.
func (s Server) Output() <-chan kafka.Message {
	return s.out
}
