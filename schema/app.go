package schema

type App interface {
	CheckSum() string
	ID() string
}

type Exiter interface {
	ExitStatus() int
}

type Signaller interface {
	Signal() string
}
