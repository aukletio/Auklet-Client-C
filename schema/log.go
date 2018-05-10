package schema

import (
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// A Log represents a raw JSON message.
type Log struct {
	// Kafka conumsers do not require any particular schema for logs.
	raw []byte
}

// NewLog converts data into a Log that can be sent to topic.
func NewLog(data []byte) (l Log) {
	l.raw = data
	return
}

// Topic returns the Kafka topic to which l should be sent.
func (l Log) Topic() kafka.Topic {
	return kafka.LogTopic
}

// Bytes returns the Log as a byte slice.
func (l Log) Bytes() ([]byte, error) {
	return l.raw, nil
}
