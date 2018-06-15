// Package agent provides access to an Auklet agent.
package agent

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// message represents messages that can be received by a Server, and thus,
// would be sent by an agent.
type message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a broker.Message.
type Handler func(data []byte) (broker.Message, error)

// Server provides a Unix domain socket listener for an Auklet agent.
type Server struct {
	// local is the Unix domain socket to be served.
	local  *os.File
	remote *os.File

	// handlers is a collection of Handler functions keyed by message
	// type. When a message is received, the corresponding Handler is looked
	// up and called. The argument to the handler is the  message's
	// data. The message returned by a handler is sent via out.
	// Errors returned by a handler are logged, and do not shut down the
	// Server.
	handlers map[string]Handler
	out      chan broker.Message

	conf chan int
}

// NewServer returns a new Server for an anonymous Unix domain socket. Incoming
// messages are processed by the given handlers.
func NewServer(handlers map[string]Handler) Server {
	local, remote, err := socketpair("dataserver-")
	if err != nil {
		errorlog.Print(err)
	}
	return Server{
		local:    local,
		remote:   remote,
		out:      make(chan broker.Message),
		handlers: handlers,
		conf:     make(chan int),
	}
}

// Serve causes s to accept an incoming connection, after which s can send and
// receive messages.
func (s Server) Serve() {
	defer close(s.out)
	log.Printf("accepted connection on %v", s.local.Name())
	defer log.Printf("connection on %v closed", s.local.Name())
	go s.requestProfiles(s.local)
	dec := json.NewDecoder(s.local)
	for {
		msg := &message{}
		if err := dec.Decode(msg); err == io.EOF {
			return
		} else if err != nil {
			// There was a problem decoding the JSON into
			// message format.
			buf, _ := ioutil.ReadAll(dec.Buffered())
			errorlog.Print(err, string(buf))
			dec = json.NewDecoder(s.local)
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
				errorlog.Println("requestProfiles:", err)
			}
		case dur := <-s.conf:
			emit.Stop()
			emit = time.NewTicker(time.Duration(dur) * time.Second)
		}
	}
}

// Output returns s's output stream.
func (s Server) Output() <-chan broker.Message {
	return s.out
}

// Remote returns the socket to be inherited by the child process.
func (s Server) Remote() *os.File {
	return s.remote
}
