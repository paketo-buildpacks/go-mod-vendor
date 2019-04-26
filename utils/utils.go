package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Command struct {
}

func (c Command) Run(bin, dir string, quiet bool, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if quiet {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func (c Command) RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error) {
	logs := &bytes.Buffer{}

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if quiet {
		cmd.Stdout = io.MultiWriter(ioutil.Discard, logs)
		cmd.Stderr = io.MultiWriter(ioutil.Discard, logs)
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, logs)
		cmd.Stderr = io.MultiWriter(os.Stderr, logs)
	}
	err := cmd.Run()

	return strings.TrimSpace(logs.String()), err
}

func (c Command) SetEnv(variableName string, path string) error {
	return os.Setenv(variableName, path)
}

func (c Command) Rename(existingPath string, newPath string) error {
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, os.ModePerm); err != nil {
		return err
	}
	return os.Rename(existingPath, newPath)
}
