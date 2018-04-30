// Package message implements interfaces for manipulating streams of Kafka
// messages.
package message

// Message is implemented by types that can be sent as Kafka messages.
type Message interface {
	Topic() string
	Bytes() ([]byte, error)
}

// Source is implemented by types that can generate a Message stream.
type Source interface {
	// Output returns a channel of Messages provided by a Source. A source
	// indicates when it has no more Messages to send by closing the
	// channel.
	Output() <-chan Message
}

// SourceError is implemented by Sources that can respond to error values
// generated by their clients while processing a Message.
type SourceError interface {
	Source
	// Err returns a channel on which clients can send values. Clients
	// close the channel to indicate that they have nothing more to send.
	Err() chan<- error
}

// A serverState, when executed, returns the next server state. A nil
// serverState signifies a server's termination.
type serverState func() serverState
