package main

import (
	"os"

	"github.com/cloudfoundry/packit/pexec"
	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	logEmitter := gomodvendor.NewLogEmitter(os.Stdout)
	packit.Run(
		gomodvendor.Detect(),
		gomodvendor.Build(
			gomodvendor.NewModVendor(pexec.NewExecutable("go")),
			logEmitter,
		),
	)
}
