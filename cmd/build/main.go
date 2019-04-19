package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/go-mod-cnb/utils"

	"github.com/cloudfoundry/go-mod-cnb/mod"

	"github.com/buildpack/libbuildpack/buildplan"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	runner := utils.Command{}

	goModContributor, willContribute, err := mod.NewContributor(context, runner)
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := goModContributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildplan.BuildPlan{})
}
