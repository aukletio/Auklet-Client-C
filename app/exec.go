// +build linux

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
	"syscall"
)

// Exec represents an executable.
type Exec struct {
	hash string
	cmd  *exec.Cmd

	// state initialized after confirming that the application is released
	AppLogs   io.Reader
	AgentData io.ReadWriter // raw data stream from the agent

	// state initialized after the process starts
	agentVersion string
	Decoder      *json.Decoder // reading from AgentData
}

// NewExec creates a new executable from one or more arguments.
func NewExec(name string, args ...string) (*Exec, error) {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return &Exec{
		hash: fmt.Sprintf("%x", sha512.Sum512_224(bytes)),
		cmd:  cmd,
	}, nil
}

var socketPair = socketpair

// AddSockets adds sockets to the executable so that we can communicate with
// it. AddSockets must be called before starting the executable.
//
// WARNING: Do not call this function on an unreleased executable!
func (exec *Exec) AddSockets() error {
	// If we fail to create sockets, we can't communicate with the running
	// process. But we should try to send these errors to somebody.
	appLogs, err := socketPair("appLogs")
	if err != nil {
		return err
	}

	agentData, err := socketPair("agentData")
	if err != nil {
		return err
	}

	// It's important that the files be given in this order, because it
	// determines what numbers they get in the child process.
	exec.cmd.ExtraFiles = append(exec.cmd.ExtraFiles,
		appLogs.remote,   // fd 3
		agentData.remote, // fd 4
	)

	exec.AppLogs = appLogs.local
	exec.AgentData = agentData.local

	return nil
}

// Start starts the OS process.
func (exec *Exec) Start() error {
	// These files must be closed after the process is started. We do not
	// use them, but if we fail to close them, our listeners might not
	// terminate when the process closes its copies of them.
	for _, file := range exec.cmd.ExtraFiles {
		defer file.Close()
	}
	return exec.cmd.Start()
}

var (
	errEncoding  = errors.New("incorrect agent version syntax")
	errEOF       = errors.New("expected agent version, got EOF")
	errNoVersion = errors.New("empty agent version")
)

// GetAgentVersion reads from the AgentData stream and reads the agentVersion.
// This function must be called after starting the executable.
//
// WARNING: Do not call this function on an unreleased executable!
func (exec *Exec) GetAgentVersion() error {
	var msg struct {
		Version string `json:"version"`
	}

	// We should use a timeout here to avoid blocking indefinitely.

	dec := json.NewDecoder(exec.AgentData)
	if err := dec.Decode(&msg); err == io.EOF {
		// The process died before it could convey its agentVersion.
		return errEOF
	} else if err != nil {
		// The process failed to speak versionMsg.
		return errEncoding
	}

	if msg.Version == "" {
		return errNoVersion
	}

	exec.agentVersion = msg.Version
	exec.Decoder = dec

	return nil
}

// Wait waits for the process to exit.
func (exec *Exec) Wait() { exec.cmd.Wait() }

// CheckSum returns the executable file's SHA512/224 sum.
func (exec *Exec) CheckSum() string { return exec.hash }

// ExitStatus returns the process's exit status.
func (exec *Exec) ExitStatus() int {
	exec.Wait()
	return exec.cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

// Signal returns the text description of the signal that killed the process, if
// any.
func (exec *Exec) Signal() string {
	exec.Wait()
	ws := exec.cmd.ProcessState.Sys().(syscall.WaitStatus)
	sig := ""
	if ws.Signaled() {
		sig = ws.Signal().String()
	}
	return sig
}

// String returns the exectuable path and agent version as a formatted string.
func (exec *Exec) String() string {
	return fmt.Sprintf("%s %s", exec.cmd.Path, exec.agentVersion)
}

// AgentVersion returns the agent version running in the process. It may be
// called only after GetAgentVersion succeeds.
func (exec *Exec) AgentVersion() string {
	return exec.agentVersion
}
