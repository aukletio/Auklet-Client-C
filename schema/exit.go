package schema

import (
	"github.com/ESG-USA/Auklet-Client-C/broker"
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

// SignalExitApp is an app with an exit status and signal description.
type SignalExitApp interface {
	App
	Exiter
	Signaller
}

// NewExit creates an exit for app. It assumes that app.Wait() has returned.
func NewExit(app SignalExitApp) broker.Message {
	e := exit{
		metadata: newMetadata(app),
		Status:   app.ExitStatus(),
		Signal:   app.Signal(),
		MacHash:  device.MacHash,
		Metrics:  device.GetMetrics(),
	}
	return marshal(e, broker.Event)
}
