package broker

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/ESG-USA/Auklet-Client-C/fsutil"
)

var errMockFs = errors.New("filesystem error")

type mockFs struct {
	fs Fs

	errOpenFile error
	errMkdirAll error
	errOpen     error
}

func (m mockFs) MkdirAll(path string, mode os.FileMode) error {
	if m.errMkdirAll != nil {
		return m.errMkdirAll
	}
	return m.fs.MkdirAll(path, mode)
}

func (m mockFs) OpenFile(path string, flag int, perm os.FileMode) (afero.File, error) {
	if m.errOpenFile != nil {
		return nil, m.errOpenFile
	}
	return m.fs.OpenFile(path, flag, perm)
}

func (m mockFs) Stat(path string) (os.FileInfo, error) {
	return m.fs.Stat(path)
}

func (m mockFs) Open(path string) (afero.File, error) {
	if m.errOpen != nil {
		return nil, m.errOpen
	}
	return m.fs.Open(path)
}

func (m mockFs) Remove(path string) error { return m.Remove(path) }

func TestSave(t *testing.T) {
	cases := []struct {
		fs Fs
		ok bool
	}{
		{
			fs: mockFs{
				fs:          afero.NewMemMapFs(),
				errMkdirAll: errMockFs,
				errOpenFile: nil,
			},
			ok: false,
		},
		{
			fs: mockFs{
				fs:          afero.NewMemMapFs(),
				errMkdirAll: nil,
				errOpenFile: errMockFs,
			},
			ok: false,
		},
		{
			fs: mockFs{
				fs:          afero.NewMemMapFs(),
				errMkdirAll: nil,
				errOpenFile: nil,
			},
			ok: true,
		},
	}

	for i, c := range cases {
		err := (Message{
			path: "",
			fs:   c.fs,
		}).save()
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func touch(path string) Fs {
	fs := afero.NewMemMapFs()
	f, err := fs.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return fs
}

func TestFilepaths(t *testing.T) {
	cases := []struct {
		fs  Fs
		dir string
		ok  bool
	}{
		{
			// dir doesn't exist
			fs:  afero.NewMemMapFs(),
			dir: "noexist",
			ok:  true,
		},
		{
			// dir exists, but not readable
			fs: mockFs{
				fs: func() Fs {
					fs := afero.NewMemMapFs()
					if err := fs.MkdirAll("dir", 0777); err != nil {
						panic(err)
					}
					return fs
				}(),
				errOpen: errMockFs, // as if can't read entries; bad permissions?
			},
			dir: "dir",
			ok:  false,
		},
		{
			// file exists but is not a dir
			fs:  touch("notdir"),
			dir: "notdir",
			ok:  false,
		},
	}

	for i, c := range cases {
		_, err := filepaths(c.dir, c.fs)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func oneFile(path string) Fs {
	fs := afero.NewMemMapFs()
	if err := fs.Mkdir(path, 0777); err != nil {
		panic(err)
	}
	if err := fsutil.WriteFile(fs.OpenFile, path+"/file", []byte("{}")); err != nil {
		panic(err)
	}
	return fs
}

func TestMessageLoader(t *testing.T) {
	cases := []struct {
		fs  Fs
		dir string
		ok  bool
	}{
		{fs: touch("notdir"), dir: "notdir", ok: false},
		{fs: oneFile("dir"), dir: "dir", ok: true},
	}

	for i, c := range cases {
		l := NewMessageLoader(c.dir, c.fs)
		m := <-l.Output()
		ok := m.Error == ""
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, m.Error)
		}
	}
}

func TestSize(t *testing.T) {
	cases := []struct {
		dir string
		fs  Fs
		ok  bool
	}{
		{dir: "notdir", fs: touch("notdir"), ok: false},
		{dir: "dir", fs: oneFile("dir"), ok: true},
	}

	for i, c := range cases {
		_, err := size(c.dir, c.fs)
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
	}
}

func TestChannels(t *testing.T) {
	cfg := make(chan *int64)
	p := NewPersistor("", afero.NewMemMapFs(), cfg)
	defer close(p.done)
	var l int64 = 42
	cfg <- &l
	if lim := <-p.currentLimit; lim != &l {
		t.Fail()
	}
}

func TestCreateMessage(t *testing.T) {
	cases := []struct {
		fs  Fs
		dir string
		ok  bool
	}{
		{fs: touch("notdir"), dir: "notdir", ok: false},
		{fs: afero.NewMemMapFs(), dir: "", ok: true},
	}

	for i, c := range cases {
		p := NewPersistor(c.dir, c.fs, nil)
		err := p.CreateMessage(&Message{})
		ok := err == nil
		if ok != c.ok {
			t.Errorf("case %v: expected %v, got %v: %v", i, c.ok, ok, err)
		}
		close(p.done)
	}
}
