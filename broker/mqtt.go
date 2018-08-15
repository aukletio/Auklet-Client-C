package broker

import (
	"crypto/tls"
	"log"

	"github.com/eclipse/paho.mqtt.golang"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// MQTTProducer wraps an MQTT client.
type MQTTProducer struct {
	c  mqtt.Client
}

// wait turns Paho's async API into a sync API.
func wait(t mqtt.Token) error {
	t.Wait()
	return t.Error()
}

// NewMQTTProducer returns a new producer for the given input.
func NewMQTTProducer(addr string, t *tls.Config, creds func() (string, string)) (*MQTTProducer, error) {
	opt := mqtt.NewClientOptions()
	opt.AddBroker(addr)
	opt.SetTLSConfig(t)
	opt.SetClientID("C")
	opt.SetCredentialsProvider(creds)
	c := mqtt.NewClient(opt)

	if err := wait(c.Connect()); err != nil {
		return nil, err
	}
	log.Print("producer: connected")

	return &MQTTProducer{
		c:  c,
	}, nil
}

// Serve launches p, enabling it to send and receive messages.
func (p MQTTProducer) Serve(in MessageSource) {
	defer func() {
		p.c.Disconnect(0)
		log.Print("producer: disconnected")
	}()

	topic := map[Topic]string{
		Profile: "c/profiler/superfluous",
		Event:   "c/events/superfluous",
		Log:     "c/logs/superfluous",
	}

	for msg := range in.Output() {
		if err := wait(p.c.Publish(topic[msg.Topic], 1, false, []byte(msg.Bytes))); err != nil {
			errorlog.Print("producer:", err)
			continue
		}
		log.Printf("producer: sent %+q", msg.Bytes)
		msg.Remove()
	}
}
