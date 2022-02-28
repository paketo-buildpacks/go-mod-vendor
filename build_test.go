package gomodvendor_test

import (
	"bytes"
	"errors"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"os"
	"path/filepath"
	"testing"
	"time"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/go-mod-vendor/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir     string
		workingDir    string
		logs          *bytes.Buffer
		buildProcess  *fakes.BuildProcess
		sbomGenerator *fakes.SBOMGenerator
		clock         chronos.Clock

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers-dir")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		logs = bytes.NewBuffer(nil)

		buildProcess = &fakes.BuildProcess{}
		buildProcess.ShouldRunCall.Returns.Ok = true

		now := time.Now()
		clock = chronos.NewClock(func() time.Time {
			return now
		})

		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateCall.Returns.SBOM = sbom.SBOM{}

		build = gomodvendor.Build(
			buildProcess,
			scribe.NewEmitter(logs),
			clock,
			sbomGenerator,
		)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("builds", func() {
		result, err := build(packit.BuildContext{
			Layers:     packit.Layers{Path: layersDir},
			WorkingDir: workingDir,
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{"application/vnd.cyclonedx+json", "application/spdx+json", "application/vnd.syft+json"},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Layers[0].Name).To(Equal("mod-cache"))
		Expect(result.Layers[0].Path).To(Equal(filepath.Join(layersDir, "mod-cache")))
		Expect(result.Layers[0].SharedEnv).To(Equal(packit.Environment{}))
		Expect(result.Layers[0].BuildEnv).To(Equal(packit.Environment{}))
		Expect(result.Layers[0].LaunchEnv).To(Equal(packit.Environment{}))
		Expect(result.Layers[0].ProcessLaunchEnv).To(Equal(map[string]packit.Environment{}))
		Expect(result.Layers[0].Build).To(BeFalse())
		Expect(result.Layers[0].Launch).To(BeTrue())
		Expect(result.Layers[0].Cache).To(BeTrue())

		Expect(result.Layers[0].SBOM.Formats()).To(Equal([]packit.SBOMFormat{
			{
				Extension: "cdx.json",
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.CycloneDXFormat),
			},
			{
				Extension: "spdx.json",
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SPDXFormat),
			},
			{
				Extension: "syft.json",
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SyftFormat),
			},
		}))

		Expect(buildProcess.ShouldRunCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(buildProcess.ExecuteCall.Receives.Path).To(Equal(filepath.Join(layersDir, "mod-cache")))
		Expect(buildProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(logs.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(logs.String()).NotTo(ContainSubstring("Skipping build process: module graph is empty"))
	})

	context("when there are no modules in go.mod", func() {
		it.Before(func() {
			buildProcess.ShouldRunCall.Returns.Ok = false
			buildProcess.ShouldRunCall.Returns.Reason = "module graph is empty"
		})

		it("does not include the module cache layer in the build result", func() {
			result, err := build(packit.BuildContext{
				Layers:     packit.Layers{Path: layersDir},
				WorkingDir: workingDir,
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{}))

			Expect(logs.String()).To(ContainSubstring("Skipping build process: module graph is empty"))
		})
	})

	context("failure cases", func() {
		context("build process fails to check if it should run", func() {
			it.Before(func() {
				buildProcess.ShouldRunCall.Returns.Err = errors.New("build process failed to check")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers:     packit.Layers{Path: layersDir},
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("build process failed to check")))
			})
		})

		context("modCacheLayer cannot be retrieved", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layersDir, "mod-cache.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers:     packit.Layers{Path: layersDir},
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("build process fails to execute", func() {
			it.Before(func() {
				buildProcess.ExecuteCall.Stub = nil
				buildProcess.ExecuteCall.Returns.Error = errors.New("build process failed to execute")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers:     packit.Layers{Path: layersDir},
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("build process failed to execute")))
			})
		})
		context("when the BOM cannot be generated", func() {
			it.Before(func() {
				sbomGenerator.GenerateCall.Returns.Error = errors.New("failed to generate SBOM")
			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						SBOMFormats: []string{"application/vnd.cyclonedx+json", "application/spdx+json", "application/vnd.syft+json"},
					},
					WorkingDir: workingDir,
					Layers:     packit.Layers{Path: layersDir},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{{Name: "node_modules"}},
					},
					Stack: "some-stack",
				})
				Expect(err).To(MatchError("failed to generate SBOM"))
			})
		})

		context("when the BOM cannot be formatted", func() {
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					BuildpackInfo: packit.BuildpackInfo{
						SBOMFormats: []string{"random-format"},
					},
				})
				Expect(err).To(MatchError("\"random-format\" is not a supported SBOM format"))
			})
		})
	})
}
