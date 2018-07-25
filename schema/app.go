package schema

// This file defines interfaces needed by schema conversion functions.

// App is anything that can return a checksum and an ID.
type App interface {
	CheckSum() string
	ID() string
}

// Exiter is anything that can return an exit status.
type Exiter interface {
	ExitStatus() int
}

// Signaller is anything that can return a signal description.
type Signaller interface {
	Signal() string
}
