package agent

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestRequester(t *testing.T) {
	r, w := io.Pipe()
	conf := make(chan int)
	req := NewPeriodicRequester(w, nil, conf)
	conf <- 1
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

func TestRequesterDone(t *testing.T) {
	done := make(chan struct{})
	req := NewPeriodicRequester(&bytes.Buffer{}, done, nil)
	// terminate the requester
	close(done)
	if _, open := <-req.Output(); open {
		t.Fail()
	}
}
