package schema

import (
	"encoding/json"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// NewAgentLog converts data into a broker message.
func NewAgentLog(data []byte) (m broker.Message, err error) {
	return broker.StdPersistor.CreateMessage(json.RawMessage(data), broker.Log)
}
