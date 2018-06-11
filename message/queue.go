package message

import (
	"log"

	"github.com/ESG-USA/Auklet-Client-C/kafka"
)

// Queue provides an "infinite" buffer for outgoing Messages.
type Queue struct {
	source kafka.MessageSource
	q      []kafka.Message
	out    chan kafka.Message
	err    chan error
}

// NewQueue creates a new Queue that buffers Messages. Any existing persisted
// Messages are enqueued.
func NewQueue(in kafka.MessageSource) *Queue {
	return &Queue{
		source: in,
		q:      kafka.StdPersistor.Load(),
		out:    make(chan kafka.Message),
		err:    make(chan error),
	}
}

// Output returns a channel from which enqueued Messages can be received.
// Messages sent on this channel are not automatically dequeued. The channel
// closes when q's input closes.
func (q *Queue) Output() <-chan kafka.Message {
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

func (q *Queue) initial() serverState {
	if len(q.q) == 0 {
		return q.empty
	}
	return q.nonEmpty
}

// empty serves q for when q is empty. In that case, the only operation
// that should be served is push; it's not possible to pop or peek an empty
// queue.
//
// We do not serve q.err, because there is nothing to dequeue.
func (q *Queue) empty() serverState {
	m, open := <-q.source.Output()
	if !open {
		// The queue's input has closed. We enter the final state,
		// waiting for our client to shut us down.
		return q.final
	}
	q.push(m)
	return q.nonEmpty
}

// nonEmpty serves q for when q has at least one element.
func (q *Queue) nonEmpty() serverState {
	select {
	case m, open := <-q.source.Output():
		if !open {
			// The queue's input has closed. Since the queue is not
			// empty, we enter the final state.
			return q.final
		}
		q.push(m)
	case q.out <- q.peek():
	case <-q.err:
		// We assume that q.err is not closed, because we have not
		// closed q.out.
		q.pop()
		if len(q.q) == 0 {
			return q.empty
		}
	}
	return q.nonEmpty
}

// The queue's input has closed. This implies that the pipeline is shutting
// down. Our policy is to not send any more messages, even if the queue is not
// empty.
//
// We need to wait for our client to close q.err, which indicates that it will
// not send any more dequeue requests. In the meantime, we handle incoming
// dequeue requests.
func (q *Queue) final() serverState {
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

func (q *Queue) push(m kafka.Message) {
	q.q = append(q.q, m)
}

func (q *Queue) peek() kafka.Message {
	return q.q[0]
}

func (q *Queue) pop() {
	q.q[0].Remove()
	q.q = q.q[1:]
}
