// +build linux

package agent

import (
	"fmt"
	"os"
	"syscall"
)

func socketpair(prefix string) (local, remote *os.File, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	var f [2]*os.File
	name := [2]string{"local", "remote"}
	for i := range fd {
		f[i] = os.NewFile(uintptr(fd[i]), prefix+name[i])
		if f[i] == nil {
			err = fmt.Errorf("socketpair: invalid file descriptor %v", fd[i])
			return
		}
	}
	local = f[0]
	remote = f[1]
	return
}
