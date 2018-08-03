package broker

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestPersistor(t *testing.T) {
	fs = afero.NewMemMapFs()
	p := NewPersistor(".auklet/message")
	var limit int64 = 900
	p.Configure() <- &limit
	m := Message{
		Bytes: make([]byte, 500),
	}
	if err := p.CreateMessage(m); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	defer m.Remove()
	err := p.CreateMessage(m)
	if _, is := err.(ErrStorageFull); !is {
		t.Errorf("expected ErrStorageFull, got %v", err)
	}
	exp := "persistor: storage full: 702 used of 900 limit"
	if err == nil || err.Error() != exp {
		t.Errorf("expected %q, got %v", exp, err)
	}
	defer m.Remove()
}

func TestPersistorLoad(t *testing.T) {
	fs = afero.NewMemMapFs()
	p := NewPersistor(".auklet/message")
	m := Message{
		Bytes: make([]byte, 500),
	}
	if err := p.CreateMessage(m); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	loader := NewMessageLoader(p.dir)
	count := 0
	for m = range loader.Output() {
		count++
	}
	if count != 1 {
		t.Errorf("expected one persistent message, got %v", count)
	}
}

var errMockFs = errors.New("filesystem error")

type mockFs struct {
	mkdirAll func(string, os.FileMode) error
	open     func(string) (afero.File, error)
	openFile func(string, int, os.FileMode) (afero.File, error)
	stat     func(string) (os.FileInfo, error)
}

func (fs mockFs) MkdirAll(path string, perm os.FileMode) error {
	return fs.mkdirAll(path, perm)
}

func (fs mockFs) Open(name string) (afero.File, error) {
	return fs.open(name)
}

func (fs mockFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return fs.openFile(name, flag, perm)
}

func (fs mockFs) Stat(name string) (os.FileInfo, error) {
	return fs.stat(name)
}

// Package broker doesn't call these functions, but we have to implement them to
// satisfy afero.Fs.
func (mockFs) Name() string                               { panic(nil) }
func (mockFs) Create(string) (afero.File, error)          { panic(nil) }
func (mockFs) Mkdir(string, os.FileMode) error            { panic(nil) }
func (mockFs) Remove(string) error                        { panic(nil) }
func (mockFs) RemoveAll(string) error                     { panic(nil) }
func (mockFs) Rename(string, string) error                { panic(nil) }
func (mockFs) Chmod(string, os.FileMode) error            { panic(nil) }
func (mockFs) Chtimes(string, time.Time, time.Time) error { panic(nil) }

func TestSave(t *testing.T) {
	fs = mockFs{
		mkdirAll: func(string, os.FileMode) error { return errMockFs },
	}
	err := Message{}.save()
	exp := "save: unable to save message to .: filesystem error"
	if err == nil || err.Error() != exp {
		t.Errorf("expected %v, got %v", exp, err)
	}
}

func TestOpenFile(t *testing.T) {
	fs = mockFs{
		mkdirAll: func(string, os.FileMode) error { return nil },
		openFile: func(string, int, os.FileMode) (afero.File, error) { return nil, errMockFs },
	}
	err := Message{}.save()
	exp := "filesystem error"
	if err == nil || err.Error() != exp {
		t.Errorf("expected %v, got %v", exp, err)
	}
}

func TestStat(t *testing.T) {
	fs = mockFs{
		stat: func(string) (os.FileInfo, error) { return nil, errMockFs },
	}
	paths, err := filepaths(".auklet/messages")
	if len(paths) != 0 || err != nil {
		t.Errorf("expected empty slice, nil; got %v, %v", paths, err)
	}
}

func TestOpen(t *testing.T) {
	mmfs := afero.NewMemMapFs()
	fs = mockFs{
		stat: mmfs.Stat,
		open: func(string) (afero.File, error) { return nil, errMockFs },
	}
	paths, err := filepaths(".auklet/messages")
	exp := "filepaths: failed to open message directory: filesystem error"
	if len(paths) != 0 || (err != nil && err.Error() != exp) {
		t.Errorf("expected empty slice, %v; got %v, %v", exp, paths, err)
	}
}

func TestReaddirnames(t *testing.T) {
	mmfs := afero.NewMemMapFs()
	fs = mockFs{
		stat: mmfs.Stat,
		open: mmfs.Open,
	}
	// Create a regular file, then try to read it as a directory.
	f, err := mmfs.Create(".auklet")
	if err != nil {
		t.Error(err)
	}
	f.Close()

	paths, err := filepaths(".auklet/messages")
	exp := "filepaths: failed to open message directory: filesystem error"
	if len(paths) != 0 || (err != nil && err.Error() != exp) {
		t.Errorf("expected empty slice, %v; got %v, %v", exp, paths, err)
	}
}
