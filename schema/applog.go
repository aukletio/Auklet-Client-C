package schema

import (
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// appLog represents custom log data as expected by broker consumers.
type appLog struct {
	metadata
	// Message is the log message sent by the application.
	Message []byte         `json:"message"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

// newAppLog converts msg into a custom log message.
func newAppLog(msg []byte, app App) appLog {
	return appLog{
		metadata: newMetadata(app),
		MacHash:  device.MacHash,
		Metrics:  device.GetMetrics(),
		Message:  msg,
	}
}
