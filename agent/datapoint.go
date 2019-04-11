package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type DataPointServer struct {
	in  io.Reader
	out chan Message
}

func NewDataPointServer(in io.Reader) *DataPointServer {
	s := &DataPointServer{
		in:  in,
		out: make(chan Message),
	}
	go s.serve()
	return s
}

func (s *DataPointServer) serve() {
	defer close(s.out)
	dec := json.NewDecoder(s.in)
	for {
		msg := Message{Type: "datapoint"}
		// Decode the stream into the Data field,
		// since "data point" can be arbitrary JSON.
		switch err := dec.Decode(&msg.Data); err {
		case nil:
			s.out <- msg

		case io.EOF:
			return

		default:
			buf, _ := ioutil.ReadAll(dec.Buffered())
			s.out <- Message{
				Type:  "log",
				Error: fmt.Sprintf("%v in %v", err.Error(), string(buf)),
			}
			dec = json.NewDecoder(s.in)
		}
	}
}

// Output returns s's output stream.
func (s *DataPointServer) Output() <-chan Message {
	return s.out
}
