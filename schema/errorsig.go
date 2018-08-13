package schema

import (
	"encoding/json"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// errorSig represents the exit of an app in which an agent handled an "error
// signal" and produced a stacktrace.
type errorSig struct {
	metadata
	// Time is the time at which the event was received.
	Time time.Time `json:"timestamp"`
	// Status is the exit status of the application.
	Status int `json:"exitStatus"`
	// Signal is an integer value provided by an agent. As an output, it is
	// encoded as a string.
	Signal string `json:"signal"`
	// Trace is a stacktrace provided by an agent.
	Trace   []frame        `json:"stackTrace"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

type frame struct {
	Fn int64 `json:"functionAddress"`
	Cs int64 `json:"callSiteAddress"`
}

// ExitApp is an App that has an exit status.
type ExitApp interface {
	App
	Exiter
}

// NewErrorSig creates an ErrorSig for app out of raw message data. It assumes
// that app.Wait() has returned.
func NewErrorSig(data []byte, app ExitApp) broker.Message {
	var e errorSig
	err := json.Unmarshal(data, &e)
	if err != nil {
		e.Error = err.Error()
	}
	e.metadata = newMetadata(app)
	e.Time = time.Now()
	e.Status = app.ExitStatus()
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	return marshal(e, broker.Event)
}
