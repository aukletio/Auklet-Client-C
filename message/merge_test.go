package message

import (
	"testing"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

type channel chan broker.Message

func (s channel) Output() <-chan broker.Message { return s }

func TestMerger(t *testing.T) {
	c := make(channel)
	merger := NewMerger(c)
	c <- broker.Message{}
	close(c)
	<-merger.Output()
}
