package gomodvendor_test

import (
	"bytes"
	"errors"
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
		layersDir, err = ioutil.TempDir("", "layers-dir")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		logs = bytes.NewBuffer(nil)

		buildProcess = &fakes.BuildProcess{}

		build = gomodvendor.Build(
			buildProcess,
			gomodvendor.NewLogEmitter(logs),
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
					Name:      "mod-cache",
					Path:      filepath.Join(layersDir, "mod-cache"),
					SharedEnv: packit.Environment{},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    false,
					Cache:     true,
					Metadata:  nil,
				},
			},
		}))

		Expect(buildProcess.ExecuteCall.Receives.Path).To(Equal(filepath.Join(layersDir, "mod-cache")))
		Expect(buildProcess.ExecuteCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(logs.String()).To(ContainSubstring("Some Buildpack some-version"))
	})

	context("failure cases", func() {
		context("modCacheLayer cannot be retrieved", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(filepath.Join(layersDir, "mod-cache.toml"), nil, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers:     packit.Layers{Path: layersDir},
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("build process fails", func() {
			it.Before(func() {
				buildProcess.ExecuteCall.Returns.Error = errors.New("build process failed")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers:     packit.Layers{Path: layersDir},
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("build process failed")))
			})
		})
	})
}
