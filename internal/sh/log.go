package sh

import (
	"log"
	"os"
)

// LogFn is the logger type
type LogFn func(format string, args ...interface{})

// NewLog creates a new nash logger
func NewLog(ns string, enable bool) LogFn {
	logger := log.New(os.Stderr, "", 0)

	return func(format string, args ...interface{}) {
		if enable {
			logger.Printf("["+ns+"] "+format, args...)
		}
	}
}
