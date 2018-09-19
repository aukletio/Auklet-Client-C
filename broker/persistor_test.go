package broker

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

var errMockFs = errors.New("filesystem error")

func TestSave(t *testing.T) {
	osMkdirAll = func(string, os.FileMode) error { return errMockFs }
	defer func() { osMkdirAll = os.MkdirAll }()

	if err := (Message{}).save(); err == nil {
		t.Fail()
	}
}

func TestOpenFile(t *testing.T) {
	osMkdirAll = func(string, os.FileMode) error { return nil }
	osOpenFile = func(string, int, os.FileMode) (*os.File, error) { return nil, errMockFs }
	defer func() {
		osMkdirAll = os.MkdirAll
		osOpenFile = os.OpenFile
	}()

	if err := (Message{}).save(); err == nil {
		t.Fail()
	}
}

func TestStat(t *testing.T) {
	orig := osStat
	osStat = func(string) (sizer, error) { return nil, errMockFs }
	defer func() { osStat = orig }()

	paths, err := filepaths(".auklet/messages")
	if len(paths) != 0 || err != nil {
		t.Errorf("expected empty slice, nil; got %v, %v", paths, err)
	}
}

func TestOpen(t *testing.T) {
	osOpen = func(string) (*os.File, error) { return nil, errMockFs }
	defer func() { osOpen = os.Open }()

	paths, err := filepaths(".auklet/messages")
	exp := "filepaths: failed to open message directory: filesystem error"
	if len(paths) != 0 || (err != nil && err.Error() != exp) {
		t.Errorf("expected empty slice, %v; got %v, %v", exp, paths, err)
	}
}

func TestReaddirnames(t *testing.T) {
	paths, err := filepaths("testdata/notdir")
	if err == nil {
		t.Fail()
	}
	if len(paths) != 0 {
		t.Fail()
	}
}

func TestLoad(t *testing.T) {
	cases := []struct {
		dir string
		ok  bool
	}{
		{
			dir: "testdata/notdir",
			ok:  false,
		},
		{
			dir: "testdata/r",
			ok:  true,
		},
	}

	for i, c := range cases {
		m := <-load(c.dir)
		ok := m.Error == ""
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, m.Error)
		}
	}
}

func TestLoadMessage(t *testing.T) {
	cases := []struct {
		readFile func(string) ([]byte, error)
		ok       bool
	}{
		{
			readFile: func(string) ([]byte, error) { return nil, errors.New("error") },
			ok:       false,
		},
		{
			readFile: func(string) ([]byte, error) { return []byte("}"), nil },
			ok:       false,
		},
		{
			readFile: func(string) ([]byte, error) { return []byte("{}"), nil },
			ok:       true,
		},
	}

	defer func() { ioutilReadFile = ioutil.ReadFile }()

	for i, c := range cases {
		ioutilReadFile = c.readFile
		m := loadMessage("")
		ok := m.Error == ""
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, m.Error)
		}
	}
}

func TestSize(t *testing.T) {
	origfp := filepaths
	origstat := osStat
	defer func() {
		osStat = origstat
		filepaths = origfp
	}()

	cases := []struct {
		dir       string
		stat      func(string) (sizer, error)
		filepaths func(string) ([]string, error)
		ok        bool
	}{
		{
			dir:       "",
			stat:      origstat,
			filepaths: func(string) ([]string, error) { return nil, errMockFs },
			ok:        false,
		},
		{
			dir:       "",
			stat:      origstat,
			filepaths: func(string) ([]string, error) { return []string{""}, nil },
			ok:        false,
		},
		{
			dir:       "testdata/r",
			stat:      func(string) (sizer, error) { return mockFileInfo{}, nil },
			filepaths: origfp,
			ok:        true,
		},
	}

	for i, c := range cases {
		filepaths = c.filepaths
		osStat = c.stat
		_, err := size(c.dir)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

type mockFileInfo struct{}

func (mockFileInfo) Size() int64 { return 0 }

func TestChannels(t *testing.T) {
	p := NewPersistor("testdata/w")
	p.Configure() <- nil
	if lim := <-p.currentLimit; lim != nil {
		t.Fail()
	}
	close(p.done)
}

func TestCreateMessage(t *testing.T) {
	cases := []struct {
		dir string
		m   Message
		ok  bool
	}{
		{
			dir: "testdata/notdir",
			ok:  false,
		},
		{
			dir: "testdata/w",
			ok:  true,
		},
	}

	for i, c := range cases {
		p := NewPersistor(c.dir)
		err := p.CreateMessage(&c.m)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
		close(p.done)
	}
}

func TestMain(m *testing.M) {
	status := m.Run()

	// remove everything in w
	paths, _ := filepaths("testdata/w")
	for _, path := range paths {
		os.Remove(path)
	}

	os.Exit(status)
}
