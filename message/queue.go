package message

import (
	"log"
	"os"
)

type Queue struct {
	source Source
	dir    string
	q      []Persistent
	out    chan Message
	err    chan error
}

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

func (q *Queue) Output() <-chan Message {
	return q.out
}

func (q *Queue) Err() chan<- error {
	return q.err
}

func (q *Queue) Serve() {
	defer close(q.out)
	for {
		log.Print("queued messages: ", len(q.q))
		select {
		case m, open := <-q.source.Output():
			if !open {
				return
			}
			p, err := toPersistent(m, q.dir)
			if err != nil {
				log.Print(err)
				continue
			}
			q.push(p)
		case q.outReady() <- q.peek():
		case err, open := <-q.err:
			if err == nil {
				q.pop()
			}
			if !open {
				return
			}
		}
	}
}

func (q *Queue) outReady() chan<- Message {
	if len(q.q) == 0 {
		return nil
	}
	return q.out
}

func (q *Queue) push(p Persistent) {
	q.q = append(q.q, p)
}

func (q *Queue) peek() (_ Persistent) {
	if len(q.q) == 0 {
		return
	}
	return q.q[0]
}

func (q *Queue) pop() {
	if len(q.q) == 0 {
		return
	}
	q.q[0].remove()
	q.q = q.q[1:]
}
