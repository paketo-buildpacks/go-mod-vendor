package gomodvendor

import (
	"fmt"
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
			if os.IsNotExist(err) {
				return packit.DetectResult{}, packit.Fail
			}
			return packit.DetectResult{}, err
		}

		_, err = os.Stat(filepath.Join(context.WorkingDir, "vendor"))
		if err == nil {
			return packit.DetectResult{}, packit.Fail
		} else {
			if !os.IsNotExist(err) {
				return packit.DetectResult{}, fmt.Errorf("failed to stat vendor directory: %w", err)
			}
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
