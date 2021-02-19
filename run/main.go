package main

import (
	"os"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout)
	goModParser := gomodvendor.NewGoModParser()
	packit.Run(
		gomodvendor.Detect(goModParser),
		gomodvendor.Build(
			gomodvendor.NewModVendor(pexec.NewExecutable("go"), logEmitter, chronos.DefaultClock),
			logEmitter,
		),
	)
}
