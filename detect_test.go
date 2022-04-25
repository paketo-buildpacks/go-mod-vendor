package gomodvendor_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/go-mod-vendor/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		goModParser *fakes.VersionParser

		detect        packit.DetectFunc
		detectContext packit.DetectContext
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		goModParser = &fakes.VersionParser{}

		Expect(os.WriteFile(filepath.Join(workingDir, "go.mod"), []byte{}, os.ModePerm)).To(Succeed())

		detect = gomodvendor.Detect(goModParser)

		detectContext = packit.DetectContext{WorkingDir: workingDir}
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it.Before(func() {
		goModParser.ParseVersionCall.Returns.Version = ">= 1.15"
	})

	it("detects", func() {
		result, err := detect(detectContext)
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
			Expect(os.Remove(filepath.Join(workingDir, "go.mod"))).To(Succeed())
		})

		it("fails detection", func() {
			_, err := detect(detectContext)
			Expect(err).To(MatchError(packit.Fail.WithMessage("go.mod file is not present")))
		})
	})

	context("when there is a vendor directory in the working directory", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(workingDir, "vendor"), os.ModePerm)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(detectContext)
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
		context("there is an error determining if the go.mod file exists", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(detectContext)
				Expect(err).To(HaveOccurred())
			})
		})

		context("the go.mod file cannot be read", func() {
			it.Before(func() {
				goModParser.ParseVersionCall.Returns.Err = errors.New("some error")
			})

			it("returns an error", func() {
				_, err := detect(detectContext)
				Expect(err).To(MatchError("some error"))
			})
		})
	})
}
