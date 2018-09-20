package main

import "testing"

func Test(t *testing.T) {
	var status int
	osExit = func(i int) { status = i }
	cases := []struct {
		args   []string
		status int
	}{
		{args: []string{"-badflag"}, status: 1},
		{args: []string{"-licenses"}, status: 0},
		{args: []string{}, status: 1},
		{args: []string{"testdata/noexist"}, status: 1},
		{args: []string{"testdata/ls"}, status: 1},
	}

	for i, c := range cases {
		osArgs = c.args
		main()
		if status != c.status {
			t.Errorf("case %v: expected %v, got %v", i, c.status, status)
		}
	}
}
