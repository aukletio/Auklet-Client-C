// +build linux

// Package app provides a model of applications using Auklet.
package app

import (
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/ESG-USA/Auklet-Client-C/api"
	"github.com/ESG-USA/Auklet-Client-C/config"
	"github.com/ESG-USA/Auklet-Client-C/errorlog"
)

// An App represents an application using Auklet.
type App struct {
	cmd         *exec.Cmd
	logs        pair // fd 3 - application use
	data        pair // fd 4 - agent use
	checkSum    string
	id          string
	isReleased  bool
	haveChecked bool // whether we have checked if this app is released
}

// New returns an App that would execute args.
func New(args []string) *App {
	// fd 3
	logs, err := socketpair("logs")
	if err != nil {
		errorlog.Print(err)
	}
	// fd 4
	data, err := socketpair("data")
	if err != nil {
		errorlog.Print(err)
	}

	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return &App{
		cmd:      c,
		logs:     logs,
		data:     data,
		checkSum: sum(c.Path),
		id:       config.AppID(),
	}
}

func (a App) String() string {
	return a.cmd.Path
}

// IsReleased returns whether the app is known to have been released.
func (a *App) IsReleased() bool {
	if a.haveChecked {
		return a.isReleased
	}
	a.isReleased = api.Release(a.checkSum)
	a.haveChecked = true
	return a.isReleased
}

// Start wraps the underlying call to Cmd.Start and logs any errors.
func (a *App) Start() (err error) {
	if a.isReleased {
		// only pass on files if a has been released. It's critical that
		// we append the files with logs first, then data, to ensure
		// that the child process inherits them as fd 3 and 4,
		// respectively.
		a.cmd.ExtraFiles = append(a.cmd.ExtraFiles, a.logs.remote, a.data.remote)
	}
	// We will not be using our copies of any file descriptors we passed on
	// to the app. We close them to ensure that our servers receive EOF when
	// the app's copies are closed.
	for _, f := range a.cmd.ExtraFiles {
		defer f.Close()
	}
	err = a.cmd.Start()
	if err == nil {
		log.Printf("%v started", a)
	} else {
		errorlog.Print(err)
	}
	return
}

// Wait wraps the underlying call to Cmd.Wait and logs the exit of app.
func (a *App) Wait() {
	a.cmd.Wait()
	log.Printf("%v exited", a)
}

// CheckSum is the SHA512/224 hash of the executable file with which we identify
// a build.
func (a *App) CheckSum() string { return a.checkSum }

// ID returns the application ID.
func (a *App) ID() string { return a.id }

// ExitStatus returns the app's exit status.
func (a *App) ExitStatus() int {
	return a.cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

// Signal returns a non-empty string describing the signal that killed the app.
// If no signal killed the app, an empty string is returned.
func (a *App) Signal() string {
	ws := a.cmd.ProcessState.Sys().(syscall.WaitStatus)
	if ws.Signaled() {
		return ws.Signal().String()
	}
	return ""
}

// sum calculates the SHA512/224 hash of the file located at path.
func sum(path string) (hash string) {
	f, err := os.Open(path)
	if err != nil {
		errorlog.Print(err)
		return
	}
	defer f.Close()
	h := sha512.New512_224()
	if _, err = io.Copy(h, f); err != nil {
		errorlog.Print(err)
		return
	}
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}

// Logs returns a's custom log connection.
func (a *App) Logs() *os.File {
	return a.logs.local
}

// Data returns a's instrumentation data connection.
func (a *App) Data() *os.File {
	return a.data.local
}
