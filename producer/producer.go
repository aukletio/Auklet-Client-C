// Package producer provides a simple wrapper around sarama.SyncProducer.
package producer

import (
	"crypto/tls"
	"log"

	"github.com/Shopify/sarama"
)

// Message is implemented by types that can be sent as Kafka messages.
type Message interface {
	// Topic returns the topic on which to send the Message.
	Topic() string

	// Bytes returns the Message as a byte slice. If err != nil, Send
	// logs the error and aborts.
	Bytes() ([]byte, error)
}

// Producer provides a simple Kafka producer.
type Producer struct {
	sarama.SyncProducer

	// LogTopic determines the Kafka topic on which Write is to send
	// messages.
	LogTopic string
}

// New creates a Kafka producer with TLS config tc, broker list brokers,
// and certain default settings.
func New(brokers []string, tc *tls.Config) (p *Producer) {
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
		SyncProducer: sp,
	}
	return
}

// Send causes p to send m.
func (p *Producer) Send(m Message) (err error) {
	b, err := m.Bytes()
	if err != nil {
		log.Print(err)
		return
	}
	_, _, err = p.SendMessage(&sarama.ProducerMessage{
		Topic: m.Topic(),
		Value: sarama.ByteEncoder(b),
	})
	if err == nil {
		log.Print(string(b))
	}
	return
}

// Write allows p to be used as a logging service.
func (p *Producer) Write(q []byte) (n int, err error) {
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
