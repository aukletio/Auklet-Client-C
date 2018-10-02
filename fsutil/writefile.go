package fsutil

import (
	"os"

	"github.com/spf13/afero"
)

// OpenFileFunc is a function that can open or create a file.
type OpenFileFunc func(string, int, os.FileMode) (afero.File, error)

// WriteFile opens or creates the file at path and writes b to it.
func WriteFile(openFile OpenFileFunc, path string, b []byte) error {
	f, err := openFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}
