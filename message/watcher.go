package message

import (
	"github.com/ESG-USA/Auklet-Client-C/app"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/schema"
)

// ExitWatcher passes through all incoming messages. If it never sees a message
// of type schema.ErrorSig, it will generate a schema.Exit when its input
// closes.
//
// This ensures that we generate Exit events in situations where the agent did
// not generate a stacktrace.
type ExitWatcher struct {
	app        *app.App
	source     broker.MessageSource
	out        chan broker.Message
	errd       bool
	eventTopic string
}

// NewExitWatcher returns a new ExitWatcher for the given input and app.
func NewExitWatcher(in broker.MessageSource, app *app.App) *ExitWatcher {
	return &ExitWatcher{
		app:    app,
		source: in,
		out:    make(chan broker.Message),
		errd:   false,
	}
}

// Serve activates e, causing it to send and receive Messages.
func (e *ExitWatcher) Serve() {
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
	m, err := schema.NewExit(e.app)
	if err != nil {
		errorlog.Print(err)
		return
	}
	e.out <- m
}

// Output returns e's output stream.
func (e *ExitWatcher) Output() <-chan broker.Message {
	return e.out
}
