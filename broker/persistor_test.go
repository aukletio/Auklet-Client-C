package broker

import (
	"testing"

	"github.com/spf13/afero"
)

func init() {
	fs = afero.NewMemMapFs()
}

func TestPersistor(t *testing.T) {
	p := NewPersistor(".auklet/message")
	var limit int64 = 900
	p.Configure() <- &limit
	m := Message{
		Bytes: make([]byte, 500),
		Topic: 0,
		Error: "",
	}
	err := p.CreateMessage(m)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	defer m.Remove()
	err = p.CreateMessage(m)
	if _, is := err.(ErrStorageFull); !is {
		t.Errorf("expected ErrStorageFull, got %v", err)
	}
	defer m.Remove()
}
