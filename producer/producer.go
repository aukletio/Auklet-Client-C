// Package producer provides a simple wrapper around sarama.SyncProducer.
package producer

import (
	"crypto/tls"
	"log"

	"github.com/Shopify/sarama"

	"github.com/ESG-USA/Auklet-Client/message"
)

// Producer provides a simple Kafka producer.
type Producer struct {
	source message.SourceError
	sarama.SyncProducer

	// LogTopic determines the Kafka topic on which Write is to send
	// messages.
	LogTopic string
}

// New creates a Kafka producer with TLS config tc, broker list brokers,
// and certain default settings.
func New(input message.SourceError, brokers []string, tc *tls.Config) (p *Producer) {
	c := sarama.NewConfig()
	c.ClientID = "ProfileTest"
	c.Producer.Return.Successes = true
	c.Net.TLS.Enable = true
	c.Net.TLS.Config = tc
	client, err := sarama.NewClient(brokers, c)
	if err != nil {
		log.Print(err)
		return
	}
	sp, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		log.Print(err)
		return
	}
	for _, b := range client.Brokers() {
		log.Printf("broker address: %v", b.Addr())
	}
	p = &Producer{
		source:       input,
		SyncProducer: sp,
	}
	return
}

// Serve activates p, causing it to send and receive messages.
func (p *Producer) Serve() {
	defer close(p.source.Err())
	defer p.Close()
	for m := range p.source.Output() {
		if err := p.send(m); err != nil {
			// We can't log the error to the remote logger because
			// we ARE the remote logger.
			log.Print(err)
			continue
		}
		p.source.Err() <- nil
	}
}

// send causes p to send m.
func (p *Producer) send(m message.Message) (err error) {
	if p == nil {
		return
	}
	b, err := m.Bytes()
	if err != nil {
		log.Print(err)
		return
	}
	log.Print("producer: sending message...")
	_, _, err = p.SendMessage(&sarama.ProducerMessage{
		Topic: m.Topic(),
		Value: sarama.ByteEncoder(b),
	})
	log.Print("producer: message sent")
	if err == nil {
		log.Print(string(b))
	}
	return
}

// Write allows p to be used as a logging service.
func (p *Producer) Write(q []byte) (n int, err error) {
	if p == nil {
		return
	}
	_, _, err = p.SendMessage(&sarama.ProducerMessage{
		Topic: p.LogTopic,
		Value: sarama.ByteEncoder(q),
		Key:   sarama.ByteEncoder("c"),
	})
	if err == nil {
		n = len(q)
	}
	return
}
