package device

import (
	"testing"
	"time"
)

// This test covers the Monitor implementation,
// but does not check for correctness.
func TestMonitor(t *testing.T) {
	m := NewMonitor()
	time.Sleep(2 * time.Second)
	m.GetMetrics()
	m.Close()
}
