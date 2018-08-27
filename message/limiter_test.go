package message

import (
	"testing"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type source chan broker.Message

// Output returns s's output. The channel closes when s shuts down.
func (s source) Output() <-chan broker.Message {
	return s
}

func intPtr(value int) *int {
	return &value
}

// newLimiter creates a DataLimiter with a budget of 4kB whose current period
// expires five seconds after its creation.
func newConfig(budget *int, periodEnd time.Time) Persistor {
	store := new(MemPersistor)
	store.Save(&DataLimiter{
		Budget:    budget,
		PeriodEnd: periodEnd,
	})

	return store
}

func consume(s broker.MessageSource) (count int) {
	for m := range s.Output() {
		count += len(m.Bytes)
	}
	return
}

func TestDataLimiter(t *testing.T) {
	cases := []struct {
		conf     func() Persistor
		generate func(source, *DataLimiter)
		consume  func(*DataLimiter)
	}{
		{
			conf: func() Persistor {
				return newConfig(intPtr(4000), time.Now().Add(50*time.Millisecond))
			},
			generate: func(s source, _ *DataLimiter) {
				defer close(s)
				for i := 0; i < 4; i++ {
					s <- broker.Message{Bytes: make([]byte, 1100)}
					time.Sleep(10 * time.Millisecond)
				}
			},
			consume: func(l *DataLimiter) {
				if count := consume(l); count > *l.Budget {
					t.Errorf("expected <= %v, got %v", *l.Budget, count)
				}
			},
		},
		{
			conf: func() Persistor {
				return newConfig(intPtr(4000), time.Now().Add(50*time.Millisecond))
			},
			generate: func(s source, l *DataLimiter) {
				defer close(s)
				for i := 0; i < 2; i++ {
					s <- broker.Message{Bytes: make([]byte, 1100)}
					time.Sleep(10 * time.Millisecond)
				}
				l.Configure() <- api.CellularConfig{
					Limit: nil,
					Date:  time.Now().Day(),
				}
				for i := 0; i < 2; i++ {
					s <- broker.Message{Bytes: make([]byte, 1100)}
					time.Sleep(10 * time.Millisecond)
				}
			},
			consume: func(l *DataLimiter) { consume(l) },
		},
	}

	for _, c := range cases {
		s := make(source)
		l := NewDataLimiter(s, c.conf())
		go c.generate(s, l)
		c.consume(l)
	}
}

func TestEnsureFuture(t *testing.T) {
	cases := []struct {
		day         int
		now, expect time.Time
	}{
		{
			day:    1,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    1,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 1, time.UTC),
			expect: time.Date(2000, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    12,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			day:    15,
			now:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			expect: time.Date(2000, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for i, c := range cases {
		if d := ensureFuture(dayThisMonth(c.day, c.now), c.now); !c.expect.Equal(d) {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, d)
		}
	}
}

func expiredTimer() *time.Timer {
	c := make(chan time.Time)
	close(c)
	return &time.Timer{C: c}
}

func closedChan() <-chan broker.Message {
	c := make(chan broker.Message)
	close(c)
	return c
}

func sendOne() <-chan broker.Message {
	c := make(chan broker.Message)
	go func() { c <- broker.Message{} }()
	return c
}

func sendConf() chan api.CellularConfig {
	c := make(chan api.CellularConfig)
	go func() { c <- api.CellularConfig{} }()
	return c
}

func TestStateFuncs(t *testing.T) {
	cases := []struct {
		state  state // which state to test
		l      *DataLimiter
		expect state
	}{
		{
			state:  initial,
			l:      &DataLimiter{},
			expect: underBudget,
		},
		{
			state: initial,
			l: &DataLimiter{
				Budget: intPtr(10),
				Count:  10,
			},
			expect: overBudget,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: expiredTimer(),
				store:       new(MemPersistor),
			},
			expect: initial,
		},
		{
			state:  underBudget,
			l:      &DataLimiter{periodTimer: new(time.Timer), in: closedChan()},
			expect: cleanup,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: new(time.Timer),
				in:          sendOne(),
				out:         make(chan broker.Message, 1),
				store:       new(MemPersistor),
			},
			expect: underBudget,
		},
		{
			state: underBudget,
			l: &DataLimiter{
				periodTimer: new(time.Timer),
				store:       new(MemPersistor),
				conf:        sendConf(),
			},
			expect: initial,
		},
		{
			state:  overBudget,
			l:      &DataLimiter{
				periodTimer: expiredTimer(),
				store: new(MemPersistor),
			},
			expect: initial,
		},
		{
			state:  overBudget,
			l:      &DataLimiter{in: closedChan(), periodTimer: new(time.Timer)},
			expect: cleanup,
		},
		{
			state: overBudget,
			l: &DataLimiter{
				in:          sendOne(),
				periodTimer: new(time.Timer),
			},
			expect: overBudget,
		},
		{
			state: overBudget,
			l: &DataLimiter{
				conf:        sendConf(),
				store:       new(MemPersistor),
				periodTimer: new(time.Timer),
			},
			expect: initial,
		},
		{
			state:  cleanup,
			l:      &DataLimiter{out: make(chan broker.Message)},
			expect: terminal,
		},
	}

	for i, c := range cases {
		if got := c.l.lookup(c.state)(); got != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, got)
		}
	}
}
