package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
}

func (r Command) Run(bin, dir string, quiet bool, args ...string) error {
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

func (r Command) RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error) {
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
