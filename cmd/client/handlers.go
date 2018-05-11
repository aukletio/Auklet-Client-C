package main

import (
	"log"

	"github.com/ESG-USA/Auklet-Client/agent"
	"github.com/ESG-USA/Auklet-Client/kafka"
	"github.com/ESG-USA/Auklet-Client/schema"
)

func profileHandler(data []byte) (m kafka.Message, err error) {
	return schema.NewProfile(data, app)
}

func errorsigHandler(data []byte) (m kafka.Message, err error) {
	app.Cmd.Wait()
	log.Printf("app %v exited with error signal", app.Path)
	return schema.NewErrorSig(data, app)
}

func logHandler(data []byte) (m kafka.Message, err error) {
	m = schema.NewLog(data)
	return
}

var customHandlers = map[string]agent.Handler{
	"profile": profileHandler,
	"event":   errorsigHandler,
	"log":     logHandler,
}
