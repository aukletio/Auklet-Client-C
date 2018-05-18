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
func (p Persistent) Bytes() []byte {
	return p.BBytes
}

var count = 0

// Persist persists m to a file in dir.
func Persist(m kafka.Message, dir string) (p Persistent, err error) {
	p = Persistent{
		BBytes: json.RawMessage(m.Bytes()),
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

func (p *Persistent) Remove() error {
	return os.Remove(p.path)
}
