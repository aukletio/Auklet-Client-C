// Package message implements types for manipulating streams of broker
// messages.
package message

// A serverState, when executed, returns the next server state. A nil
// serverState signifies a server's termination.
type serverState func() serverState
