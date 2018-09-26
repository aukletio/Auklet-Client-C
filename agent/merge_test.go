package agent

import (
	"testing"
)

type source chan Message

func (s source) Output() <-chan Message { return s }

func TestMerger(t *testing.T) {
	c := make(source)
	merger := Merge(c)
	c <- Message{}
	close(c)
	<-merger.Output()
}
