package schema

import (
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/device"
)

// NewExit creates an exit for app. It assumes that app.Wait() has returned.
func NewExit(app *app.App) (m broker.Message, err error) {
	var e Exit
	e.Application = &app.ID
	e.Checksum = &app.CheckSum
	e.PublicIP = device.CurrentIP()
	id := uuid.NewV4().String()
	e.Id = &id
	t := time.Now().String()
	e.Timestamp = &t
	ws := app.ProcessState.Sys().(syscall.WaitStatus)
	status := int32(ws.ExitStatus())
	e.ExitStatus = &status
	if ws.Signaled() {
		sig := ws.Signal().String()
		e.Signal = &sig
	}
	e.MacAddressHash = &device.MacHash
	e.SystemMetrics = device.GetMetrics()
	return broker.StdPersistor.CreateMessage(e, broker.Event)
}
