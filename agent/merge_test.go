package agent

import (
	"testing"
)

type source chan Message

func (s source) Output() <-chan Message { return s }

// TestMerger proves that the merger doesn't panic in a trivial case, but
// does not test it for correctness.
func TestMerger(t *testing.T) {
	c := make(source)
	merger := Merge(c)
	c <- Message{}
	close(c)
	<-merger.Output()
}
