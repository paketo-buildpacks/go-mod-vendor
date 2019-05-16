package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"

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
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	_, wantDependency := context.BuildPlan[mod.Dependency]
	if !wantDependency {
		return context.Failure(102), nil
	}

	if err := goModContributor.Contribute(); err != nil {
		return context.Failure(103), err
	}

	if err := goModContributor.Cleanup(); err != nil {
		return context.Failure(104), err
	}

	return context.Success(buildplan.BuildPlan{})
}
