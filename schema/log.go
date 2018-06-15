package schema

import (
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// NewLog converts data into a Log that can be sent to topic.
func NewLog(data []byte) (m broker.Message, err error) {
	// logs are not formatted in any particular way.
	return broker.StdPersistor.CreateMessage(data, broker.Log)
}
