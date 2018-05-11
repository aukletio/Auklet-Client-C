package message

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ESG-USA/Auklet-Client/kafka"
)

// Persistent represents a Message in a form that can be written to and read
// from disk.
type Persistent struct {
	TTopic kafka.Topic     `json:"topic"`
	BBytes json.RawMessage `json:"bytes"`
	path   string          // location of this message in the filesystem
}

// Topic returns the Kafka topic of p.
func (p Persistent) Topic() kafka.Topic {
	return p.TTopic
}

// Bytes returns p's value as a byte slice.
func (p Persistent) Bytes() ([]byte, error) {
	return p.BBytes, nil
}

var count = 0

func toPersistent(m kafka.Message, dir string) (p Persistent, err error) {
	b, err := m.Bytes()
	if err != nil {
		return
	}
	p = Persistent{
		BBytes: json.RawMessage(b),
		TTopic: m.Topic(),
		path:   fmt.Sprintf("%v/%v-%v", dir, os.Getpid(), count),
	}
	count++
	err = p.save()
	return
}

func (p *Persistent) load() (err error) {
	f, err := os.Open(p.path)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(p)
	return
}

func (p *Persistent) save() (err error) {
	f, err := os.OpenFile(p.path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(p)
	return
}

func (p *Persistent) remove() error {
	return os.Remove(p.path)
}
