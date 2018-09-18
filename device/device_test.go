package device

import (
	"errors"
	"time"
	"testing"
)

func Test(t *testing.T) {
	getIP = func() (string, error) {
		return "", errors.New("error")
	}
	CurrentIP()
	time.Sleep(2*time.Second)
	GetMetrics()
}
