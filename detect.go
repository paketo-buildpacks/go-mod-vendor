package gomodvendor

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Version       string `toml:"version"`
	Build         bool   `toml:"build"`
}

func Detect(goModParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		goModFilepath := filepath.Join(context.WorkingDir, GoModLocation)
		exists, err := fs.Exists(goModFilepath)
		if err != nil {
			return packit.DetectResult{}, err
		}
		if !exists {
			return packit.DetectResult{}, packit.Fail.WithMessage("go.mod file is not present")
		}
		version, err := goModParser.ParseVersion(goModFilepath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: GoLayerName,
						Metadata: BuildPlanMetadata{
							VersionSource: GoModLocation,
							Build:         true,
							Version:       version,
						},
					},
				},
			},
		}, nil
	}
}
