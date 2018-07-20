package message

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// DataLimiter is a passthrough that limits the number of application-layer
// bytes transmitted per period.
type DataLimiter struct {
	source broker.MessageSource
	out    chan broker.Message
	conf   chan api.CellularConfig
	path   string

	// Budget is how many bytes can be transmitted per period. If nil, any
	// number of bytes can be transmitted.
	Budget *int `json:"budget"`

	// Count is how many bytes have been transmitted during the current
	// period.
	Count int `json:"count"`

	// PeriodEnd marks the end of the current period.
	PeriodEnd time.Time `json:"periodEnd"`
}

type Persistor interface {
	Save(Encodable) error
	Load(Decodable) error
}

// NewDataLimiter returns a DataLimiter for input whose state persists on
// the filesystem.
func NewDataLimiter(input broker.MessageSource, p Persistor) *DataLimiter {
	l := &DataLimiter{
		source: input,
		out:    make(chan broker.Message),
		conf:   make(chan api.CellularConfig),
		path:   ".auklet/datalimit.json",
	}
	l.load()
	// If load fails, there is no budget, so all messages will be sent.
	return l
}

func (l *DataLimiter) setBudget(megabytes *int) {
	if megabytes == nil {
		log.Print("limiter: setting budget to unlimited")
		l.Budget = nil
		return
	}
	*l.Budget = 1e6 * *megabytes
	log.Printf("limiter: setting budget to %v B", *l.Budget)
}

func (l *DataLimiter) Decode(r io.Reader) (err error) {
	return json.NewDecoder(r).Decode(l)
}

// save saves the data limiter's state to disk. If there is an error, it's a
// encoding error.
func (l *DataLimiter) Encode(w io.Writer) (err error) {
	return json.NewEncoder(w).Encode(l)
}

// newPeriod returns true if the current time is after the period end.
func (l *DataLimiter) newPeriod() bool {
	return time.Now().After(l.PeriodEnd)
}

func (l *DataLimiter) advancePeriodEnd() {
	now := time.Now()
	newEnd := l.PeriodEnd
	for newEnd.Before(now) {
		// advance newEnd by one month
		newEnd = newEnd.AddDate(0, 1, 0)
	}
	l.PeriodEnd = newEnd
}

func (l *DataLimiter) setPeriodDay(day int) {
	if l.PeriodEnd.Day() == day {
		return
	}
	d := toFutureDate(day)
	log.Printf("limiter: moving period day from %v to %v", l.PeriodEnd, d)
	l.PeriodEnd = d
}

func toFutureDate(day int) time.Time {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
	if t.Before(now) {
		return t.AddDate(0, 1, 0)
	}
	return t
}

func (l *DataLimiter) startThisPeriod() {
	l.advancePeriodEnd()
	l.Count = 0
	l.save()
}

func (l *DataLimiter) increment(n int) (err error) {
	l.Count += n
	return l.save()
}

// Serve activates l, causing it to receive and send Messages.
func (l *DataLimiter) Serve() {
	state := l.initial
	for state != nil {
		if l.newPeriod() {
			l.startThisPeriod()
			state = l.initial
		}
		state = state()
	}
}

func (l *DataLimiter) initial() serverState {
	if l.Budget != nil && l.Count > 9**l.Budget/10 {
		return l.overBudget
	}
	return l.underBudget
}

func (l *DataLimiter) underBudget() serverState {
	select {
	case m, open := <-l.source.Output():
		if !open {
			return l.final
		}
		return l.handleMessage(m)
	case conf := <-l.conf:
		return l.apply(conf)
	}
}

func (l *DataLimiter) handleMessage(m broker.Message) serverState {
	n := len(m.Bytes)
	if l.Budget != nil {
		if n+l.Count > *l.Budget {
			// m would put us over budget. We begin dropping messages.
			return l.overBudget
		} else if n+l.Count > 9**l.Budget/10 {
			// m would put us over 90% of the budget, but not over 100%.
			// We send it and begin to drop messages.
			l.out <- m
			l.increment(n)
			return l.overBudget
		}
	}
	// m does not put us over 90% of budget.
	l.out <- m
	if l.increment(n) != nil {
		// We had a problem persisting the counter. To be safe, we
		// start dropping data.
		if l.Budget != nil {
			return l.overBudget
		}
	}
	return l.underBudget
}

// The current period has exceeded 90% of its data budget. We drop messages.
// Note that it is still possible for the limiter to return to the underBudget
// state, when Serve notices that a new period has begun.
func (l *DataLimiter) overBudget() serverState {
	select {
	case _, open := <-l.source.Output():
		if !open {
			return l.final
		}
		return l.overBudget
	case conf := <-l.conf:
		return l.apply(conf)
	}
}

func (l *DataLimiter) apply(conf api.CellularConfig) serverState {
	l.setPeriodDay(conf.Date)
	l.setBudget(conf.Limit)
	l.startThisPeriod()
	return l.initial
}

// The input channel has closed, which implies that the pipeline is shutting
// down.
func (l *DataLimiter) final() serverState {
	close(l.out)
	return nil
}

// Output returns a channel on which messages can be received. The channel
// closes when l's input closes.
func (l *DataLimiter) Output() <-chan broker.Message {
	return l.out
}

// Configure returns a channel by which the configuration can be updated.
func (l *DataLimiter) Configure() chan<- api.CellularConfig {
	return l.conf
}
