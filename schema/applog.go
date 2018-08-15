package schema

import (
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// appLog represents custom log data as expected by broker consumers.
type appLog struct {
	metadata
	// Time is the Unix epoch time (in milliseconds) at which the
	// message was received.
	Time int64 `json:"timestamp"`
	// Message is the log message sent by the application.
	Message []byte         `json:"message"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

// NewAppLog converts msg into a custom log message.
func NewAppLog(msg []byte, app App) broker.Message {
	a := appLog{
		metadata: newMetadata(app),
		Time:     nowMilli(),
		MacHash:  device.MacHash,
		Metrics:  device.GetMetrics(),
		Message:  msg,
	}
	return marshal(a, broker.Event)
}
