package app

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// oneOrMoreArgs repesents one or more command line arguments.
type oneOrMoreArgs struct {
	first string
	rest  []string
}

var errNumArgs = errors.New("need one or more arguments")

func newOneOrMoreArgs(args []string) (*oneOrMoreArgs, error) {
	if len(args) < 1 {
		return nil, errNumArgs
	}

	return &oneOrMoreArgs{
		first: args[0],
		rest:  args[1:],
	}, nil
}

// osExecutable contains a command and a checksum identifying the associated
// file.
type osExecutable struct {
	hash string
	cmd  *exec.Cmd
}

// start starts the OS process.
func (exec osExecutable) start() error {
	for _, file := range exec.cmd.ExtraFiles {
		// These files must be closed after the process is started. We
		// do not use them, but if we fail to close them, our listeners
		// might not terminate when the process closes its copies of
		// them.
		defer file.Close()
	}
	return exec.cmd.Start()
}

func (exec osExecutable) checksum() string {
	return exec.hash
}

func (exec osExecutable) inherit(files ...*os.File) {
	exec.cmd.ExtraFiles = append(exec.cmd.ExtraFiles, files...)
}

// newExec creates a new exectuable from one or more arguments.
func newExec(args oneOrMoreArgs) (*osExecutable, error) {
	bytes, err := ioutil.ReadFile(args.first)
	if err != nil {
		return nil, err
	}

	return &osExecutable{
		hash: func() string {
			return fmt.Sprintf("%x", sha512.Sum512_224(bytes))
		}(),
		cmd: func() *exec.Cmd {
			cmd := exec.Command(args.first, args.rest...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd
		}(),
	}, nil
}

type enviro struct {
	appID  string
	apiKey string
}

var (
	errNoAppID  = errors.New("missing APP_ID")
	errNoAPIKey = errors.New("missing API_KEY")
)

func newEnviro(getenv func(string) string) (*enviro, error) {
	appID := getenv("APP_ID")
	apiKey := getenv("API_KEY")

	switch "" {
	case appID:
		return nil, errNoAppID
	case apiKey:
		return nil, errNoAPIKey
	}

	return &enviro{
		appID:  appID,
		apiKey: apiKey,
	}, nil
}

// relExec represents an executable and proof that it has been released. The
// running process can be communicated with by the given streams.
type relExec struct {
	exec  executable
	proof relProof

	// streams for communicating with the running process.
	appLogs   io.Reader
	agentData io.ReadWriter
}

// relProof constitutes proof that a release exists.
type relProof struct{}

// errBug represents an unexpected error. These errors should be reported
// remotely, if possible.
type errBug struct {
	what error
}

func (err errBug) Error() string {
	return fmt.Sprintf("unexpected error: %v", err.what)
}

type executable interface {
	start() error
	inherit(...*os.File)
	checksum() string
}

// relChecker is a type of function that can be used to check whether a release
// exists.
type relChecker func(env enviro, checksum string) (*relProof, error)

var socketPair = socketpair

// newRelExec returns a relExec if check proves that exec has been released.
func newRelExec(env enviro, check relChecker, exec executable) (*relExec, error) {
	proof, err := check(env, exec.checksum())
	if err != nil {
		return nil, err
	}

	// If we fail to create sockets, we can't communicate with the running
	// process, so we shouldn't return a relExec. But we should try to send
	// these errors to somebody.
	appLogs, err := socketPair("appLogs")
	if err != nil {
		return nil, errBug{err}
	}

	agentData, err := socketPair("agentData")
	if err != nil {
		return nil, errBug{err}
	}

	// It's important that the files be given in this order, because it
	// determines what numbers they get in the child process.
	//           fd 3            fd 4
	exec.inherit(appLogs.remote, agentData.remote)

	return &relExec{
		exec:      exec,
		proof:     *proof,
		appLogs:   appLogs.local,
		agentData: agentData.local,
	}, nil
}

func (rel relExec) start() error {
	return rel.exec.start()
}

// An instdProc is a running process that we have successfully communicated
// with (by getting its agentVersion).
type instdProc struct {
	rel          relExec
	agentVersion string
	dec          *json.Decoder
}

var errNoVersion = errors.New("empty agentVersion")

// run starts the released executable and returns a process.
func run(rel relExec) (*instdProc, error) {
	// We're not expecting starting the process to fail.
	if err := rel.start(); err != nil {
		return nil, errBug{err}
	}

	version, dec, err := getAgentVersion(rel.agentData)
	if err != nil {
		return nil, err
	}

	return &instdProc{
		rel:          rel,
		agentVersion: version,
		dec:          dec,
	}, nil
}

func getAgentVersion(r io.Reader) (string, *json.Decoder, error) {
	// Agent messages have other schemas, but the first message has to
	// match this.
	type versionMsg struct {
		Version string `json:"version"`
	}

	// We should use a timeout here to prevent hanging indefinitely.

	dec := json.NewDecoder(r)
	var msg versionMsg
	if err := dec.Decode(&msg); err == io.EOF {
		// The process died before it could convey its agentVersion.
		return "", nil, errBug{err}
	} else if err != nil {
		// The process failed to speak versionMsg.
		return "", nil, errBug{err}
	}

	if msg.Version == "" {
		return "", nil, errBug{errNoVersion}
	}
	return msg.Version, dec, nil
}
