// Package errorlog provides a logger for error messages.
package errorlog

import (
	"io"
	"log"
	"os"
)

var errorLogger = log.New(os.Stdout, "", log.Lmicroseconds|log.LstdFlags)

// Print prints to the logger as in the manner of fmt.Print.
func Print(v ...interface{}) {
	errorLogger.Print(v...)
}

// Print prints to the logger as in the manner of fmt.Println.
func Println(v ...interface{}) {
	errorLogger.Println(v...)
}

// Print prints to the logger as in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// SetOutput sets the output destination for the logger.
func SetOutput(w io.Writer) {
	errorLogger.SetOutput(w)
}
