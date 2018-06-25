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

type PBAdapter struct {
	in  broker.MessageSourceError
	out chan broker.Message
}

func NewPBAdapter(in broker.MessageSourceError) PBAdapter {
	return PBAdapter{
		in:  in,
		out: make(chan broker.Message),
	}
}

func (p PBAdapter) Output() <-chan broker.Message {
	return p.out
}

func (p PBAdapter) Err() chan<- error {
	return p.in.Err()
}

func (p PBAdapter) Serve() {
	defer close(p.out)
	for msg := range p.in.Output() {
		if err := adapt(&msg); err != nil {
			errorlog.Print(err)
			// Normally, the adapter has no reason to talk to its
			// source, because it defers this responsibility to its
			// sink. However, when a message fails to be translated,
			// it is undeliverable, and must be deleted.
			p.in.Err() <- nil
			continue
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
	// persistor, but the producers don't care. They see it as []byte.
	msg.Bytes = b
	return nil
}
