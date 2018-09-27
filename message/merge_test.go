package message

import (
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type channel chan broker.Message

func (s channel) Output() <-chan broker.Message { return s }

// TestMerger proves that the merger doesn't panic in a trivial case, but
// does not test it for correctness.
func TestMerger(t *testing.T) {
	c := make(channel)
	merger := Merge(c)
	c <- broker.Message{}
	close(c)
	<-merger.Output()
}
