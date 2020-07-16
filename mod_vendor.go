package gomodvendor

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type ModVendor struct {
	executable Executable
}

func NewModVendor(executable Executable) ModVendor {
	return ModVendor{
		executable: executable,
	}
}

func (m ModVendor) Execute(path, workingDir string) error {
	err := m.executable.Execute(pexec.Execution{
		Args: []string{"mod", "vendor"},
		Env:  append(os.Environ(), fmt.Sprintf("GOPATH=%s", path)),
		Dir:  workingDir,
	})
	if err != nil {
		return err
	}

	return nil
}
