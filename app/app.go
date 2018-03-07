// Package app provides a model of applications using Auklet.
package app

import (
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// An App represents an application using Auklet.
type App struct {
	// Cmd is the command to be executed.
	*exec.Cmd

	// Checksum is the SHA512/224 hash of the executable file (Cmd.Path)
	// with which we identify a build.
	CheckSum string

	// AppID is to be provided by config.Config.
	AppID string
}

// New returns an App that would execute args.
func New(args []string, appid string) (app *App) {
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	app = &App{
		Cmd:      c,
		CheckSum: sum(c.Path),
		AppID:    appid,
	}
	return
}

// IsReleased returns true if app is released according to the API at baseurl.
func (app *App) IsReleased(baseurl, apikey string) (ok bool) {
	url := baseurl + "/releases/?checksum=" + app.CheckSum
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Print(err)
	}
	req.Header.Add("Authentication", "JWT "+apikey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}
	switch resp.StatusCode {
	case 200:
		ok = true
	case 404:
		log.Printf(`app %v with checksum "%v" not released against %v`,
			app.Path, app.CheckSum, url)
		ok = false
	case 500:
		// Something is wrong with the API.
		fallthrough
	default:
		log.Printf("App.IsReleased: got unexpected status %v", resp.Status)
	}
	return
}

// Start wraps the underlying call to Cmd.Start and logs any errors.
func (app *App) Start() (err error) {
	err = app.Cmd.Start()
	if err == nil {
		log.Printf("app %v started", app.Path)
	} else {
		log.Print(err)
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
		log.Print(err)
		return
	}
	defer f.Close()
	h := sha512.New512_224()
	if _, err = io.Copy(h, f); err != nil {
		log.Print(err)
		return
	}
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}
