package agent

import (
	"bufio"
	"io"
	"log"

	"github.com/ESG-USA/Auklet-Client-C/broker"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// Handler converts data into a broker.Message.
type Handler func(data []byte) (broker.Message, error)

// Logger is a remote logging connection server.
type Logger struct {
	out     chan broker.Message
	conn    io.Reader
	handler Handler
}

// NewLogger returns a Logger that uses handler to convert data from conn into
// broker Messages.
func NewLogger(conn io.Reader, handler Handler) Logger {
	l := Logger{
		conn:    conn,
		out:     make(chan broker.Message),
		handler: handler,
	}
	go l.serve()
	return l
}

// serve activates l, causing it to send and receive messages.
func (l Logger) serve() {
	defer close(l.out)
	log.Printf("Logger: accepted connection")
	defer log.Printf("Logger: connection closed")
	s := bufio.NewScanner(l.conn)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		log.Printf(`got log "%v"`, s.Text())
		m, err := l.handler(s.Bytes())
		if err != nil {
			errorlog.Print(err)
			continue
		}
		l.out <- m
	}
}

// Output returns l's output channel.
func (l Logger) Output() <-chan broker.Message {
	return l.out
}
