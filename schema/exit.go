package schema

import (
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// exit represents the exit of an app in which an agent did not handle a
// signal. The app may or may not have been delivered a termination signal of
// some kind, but not one handled by an agent. See man 7 signal for details.
type exit struct {
	metadata
	// Status is the exit status of the application as accessible through
	// App.Wait.
	Status  int            `json:"exitStatus"`
	Signal  string         `json:"signal,omitempty"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

// newExit creates an exit for app.
func newExit(app App, signal string, exitStatus int) exit {
	return exit{
		metadata: newMetadata(app),
		Status:   exitStatus,
		Signal:   signal,
		MacHash:  device.MacHash,
		Metrics:  device.GetMetrics(),
	}
}
