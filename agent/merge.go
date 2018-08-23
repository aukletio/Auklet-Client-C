package agent

import (
	"sync"
)

// MessageSource is a source of messages.
type MessageSource interface {
	Output() <-chan Message
}

// Merger is a MessageSource that merges multiple other MessageSources
// into one stream.
type Merger struct {
	src []MessageSource
	out chan Message
}

// NewMerger returns a Merger that merges the streams of each element in src.
func NewMerger(src ...MessageSource) Merger {
	m := Merger{
		src: src,
		out: make(chan Message),
	}
	go m.serve()
	return m
}

// Output returns m's output channel. It closes when all input streams have
// closed.
func (m Merger) Output() <-chan Message {
	return m.out
}

// serve activates m, causing it to send and receive messages.
func (m Merger) serve() {
	var wg sync.WaitGroup
	merge := func(s MessageSource) {
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
