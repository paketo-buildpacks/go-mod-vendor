package main

import (
	"os"

	"github.com/cloudfoundry/packit/pexec"
	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
)

func main() {
	logEmitter := gomodvendor.NewLogEmitter(os.Stdout)
	goModParser := gomodvendor.NewGoModParser()
	packit.Run(
		gomodvendor.Detect(goModParser),
		gomodvendor.Build(
			gomodvendor.NewModVendor(pexec.NewExecutable("go"), logEmitter, chronos.DefaultClock),
			logEmitter,
		),
	)
}
