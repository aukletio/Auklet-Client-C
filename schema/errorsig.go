package schema

import (
	"encoding/json"

	"github.com/ESG-USA/Auklet-Client-C/device"
)

// errorSig represents the exit of an app in which an agent handled an "error
// signal" and produced a stacktrace.
type errorSig struct {
	metadata
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

// newErrorSig creates an ErrorSig for app out of raw message data.
func newErrorSig(data []byte, app App, exitStatus int) errorSig {
	var e errorSig
	err := json.Unmarshal(data, &e)
	if err != nil {
		e.Error = err.Error()
	}
	e.metadata = newMetadata(app)
	e.Status = exitStatus
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	return e
}
