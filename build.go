package gomodvendor

import (
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	Generate(dir string) (sbom.SBOM, error)
}

//go:generate faux --interface BuildProcess --output fakes/build_process.go
type BuildProcess interface {
	ShouldRun(workingDir string) (ok bool, reason string, err error)
	Execute(path, workingDir string) error
}

func Build(buildProcess BuildProcess, logs scribe.Emitter, clock chronos.Clock, sbomGenerator SBOMGenerator) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		ok, reason, err := buildProcess.ShouldRun(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if !ok {
			logs.Process("Skipping build process: %s", reason)
			logs.Break()

			return packit.BuildResult{}, nil
		}

		modCacheLayer, err := context.Layers.Get("mod-cache")
		if err != nil {
			return packit.BuildResult{}, err
		}

		modCacheLayer.Cache = true

		err = buildProcess.Execute(modCacheLayer.Path, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.GeneratingSBOM(context.WorkingDir)

		var sbomContent sbom.SBOM
		duration, err := clock.Measure(func() error {
			sbomContent, err = sbomGenerator.Generate(context.WorkingDir)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logs.Action("Completed in %s", duration.Round(time.Millisecond))
		logs.Break()

		logs.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)

		var buildMetadata packit.BuildMetadata
		buildMetadata.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   context.Plan,
			Layers: []packit.Layer{modCacheLayer},
			Build:  buildMetadata,
		}, nil
	}
}
