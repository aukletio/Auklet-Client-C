package schema

import (
	"encoding/json"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/device"
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// exit represents the exit of an app in which libauklet did not handle a
// signal. The app may or may not have been delivered a termination signal of
// some kind, but not one handled by libauklet. See man 7 signal for details.
type exit struct {
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
	Status  int            `json:"exitStatus"`
	Signal  sig            `json:"signal,omitempty"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

// NewExit creates an exit for app. It assumes that app.Wait() has returned.
func NewExit(app *app.App) (m kafka.Message, err error) {
	var e exit
	e.AppID = app.ID
	e.CheckSum = app.CheckSum
	e.IP = device.CurrentIP()
	e.UUID = uuid.NewV4().String()
	e.Time = time.Now()
	ws := app.ProcessState.Sys().(syscall.WaitStatus)
	e.Status = ws.ExitStatus()
	if ws.Signaled() {
		e.Signal = sig(ws.Signal())
	}
	e.MacHash = device.MacHash
	e.Metrics = device.GetMetrics()
	b, err := json.MarshalIndent(e, "", "\t")
	if err != nil {
		return
	}
	return kafka.StdPersistor.CreateMessage(b, kafka.Event)
}
