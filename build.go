package gomodvendor

import (
	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface BuildProcess --output fakes/build_process.go
type BuildProcess interface {
	ShouldRun(workingDir string) (bool, error)
	Execute(path, workingDir string) error
}

//go:generate faux --interface ChecksumCalculator --output fakes/checksum_calculator.go
type ChecksumCalculator interface {
	Sum(...string) (sha string, err error)
}

//go:generate faux --interface PlanRefinery --output fakes/plan_refinery.go
type PlanRefinery interface {
	BillOfMaterials(workingDir string) (packit.BuildpackPlanEntry, error)
}

func Build(buildProcess BuildProcess, planRefinery PlanRefinery, logs LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		ok, err := buildProcess.ShouldRun(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if !ok {
			logs.Process("Skipping build process: module graph is empty")
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

		bom, err := planRefinery.BillOfMaterials(context.WorkingDir)
		if err != nil {
			panic(err)
		}

		return packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{bom},
			},
			Layers: []packit.Layer{modCacheLayer},
		}, nil
	}
}
