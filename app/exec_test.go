package app

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestArgs(t *testing.T) {
	cases := []struct {
		given  []string
		expect error
	}{
		{
			given:  []string{},
			expect: errNumArgs,
		}, {
			given:  []string{"ls"},
			expect: nil,
		},
	}

	for i, c := range cases {
		_, err := newOneOrMoreArgs(c.given)
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}

func TestExec(t *testing.T) {
	cases := []struct {
		given  oneOrMoreArgs
		expect string
	}{
		{
			given: oneOrMoreArgs{
				first: "ls",
				rest:  []string{},
			},
			expect: "open ls: no such file or directory",
		}, {
			given: oneOrMoreArgs{
				first: "testdata/ls",
				rest:  []string{},
			},
			expect: "",
		},
	}

	for i, c := range cases {
		_, err := newExec(c.given)
		if (err != nil && err.Error() != c.expect) || (err == nil && c.expect != "") {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}

func TestEnviro(t *testing.T) {
	cases := []struct {
		given  func(string) string
		expect error
	}{
		{
			given:  func(string) string { return "" },
			expect: errNoAppID,
		}, {
			given: func(key string) string {
				if key == "APP_ID" {
					return "something"
				}
				return ""
			},
			expect: errNoAPIKey,
		}, {
			given:  func(string) string { return "something" },
			expect: nil,
		},
	}

	for i, c := range cases {
		_, err := newEnviro(c.given)
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}

var errNotReleased = errors.New("not released")

type mockExec struct{}

func (mockExec) start() error        { return nil }
func (mockExec) inherit(...*os.File) {}
func (mockExec) checksum() string    { return "" }

var errSocketPair = errors.New("socketpair failed")

func TestRelExec(t *testing.T) {
	notReleased := func(enviro, string) (*relProof, error) {
		return nil, errNotReleased
	}
	released := func(enviro, string) (*relProof, error) {
		return &relProof{}, nil
	}
	cases := []struct {
		socketpair func(string) (pair, error)
		check      relChecker
		expect     error
	}{
		{
			socketpair: socketpair,
			check:      notReleased,
			expect:     errNotReleased,
		},
		{
			socketpair: socketpair,
			check:      released,
			expect:     nil,
		},
		{
			socketpair: func(string) (pair, error) {
				return pair{}, errSocketPair
			},
			check:  released,
			expect: errBug{errSocketPair},
		},
		{
			socketpair: func(name string) (pair, error) {
				if name == "agentData" {
					return pair{}, errSocketPair
				}
				return pair{}, nil
			},
			check:  released,
			expect: errBug{errSocketPair},
		},
	}

	for i, c := range cases {
		socketPair = c.socketpair
		_, err := newRelExec(enviro{}, c.check, mockExec{})
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
		socketPair = socketpair
	}
}

func TestRun(t *testing.T) {
	cases := []struct {
		rel    relExec
		expect error
	}{
		{
			rel: relExec{
				exec:      mockExec{},
				agentData: bytes.NewBufferString(`{"version":"something"}`),
			},
			expect: nil,
		},
	}

	for i, c := range cases {
		_, err := run(c.rel)
		if err != c.expect {
			format := "case %v: expected %v, got %v"
			t.Errorf(format, i, c.expect, err)
		}
	}
}
