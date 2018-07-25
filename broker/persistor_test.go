package broker

import (
	"testing"

	"github.com/spf13/afero"
)

func init() {
	fs = afero.NewMemMapFs()
}

func TestPersistor(t *testing.T) {
	p := NewPersistor("")
	var limit int64 = 900
	p.Configure() <- &limit
	m := make([]byte, 500)
	m1, err := p.CreateMessage(m, 0)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	defer m1.Remove()
	m2, err = p.CreateMessage(m, 0)
	if _, is := err.(ErrStorageFull); !is {
		t.Errorf("expected ErrStorageFull, got %v", err)
	}
	defer m2.Remove()
}
