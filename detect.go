package gomodvendor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, "go.mod"))
		if err != nil {
			if os.IsNotExist(err) {
				return packit.DetectResult{}, packit.Fail
			}
			return packit.DetectResult{}, fmt.Errorf("Failed to stat go.mod : %w", err)
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "go-dist",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
				},
			},
		}, nil
	}
}
