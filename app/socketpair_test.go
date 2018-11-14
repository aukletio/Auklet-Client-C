// +build linux

package app

import (
	"testing"
)

func TestSocketpair(t *testing.T) {
	// We test this to demonstrate that our arguments
	// to the underlying system call are reasonable.
	p, err := socketpair("test")
	if err != nil {
		t.Error(err)
	}
	defer p.local.Close()
	defer p.remote.Close()
}
