package schema

import (
	"fmt"
	"log"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// Converter converts a stream of agent.Message to a stream of broker.Message.
type Converter struct {
	in        MessageSource
	out       chan broker.Message
	persistor Persistor
	app       ExitWaitApp
}

// ExitWaitApp is an ExitApp for which we can wait to exit.
type ExitWaitApp interface {
	ExitApp
	Wait()
}

// MessageSource is a source of agent messages.
type MessageSource interface {
	Output() <-chan agent.Message
}

// Persistor provides a persistor interface.
type Persistor interface {
	CreateMessage(*broker.Message) error
}

// NewConverter returns a converter for the given input stream that uses the
// given persistor and app.
func NewConverter(in MessageSource, persistor Persistor, app ExitWaitApp) Converter {
	c := Converter{
		in:        in,
		out:       make(chan broker.Message),
		persistor: persistor,
		app:       app,
	}
	go c.serve()
	return c
}

// Output returns the converter's output stream.
func (c Converter) Output() <-chan broker.Message {
	return c.out
}

func (c Converter) serve() {
	defer close(c.out)
	for agentMsg := range c.in.Output() {
		brokerMsg := convert(agentMsg, c.app)
		if err := c.persistor.CreateMessage(&brokerMsg); err != nil {
			// Let the backend know we ran out of local storage.
			c.out <- broker.Message{
				Error: err.Error(),
				Topic: broker.Log,
			}
			continue
		}
		c.out <- brokerMsg
	}
}

func convert(m agent.Message, app ExitWaitApp) broker.Message {
	switch m.Type {
	case "applog":
		return NewAppLog(m.Data, app)
	case "profile":
		return NewProfile(m.Data, app)
	case "event":
		app.Wait()
		log.Printf("%v exited with error signal", app)
		return NewErrorSig(m.Data, app)
	case "log":
		return broker.Message{
			Bytes: m.Data,
			Topic: broker.Log,
		}
	}
	return broker.Message{
		Error: fmt.Sprintf("message of type %q not handled", m.Type),
		Topic: broker.Log,
	}
}
