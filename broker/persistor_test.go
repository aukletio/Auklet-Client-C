package broker

import (
	"errors"
	"os"
	"testing"
)

var errMockFs = errors.New("filesystem error")

func TestSave(t *testing.T) {
	defer func() {
		osMkdirAll = os.MkdirAll
		osOpenFile = os.OpenFile
	}()

	cases := []struct {
		mkdirall func(string, os.FileMode) error
		openfile func(string, int, os.FileMode) (*os.File, error)
		ok       bool
	}{
		{
			mkdirall: func(string, os.FileMode) error { return errMockFs },
			openfile: os.OpenFile,
			ok:       false,
		},
		{
			mkdirall: func(string, os.FileMode) error { return nil },
			openfile: func(string, int, os.FileMode) (*os.File, error) { return nil, errMockFs },
			ok:       false,
		},
	}

	for i, c := range cases {
		osMkdirAll = c.mkdirall
		osOpenFile = c.openfile
		err := (Message{}).save()
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestFilepaths(t *testing.T) {
	stat := osStat
	open := osOpen
	defer func() {
		osStat = stat
		osOpen = open
	}()
	statfail := func(string) (sizer, error) { return nil, errMockFs }
	openfail := func(string) (*os.File, error) { return nil, errMockFs }
	cases := []struct {
		stat func(string) (sizer, error)
		open func(string) (*os.File, error)
		dir  string
		ok   bool
	}{
		{stat: statfail, open: open, dir: "testdata/noexist", ok: true},
		{stat: stat, open: openfail, dir: "testdata/noexist", ok: true},
		{stat: stat, open: open, dir: "testdata/notdir", ok: false},
		{stat: stat, open: openfail, dir: "testdata/r", ok: false},
	}

	for i, c := range cases {
		osStat = c.stat
		osOpen = c.open
		_, err := filepaths(c.dir)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestMessageLoader(t *testing.T) {
	cases := []struct {
		dir string
		ok  bool
	}{
		{dir: "testdata/notdir", ok: false},
		{dir: "testdata/r", ok: true},
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

func mkReadFile(b []byte, err error) func(string) ([]byte, error) {
	return func(string) ([]byte, error) { return b, err }
}

func TestLoadMessage(t *testing.T) {
	orig := ioutilReadFile
	defer func() { ioutilReadFile = orig }()

	cases := []struct {
		readFile func(string) ([]byte, error)
		ok       bool
	}{
		{readFile: mkReadFile(nil, errMockFs), ok: false},
		{readFile: mkReadFile([]byte("}"), nil), ok: false},
		{readFile: mkReadFile([]byte("{}"), nil), ok: true},
	}

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
	var n int64 = 10
	cases := []struct {
		dir string
		lim *int64
		ok  bool
	}{
		{dir: "testdata/notdir", ok: false},
		{dir: "testdata/w", ok: true},
		{dir: "testdata/w", lim: &n, ok: false},
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

func TestErrorStorageFull(t *testing.T) {
	(ErrStorageFull{}).Error()
}

func TestServe(t *testing.T) {
	c := make(chan struct{})
	close(c)
	p := Persistor{done: c}
	p.serve()
}
