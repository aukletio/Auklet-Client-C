package schema

import (
	"fmt"
	"log"

	"github.com/ESG-USA/Auklet-Client-C/agent"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type Converter struct {
	in        MessageSource
	out       chan broker.Message
	persistor *broker.Persistor
	app       ExitWaitApp
}

type ExitWaitApp interface {
	ExitApp
	Wait()
}

type MessageSource interface {
	Output() <-chan agent.Message
}

func NewConverter(in MessageSource, persistor *broker.Persistor, app ExitWaitApp) Converter {
	c := Converter{
		in:        in,
		out:       make(chan broker.Message),
		persistor: persistor,
		app:       app,
	}
	go c.serve()
	return c
}

func (c Converter) Output() <-chan broker.Message {
	return c.out
}

func (c Converter) serve() {
	defer close(c.out)
	for agentMsg := range c.in.Output() {
		brokerMsg := convert(agentMsg, c.app)
		if err := c.persistor.CreateMessage(brokerMsg); err != nil {
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
