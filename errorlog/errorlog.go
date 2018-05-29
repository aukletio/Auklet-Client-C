// Package errorlog provides a logger for error messages.
package errorlog

import (
	"io"
	"log"
	"os"
)

var errorLogger = log.New(os.Stdout, "", log.Lmicroseconds|log.LstdFlags)

func Print(v ...interface{}) {
	errorLogger.Print(v...)
}

func Println(v ...interface{}) {
	errorLogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

func SetOutput(w io.Writer) {
	errorLogger.SetOutput(w)
}
