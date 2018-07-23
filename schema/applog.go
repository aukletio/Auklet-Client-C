package schema

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// AppLog represents custom log data as expected by broker consumers.
type AppLog struct {
	// AppID is a long string uniquely associated with a particular app.
	AppID string `json:"application"`

	// CheckSum is the SHA512/224 hash of the executable, used to associate
	// message data with a particular release.
	CheckSum string `json:"checksum"`

	// IP is the public IP address of the device on which we are running,
	// used to associate message data with an estimated geographic
	// location.
	IP string `json:"publicIP"`

	// UUID is a unique identifier for a particular message.
	UUID string `json:"id"`

	// Time is the Unix epoch time (in milliseconds) at which the
	// message was received.
	Time time.Time `json:"timestamp"`

	// Message is the log message sent by the application.
	Message []byte         `json:"message"`
	MacHash string         `json:"macAddressHash"`
	Metrics device.Metrics `json:"systemMetrics"`
}

// NewAppLog converts msg into a custom log message.
func NewAppLog(msg []byte, app *app.App) (m broker.Message, err error) {
	var a AppLog
	a.AppID = app.ID
	a.CheckSum = app.CheckSum
	a.IP = device.CurrentIP()
	a.UUID = uuid.NewV4().String()
	a.Time = time.Now()
	a.MacHash = device.MacHash
	a.Metrics = device.GetMetrics()
	a.Message = msg
	return broker.StdPersistor.CreateMessage(a, broker.Log)
}