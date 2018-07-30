package agent

import (
	"io"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// PeriodicRequester periodically sends emission requests over a connection.
type PeriodicRequester struct {
	conf chan int  // provides the period in seconds
	out  io.Writer // the connection on which to emit requests
}

// NewPeriodicRequester creates a PeriodicRequester that sends rqeuests over
// conn.
func NewPeriodicRequester(conn io.Writer) PeriodicRequester {
	r := PeriodicRequester{
		conf: make(chan int),
		out:  conn,
	}
	go r.run()
	return r
}

// Configure returns a channel on which the request period in seconds can be
// set.
func (r PeriodicRequester) Configure() chan<- int {
	return r.conf
}

func (r PeriodicRequester) run() {
	emit := time.NewTicker(time.Second)
	for {
		select {
		case <-emit.C:
			if _, err := r.out.Write([]byte{0}); err != nil {
				errorlog.Println("PeriodicRequester.run:", err)
			}
		case dur := <-r.conf:
			emit.Stop()
			if dur > 0 {
				emit = time.NewTicker(time.Duration(dur) * time.Second)
			}
		}
	}
}
