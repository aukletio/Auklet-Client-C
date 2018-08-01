package agent

import (
	"bufio"
	"io"
	"log"
)

// Logger is a remote logging connection server.
type Logger struct {
	line *bufio.Scanner
	out  chan Message
}

// NewLogger returns a Logger that uses handler to convert data from in into
// broker Messages.
func NewLogger(in io.Reader) Logger {
	l := Logger{
		line: bufio.NewScanner(in),
		out:  make(chan Message),
	}
	go l.serve()
	return l
}

// serve activates l, causing it to send and receive messages.
func (l Logger) serve() {
	defer close(l.out)
	log.Printf("Logger: accepted connection")
	defer log.Printf("Logger: connection closed")
	for l.line.Scan() {
		l.out <- Message{
			Type: "applog",
			Data: l.line.Bytes(),
		}
	}
	if err := l.line.Err(); err != nil {
		l.out <- Message{
			Type:  "log",
			Error: err.Error(),
		}
	}
}

// Output returns l's output channel.
func (l Logger) Output() <-chan Message {
	return l.out
}
