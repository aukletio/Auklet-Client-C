// +build linux

package app

import (
	"errors"
	"os"
	"syscall"
)

type pair struct {
	local, remote *os.File
}

var errInvalidFD = errors.New("invalid file descriptor")

// socketpair returns a pair of sockets, already connected.
func socketpair(prefix string) (p pair, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return
	}

	p = pair{
		local:  os.NewFile(uintptr(fd[0]), prefix+"-local"),
		remote: os.NewFile(uintptr(fd[1]), prefix+"-remote"),
	}

	if p.local == nil || p.remote == nil {
		err = errInvalidFD
	}
	return
}
