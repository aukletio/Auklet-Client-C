package agent

import (
	"io"
	"time"

	"github.com/aukletio/Auklet-Client-C/broker"
)

// PeriodicRequester periodically sends emission requests over a connection.
type PeriodicRequester struct {
	conf <-chan int // provides the period in seconds; should never be closed
	conn io.Writer  // the connection on which to emit requests
	out  chan broker.Message
	done <-chan struct{} // cancellation requests
}

// NewPeriodicRequester creates a PeriodicRequester that sends requests over
// conn. When done closes, the requester closes its output and terminates.
func NewPeriodicRequester(conn io.Writer, done <-chan struct{}, conf <-chan int) PeriodicRequester {
	r := PeriodicRequester{
		conf: conf,
		conn: conn,
		out:  make(chan broker.Message),
		done: done,
	}
	go r.run()
	return r
}

// Output returns r's output channel, which might generate an error message.
func (r PeriodicRequester) Output() <-chan broker.Message {
	return r.out
}

func (r PeriodicRequester) run() {
	defer close(r.out)
	emit := time.NewTicker(time.Minute)
	var prevErr error
	for {
		select {
		case <-r.done:
			return
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
