package gomodvendor_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

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

		layersDir    string
		workingDir   string
		logs         *bytes.Buffer
		buildProcess *fakes.BuildProcess

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

		build = gomodvendor.Build(
			buildProcess,
			scribe.NewEmitter(logs),
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
				Name:    "Some Buildpack",
				Version: "some-version",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Layers: []packit.Layer{
				{
					Name:             "mod-cache",
					Path:             filepath.Join(layersDir, "mod-cache"),
					SharedEnv:        packit.Environment{},
					BuildEnv:         packit.Environment{},
					LaunchEnv:        packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
					Build:            false,
					Launch:           false,
					Cache:            true,
					Metadata:         nil,
				},
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
	})
}
