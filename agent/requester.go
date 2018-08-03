package agent

import (
	"io"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// PeriodicRequester periodically sends emission requests over a connection.
type PeriodicRequester struct {
	conf chan int  // provides the period in seconds
	conn io.Writer // the connection on which to emit requests
	out  chan broker.Message
}

// NewPeriodicRequester creates a PeriodicRequester that sends rqeuests over
// conn.
func NewPeriodicRequester(conn io.Writer) PeriodicRequester {
	r := PeriodicRequester{
		conf: make(chan int),
		conn: conn,
		out:  make(chan broker.Message),
	}
	go r.run()
	return r
}

// Configure returns a channel on which the request period in seconds can be
// set.
func (r PeriodicRequester) Configure() chan<- int {
	return r.conf
}

// Output returns r's output channel, which might generate an error message.
func (r PeriodicRequester) Output() <-chan broker.Message {
	return r.out
}

func (r PeriodicRequester) run() {
	defer close(r.out)
	emit := time.NewTicker(time.Second)
	var prevErr error
	for {
		select {
		case <-emit.C:
			if _, err := r.conn.Write([]byte{0}); err != nil {
				if prevErr != nil {
					// This is our second write error. A
					// single write error sometimes happens
					// when the application exits just
					// before emit fires. But two errors in
					// a row suggests that we've
					// unexpectedly lost our connection to
					// the agent.
					r.out <- broker.Message{
						Error: err.Error(),
						Topic: broker.Log,
					}
					return
				}
				prevErr = err
			}
		case dur := <-r.conf:
			emit.Stop()
			if dur > 0 {
				emit = time.NewTicker(time.Duration(dur) * time.Second)
			}
		}
	}
}
