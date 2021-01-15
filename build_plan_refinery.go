package gomodvendor

import (
	"bytes"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
)

type BuildPlanRefinery struct {
	executable         Executable
	logs               LogEmitter
	clock              chronos.Clock
	checksumCalculator ChecksumCalculator
}

func NewBuildPlanRefinery(executable Executable, logs LogEmitter, clock chronos.Clock, checksumCalculator ChecksumCalculator) BuildPlanRefinery {
	return BuildPlanRefinery{
		executable:         executable,
		logs:               logs,
		clock:              clock,
		checksumCalculator: checksumCalculator,
	}
}

func (r BuildPlanRefinery) BillOfMaterials(workingDir string) (packit.BuildpackPlanEntry, error) {

	buffer := bytes.NewBuffer(nil)
	args := []string{"mod", "graph"}

	r.logs.Process("Adding go modules to Bill of Materials")
	r.logs.Subprocess("Running 'go %s'", strings.Join(args, " "))

	duration, err := r.clock.Measure(func() error {
		return r.executable.Execute(pexec.Execution{
			Args:   args,
			Dir:    workingDir,
			Stdout: buffer,
			Stderr: buffer,
		})
	})

	if err != nil {
		r.logs.Action("Failed after %s", duration.Round(time.Millisecond))
		r.logs.Detail(buffer.String())

		return packit.BuildpackPlanEntry{}, err
	}

	r.logs.Action("Completed in %s", duration.Round(time.Millisecond))
	r.logs.Break()

	rawModuleStrings := strings.Split(buffer.String(), "\n")

	var modules []map[string]string
	for _, module := range rawModuleStrings {
		if module == "" {
			continue
		}
		// From go help mod graph:  Each line in the output has two space-separated
		// fields: a module and one of its requirements.
		moduleWithCommitish := strings.Split(module, " ")[1]

		moduleName := strings.Split(moduleWithCommitish, "@")[0]
		commitish := strings.Split(moduleWithCommitish, "@")[1]
		moduleStructure := map[string]string{
			"module":  moduleName,
			"version": commitish,
		}
		modules = append(modules, moduleStructure)
	}

	checksum, err := r.checksumCalculator.Sum(filepath.Join(workingDir, "go.mod"))
	if err != nil {
		return packit.BuildpackPlanEntry{}, err
	}

	return packit.BuildpackPlanEntry{
		Name: "go-mod",
		Metadata: map[string]interface{}{
			"go.mod sha256": checksum,
			"modules":       modules,
		},
	}, nil
}
