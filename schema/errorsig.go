// +build linux

package schema

import (
	"encoding/json"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// errorSig represents the exit of an app in which libauklet handled an "error
// signal" and produced a stacktrace.
type errorSig struct {
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

	// Status is the exit status of the application as accessible through
	// App.Wait.
	Status int `json:"exitStatus"`

	// Signal is an integer value provided by libauklet. In JSON output, it
	// is represented as a string.
	Signal sig `json:"signal"`

	// Trace is a stacktrace provided by libauklet.
	Trace   json.RawMessage `json:"stackTrace"`
	MacHash string          `json:"macAddressHash"`
	Metrics device.Metrics  `json:"systemMetrics"`
}

// NewErrorSig creates an ErrorSig for app out of JSON data. It assumes that
// app.Wait() has returned.
func NewErrorSig(data []byte, app *app.App) (m broker.Message, err error) {
	var e errorSig
	err = json.Unmarshal(data, &e)
	if err != nil {
		return
	}
	e.AppID = app.ID
	e.CheckSum = app.CheckSum
	e.IP = device.CurrentIP()
	e.UUID = uuid.NewV4().String()
	e.Time = time.Now()
	e.Status = app.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	b, err := json.MarshalIndent(e, "", "\t")
	if err != nil {
		return
	}
	return broker.StdPersistor.CreateMessage(b, broker.Event)
}

type sig syscall.Signal

// String returns s represented as a human-readable string.
func (s sig) String() string {
	return syscall.Signal(s).String()
}

// MarshalText encodes sig as a human-readable string.
func (s sig) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}
