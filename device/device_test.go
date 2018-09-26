package device

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	time.Sleep(2 * time.Second)
	GetMetrics()
}
