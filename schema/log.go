package schema

import (
	"encoding/json"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// NewAgentLog converts data into an AgentLog and returns it as a broker
// message.
func NewAgentLog(data []byte) (m broker.Message, err error) {
	var a AgentLog
	err = json.Unmarshal(data, &a)
	if err != nil {
		return
	}
	return broker.StdPersistor.CreateMessage(a, broker.Log)
}
