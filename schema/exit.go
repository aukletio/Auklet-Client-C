package schema

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
	"github.com/ESG-USA/Auklet-Client-C/version"
)

// exit represents the exit of an app in which an agent did not handle a
// signal. The app may or may not have been delivered a termination signal of
// some kind, but not one handled by an agent. See man 7 signal for details.
type exit struct {
	Version string `json:"clientVersion"`
	AppID   string `json:"application"`
	// CheckSum is the SHA512/224 hash of the executable, used to associate
	// event data with a particular release.
	CheckSum string `json:"checksum"`

	// IP is the public IP address of the device on which we are running,
	// used to associate event data with an estimated geographic location.
	IP string `json:"publicIP"`

	// UUID is a unique identifier for a particular event.
	UUID string `json:"id"`

	// Time is the time at which the event was received.
	Time time.Time `json:"timestamp"`

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
		Version:  version.Version,
		AppID:    app.ID(),
		CheckSum: app.CheckSum(),
		IP:       device.CurrentIP(),
		UUID:     uuid.NewV4().String(),
		Time:     time.Now(),
		Status:   app.ExitStatus(),
		Signal:   app.Signal(),
		MacHash:  device.MacHash,
		Metrics:  device.GetMetrics(),
	}
	return marshal(e, broker.Event)
}
