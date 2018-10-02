package broker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/fsutil"
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

	fs Fs
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

// Fs provides file system functions.
type Fs interface {
	Open(string) (afero.File, error)
	Stat(string) (os.FileInfo, error)
	MkdirAll(string, os.FileMode) error
	OpenFile(string, int, os.FileMode) (afero.File, error)
	Remove(path string) error
}

// Persistor controls a persistence layer for Messages.
type Persistor struct {
	limit        *int64        // storage limit in bytes; no limit if nil
	newLimit     <-chan *int64 // incoming new values for limit
	currentLimit chan *int64   // outgoing current values for limit
	dir          string
	count        int // counter to give Messages unique names
	done         chan struct{}

	fs Fs
}

// NewPersistor creates a new Persistor in dir.
func NewPersistor(dir string, fs Fs, conf <-chan *int64) *Persistor {
	p := &Persistor{
		dir:          dir,
		newLimit:     conf,
		currentLimit: make(chan *int64),
		done:         make(chan struct{}),
		fs:           fs,
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

// MessageLoader generates a stream of messages from the filesystem.
type MessageLoader struct {
	out <-chan Message
}

// NewMessageLoader reads dir for messages and returns them as a stream.
func NewMessageLoader(dir string, fs Fs) MessageLoader {
	return MessageLoader{load(dir, fs)}
}

// Output returns l's output stream.
func (l MessageLoader) Output() <-chan Message { return l.out }

// load loads the output channel with messages from the filesystem.
func load(dir string, fs Fs) <-chan Message {
	out := make(chan Message)
	go func() {
		defer close(out)
		paths, err := filepaths(dir, fs)
		if err != nil {
			out <- Message{
				Error: err.Error(),
				Topic: Log,
			}
		}
		for _, path := range paths {
			out <- loadMessage(path, fs)
		}
	}()
	return out
}

// loadMessage decodes the file at path into a Message.
func loadMessage(path string, fs Fs) (m Message) {
	m.path = path
	m.fs = fs
	b, err := readFile(path, fs.Open)
	if err != nil {
		m.Error = err.Error()
		return
	}
	if err := json.Unmarshal(b, &m); err != nil {
		m.Error = err.Error()
	}
	return
}

// openFunc provides a way of opening files.
type openFunc func(string) (afero.File, error)

func readFile(path string, open openFunc) ([]byte, error) {
	f, err := open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

// CreateMessage creates a new Message under p.
func (p *Persistor) CreateMessage(m *Message) (err error) {
	lim := <-p.currentLimit
	totalSize, err := size(p.dir, p.fs)
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
	m.fs = p.fs
	p.count++
	return m.save()
}

func size(dir string, fs Fs) (int64, error) {
	var n int64
	paths, err := filepaths(dir, fs)
	if err != nil {
		return n, err
	}
	for _, path := range paths {
		f, err2 := fs.Stat(path)
		if err2 != nil {
			err = fmt.Errorf("size: failed to calculate storage size of message %v: %v", path, err2)
			continue
		}
		n += f.Size()
	}
	return n, err
}

// filepaths returns a list of paths of messages.
func filepaths(dir string, fs Fs) ([]string, error) {
	var paths []string
	if _, err := fs.Stat(dir); err != nil {
		// no directory; this is not necessarily an error.
		return paths, nil
	}
	d, err := fs.Open(dir)
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
	if err := m.fs.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("save: unable to save message to %v: %v", dir, err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		// Message doesn't currently export anything that can't be
		// encoded to JSON; this check is here to keep it that way.
		return fmt.Errorf("save: could not marshal JSON: %v", err)
	}
	return fsutil.WriteFile(m.fs.OpenFile, m.path, b)
}

// Remove deletes m from the persistence layer.
func (m Message) Remove() {
	if m.fs == nil {
		return
	}
	if err := m.fs.Remove(m.path); err != nil {
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
