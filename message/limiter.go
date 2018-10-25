package message

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// DataLimiter is a passthrough that limits the number of application-layer
// bytes transmitted per period.
type DataLimiter struct {
	in  <-chan broker.Message
	out chan broker.Message
	// conf is a channel by which the configuration can be updated.
	conf  <-chan api.CellularConfig
	store Persistor

	// Budget is how many bytes can be transmitted per period.
	// If HasBudget is false, any number of bytes can be transmitted.
	Budget    int  `json:"budget"`
	HasBudget bool `json:"hasBudget"`

	// Count is how many bytes have been transmitted during the current
	// period.
	Count int `json:"count"`

	// PeriodEnd marks the end of the current period.
	PeriodEnd time.Time `json:"periodEnd"`

	// initialized in the initial state
	periodTimer *time.Timer
}

// NewDataLimiter returns a DataLimiter for input whose state persists on
// the filesystem.
func NewDataLimiter(
	store Persistor,
	conf <-chan api.CellularConfig,
	src ...broker.MessageSource,
) *DataLimiter {
	l := &DataLimiter{
		in:    Merge(src...).Output(),
		out:   make(chan broker.Message),
		conf:  conf,
		store: store,
	}
	l.store.Load(l)
	// If Load fails, there is no budget, so all messages will be sent.
	go l.serve()
	return l
}

func (l *DataLimiter) setBudget(megabytes int, hasBudget bool) {
	l.Budget = 1e6 * megabytes
	l.HasBudget = hasBudget
	if !hasBudget {
		log.Print("limiter: setting budget to unlimited")
		return
	}
	log.Printf("limiter: setting budget to %v MB", megabytes)
}

// Decode updates l's state by reading bytes from r.
func (l *DataLimiter) Decode(r io.Reader) (err error) {
	return json.NewDecoder(r).Decode(l)
}

// Encode writes the l's state to w.
func (l *DataLimiter) Encode(w io.Writer) (err error) {
	return json.NewEncoder(w).Encode(l)
}

// ensureFuture ensures that the period end is in the future. If the
// current period end is in the past (implying that we're in a new period)
// the period end is advanced by one month.
func ensureFuture(periodEnd, now time.Time) time.Time {
	for periodEnd.Before(now) {
		// advance newEnd by one month
		periodEnd = periodEnd.AddDate(0, 1, 0)
	}

	return periodEnd
}

// dayThisMonth moves the boundary between periods to the given day of the
// month.
func dayThisMonth(dayOfMonth int, now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), dayOfMonth, 0, 0, 0, 0, now.Location())
}

// startThisPeriod moves the PeriodEnd forward (if necessary) and resets the
// counter.
func (l *DataLimiter) startThisPeriod() {
	// Make sure that the period end is in the future.
	l.PeriodEnd = ensureFuture(l.PeriodEnd, time.Now())
	l.reset()
}

func (l *DataLimiter) reset() (err error) {
	l.Count = 0
	return l.store.Save(l)
}

func (l *DataLimiter) increment(n int) (err error) {
	l.Count += n
	return l.store.Save(l)
}

// serve activates l, causing it to receive and send Messages.
func (l *DataLimiter) serve() {
	for s := initial; s != terminal; s = l.lookup(s)() {
	}
}

type state int

const (
	terminal state = iota
	initial
	underBudget
	overBudget
	cleanup
)

func (l *DataLimiter) lookup(s state) func() state {
	return map[state]func() state{
		initial:     l.initial,
		underBudget: l.underBudget,
		overBudget:  l.overBudget,
		cleanup:     l.cleanup,
	}[s]
}

// initial starts a timer that expires at the period end, then
// returns either overBudget or underBudget.
func (l *DataLimiter) initial() state {
	l.periodTimer = time.NewTimer(time.Until(l.PeriodEnd))
	if l.HasBudget && l.Count > 9*l.Budget/10 {
		return overBudget
	}
	return underBudget
}

func (l *DataLimiter) underBudget() state {
	select {
	case <-l.periodTimer.C:
		l.startThisPeriod()
		return initial
	case m, open := <-l.in:
		if !open {
			return cleanup
		}
		return l.handleMessage(m)
	case conf := <-l.conf:
		return l.apply(conf)
	}
}

func (l *DataLimiter) handleMessage(m broker.Message) state {
	n := len(m.Bytes)
	if l.HasBudget {
		if n+l.Count > l.Budget {
			// m would put us over budget. We begin dropping messages.
			return overBudget
		} else if n+l.Count > 9*l.Budget/10 {
			// m would put us over 90% of the budget, but not over 100%.
			// We send it and begin to drop messages.
			l.out <- m
			l.increment(n)
			return overBudget
		}
	}
	// m does not put us over 90% of budget.
	l.out <- m
	if l.increment(n) != nil {
		// We had a problem persisting the counter. To be safe, we
		// start dropping data.
		if l.HasBudget {
			return overBudget
		}
	}
	return underBudget
}

// The current period has exceeded 90% of its data budget. We drop messages.
// Note that it is still possible for the limiter to return to the underBudget
// state, when Serve notices that a new period has begun.
func (l *DataLimiter) overBudget() state {
	select {
	case <-l.periodTimer.C:
		l.startThisPeriod()
		return initial
	case _, open := <-l.in:
		if !open {
			return cleanup
		}
		return overBudget
	case conf := <-l.conf:
		return l.apply(conf)
	}
}

// apply applies the configuration and returns the initial state.
func (l *DataLimiter) apply(conf api.CellularConfig) state {
	old := l.PeriodEnd
	now := time.Now()
	l.PeriodEnd = ensureFuture(dayThisMonth(conf.Date, now), now)
	log.Printf(`limiter: moving period day
	from %v
	to   %v`, old, l.PeriodEnd)
	l.setBudget(conf.Limit, conf.Defined)
	l.reset()
	return initial
}

// The input channel has closed, which implies that the pipeline is shutting
// down.
func (l *DataLimiter) cleanup() state {
	close(l.out)
	return terminal
}

// Output returns a channel on which messages can be received. The channel
// closes when l's input closes.
func (l *DataLimiter) Output() <-chan broker.Message {
	return l.out
}

// Persistor can save and load an object to some kind of storage.
type Persistor interface {
	Save(Encodable) error
	Load(Decodable) error
}
