// +build linux

// Package app provides a model of applications using Auklet.
package app

import (
	"crypto/sha512"
	"syscall"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

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

// Checksum is the SHA512/224 hash of the executable file (Cmd.Path)
// with which we identify a build.
func (a *App) CheckSum() string { return a.checkSum }
func (a *App) ID() string       { return a.id }
func (a *App) ExitStatus() int {
	return a.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}
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
func (app *App) Start() (err error) {
	err = app.Cmd.Start()
	if err == nil {
		log.Printf("app %v started", app.Path)
	} else {
		errorlog.Print(err)
	}
	// We will not be using our copies of any file descriptors we
	// passed on to the app. We close them to ensure that our
	// servers receive EOF when the app's copies are closed.
	for _, f := range app.ExtraFiles {
		log.Println("closing ", f.Name())
		f.Close()
	}
	return
}

// Wait wraps the underlying call to Cmd.Wait and logs the exit of app.
func (app *App) Wait() {
	app.Cmd.Wait()
	log.Printf("app %v exited", app.Path)
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
