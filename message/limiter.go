package message

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ESG-USA/Auklet-Client/api"
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// DataLimiter is a passthrough that limits the number of application-layer
// bytes transmitted per period.
type DataLimiter struct {
	source kafka.MessageSource
	out    chan kafka.Message
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

// NewDataLimiter returns a DataLimiter for input whose state persists on
// the filesystem.
func NewDataLimiter(input kafka.MessageSource, appID string) *DataLimiter {
	l := &DataLimiter{
		source: input,
		out:    make(chan kafka.Message),
		conf:   make(chan api.CellularConfig),
		path:   ".auklet/limit.json",
	}
	if err := l.load(); err != nil {
		log.Println(err)
	}
	// If load fails, there is no budget, so all messages will be sent.
	return l
}

func (l *DataLimiter) setBudget(megabytes *int) {
	if megabytes == nil {
		l.Budget = nil
		return
	}
	*l.Budget = 1e6 * *megabytes
}

func (l *DataLimiter) load() (err error) {
	f, err := os.Open(l.path)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(l)
}

func (l *DataLimiter) save() (err error) {
	f, err := os.Create(l.path)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(l)
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
	l.PeriodEnd = toFutureDate(day)
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
	if err := l.save(); err != nil {
		log.Println(err)
	}
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

func (l *DataLimiter) handleMessage(m kafka.Message) serverState {
	n := len(m.Bytes())
	if l.Budget != nil {
		if n+l.Count > *l.Budget {
			// m would put us over budget. We begin dropping messages.
			return l.overBudget
		} else if n+l.Count > 9**l.Budget/10 {
			// m would put us over 90% of the budget, but not over 100%.
			// We send it and begin to drop messages.
			l.out <- m
			if err := l.increment(n); err != nil {
				log.Print(err)
			}
			return l.overBudget
		}
	}
	// m does not put us over 90% of budget.
	l.out <- m
	if err := l.increment(n); err != nil {
		log.Print(err)
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
	log.Print("limiter: got new config", conf)
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
func (l *DataLimiter) Output() <-chan kafka.Message {
	return l.out
}

func (l *DataLimiter) Configure() chan<- api.CellularConfig {
	return l.conf
}
