package message

import (
	"testing"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type source struct {
	size, num int
	out       chan broker.Message
}

// newsource creates a source that generates num messages of the given size in
// bytes.
func newsource(num, size int) source {
	s := source{
		size: size,
		num:  num,
		out:  make(chan broker.Message),
	}
	go s.serve()
	return s
}

// serve causes s to generate one message per second.
func (s source) serve() {
	defer close(s.out)
	for i := 0; i < s.num; i++ {
		s.out <- broker.Message{
			Bytes: make([]byte, s.size),
		}
		time.Sleep(time.Second)
	}
}

// Output returns s's output. The channel closes when s shuts down.
func (s source) Output() <-chan broker.Message {
	return s.out
}

// newLimiter creates a DataLimiter with a budget of 4kB whose current period
// expires five seconds after its creation.
func newLimiter(s broker.MessageSource) *DataLimiter {
	l := NewDataLimiter(s, new(MemPersistor))
	l.Budget = new(int)
	*l.Budget = 4000
	l.PeriodEnd = time.Now().Add(5 * time.Second)
	return l
}

func consume(s broker.MessageSource) (count int) {
	for m := range s.Output() {
		count += len(m.Bytes)
	}
	return
}

func TestDataLimiter(t *testing.T) {
	s := newsource(4, 1100)
	l := newLimiter(s)

	if count := consume(l); count > *l.Budget {
		t.Errorf("expected <= %v, got %v", *l.Budget, count)
	}
}