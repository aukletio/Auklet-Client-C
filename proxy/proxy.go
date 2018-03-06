// Package proxy implements a server that reads typed data from a Unix domain
// socket, transforms it according to client-defined functions, and sends it via
// a Kafka producer.
package proxy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/ESG-USA/apc/producer"
)

// sockMessage represents the JSON schema of messages that can be received by a
// Proxy.
type sockMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Handler transforms a byte slice into a producer message.
type Handler func(data []byte) (producer.Message, error)

// Proxy serves a single-client, simplex connection from a Unix domain
// socket to a Kafka producer.
type Proxy struct {
	// Listener is the Unix domain socket on which the Proxy waits for
	// an incoming connection.
	net.Listener

	// Producer is the Kafka producer to which the Proxy sends messages
	// returned by Handlers.
	*producer.Producer

	// Handlers is a collection of Handler functions keyed by socket message
	// type. When a message is received, the corresponding Handler is looked
	// up and called. The argument to the handler is the socket message's
	// data. The producer.Message returned by a handler is sent via Producer.
	// Errors returned by a handler are logged, and do not shut down the
	// Proxy.
	Handlers map[string]Handler
}

// Serve waits for proxy to accept an incoming connection, then serves the
// connection.
func (proxy Proxy) Serve() {
	if proxy.Producer == nil {
		log.Print("Proxy.Serve: called with nil Producer")
		return
	}
	defer proxy.Close()
	conn, err := proxy.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("accepted connection on %v", proxy.Addr())
	d := json.NewDecoder(conn)
	for {
		sm := &sockMessage{}
		if err := d.Decode(sm); err == io.EOF {
			log.Printf("connection on %v closed", proxy.Addr())
			break
		} else if err != nil {
			// There was a problem decoding the JSON into
			// sockMessage format.
			b, _ := ioutil.ReadAll(d.Buffered())
			log.Print(err, string(b))
			continue
		}

		if handler, in := proxy.Handlers[sm.Type]; in {
			pm, err := handler(sm.Data)
			if err != nil {
				log.Print(err)
				continue
			}
			if err := proxy.Send(pm); err != nil {
				log.Print(err)
			}
		} else {
			log.Printf(`socket message of type "%v" not handled`, sm.Type)
		}
	}
}
