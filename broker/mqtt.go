package broker

import (
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT client.
type MQTTProducer struct {
	in MessageSource
	c  mqtt.Client
}

// wait turns Paho's async API into a sync API.
func wait(t mqtt.Token) error {
	t.Wait()
	return t.Error()
}

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(in MessageSource) MQTTProducer {
	opt := mqtt.NewClientOptions()
	opt.AddBroker("tcp://:8080")
	opt.SetClientID("C")
	c := mqtt.NewClient(opt)

	if err := wait(c.Connect()); err != nil {
		errorlog.Print("producer:", err)
		return MQTTProducer{}
	}
	log.Print("producer: connected")

	return MQTTProducer{
		in: in,
		c:  c,
	}
}

// Serve launches p, enabling it to send and receive messages.
func (p MQTTProducer) Serve() {
	defer func() {
		p.c.Disconnect(0)
		log.Print("producer: disconnected")
	}()

	for msg := range p.in.Output() {
		if err := wait(p.c.Publish("test-topic", 1, false, []byte(msg.Bytes))); err != nil {
			errorlog.Print("producer:", err)
			continue
		}
		log.Printf("producer: sent %+q", msg.Bytes)
		msg.Remove()
	}
}
