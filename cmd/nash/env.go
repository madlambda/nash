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
	h, err := home()
	return filepath.Join(h, "nash"), err
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
	
	h, err := home()
	return filepath.Join(h, "nashroot"), err
}

func home() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if usr.HomeDir == "" {
		return "", fmt.Errorf("user[%v] has empty home dir", usr)
	}
	return usr.HomeDir, nil
}