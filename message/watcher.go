package message

import (
	"github.com/ESG-USA/Auklet-Client/app"
	"github.com/ESG-USA/Auklet-Client/kafka"
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
	source     kafka.MessageSource
	out        chan kafka.Message
	errd       bool
	eventTopic string
}

// NewExitWatcher returns a new ExitWatcher for the given input and app.
func NewExitWatcher(in kafka.MessageSource, app *app.App) *ExitWatcher {
	return &ExitWatcher{
		app:    app,
		source: in,
		out:    make(chan kafka.Message),
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
		e.out <- schema.NewExit(e.app)
	}
}

// Output returns e's output stream.
func (e *ExitWatcher) Output() <-chan kafka.Message {
	return e.out
}
