package schema

import (
	"encoding/json"
	"errors"

	"github.com/vmihailenco/msgpack"
)

// RawMessage represents a raw JSON message that can be encoded to MessagePack.
type RawMessage []byte

// MarshalJSON returns m as a byte slice.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets m to data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

// EncodeMsgpack uses enc to encode m as MessagePack.
func (m RawMessage) EncodeMsgpack(enc *msgpack.Encoder) error {
	var v interface{}
	err := json.Unmarshal(m, &v)
	if err != nil {
		return err
	}
	return enc.Encode(v)
}
