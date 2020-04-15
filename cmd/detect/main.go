package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/paketo-buildpacks/go-mod/mod"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

const GoDependency = "go"

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	goModFile := filepath.Join(context.Application.Root, "go.mod")
	if exists, err := helper.FileExists(goModFile); err != nil {
		return detect.FailStatusCode, errors.Wrap(err, fmt.Sprintf("error checking filepath: %s", goModFile))
	} else if !exists {
		return detect.FailStatusCode, fmt.Errorf(`no "go.mod" found at: %s`, goModFile)
	}

	return context.Pass(buildplan.Plan{
		Provides: []buildplan.Provided{{Name: mod.Dependency}},
		Requires: []buildplan.Required{{
			Name: mod.Dependency,
			Metadata: buildplan.Metadata{
				"build": true,
			},
		}, {
			Name: GoDependency,
		}},
	})

}
