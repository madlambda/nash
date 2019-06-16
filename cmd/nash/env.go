package main

import (
	"errors"
	"os"
	"path/filepath"
)

func NashPath() (string, error) {
	nashpath := os.Getenv("NASHPATH")
	if nashpath != "" {
		return nashpath, nil
	}
	h, err := home()
	return filepath.Join(h, "nash"), err
}

func NashRoot() (string, error) {
	nashroot, ok := os.LookupEnv("NASHROOT")
	if ok {
		return nashroot, nil
	}

	h, err := home()
	return filepath.Join(h, "nashroot"), err
}

func home() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if homedir == "" {
		return "", errors.New("invalid empty home dir")
	}
	return homedir, nil
}
