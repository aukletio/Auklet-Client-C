package agent

import (
	"errors"
	"io"
	"testing"
)

func TestLogger(t *testing.T) {
	data := `hello
world`
	r, w := io.Pipe()
	logger := NewLogger(r)
	if _, err := w.Write([]byte(data)); err != nil {
		t.Error(err)
	}
	w.CloseWithError(errors.New("io error"))

	m := <-logger.Output()
	exp := Message{
		Data: []byte("hello"),
		Type: "applog",
	}
	if !compare(m, exp) {
		t.Errorf("expected %v, got %v", exp, m.Data)
	}

	m = <-logger.Output()
	exp = Message{
		Data: []byte("world"),
		Type: "applog",
	}
	if !compare(m, exp) {
		t.Errorf("expected %v, got %v", exp, m.Data)
	}

	m = <-logger.Output()
	exp = Message{
		Error: "io error",
		Type:  "log",
	}
	if !compare(m, exp) {
		t.Errorf("expected %v, got %v", exp, m.Data)
	}
}
