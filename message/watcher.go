package message

import (
	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/schema"
)

type ExitWatcher struct {
	app        *app.App
	source     Source
	out        chan Message
	errd       bool
	eventTopic string
}

func NewExitWatcher(in Source, app *app.App, eventTopic string) *ExitWatcher {
	return &ExitWatcher{
		app:    app,
		source: in,
		out:    make(chan Message),
		errd:   false,
	}
}

func (e *ExitWatcher) Serve() {
	defer close(e.out)
	for m := range e.source.Output() {
		if _, is := m.(schema.ErrorSig); is {
			e.errd = true
		}
		e.out <- m
	}
	if !e.errd {
		e.app.Wait()
		e.out <- schema.NewExit(e.app, e.eventTopic)
	}
}

func (e *ExitWatcher) Output() <-chan Message {
	return e.out
}
