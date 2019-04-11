package schema

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"

	"github.com/aukletio/Auklet-Client-C/device"
	"github.com/aukletio/Auklet-Client-C/version"
)

type metadata struct {
	Version       string `json:"version"` // user-defined version
	Username      string `json:"device"`
	ClientVersion string `json:"clientVersion"`
	AgentVersion  string `json:"agentVersion"`
	AppID         string `json:"application"`
	CheckSum      string `json:"release"` // SHA512/224 hash of the executable
	MacHash       string `json:"macAddressHash"`
	IP            string `json:"publicIP"`  // current public IP address
	UUID          string `json:"id"`        // identifier for this message
	Time          int64  `json:"timestamp"` // Unix milliseconds
	Error         string `json:"error,omitempty"`
}

func nowMilli() int64 {
	return time.Now().UnixNano() / 1000000 // milliseconds
}

func (c Converter) metadata() metadata {
	return metadata{
		Version:       c.UserVersion,
		Username:      c.Username,
		ClientVersion: version.Version,
		AgentVersion:  c.App.AgentVersion(),
		AppID:         c.AppID,
		CheckSum:      c.App.CheckSum(),
		MacHash:       c.MacHash,
		IP:            device.CurrentIP(),
		UUID:          uuid.NewV4().String(),
		Time:          nowMilli(),
	}
}

// appLog represents custom log data as expected by broker consumers.
type appLog struct {
	metadata
	// Message is the log message sent by the application.
	Message []byte         `json:"message"`
	Metrics device.Metrics `json:"systemMetrics"`
}

func (c Converter) appLog(msg []byte) appLog {
	return appLog{
		metadata: c.metadata(),
		Metrics:  c.Monitor.GetMetrics(),
		Message:  msg,
	}
}

// profile represents profile data as expected by broker consumers.
type profile struct {
	metadata
	// Tree represents the profile tree data generated by an agent.
	Tree node `json:"tree"`
}

type node struct {
	Fn       *int64 `json:"functionAddress"`
	Cs       int64  `json:"callSiteAddress"`
	Ncalls   int    `json:"nCalls"`
	Nsamples int    `json:"nSamples"`
	Callees  []node `json:"callees"`
}

func (c Converter) profile(data []byte) profile {
	var p profile
	err := json.Unmarshal(data, &p)
	if err != nil {
		p.Error = err.Error()
	}
	p.metadata = c.metadata()
	return p
}

// errorSig represents the exit of an app in which an agent handled an "error
// signal" and produced a stacktrace.
type errorSig struct {
	metadata
	Status  int            `json:"exitStatus"`
	Signal  string         `json:"signal"`
	Trace   []frame        `json:"stackTrace"`
	Metrics device.Metrics `json:"systemMetrics"`
}

type frame struct {
	Fn *int64 `json:"functionAddress"`
	Cs int64  `json:"callSiteAddress"`
}

func (c Converter) errorSig(data []byte) errorSig {
	var e errorSig
	err := json.Unmarshal(data, &e)
	if err != nil {
		e.Error = err.Error()
	}
	e.metadata = c.metadata()
	e.Status = c.App.ExitStatus()
	e.Metrics = c.Monitor.GetMetrics()
	return e
}

// exit represents the exit of an app in which an agent did not handle a
// signal. The app may or may not have been delivered a termination signal of
// some kind, but not one handled by an agent. See man 7 signal for details.
type exit struct {
	metadata
	Status  int            `json:"exitStatus"`
	Signal  string         `json:"signal"`
	Metrics device.Metrics `json:"systemMetrics"`
}

func (c Converter) exit() exit {
	return exit{
		metadata: c.metadata(),
		Status:   c.App.ExitStatus(),
		Signal:   c.App.Signal(),
		Metrics:  c.Monitor.GetMetrics(),
	}
}

type dataPoint struct {
	metadata
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (c Converter) dataPoint(data []byte) dataPoint {
	var d dataPoint
	if err := json.Unmarshal(data, &d); err != nil {
		d.Error = err.Error()
	}
	d.metadata = c.metadata()
	return d
}
