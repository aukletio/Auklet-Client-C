package agent

import (
	"io"
	"testing"
	"time"
)

func TestRequester(t *testing.T) {
	r, w := io.Pipe()
	req := NewPeriodicRequester(w)
	req.Configure() <- 1
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil || n != 1 {
		t.Fail()
	}
	r.Close()
	// this will generate an error, if we wait long enough
	timeout := time.After(time.Second * 4)
	select {
	case <-req.Output():
	case <-timeout:
		t.Fail()
	}
}
