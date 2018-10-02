package message

import (
	"io"
	"testing"
)

type mockObj struct{}

func (mockObj) Encode(io.Writer) error { return nil }
func (mockObj) Decode(io.Reader) error { return nil }

func TestFilePersistor(t *testing.T) {
	cases := []struct {
		method, path  string
		shouldSucceed bool
	}{
		{
			// dir doesn't exist
			method:        "save",
			path:          "testdata/noexist/file",
			shouldSucceed: false,
		},
		{
			// file doesn't exist
			method:        "load",
			path:          "testdata/noexist/file",
			shouldSucceed: false,
		},
		{
			// have write permissions for this dir
			method:        "save",
			path:          "testdata/file",
			shouldSucceed: true,
		},
		{
			// file exists
			method:        "load",
			path:          "testdata/file",
			shouldSucceed: true,
		},
	}

	for i, c := range cases {
		p := FilePersistor{c.path}
		var err error
		switch c.method {
		case "save":
			err = p.Save(mockObj{})
		case "load":
			err = p.Load(mockObj{})
		}
		if c.shouldSucceed && err != nil {
			t.Errorf("case %v: %v: expected success, got %v", i, c.method, err)
		} else if !c.shouldSucceed && err == nil {
			t.Errorf("case %v: %v: expected failure, got no error", i, c.method)
		}
	}
}
