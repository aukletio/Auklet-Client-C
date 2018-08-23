// +build linux

package app

import (
	"errors"
	"syscall"
	"testing"
)

var errSockp = errors.New("socketpair error")

func TestSocketpair(t *testing.T) {
	cases := []struct {
		socketpair func(domain, typ, proto int) (fd [2]int, err error)
		expect     error
	}{
		{
			socketpair: func(_, _, _ int) (fd [2]int, err error) {
				err = errSockp
				return
			},
			expect: errSockp,
		}, {
			socketpair: func(_, _, _ int) ([2]int, error) {
				return [2]int{42, -167}, nil
			},
			expect: errInvalidFD,
		},
	}
	for i, c := range cases {
		sockp = c.socketpair
		if _, err := socketpair(""); err != c.expect {
			t.Errorf("case %v: expected %v, got %v", i, c.expect, err)
		}
		sockp = syscall.Socketpair
	}
}
