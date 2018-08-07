package schema

import (
	"bytes"

	"github.com/vmihailenco/msgpack"

	"github.com/ESG-USA/Auklet-Client-C/broker"
)

// marshaler determines which transport encoding is used for messages.
var marshaler = msgpackMarshal

// msgpackMarshal has the same signature as json.Marshal, so that the two
// functions can be interchanged.
func msgpackMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseJSONTag(true)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func marshal(v interface{}, topic broker.Topic) broker.Message {
	bytes, err := marshaler(v)
	return broker.Message{
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
		Bytes: bytes,
		Topic: topic,
	}
}