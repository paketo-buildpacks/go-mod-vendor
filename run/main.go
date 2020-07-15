package main

import (
	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	packit.Run(
		gomodvendor.Detect(),
		gomodvendor.Build(),
	)
}
