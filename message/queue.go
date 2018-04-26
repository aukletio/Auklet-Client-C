package message

import (
	"log"
	"os"
)

// Queue provides an "infinite" buffer for outgoing Messages.
// Enqueued Messages persist via the filesystem.
type Queue struct {
	source Source
	dir    string
	q      []Persistent
	out    chan Message
	err    chan error
}

// NewQueue creates a new Queue that receives Messages from in and persists them
// to the path dir. If dir contains persisted Messages, they are enqueued.
func NewQueue(in Source, dir string) (q *Queue) {
	q = &Queue{
		source: in,
		dir:    dir,
		q:      make([]Persistent, 0),
		out:    make(chan Message),
		err:    make(chan error),
	}
	if err := q.load(); err != nil {
		log.Print(err)
	}
	return
}

// load enqueues q with Persistent messages from the filesystem.
func (q *Queue) load() (err error) {
	f, err := os.Open(q.dir)
	if err != nil {
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return
	}
	for _, name := range names {
		p := Persistent{path: q.dir + "/" + name}
		if err := p.load(); err != nil {
			log.Print(err)
			continue
		}
		q.push(p)
	}
	return
}

// Output returns a channel from which enqueued Messages can be received.
// Messages sent on this channel are not automatically dequeued. The channel
// closes when q's input closes.
func (q *Queue) Output() <-chan Message {
	return q.out
}

// Err returns a channel on which clients can send values. Any value causes
// q to dequeue its head. Sends on this channel will block if q is empty.
// Closing the channel shuts down q. Err() must not be closed before Output()
// closes.
func (q *Queue) Err() chan<- error {
	return q.err
}

// Serve activates q, causing it to receive and send Messages. Serve returns
// when q shuts down.
func (q *Queue) Serve() {
	for state := q.initial; state != nil; state = state() {
		log.Print("queued messages: ", len(q.q))
	}
}

// A queueState, when executed, returns the next queue state. A nil queueState
// signifies a Queue's termination.
type queueState func() queueState

func (q *Queue) initial() queueState {
	if len(q.q) == 0 {
		return q.serveEmpty
	}
	return q.serveNonEmpty
}

// serveEmpty serves q for when q is empty. In that case, the only operation
// that should be served is push; it's not possible to pop or peek an empty
// queue.
//
// We do not serve q.err, because there is nothing to dequeue.
func (q *Queue) serveEmpty() queueState {
	m, open := <-q.source.Output()
	if !open {
		// The queue's input has closed. We enter the final state,
		// waiting for our client to shut us down.
		return q.final
	}
	p, err := toPersistent(m, q.dir)
	if err != nil {
		log.Print(err)
		// The message could not be made persistent, so the queue is
		// still empty.
		return q.serveEmpty
	}
	q.push(p)
	return q.serveNonEmpty
}

// serveNonEmpty serves q for when q has at least one element.
func (q *Queue) serveNonEmpty() queueState {
	select {
	case m, open := <-q.source.Output():
		if !open {
			// The queue's input has closed. Since the queue is not
			// empty, we enter the final state.
			return q.final
		}
		if p, err := toPersistent(m, q.dir); err != nil {
			log.Print(err)
		} else {
			q.push(p)
		}
	case q.out <- q.peek():
	case <-q.err:
		// We assume that q.err is not closed, because we have not
		// closed q.out.
		q.pop()
		if len(q.q) == 0 {
			return q.serveEmpty
		}
	}
	return q.serveNonEmpty
}

// The queue's input has closed. This implies that the pipeline is shutting
// down. Our policy is to not send any more messages, even if the queue is not
// empty. This is OK because messages are persistent and thus can be sent upon
// the next startup.
//
// We need to wait for our client to close q.err, which indicates that it will
// not send any more dequeue requests. In the meantime, we handle incoming
// dequeue requests.
func (q *Queue) final() queueState {
	// We inform our client that we won't be sending any more values. We
	// expect them to shut us down soon.
	close(q.out)
	if _, open := <-q.err; !open {
		// Our client has indicated that they have no more requests to
		// send. We can shut down immediately.
		return nil
	}
	// Our client sent a dequeue request, which we dutifully process.
	q.pop()
	if len(q.q) == 0 {
		// If the q is empty, we're assume the client won't
		// send any more requests.
		return nil
	}
	// Our client might need to send more dequeue requests.
	return q.final
}

func (q *Queue) push(p Persistent) {
	q.q = append(q.q, p)
}

func (q *Queue) peek() (_ Persistent) {
	return q.q[0]
}

func (q *Queue) pop() {
	q.q[0].remove()
	q.q = q.q[1:]
}
