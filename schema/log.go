package schema

import (
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// NewLog converts data into a Log that can be sent to topic.
func NewLog(data []byte) (m kafka.Message, err error) {
	// logs are not formatted in any particular way.
	return kafka.StdPersistor.CreateMessage(data, kafka.Log)
}
