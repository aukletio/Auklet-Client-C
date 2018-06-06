package agent

import (
	"bufio"
	"log"
	"net"

	"github.com/ESG-USA/Auklet-Client/errorlog"
	"github.com/ESG-USA/Auklet-Client/kafka"
)

// Logger is a remote logging connection server. Applications using Auklet can
// write newline-delimited messages to the logger's socket to send logs to
// Auklet's backend.
type Logger struct {
	l       net.Listener
	out     chan kafka.Message
	handler Handler
}

// NewLogger opens a socket at addr and returns a Logger that uses handler to
// convert socket messages into kafka Messages.
func NewLogger(addr string, handler Handler) Logger {
	l, err := net.Listen("unix", addr)
	if err != nil {
		errorlog.Print(err)
	}
	return Logger{
		l:       l,
		out:     make(chan kafka.Message),
		handler: handler,
	}
}

// Serve activates l, causing it to send and receive messages.
func (l Logger) Serve() {
	defer l.l.Close()
	conn, err := l.l.Accept()
	if err != nil {
		errorlog.Print(err)
	}
	log.Printf("accepted connection on %v", l.l.Addr())
	defer log.Printf("connection on %v closed", l.l.Addr())
	s := bufio.NewScanner(conn)
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
func (l Logger) Output() <-chan kafka.Message {
	return l.out
}
