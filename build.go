package gomodvendor

import (
	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface BuildProcess --output fakes/build_process.go
type BuildProcess interface {
	Execute(path, workingDir string) error
}

func Build(
	buildProcess BuildProcess,
	logs LogEmitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		modCacheLayer, err := context.Layers.Get("mod-cache", packit.CacheLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.Process("Executing build process")

		err = buildProcess.Execute(modCacheLayer.Path, context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   context.Plan,
			Layers: []packit.Layer{modCacheLayer},
		}, nil
	}
}
