package schema

import (
	"encoding/json"
)

type Log struct {
	// Kafka conumsers do not require any particular schema for logs.
	raw        json.RawMessage
	kafkaTopic string `json:"-"`
}

func NewLog(data []byte, topic string) (l Log) {
	l.raw = json.RawMessage(data)
	l.kafkaTopic = topic
	return
}

func (l Log) Topic() string {
	return l.kafkaTopic
}

func (l Log) Bytes() (b []byte, err error) {
	return json.Marshal(l.raw)
}
