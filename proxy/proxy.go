// Package proxy implements a server that reads typed data from a Unix domain
// socket, transforms it according to client-defined functions, and relays it to
// a message sink.
package proxy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/ESG-USA/Auklet-Client/message"
)

// sockMessage represents the JSON schema of messages that can be received by a
// Proxy.
type sockMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a producer message.
type Handler func(data []byte) (message.Message, error)

type Proxy struct {
	// in is the Unix domain socket on which the Proxy waits for
	// an incoming connection.
	in net.Listener

	// handlers is a collection of Handler functions keyed by socket message
	// type. When a message is received, the corresponding Handler is looked
	// up and called. The argument to the handler is the socket message's
	// data. The message returned by a handler is sent via out.
	// Errors returned by a handler are logged, and do not shut down the
	// Proxy.
	handlers map[string]Handler
	out      chan message.Message
}

func New(addr string, handlers map[string]Handler) Proxy {
	l, err := net.Listen("unixpacket", addr)
	if err != nil {
		log.Print(err)
	}
	p := Proxy{
		in:       l,
		out:      make(chan message.Message),
		handlers: handlers,
	}
	return p
}

// Serve waits for proxy to accept an incoming connection, then serves the
// connection.
func (proxy Proxy) Serve() {
	defer close(proxy.out)
	defer proxy.in.Close()
	conn, err := proxy.in.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("accepted connection on %v", proxy.in.Addr())
	d := json.NewDecoder(conn)
	for {
		sm := &sockMessage{}
		if err := d.Decode(sm); err == io.EOF {
			log.Printf("connection on %v closed", proxy.in.Addr())
			break
		} else if err != nil {
			// There was a problem decoding the JSON into
			// sockMessage format.
			b, _ := ioutil.ReadAll(d.Buffered())
			log.Print(err, string(b))
			continue
		}

		if handler, in := proxy.handlers[sm.Type]; in {
			pm, err := handler(sm.Data)
			if err != nil {
				log.Print(err)
				continue
			}
			proxy.out <- pm
		} else {
			log.Printf(`socket message of type "%v" not handled`, sm.Type)
		}
	}
}

func (proxy Proxy) Output() <-chan message.Message {
	return proxy.out
}
