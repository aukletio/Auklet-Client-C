package agent

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/ESG-USA/Auklet-Client-C/errorlog"
	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// Logger is a remote logging connection server. Applications using Auklet can
// write newline-delimited messages to the logger's socket to send logs to
// Auklet's backend.
type Logger struct {
	local, remote *os.File
	out           chan broker.Message
	handler       Handler
}

// NewLogger opens an anonymous socket and returns a Logger that uses handler to
// convert socket messages into broker Messages.
func NewLogger(handler Handler) Logger {
	local, remote, err := socketpair("logserver-")
	if err != nil {
		errorlog.Print(err)
	}
	return Logger{
		local:   local,
		remote:  remote,
		out:     make(chan broker.Message),
		handler: handler,
	}
}

// Serve activates l, causing it to send and receive messages.
func (l Logger) Serve() {
	defer close(l.out)
	conn, err := net.FileConn(l.local)
	if err != nil {
		errorlog.Print(err)
	}
	log.Printf("accepted connection on %v", l.local.Name())
	defer log.Printf("connection on %v closed", l.local.Name())
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
func (l Logger) Output() <-chan broker.Message {
	return l.out
}

// Remote returns the socket to be inherited by the child process.
func (l Logger) Remote() *os.File {
	return l.remote
}
