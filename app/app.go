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
	// Cmd is the command to be executed.
	*exec.Cmd
	checkSum   string
	id         string
	IsReleased bool
}

// CheckSum is the SHA512/224 hash of the executable file (Cmd.Path)
// with which we identify a build.
func (a *App) CheckSum() string { return a.checkSum }

// ID returns the application ID.
func (a *App) ID() string { return a.id }

// ExitStatus returns the app's exit status.
func (a *App) ExitStatus() int {
	return a.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

// Signal returns a non-empty string describing the signal that killed the app.
// If no signal killed the app, an empty string is returned.
func (a *App) Signal() string {
	ws := a.ProcessState.Sys().(syscall.WaitStatus)
	if ws.Signaled() {
		return ws.Signal().String()
	}
	return ""
}

// New returns an App that would execute args.
func New(args []string) (app *App) {
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	s := sum(c.Path)
	app = &App{
		Cmd:        c,
		checkSum:   s,
		id:         config.AppID(),
		IsReleased: api.Release(s),
	}
	return
}

// Start wraps the underlying call to Cmd.Start and logs any errors.
func (a *App) Start() (err error) {
	err = a.Cmd.Start()
	if err == nil {
		log.Printf("app %v started", a.Path)
	} else {
		errorlog.Print(err)
	}
	// We will not be using our copies of any file descriptors we
	// passed on to the app. We close them to ensure that our
	// servers receive EOF when the app's copies are closed.
	for _, f := range a.ExtraFiles {
		log.Println("closing ", f.Name())
		f.Close()
	}
	return
}

// Wait wraps the underlying call to Cmd.Wait and logs the exit of app.
func (a *App) Wait() {
	a.Cmd.Wait()
	log.Printf("app %v exited", a.Path)
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
