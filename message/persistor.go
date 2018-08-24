package message

import (
	"bytes"
	"io"
	"os"
)

// FilePersistor permits saving and loading data from a file.
type FilePersistor struct {
	Path string
}

// Save encodes object to the underlying file.
func (fp FilePersistor) Save(object Encodable) error {
	f, err := os.Create(fp.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return object.Encode(f)
}

// Load decodes object from the underlying file.
func (fp FilePersistor) Load(object Decodable) error {
	f, err := os.Open(fp.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	return object.Decode(f)
}

// MemPersistor permits saving and loading data from an internal memory buffer.
type MemPersistor struct {
	Buf bytes.Buffer
}

// Save encodes object to mp's internal buffer.
func (mp *MemPersistor) Save(object Encodable) error {
	return object.Encode(&mp.Buf)
}

// Load decodes object from mp's internal buffer.
func (mp *MemPersistor) Load(object Decodable) error {
	return object.Decode(&mp.Buf)
}

// Encodable is an object that can encode itself to a byte stream.
type Encodable interface {
	Encode(io.Writer) error
}

// Decodable is an object that can decode itself from a byte stream.
type Decodable interface {
	Decode(io.Reader) error
}
