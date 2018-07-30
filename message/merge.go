package message

import (
	"sync"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// Merger is a broker.MessageSource that merges multiple other MessageSources
// into one stream.
type Merger struct {
	src []broker.MessageSource
	out chan broker.Message
}

// NewMerger returns a Merger that merges the streams of each element in src.
func NewMerger(src ...broker.MessageSource) Merger {
	m := Merger{
		src: src,
		out: make(chan broker.Message, 10),
	}
	go m.serve()
	return m
}

// Output returns m's output channel. It closes when all input streams have
// closed.
func (m Merger) Output() <-chan broker.Message {
	return m.out
}

// serve activates m, causing it to send and receive messages.
func (m Merger) serve() {
	var wg sync.WaitGroup
	merge := func(s broker.MessageSource) {
		defer wg.Done()
		for msg := range s.Output() {
			m.out <- msg
		}
	}
	wg.Add(len(m.src))
	for _, src := range m.src {
		go merge(src)
	}
	wg.Wait()
	close(m.out)
}
