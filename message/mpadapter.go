package message

import (
	"bytes"
	"encoding/json"

	"github.com/vmihailenco/msgpack"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MPAdapter translates messages from JSON to MessagePack.
type MPAdapter struct {
	in  broker.MessageSource
	out chan broker.Message
}

// NewMPAdapter returns an MPAdapter that translates messages from in.
func NewMPAdapter(in broker.MessageSource) MPAdapter {
	a := MPAdapter{
		in:  in,
		out: make(chan broker.Message),
	}
	go a.serve()
	return a
}

// Output returns the adapter's output channel. It closes when its source
// closes.
func (a MPAdapter) Output() <-chan broker.Message {
	return a.out
}

// serve runs the adapter, causing it to send and receive messages. Serve
// returns when the adapter's input closes.
func (a MPAdapter) serve() {
	defer close(a.out)
	for msg := range a.in.Output() {
		if err := adapt(&msg); err != nil {
			errorlog.Print(err)
		}
		a.out <- msg
	}
}

// adapt translates msg from JSON to MessagePack. We modify the message rather
// than create a new one in order to maintain its association with the
// persistent file (which is hidden from us).
func adapt(msg *broker.Message) error {
	var v interface{}
	if err := json.Unmarshal(msg.Bytes, &v); err != nil {
		return err
	}
	b, err := marshalMP(v)
	if err != nil {
		return err
	}
	// Bytes is of type json.RawMessage, which is necessary for the
	// persistor, but the producers don't care. They see it as []byte.
	msg.Bytes = b
	return nil
}

func marshalMP(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseJSONTag(true)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
