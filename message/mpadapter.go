package message

import (
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

// MPAdapter translates messages from JSON to MessagePack.
type MPAdapter struct {
	in  broker.MessageSource
	out chan broker.Message
}

// NewMPAdapter returns an MPAdapter that translates messages from in.
func NewMPAdapter(in broker.MessageSource) MPAdapter {
	return MPAdapter{
		in:  in,
		out: make(chan broker.Message),
	}
}

// Output returns the adapter's output channel. It closes when its source
// closes.
func (a MPAdapter) Output() <-chan broker.Message {
	return a.out
}

// Serve runs the adapter, causing it to send and receive messages. Serve
// returns when the adapter's input closes.
func (a MPAdapter) Serve() {
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
	v, in := map[string]interface{}{
		"[]uint8":         new([]byte),
		"schema.AppLog":   new(schema.AppLog),
		"schema.ErrorSig": new(schema.ErrorSig),
		"schema.Exit":     new(schema.Exit),
		"schema.Profile":  new(schema.Profile),
	}[msg.Type]
	if !in {
		return fmt.Errorf("adapt: can't convert %q to MessagePack format", msg.Type)
	}
	if err := json.Unmarshal(msg.Bytes, v); err != nil {
		return err
	}
	b, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	// Bytes is of type json.RawMessage, which is necessary for the
	// persistor, but the producers don't care. They see it as []byte.
	msg.Bytes = b
	return nil
}
