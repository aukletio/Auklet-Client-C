package schema

import (
	"encoding/json"
)

// A Log represents a raw JSON message.
type Log struct {
	// Kafka conumsers do not require any particular schema for logs.
	raw        json.RawMessage
	kafkaTopic string
}

// NewLog converts data into a Log that can be sent to topic.
func NewLog(data []byte, topic string) (l Log) {
	l.raw = json.RawMessage(data)
	l.kafkaTopic = topic
	return
}

// Topic returns the Kafka topic to which l should be sent.
func (l Log) Topic() string {
	return l.kafkaTopic
}

// Bytes returns the Log as a byte slice.
func (l Log) Bytes() (b []byte, err error) {
	return json.Marshal(l.raw)
}
