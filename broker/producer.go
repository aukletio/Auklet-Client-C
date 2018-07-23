// Package broker provides a simple wrapper around sarama.SyncProducer.
package broker

import (
	"log"
	"regexp"
	"time"

	"github.com/Shopify/sarama"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// Producer provides a simple broker producer.
type Producer struct {
	source MessageSource
	sarama.SyncProducer
	topic map[Topic]string
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

// NewProducer creates a broker producer.
func NewProducer(input MessageSource) (p *Producer) {
	kp := api.GetBrokerParams()
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
		topic: map[Topic]string{
			Profile: kp.ProfileTopic,
			Event:   kp.EventTopic,
			Log:     kp.LogTopic,
		},
	}
	return
}

// Serve activates p, causing it to send and receive messages.
func (p *Producer) Serve() {
	defer p.Close()
	for m := range p.source.Output() {
		p.send(m)
	}
}

// send causes p to send m. If the message fails to be sent, it will be retried
// at most ten times.
func (p *Producer) send(m Message) {
	if p == nil {
		return
	}
	for i := 0; i < 10; i++ {
		_, _, err := p.SendMessage(&sarama.ProducerMessage{
			Topic: p.topic[m.Topic],
			Value: sarama.ByteEncoder(m.Bytes),
		})
		if err == nil {
			log.Printf("producer: message sent: %+q", string(m.Bytes))
			m.Remove()
			return
		}
		errorlog.Printf("producer: send attempt %v: %v", i+1, err)
		time.Sleep(time.Second)
	}
}
