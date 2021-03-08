package gomodvendor_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/go-mod-vendor/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		goModParser *fakes.VersionParser
		detect      packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
		goModParser = &fakes.VersionParser{}

		detect = gomodvendor.Detect(goModParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it.Before(func() {
		goModParser.ParseVersionCall.Returns.Version = ">= 1.15"
	})

	it("detects", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Requires: []packit.BuildPlanRequirement{
				{
					Name: "go",
					Metadata: gomodvendor.BuildPlanMetadata{
						VersionSource: "go.mod",
						Version:       ">= 1.15",
						Build:         true,
					},
				},
			},
		}))
	})

	context("go.mod does not exist in the working directory", func() {
		it.Before(func() {
			_, err := os.Stat("/no/such/go.mod")
			goModParser.ParseVersionCall.Returns.Err = fmt.Errorf("failed to parse go.mod: %w", err)
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("go.mod file is not present")))
		})
	})

	context("when there is a vendor directory in the working directory", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(workingDir, "vendor"), os.ModePerm)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "go",
						Metadata: gomodvendor.BuildPlanMetadata{
							VersionSource: "go.mod",
							Version:       ">= 1.15",
							Build:         true,
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("the go.mod file cannot be read", func() {
			it.Before(func() {
				goModParser.ParseVersionCall.Returns.Err = errors.New("failed to read go.mod file")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("failed to read go.mod file"))
			})
		})
	})
}
