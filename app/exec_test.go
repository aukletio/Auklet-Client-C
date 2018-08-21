package app

import (
	"bytes"
	"errors"
	"testing"
)

func TestMethods(t *testing.T) {
	exec, err := NewExec("testdata/ls")
	if err != nil {
		t.Fatal(err)
	}

	// check if it's released
	if exec.CheckSum() == "" {
		// If not released, don't add sockets. Just run it and wait.
		exec.Start()
		exec.Wait()
		return
	}

	if err := exec.AddSockets(); err != nil {
		t.Fatal(err)
	}

	if err := exec.Start(); err != nil {
		t.Fatal(err)
	}

	if err := exec.GetAgentVersion(); err == nil {
		t.Fatal(err)
	}

	exec.String()
	exec.AgentVersion()

	// wait for it to exit
	exec.Wait()

	// see if it crashed
	exec.ExitStatus()
	exec.Signal()
}

func TestExec(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{
			given:  "ls",
			expect: "open ls: no such file or directory",
		}, {
			given:  "testdata/ls",
			expect: "",
		},
	}

	for i, c := range cases {
		_, err := NewExec(c.given)
		if (err != nil && err.Error() != c.expect) || (err == nil && c.expect != "") {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}

var errSocketPair = errors.New("socketpair failed")

func TestAddSockets(t *testing.T) {
	cases := []struct {
		socketpair func(string) (pair, error)
		expect     error
	}{
		{
			socketpair: socketpair,
			expect:     nil,
		}, {
			socketpair: func(string) (pair, error) {
				return pair{}, errSocketPair
			},
			expect: errSocketPair,
		}, {
			socketpair: func(name string) (pair, error) {
				if name == "agentData" {
					return pair{}, errSocketPair
				}
				return pair{}, nil
			},
			expect: errSocketPair,
		},
	}
	must := func(exec *Exec, err error) *Exec {
		if err != nil {
			panic(err)
		}
		return exec
	}
	for i, c := range cases {
		socketPair = c.socketpair
		err := must(NewExec("testdata/ls")).AddSockets()
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
		socketPair = socketpair
	}
}

func TestGetAgentVersion(t *testing.T) {
	cases := []struct {
		exec   *Exec
		expect error
	}{
		{
			exec: &Exec{
				AgentData: bytes.NewBufferString(`{"version":"something"}`),
			},
			expect: nil,
		}, {
			exec: &Exec{
				AgentData: bytes.NewBufferString(`{"version":""}`),
			},
			expect: errNoVersion,
		}, {
			exec: &Exec{
				AgentData: bytes.NewBufferString(` `),
			},
			expect: errEOF,
		}, {
			exec: &Exec{
				AgentData: bytes.NewBufferString(`}`),
			},
			expect: errEncoding,
		},
	}

	for i, c := range cases {
		err := c.exec.GetAgentVersion()
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}
