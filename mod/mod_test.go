package mod_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/helper"

	"github.com/cloudfoundry/libcfbuildpack/layers"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/platform"

	"github.com/cloudfoundry/libcfbuildpack/test"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/go-mod-cnb/mod"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=mod.go -destination=mocks_test.go -package=mod_test

const appName = "some_app_name"

func TestUnitGoMod(t *testing.T) {
	spec.Run(t, "Go Mod", testGoMod, spec.Report(report.Terminal{}))
}

func testGoMod(t *testing.T, when spec.G, it spec.S) {
	var (
		mockCtrl    *gomock.Controller
		factory     *test.BuildFactory
		mockRunner  *MockRunner
		appRoot     string
		buildPath   string
		launchPath  string
		goModLayer  layers.Layer
		launchLayer layers.Layer
		contributor mod.Contributor
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = NewMockRunner(mockCtrl)
		appRoot = factory.Build.Application.Root
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("Contribute", func() {
		it.Before(func() {
			factory.AddBuildPlan(mod.Dependency, buildplan.Dependency{})
			goModLayer = factory.Build.Layers.Layer(mod.Dependency)
			launchLayer = factory.Build.Layers.Layer(mod.Launch)
			buildPath = filepath.Join(goModLayer.Root, "bin", appName)
			launchPath = filepath.Join(launchLayer.Root, appName)

			contributor = mod.NewContributor(factory.Build, mockRunner)
		})

		it.After(func() {
			if os.Getenv("BP_GO_MOD_TARGETS") != "" {
				os.Unsetenv("BP_GO_MOD_TARGETS")
			}

			Expect(factory.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", launchPath}}}))

			Expect(goModLayer).To(test.HaveLayerMetadata(false, true, false))
			Expect(launchLayer).To(test.HaveLayerMetadata(false, false, true))
		})

		when("The app is NOT vendored", func() {
			it("runs `go install`, gets app name and contributes the start command", func() {
				mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

				mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)

				mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry").Do(func(_ ...interface{}) {
					Expect(helper.WriteFile(buildPath, os.ModePerm, "")).To(Succeed())
				})

				Expect(contributor.Contribute()).To(Succeed())
			})

			when("the target is not at the root directory", func() {
				when("`BP_GO_MOD_TARGETS` environment variable is set", func() {
					it("runs `go install`, gets app name and contributes the start command", func() {
						factory.Build.Platform.EnvironmentVariables = platform.EnvironmentVariables{
							"BP_GO_MOD_TARGETS": "./path/to/first:./path/to/second",
						}
						factory.Build.Platform.EnvironmentVariables.SetAll()
						mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

						mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)

						mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "./path/to/first", "./path/to/second").Do(func(_ ...interface{}) {
							Expect(helper.WriteFile(buildPath, os.ModePerm, "")).To(Succeed())
						})

						Expect(contributor.Contribute()).To(Succeed())
					})
				})

			})
		})

		when("The app is vendored", func() {
			it("runs `go install`, gets app name and contributes the start command", func() {
				vendorDir := filepath.Join(factory.Build.Application.Root, "vendor")
				os.MkdirAll(vendorDir, 0666)
				defer os.RemoveAll(vendorDir)

				mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

				mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)

				mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "-mod=vendor").Do(func(_ ...interface{}) {
					Expect(helper.WriteFile(buildPath, os.ModePerm, "")).To(Succeed())
				})

				Expect(contributor.Contribute()).To(Succeed())
			})
		})
	})

	when("Cleanup", func() {
		it("removes the contents of the app dir", func() {
			dummyFile := filepath.Join(factory.Build.Application.Root, "dummy")
			Expect(ioutil.WriteFile(dummyFile, []byte("baller"), 0777))

			contributor = mod.NewContributor(factory.Build, mockRunner)

			Expect(contributor.Cleanup()).To(Succeed())
			appDirContents, err := ioutil.ReadDir(factory.Build.Application.Root)
			Expect(err).NotTo(HaveOccurred())
			Expect(appDirContents).To(BeEmpty())
		})
	})
}
