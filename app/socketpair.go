// +build linux

package app

import (
	"fmt"
	"os"
	"syscall"
)

type pair struct {
	local, remote *os.File
}

// socketpair returns a pair of sockets, already connected.
func socketpair(prefix string) (p pair, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	p = pair {
		local:  os.NewFile(uintptr(fd[0]), prefix+"-local"),
		remote: os.NewFile(uintptr(fd[1]), prefix+"-remote"),
	}
	format := "socketpair: invalid file descriptor %v"
	if p.local == nil {
		err = fmt.Errorf(format, fd[0])
		return
	}
	if p.remote == nil {
		err = fmt.Errorf(format, fd[1])
		return
	}
	return
}
