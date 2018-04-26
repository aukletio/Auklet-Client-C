package message

import (
	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/schema"
)

// ExitWatcher passes through all incoming messages. If it never sees a message
// of type schema.ErrorSig, it will generate a schema.Exit when its input
// closes.
//
// This ensures that we generate Exit events in situations where the agent did
// not generate a stacktrace.
type ExitWatcher struct {
	app        *app.App
	source     Source
	out        chan Message
	errd       bool
	eventTopic string
}

// NewExitWatcher returns a new ExitWatcher for the given input, app, and
// eventTopic.
func NewExitWatcher(in Source, app *app.App, eventTopic string) *ExitWatcher {
	return &ExitWatcher{
		app:    app,
		source: in,
		out:    make(chan Message),
		errd:   false,
	}
}

// Serve activates e, causing it to send and receive Messages.
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

// Output returns e's output stream.
func (e *ExitWatcher) Output() <-chan Message {
	return e.out
}
