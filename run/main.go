package main

import (
	"os"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type SBOMGenerator struct{}

func (s SBOMGenerator) Generate(path string) (sbom.SBOM, error) {
	return sbom.Generate(path)
}

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	goModParser := gomodvendor.NewGoModParser()
	sbomGenerator := SBOMGenerator{}

	packit.Run(
		gomodvendor.Detect(goModParser),
		gomodvendor.Build(
			gomodvendor.NewModVendor(pexec.NewExecutable("go"), logEmitter, chronos.DefaultClock),
			logEmitter,
			chronos.DefaultClock,
			sbomGenerator,
		),
	)
}
