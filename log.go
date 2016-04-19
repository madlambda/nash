package cnt

import (
	"log"
	"os"
)

type LogFn func(format string, args ...interface{})

func NewLog(ns string, enable bool) LogFn {
	logger := log.New(os.Stderr, "", 0)

	return func(format string, args ...interface{}) {
		if enable {
			logger.Printf("["+ns+"] "+format, args...)
		}
	}
}
