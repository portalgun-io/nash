package main

import (
	"fmt"
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
	gopath, ok := os.LookupEnv("GOPATH")
	if ok {
		return filepath.Join(gopath, "src", "github.com", "NeowayLabs", "nash"), nil
	}

	h, err := home()
	return filepath.Join(h, "nashroot"), err
}

func home() (string, error) {
	homedir, err := os.UserHomeDir()
	if homedir == "" {
		return "", fmt.Errorf("user[%v] has empty home dir", usr)
	}
	return homedir, err
}
