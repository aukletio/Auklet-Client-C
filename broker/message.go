package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// This file defines interfaces for manipulating streams of broker
// messages, plus a message persistence layer.

// Topic encodes a Message topic.
type Topic int

// Profile, Event, and Log are Message types.
const (
	Profile Topic = iota
	Event
	Log
)

// Message represents a broker message.
type Message struct {
	Error string `json:"error"`
	Topic Topic  `json:"topic"`
	Bytes []byte `json:"bytes"`
	path  string
}

// ErrStorageFull indicates that the corresponding Persistor is full.
type ErrStorageFull struct {
	limit int64
	used  int64
}

// Error returns e as a string.
func (e ErrStorageFull) Error() string {
	return fmt.Sprintf("persistor: storage full: %v used of %v limit", e.used, e.limit)
}

// Persistor controls a persistence layer for Messages.
type Persistor struct {
	limit        *int64      // storage limit in bytes; no limit if nil
	newLimit     chan *int64 // incoming new values for limit
	currentLimit chan *int64 // outgoing current values for limit
	dir          string
	count        int // counter to give Messages unique names
	done         chan struct{}
}

// NewPersistor creates a new Persistor in dir.
func NewPersistor(dir string) *Persistor {
	p := &Persistor{
		dir:          dir,
		newLimit:     make(chan *int64),
		currentLimit: make(chan *int64),
		done:         make(chan struct{}),
	}
	go p.serve()
	return p
}

// serve serializes access to p.limit
func (p *Persistor) serve() {
	for {
		select {
		case <-p.done:
			return
		case p.limit = <-p.newLimit:
		case p.currentLimit <- p.limit:
		}
	}
}

// Configure returns a channel on which p's storage limit can be controlled.
func (p *Persistor) Configure() chan<- *int64 {
	return p.newLimit
}

// MessageLoader generates a stream of messages from the filesystem.
type MessageLoader struct {
	out <-chan Message
}

// NewMessageLoader reads dir for messages and returns them as a stream.
func NewMessageLoader(dir string) MessageLoader {
	return MessageLoader{load(dir)}
}

// Output returns l's output stream.
func (l MessageLoader) Output() <-chan Message { return l.out }

// load loads the output channel with messages from the filesystem.
func load(dir string) <-chan Message {
	out := make(chan Message)
	go func() {
		defer close(out)
		paths, err := filepaths(dir)
		if err != nil {
			out <- Message{
				Error: err.Error(),
				Topic: Log,
			}
		}
		for _, path := range paths {
			out <- loadMessage(path)
		}
	}()
	return out
}

// loadMessage decodes the file at path into a Message.
func loadMessage(path string) (m Message) {
	m.path = path
	b, err := ioutilReadFile(path)
	if err != nil {
		m.Error = err.Error()
		return
	}
	if err = json.Unmarshal(b, &m); err != nil {
		m.Error = err.Error()
	}
	return
}

// CreateMessage creates a new Message under p.
func (p *Persistor) CreateMessage(m *Message) (err error) {
	lim := <-p.currentLimit
	totalSize, err := size(p.dir)
	if err != nil {
		return err
	}
	if lim != nil && int64(len(m.Bytes))+totalSize > 9**lim/10 {
		return ErrStorageFull{
			limit: *lim,
			used:  totalSize,
		}
	}
	m.path = fmt.Sprintf("%v/%v-%v", p.dir, os.Getpid(), p.count)
	p.count++
	return m.save()
}

type sizer interface {
	Size() int64
}

var (
	osOpen         = os.Open
	osStat         = func(path string) (sizer, error) { return os.Stat(path) }
	osMkdirAll     = os.MkdirAll
	osOpenFile     = os.OpenFile
	ioutilReadFile = ioutil.ReadFile
)

func size(dir string) (int64, error) {
	var n int64
	paths, err := filepaths(dir)
	if err != nil {
		return n, err
	}
	for _, path := range paths {
		f, err2 := osStat(path)
		if err2 != nil {
			err = fmt.Errorf("size: failed to calculate storage size of message %v: %v", path, err2)
			continue
		}
		n += f.Size()
	}
	return n, err
}

// filepaths returns a list of paths of messages.
var filepaths = func(dir string) ([]string, error) {
	var paths []string
	if _, err := osStat(dir); err != nil {
		// no directory; this is not necessarily an error.
		return paths, nil
	}
	d, err := osOpen(dir)
	if err != nil {
		return paths, fmt.Errorf("filepaths: failed to open message directory: %v", err)
	}
	defer d.Close()
	names, err := d.Readdirnames(0)
	if err != nil {
		return paths, fmt.Errorf("filepaths: failed to read directory names in %v: %v", d.Name(), err)
	}
	for _, name := range names {
		paths = append(paths, dir+"/"+name)
	}
	return paths, nil
}

func (m Message) save() error {
	dir := filepath.Dir(m.path)
	if err := osMkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("save: unable to save message to %v: %v", dir, err)
	}
	b, _ := json.Marshal(m)
	// None of the types within Message
	// can cause Marshal to fail, so we
	// don't use the error value.
	return ioutil.WriteFile(m.path, b, 0644)
}

// Remove deletes m from the persistence layer.
func (m Message) Remove() {
	if err := os.Remove(m.path); err != nil {
		errorlog.Print(err)
	}
}

// MessageSource is implemented by types that can generate a Message stream.
type MessageSource interface {
	// Output returns a channel of Messages provided by a Source. A source
	// indicates when it has no more Messages to send by closing the
	// channel.
	Output() <-chan Message
}
