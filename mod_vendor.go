package gomodvendor

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/packit/pexec"
	"github.com/paketo-buildpacks/packit/chronos"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type ModVendor struct {
	executable Executable
	logs       LogEmitter
	clock      chronos.Clock
}

func NewModVendor(executable Executable, logs LogEmitter, clock chronos.Clock) ModVendor {
	return ModVendor{
		executable: executable,
		logs:       logs,
		clock:      clock,
	}
}

func (m ModVendor) ShouldRun(workingDir string) (bool, error) {
	buffer := bytes.NewBuffer(nil)
	args := []string{"mod", "graph"}

	m.logs.Process("Checking module graph")
	m.logs.Subprocess("Running 'go %s'", strings.Join(args, " "))

	duration, err := m.clock.Measure(func() error {
		return m.executable.Execute(pexec.Execution{
			Args:   args,
			Dir:    workingDir,
			Stdout: buffer,
			Stderr: buffer,
		})
	})
	if err != nil {
		m.logs.Action("Failed after %s", duration.Round(time.Millisecond))
		m.logs.Detail(buffer.String())

		return false, err
	}

	m.logs.Action("Completed in %s", duration.Round(time.Millisecond))
	m.logs.Break()

	return buffer.Len() > 0, nil
}

func (m ModVendor) Execute(path, workingDir string) error {
	buffer := bytes.NewBuffer(nil)
	args := []string{"mod", "vendor"}

	m.logs.Process("Executing build process")
	m.logs.Subprocess("Running 'go %s'", strings.Join(args, " "))

	duration, err := m.clock.Measure(func() error {
		return m.executable.Execute(pexec.Execution{
			Args:   args,
			Env:    append(os.Environ(), fmt.Sprintf("GOPATH=%s", path)),
			Dir:    workingDir,
			Stdout: buffer,
			Stderr: buffer,
		})
	})
	if err != nil {
		m.logs.Action("Failed after %s", duration.Round(time.Millisecond))
		m.logs.Detail(buffer.String())

		return err
	}

	m.logs.Action("Completed in %s", duration.Round(time.Millisecond))
	m.logs.Break()

	return nil
}
