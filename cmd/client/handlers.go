package main

import (
	"log"

	"github.com/ESG-USA/Auklet-Client/producer"
	"github.com/ESG-USA/Auklet-Client/proxy"
	"github.com/ESG-USA/Auklet-Client/schema"
)

func profileHandler(data []byte) (m producer.Message, err error) {
	return schema.NewProfile(data, app, kp.ProfileTopic)
}

func errorsigHandler(data []byte) (m producer.Message, err error) {
	app.Cmd.Wait()
	errsigged = true
	log.Printf("app %v exited with error signal", app.Path)
	return schema.NewErrorSig(data, app, kp.EventTopic)
}

func logHandler(data []byte) (m producer.Message, err error) {
	m = schema.NewLog(data, kp.LogTopic)
	return
}

var customHandlers = map[string]proxy.Handler{
	"profile": profileHandler,
	"event":   errorsigHandler,
	"log":     logHandler,
}
