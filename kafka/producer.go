// Package kafka provides a simple wrapper around sarama.SyncProducer.
package kafka

import (
	"log"
	"regexp"

	"github.com/Shopify/sarama"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// Producer provides a simple Kafka producer.
type Producer struct {
	source MessageSourceError
	sarama.SyncProducer
	topic map[Type]string
}

func verify(brokers []*sarama.Broker) bool {
	pattern, err := regexp.Compile(`[^\.]+\.feeds\.auklet\.io:9093`)
	if err != nil {
		errorlog.Println("producer:", err)
		return false
	}
	for _, b := range brokers {
		addr := b.Addr()
		if !pattern.MatchString(addr) {
			errorlog.Printf("producer: failed to verify broker address %v", addr)
			return false
		}
		log.Printf("producer: broker address: %v", addr)
	}
	return true
}

// NewProducer creates a Kafka producer.
func NewProducer(input MessageSourceError) (p *Producer) {
	kp := api.GetKafkaParams()
	c := sarama.NewConfig()
	c.ClientID = "ProfileTest"
	c.Producer.Return.Successes = true
	c.Net.TLS.Enable = true
	c.Net.TLS.Config = api.Certificates()
	client, err := sarama.NewClient(kp.Brokers, c)
	if err != nil {
		errorlog.Print("producer:", err)
		return
	}
	sp, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		errorlog.Print("producer:", err)
		return
	}
	if !verify(client.Brokers()) {
		return
	}

	p = &Producer{
		source:       input,
		SyncProducer: sp,
		topic: map[Type]string{
			Profile: kp.ProfileTopic,
			Event:   kp.EventTopic,
			Log:     kp.LogTopic,
		},
	}
	return
}

// Serve activates p, causing it to send and receive messages.
func (p *Producer) Serve() {
	defer close(p.source.Err())
	defer p.Close()
	for m := range p.source.Output() {
		if err := p.send(m); err != nil {
			errorlog.Println("producer:", err)
			continue
		}
		p.source.Err() <- nil
	}
}

// send causes p to send m.
func (p *Producer) send(m Message) (err error) {
	if p == nil {
		return
	}
	b := m.Bytes
	log.Print("producer: sending message...")
	_, _, err = p.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic[m.Type],
		Value: sarama.ByteEncoder(b),
	})
	if err == nil {
		log.Println("producer: message sent:", string(b))
	}
	return
}
