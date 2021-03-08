package gomodvendor

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
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
		version, err := goModParser.ParseVersion(filepath.Join(context.WorkingDir, GoModLocation))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return packit.DetectResult{}, packit.Fail.WithMessage("go.mod file is not present")
			}
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
