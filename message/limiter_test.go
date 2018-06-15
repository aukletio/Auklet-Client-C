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
	return source{
		size: size,
		num:  num,
		out:  make(chan broker.Message),
	}
}

// Serve causes s to generate one message per second.
func (s source) Serve() {
	m := make(msg, s.size)
	defer close(s.out)
	for i := 0; i < s.num; i++ {
		s.out <- m
		time.Sleep(time.Second)
	}
}

// Output returns s's output. The channel closes when s shuts down.
func (s source) Output() <-chan broker.Message {
	return s.out
}

type msg []byte

// Topic provides the topic of m.
func (m msg) Topic() (t broker.Topic) {
	return
}

// Bytes generates the byte content of m.
func (m msg) Bytes() ([]byte, error) {
	return []byte(m), nil
}

// newLimiter creates a DataLimiter with a budget of 4kB whose current period
// expires five seconds after its creation.
func newLimiter(s Source) *DataLimiter {
	l := NewDataLimiter(s, "limiter_test.json")
	l.Budget = 4000
	l.PeriodEnd = time.Now().Add(5 * time.Second)
	return l
}

func consume(s Source) (count int) {
	for m := range s.Output() {
		b, _ := m.Bytes()
		count += len(b)
	}
	return
}

func TestDataLimiter(t *testing.T) {
	s := newsource(15, 900)
	l := newLimiter(s)
	go s.Serve()
	go l.Serve()

	if count := consume(l); count > l.Budget {
		t.Errorf("expected <= %v, got %v", l.Budget, count)
	}
}
