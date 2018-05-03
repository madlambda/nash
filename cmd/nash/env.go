package main

import (
	"os"
	"fmt"
	"os/user"
	"path/filepath"
)

func NashPath() (string, error) {
	nashpath := os.Getenv("NASHPATH")
	if nashpath != "" {
		return nashpath, nil
	}
	
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error[%s] getting current user (no NASHPATH env var set)", err)
	}
	if usr.HomeDir == "" {
		return "", fmt.Errorf("user[%v] has an empty home dir (no NASHPATH env var set)", err)
	}
	return filepath.Join(usr.HomeDir, "nash"), nil
}

func NashRoot() (string, error) {
	nashroot, ok := os.LookupEnv("NASHROOT")
	if ok {
		return nashroot, nil
	}
	gopath, ok := os.LookupEnv("GOPATH")
	if ok {
		return filepath.Join(gopath, "src", "github.com", "NeowayLabs", "nash"), nil
	}
	return "", nil
}