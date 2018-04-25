package message

import (
	"encoding/json"
	"fmt"
	"os"
)

type Persistent struct {
	Topic_ string          `json:"topic"`
	Bytes_ json.RawMessage `json:"bytes"`
	path   string          // location of this message in the filesystem
}

func (p Persistent) Topic() string {
	return p.Topic_
}

func (p Persistent) Bytes() ([]byte, error) {
	return p.Bytes_, nil
}

var count = 0

func toPersistent(m Message, dir string) (p Persistent, err error) {
	b, err := m.Bytes()
	if err != nil {
		return
	}
	p = Persistent{
		Bytes_: json.RawMessage(b),
		Topic_: m.Topic(),
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
