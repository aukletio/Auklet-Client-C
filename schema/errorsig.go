package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// ErrorSig represents the exit of an app in which an agent handled an "error
// signal" and produced a stacktrace.
type ErrorSig struct {
	AppID string `json:"application"`
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

	// Status is the exit status of the application.
	Status int `json:"exitStatus"`

	// Signal is an integer value provided by an agent. As an output, it is
	// encoded as a string.
	Signal string `json:"signal"`

	// Trace is a stacktrace provided by an agent.
	Trace   json.RawMessage `json:"stackTrace"`
	MacHash string          `json:"macAddressHash"`
	Metrics device.Metrics  `json:"systemMetrics"`
}

// ExitApp is an App that has an exit status.
type ExitApp interface {
	App
	Exiter
}

// NewErrorSig creates an ErrorSig for app out of raw message data. It assumes
// that app.Wait() has returned.
func NewErrorSig(data []byte, app ExitApp) (m broker.Message, err error) {
	var e ErrorSig
	err = json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	e.AppID = app.ID()
	e.CheckSum = app.CheckSum()
	e.IP = device.CurrentIP()
	e.UUID = uuid.NewV4().String()
	e.Time = time.Now()
	e.Status = app.ExitStatus()
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	return broker.StdPersistor.CreateMessage(e, broker.Event)
}
