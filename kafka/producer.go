// Package kafka provides a simple wrapper around sarama.SyncProducer.
package kafka

import (
	"log"

	"github.com/Shopify/sarama"

	"github.com/ESG-USA/Auklet-Client/api"
)

type Topic int

const (
	ProfileTopic Topic = iota
	EventTopic
	LogTopic
)

// Message is implemented by types that can be sent as Kafka messages.
type Message interface {
	// Topic returns the topic on which to send the Message.
	Topic() Topic

	// Bytes returns the Message as a byte slice. If err != nil, Send
	// logs the error and aborts.
	Bytes() ([]byte, error)
}

// Producer provides a simple Kafka producer.
type Producer struct {
	sarama.SyncProducer
	api.KafkaParams
	topic map[Topic]string
}

// New creates a Kafka producer.
func NewProducer() (p *Producer) {
	kp := api.GetKafkaParams()
	c := sarama.NewConfig()
	c.ClientID = "ProfileTest"
	c.Producer.Return.Successes = true
	c.Net.TLS.Enable = true
	c.Net.TLS.Config = api.Certificates()
	client, err := sarama.NewClient(kp.Brokers, c)
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
		KafkaParams:  kp,
		topic: map[Topic]string{
			ProfileTopic: kp.ProfileTopic,
			EventTopic:   kp.EventTopic,
			LogTopic:     kp.LogTopic,
		},
	}
	return
}

// Send causes p to send m.
func (p *Producer) Send(m Message) (err error) {
	if p == nil {
		return
	}
	b, err := m.Bytes()
	if err != nil {
		log.Print(err)
		return
	}
	_, _, err = p.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic[m.Topic()],
		Value: sarama.ByteEncoder(b),
	})
	if err == nil {
		log.Print(string(b))
	}
	return
}
