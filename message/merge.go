package message

import (
	"sync"

	"github.com/ESG-USA/Auklet-Client-C/kafka"
)

// Merger is a kafka.MessageSource that merges multiple other MessageSources
// into one stream.
type Merger struct {
	src []kafka.MessageSource
	out chan kafka.Message
}

// NewMerger returns a Merger that merges the streams of each element in src.
func NewMerger(src ...kafka.MessageSource) Merger {
	return Merger{
		src: src,
		out: make(chan kafka.Message),
	}
}

// Output returns m's output channel. It closes when all input streams have
// closed.
func (m Merger) Output() <-chan kafka.Message {
	return m.out
}

// Serve activates m, causing it to send and receive messages.
func (m Merger) Serve() {
	var wg sync.WaitGroup
	merge := func(s kafka.MessageSource) {
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
