package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/cloudfoundry/go-mod-cnb/utils"

	"github.com/cloudfoundry/go-mod-cnb/mod"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

type GoModContributor interface {
	Contribute() error
	Cleanup() error
}

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	runner := utils.Command{}
	var goModContributor GoModContributor = mod.NewContributor(context, runner)
	code, err := runBuild(context, goModContributor)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runBuild(context build.Build, goModContributor GoModContributor) (int, error) {
	context.Logger.Title(context.Buildpack)

	_, wantDependency, err := context.Plans.GetShallowMerged(mod.Dependency)
	if err != nil {
		return context.Failure(105), err
	}

	if !wantDependency {
		return context.Failure(102), nil
	}

	if err := goModContributor.Contribute(); err != nil {
		return context.Failure(103), err
	}

	if err := goModContributor.Cleanup(); err != nil {
		return context.Failure(104), err
	}

	return context.Success(buildpackplan.Plan{})
}
