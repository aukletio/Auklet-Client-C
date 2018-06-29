package schema

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// NewAppLog converts msg into a custom log message.
func NewAppLog(msg []byte, app *app.App) (m broker.Message, err error) {
	var a AppLog
	a.Application = &app.ID
	a.Checksum = &app.CheckSum
	a.PublicIP = device.CurrentIP()
	id := uuid.NewV4().String()
	a.Id = &id
	t := time.Now().String()
	a.Timestamp = &t
	a.MacAddressHash = &device.MacHash
	a.SystemMetrics = device.GetMetrics()
	a.Message = msg
	return broker.StdPersistor.CreateMessage(a, broker.Event)
}
