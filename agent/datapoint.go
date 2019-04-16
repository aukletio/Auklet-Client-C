package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aukletio/Auklet-Client-C/errorlog"
)

// DataPointServer reads a stream of data point JSON messages.
type DataPointServer struct {
	in  io.Reader
	dec *json.Decoder
	out chan Message
	msg Message
}

// NewDataPointServer returns a new DataPointServer.
func NewDataPointServer(in io.Reader) *DataPointServer {
	s := &DataPointServer{
		in:  in,
		dec: json.NewDecoder(in),
		out: make(chan Message),
	}
	go s.serve()
	return s
}

func (s *DataPointServer) scan() bool {
	msg := Message{Type: "datapoint"}
	// Decode the stream into the Data field,
	// since "data point" can be arbitrary JSON.
	switch err := s.dec.Decode(&msg.Data); err {
	case nil:
		s.msg = msg
		return true

	case io.EOF:
		return false

	default:
		buf, _ := ioutil.ReadAll(s.dec.Buffered())
		err := fmt.Errorf("%v in %v", err.Error(), string(buf))
		msg := Message{
			Type:  "log",
			Error: err.Error(),
		}
		s.msg = msg
		s.dec = json.NewDecoder(s.in)
		errorlog.Printf("DataPointServer.serve: %v in %q", err, string(buf))
		return true
	}
}

func (s *DataPointServer) serve() {
	defer close(s.out)
	for s.scan() {
		s.out <- s.msg
	}
}

// Output returns s's output stream.
func (s *DataPointServer) Output() <-chan Message {
	return s.out
}
