package device

import (
	"errors"
	"testing"
	"time"
)

func Test(t *testing.T) {
	getIP = func() (string, error) {
		return "", errors.New("error")
	}
	CurrentIP()
	time.Sleep(2 * time.Second)
	GetMetrics()
}
