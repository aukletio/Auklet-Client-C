package message

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// DataLimiter is a passthrough that limits the number of application-layer
// bytes transmitted per period.
type DataLimiter struct {
	source Source
	out    chan Message
	path   string

	// Budget is how many bytes can be transmitted per period.
	Budget int `json:"budget"`

	// Count is how many bytes have been transmitted during the current
	// period.
	Count int `json:"count"`

	// PeriodEnd marks the end of the current period.
	PeriodEnd time.Time `json:"periodEnd"`
}

// NewDataLimiter returns a DataLimiter for input whose state is
// associated with the given configuration path.
func NewDataLimiter(input Source, configpath string) *DataLimiter {
	l := &DataLimiter{
		source: input,
		out:    make(chan Message),
		path:   configpath,
	}
	if err := l.load(); err != nil {
		log.Println(err)
	}
	// If load fails, the budget is 0, and no messages will be sent.
	return l
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

func (l *DataLimiter) updatePeriodEnd() {
	now := time.Now()
	newEnd := l.PeriodEnd
	for newEnd.Before(now) {
		// advance newEnd by one month
		newEnd = newEnd.AddDate(0, 1, 0)
	}
	l.PeriodEnd = newEnd
}

func (l *DataLimiter) startThisPeriod() {
	l.updatePeriodEnd()
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
	if l.Count > 9*l.Budget/10 {
		return l.overBudget
	}
	return l.underBudget
}

func (l *DataLimiter) underBudget() serverState {
	m, open := <-l.source.Output()
	if !open {
		return l.final
	}
	b, err := m.Bytes()
	if err != nil {
		log.Print(err)
		return l.underBudget
	}
	n := len(b)
	if n+l.Count > l.Budget {
		// m would put us over budget. We begin dropping messages.
		return l.overBudget
	} else if n+l.Count > 9*l.Budget/10 {
		// m would put us over 90% of the budget, but not over 100%.
		// We send it and begin to drop messages.
		l.out <- m
		if err := l.increment(n); err != nil {
			log.Print(err)
		}
		return l.overBudget
	}
	// m does not put us over 90% of budget.
	l.out <- m
	if err := l.increment(n); err != nil {
		log.Print(err)
		// We had a problem persisting the counter. To be safe, we
		// start dropping data.
		return l.overBudget
	}
	return l.underBudget
}

// The current period has exceeded 90% of its data budget. We drop messages.
// Note that it is still possible for the limiter to return to the underBudget
// state, when Serve notices that a new period has begun.
func (l *DataLimiter) overBudget() serverState {
	if _, open := <-l.source.Output(); !open {
		return l.final
	}
	return l.overBudget
}

// The input channel has closed, which implies that the pipeline is shutting
// down.
func (l *DataLimiter) final() serverState {
	close(l.out)
	return nil
}

// Output returns a channel on which messages can be received. The channel
// closes when l's input closes.
func (l *DataLimiter) Output() <-chan Message {
	return l.out
}
