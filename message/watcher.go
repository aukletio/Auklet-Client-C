package message

import (
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

// ExitWatcher passes through all incoming messages. If it never sees a message
// of type schema.ErrorSig, it will generate a schema.Exit when its input
// closes.
//
// This ensures that we generate Exit events in situations where the agent did
// not generate a stacktrace.
type ExitWatcher struct {
	p          *broker.Persistor
	app        Watchable
	source     broker.MessageSource
	out        chan broker.Message
	errd       bool
	eventTopic string
}

// Watchable is an app that we can wait on to exit.
type Watchable interface {
	schema.SignalExitApp
	Wait()
}

// NewExitWatcher returns a new ExitWatcher for the given input and app.
func NewExitWatcher(in broker.MessageSource, app Watchable, p *broker.Persistor) *ExitWatcher {
	e := &ExitWatcher{
		p:      p,
		app:    app,
		source: in,
		out:    make(chan broker.Message),
		errd:   false,
	}
	go e.serve()
	return e
}

// serve activates e, causing it to send and receive Messages.
func (e *ExitWatcher) serve() {
	defer close(e.out)
	for m := range e.source.Output() {
		if m.Topic == broker.Event {
			e.errd = true
		}
		e.out <- m
	}
	if e.errd {
		return
	}
	e.app.Wait()
	m := schema.NewExit(e.app)
	if err := e.p.CreateMessage(&m); err != nil {
		// Let the backend know we ran out of local storage.
		e.out <- broker.Message{
			Error: err.Error(),
			Topic: broker.Log,
		}
		return
	}
	e.out <- m
}

// Output returns e's output stream.
func (e *ExitWatcher) Output() <-chan broker.Message {
	return e.out
}
