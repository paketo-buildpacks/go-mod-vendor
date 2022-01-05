package gomodvendor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type ModVendor struct {
	executable Executable
	logs       scribe.Emitter
	clock      chronos.Clock
}

func NewModVendor(executable Executable, logs scribe.Emitter, clock chronos.Clock) ModVendor {
	return ModVendor{
		executable: executable,
		logs:       logs,
		clock:      clock,
	}
}

func (m ModVendor) ShouldRun(workingDir string) (bool, string, error) {
	ok, err := m.hasVendorDirectory(workingDir)
	if err != nil {
		return false, "", err
	}
	if ok {
		return false, "modules are already vendored", nil
	}

	ok, err = m.hasModuleGraph(workingDir)
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, "module graph is empty", nil
	}

	return true, "", nil
}

func (m ModVendor) hasVendorDirectory(workingDir string) (bool, error) {
	_, err := os.Stat(filepath.Join(workingDir, "vendor"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (m ModVendor) hasModuleGraph(workingDir string) (bool, error) {
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
