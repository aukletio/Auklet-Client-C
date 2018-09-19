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

func TestMessageLoader(t *testing.T) {
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
		l := NewMessageLoader(c.dir)
		m := <-l.Output()
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
	defer close(p.done)
	p.Configure() <- nil
	if lim := <-p.currentLimit; lim != nil {
		t.Fail()
	}
}

func TestCreateMessage(t *testing.T) {
	cases := []struct {
		dir string
		lim *int64
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
		{
			dir: "testdata/w",
			lim: func(n int64) *int64 { return &n }(-10),
			ok:  false,
		},
	}

	for i, c := range cases {
		p := NewPersistor(c.dir)
		p.Configure() <- c.lim
		err := p.CreateMessage(&Message{})
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
		close(p.done)
	}
}

func TestMain(m *testing.M) {
	clean := func() {
		// remove everything in w
		paths, _ := filepaths("testdata/w")
		for _, path := range paths {
			os.Remove(path)
		}
	}

	clean()
	status := m.Run()
	clean()

	os.Exit(status)
}

func TestFilepaths(t *testing.T) {
	osOpen = func(string) (*os.File, error) { return nil, errMockFs }
	defer func() { osOpen = os.Open }()
	paths, err := filepaths("testdata/r")
	if err == nil {
		t.Fail()
	}
	if len(paths) != 0 {
		t.Fail()
	}
}

func TestErrorStorageFull(t *testing.T) {
	(ErrStorageFull{}).Error()
}

func TestServe(t *testing.T) {
	c := make(chan struct{})
	close(c)
	p := Persistor{done: c}
	p.serve()
}
