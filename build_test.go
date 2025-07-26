package gomodvendor_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"

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

		Expect(os.WriteFile(filepath.Join(workingDir, "go.mod"), nil, os.ModePerm))

		logs = bytes.NewBuffer(nil)

		buildProcess = &fakes.BuildProcess{}
		buildProcess.ShouldRunCall.Returns.Ok = true

		err = os.MkdirAll(filepath.Join(layersDir, "mod-cache"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(filepath.Join(layersDir, "mod-cache"), "cache"), nil, os.ModePerm))

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
				SBOMFormats: []string{"application/vnd.cyclonedx+json", "application/spdx+json"},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("mod-cache"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "mod-cache")))

		Expect(result.Build.SBOM.Formats()).To(HaveLen(2))
		cdx := result.Build.SBOM.Formats()[0]
		spdx := result.Build.SBOM.Formats()[1]

		Expect(cdx.Extension).To(Equal("cdx.json"))
		content, err := io.ReadAll(cdx.Content)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(MatchJSON(`{
			"$schema": "http://cyclonedx.org/schema/bom-1.3.schema.json",
			"bomFormat": "CycloneDX",
			"metadata": {
				"tools": [
					{
						"name": "",
						"vendor": "anchore"
					}
				]
			},
			"specVersion": "1.3",
			"version": 1
		}`))

		Expect(spdx.Extension).To(Equal("spdx.json"))
		content, err = io.ReadAll(spdx.Content)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(MatchJSON(`{
			"SPDXID": "SPDXRef-DOCUMENT",
			"creationInfo": {
				"created": "0001-01-01T00:00:00Z",
				"creators": [
					"Organization: Anchore, Inc",
					"Tool: -"
				],
				"licenseListVersion": "3.25"
			},
			"dataLicense": "CC0-1.0",
			"documentNamespace": "https://paketo.io/unknown-source-type/unknown-9ecf240a-d971-5a3c-8e7b-6d3f3ea4d9c2",
			"name": "unknown",
			"packages": [
				{
					"SPDXID": "SPDXRef-DocumentRoot-Unknown-",
					"copyrightText": "NOASSERTION",
					"downloadLocation": "NOASSERTION",
					"filesAnalyzed": false,
					"licenseConcluded": "NOASSERTION",
					"licenseDeclared": "NOASSERTION",
					"name": "",
					"supplier": "NOASSERTION"
				}
			],
			"relationships": [
				{
					"relatedSpdxElement": "SPDXRef-DocumentRoot-Unknown-",
					"relationshipType": "DESCRIBES",
					"spdxElementId": "SPDXRef-DOCUMENT"
				}
			],
			"spdxVersion": "SPDX-2.2"
		}`))

		Expect(buildProcess.ShouldRunCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(buildProcess.ExecuteCall.Receives.Path).To(Equal(filepath.Join(layersDir, "mod-cache")))
		Expect(buildProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(sbomGenerator.GenerateCall.Receives.Dir).To(Equal(filepath.Join(workingDir, "go.mod")))

		Expect(logs.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(logs.String()).NotTo(ContainSubstring("Skipping build process: module graph is empty"))
	})

	context("when the mod cache layer does not exist", func() {
		it.Before(func() {
			err := os.RemoveAll(filepath.Join(layersDir, "mod-cache"))
			Expect(err).NotTo(HaveOccurred())
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

			Expect(result.Layers).To(BeEmpty())
		})
	})

	context("when the mod cache layer is empty", func() {
		it.Before(func() {
			err := os.RemoveAll(filepath.Join(layersDir, "mod-cache", "cache"))
			Expect(err).NotTo(HaveOccurred())
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

			Expect(result.Layers).To(BeEmpty())
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
		context("when checking for go.mod fails", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
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
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})

		})
		context("when go.mod is missing", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "go.mod"))).To(Succeed())
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
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("failed to generate SBOM: '%s' does not exist", filepath.Join(workingDir, "go.mod")))))
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
				Expect(err).To(MatchError("unsupported SBOM format: 'random-format'"))
			})
		})
	})
}
