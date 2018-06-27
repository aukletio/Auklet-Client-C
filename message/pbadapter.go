package message

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

// This file implements a JSON-to-protobuf adapter stage.

// PBAdapter is a stage that translates incoming JSON-encoded messages to
// Protocol Buffer encoding.
type PBAdapter struct {
	in  broker.MessageSource
	out chan broker.Message
}

// NewPBAdapter returns a PBAdapter that translates messages from in.
func NewPBAdapter(in broker.MessageSource) PBAdapter {
	return PBAdapter{
		in:  in,
		out: make(chan broker.Message),
	}
}

// Output returns the adapter's output channel.
func (p PBAdapter) Output() <-chan broker.Message {
	return p.out
}

// Serve runs the adapter, allowing messages to be sent and received.
func (p PBAdapter) Serve() {
	defer close(p.out)
	for msg := range p.in.Output() {
		if err := adapt(&msg); err != nil {
			errorlog.Print(err)
			// It should be impossible for a message to fail
			// translation; but if it ever happens, we want to
			// know about it. So we send it anyway, hoping that
			// backend logs will make it visible.
		}
		p.out <- msg
	}
}

// adapt translates msg from JSON to protobuf. We modify the message rather than
// create a new one in order to maintain its association with the persistent
// file (which is hidden from us).
func adapt(msg *broker.Message) error {
	pb, in := map[string]proto.Message{
		"schema.AgentLog": &schema.AgentLog{},
		"schema.AppLog":   &schema.AppLog{},
		"schema.Exit":     &schema.Exit{},
		"schema.Profile":  &schema.Profile{},
	}[msg.Type]
	if !in {
		return fmt.Errorf("adapt: can't convert %q to protobuf message", msg.Type)
	}
	if err := json.Unmarshal(msg.Bytes, pb); err != nil {
		return err
	}
	b, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	// Bytes is of type json.RawMessage, which is necessary for the
	// persistor, but the other stages don't care. They see it as []byte.
	msg.Bytes = b
	return nil
}
