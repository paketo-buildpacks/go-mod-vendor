package mod_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/platform"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/paketo-buildpacks/go-mod-vendor/mod"
)

//go:generate mockgen -source=mod.go -destination=mocks_test.go -package=mod_test

const appName = "some_app_name"

func TestUnitGoMod(t *testing.T) {
	spec.Run(t, "Go Mod", testGoMod, spec.Report(report.Terminal{}))
}

func testGoMod(t *testing.T, when spec.G, it spec.S) {
	var (
		mockCtrl     *gomock.Controller
		factory      *test.BuildFactory
		mockRunner   *MockRunner
		appRoot      string
		launchPath   string
		goModLayer   layers.Layer
		goCacheLayer layers.Layer
		launchLayer  layers.Layer
		contributor  mod.Contributor
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
			factory.AddPlan(buildpackplan.Plan{Name: mod.Dependency})
			goModLayer = factory.Build.Layers.Layer(mod.Dependency)
			goCacheLayer = factory.Build.Layers.Layer(mod.Cache)
			launchLayer = factory.Build.Layers.Layer(mod.Launch)
			launchPath = filepath.Join(launchLayer.Root, "bin", appName)

			contributor = mod.NewContributor(factory.Build, mockRunner)
		})

		it.After(func() {
			if os.Getenv("BP_GO_TARGETS") != "" {
				os.Unsetenv("BP_GO_TARGETS")
			}

			Expect(factory.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Processes: []layers.Process{
					{
						Type:    "web",
						Command: launchPath,
						Direct:  false,
					},
				},
			}))

			Expect(goModLayer).To(test.HaveLayerMetadata(false, true, false))
			Expect(goCacheLayer).To(test.HaveLayerMetadata(false, true, false))
			Expect(launchLayer).To(test.HaveLayerMetadata(false, false, true))
		})

		when("The app is NOT vendored", func() {
			it("runs `go install`, gets app name and contributes the start command", func() {
				mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

				mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
				mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
				mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

				mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

				mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry").Do(func(_ ...interface{}) {
					Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
				})

				Expect(contributor.Contribute()).To(Succeed())
			})

			when("given ldflags", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(appRoot, "buildpack.yml"),
						[]byte(`---
go:
  ldflags:
    main.linker_flag: linked_flag
    main.other_linker_flag: other_linked_flag`),
						os.FileMode(0666))).To(Succeed())
				})

				it("runs `go install` with ldflags", func() {
					mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

					mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
					mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
					mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

					mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

					mockRunner.EXPECT().
						Run(
							"go", appRoot, false, "install",
							"-buildmode", "pie",
							"-tags", "cloudfoundry",
							`-ldflags=-X 'main.linker_flag=linked_flag' -X 'main.other_linker_flag=other_linked_flag'`,
						).
						Do(func(_ ...interface{}) {
							Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
						})

					Expect(contributor.Contribute()).To(Succeed())
				})
			})

			when("given ldflags and targets", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(appRoot, "buildpack.yml"),
						[]byte(`---
go:
  targets: ["./path/to/first"]
  ldflags:
    main.linker_flag: linked_flag
    main.other_linker_flag: other_linked_flag`),
						os.FileMode(0666))).To(Succeed())

					launchPath = filepath.Join(launchLayer.Root, "bin", "first")
				})

				it("runs `go install` with ldflags before targets", func() {
					mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
					mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
					mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

					mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

					mockRunner.EXPECT().
						Run(
							"go", appRoot, false, "install",
							"-buildmode", "pie",
							"-tags", "cloudfoundry",
							`-ldflags=-X 'main.linker_flag=linked_flag' -X 'main.other_linker_flag=other_linked_flag'`,
							"./path/to/first",
						).
						Do(func(_ ...interface{}) {
							Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
						})

					Expect(contributor.Contribute()).To(Succeed())
				})
			})

			when("the target is not at the root directory", func() {
				it.Before(func() {
					launchPath = filepath.Join(launchLayer.Root, "bin", "first")
				})

				when("`BP_GO_TARGETS` environment variable is set", func() {
					when("`BP_GO_TARGETS` value contains a trailing forward slash", func() {
						it("runs `go install`, gets app name and contributes the start command", func() {

							factory.Build.Platform.EnvironmentVariables = platform.EnvironmentVariables{
								"BP_GO_TARGETS": "./path/to/first/:./path/to/second/",
							}
							Expect(factory.Build.Platform.EnvironmentVariables.SetAll()).To(Succeed())

							mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
							mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
							mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

							mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

							mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "./path/to/first/", "./path/to/second/").Do(func(_ ...interface{}) {
								Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
							})

							Expect(contributor.Contribute()).To(Succeed())
						})
					})
					it("runs `go install`, gets app name and contributes the start command", func() {
						factory.Build.Platform.EnvironmentVariables = platform.EnvironmentVariables{
							"BP_GO_TARGETS": "./path/to/first:./path/to/second",
						}
						Expect(factory.Build.Platform.EnvironmentVariables.SetAll()).To(Succeed())

						mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
						mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
						mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

						mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

						mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "./path/to/first", "./path/to/second").Do(func(_ ...interface{}) {
							Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
						})

						Expect(contributor.Contribute()).To(Succeed())
					})
				})

				when("buildpack.yml config file is present", func() {
					it("runs `go install`, gets app name and contributes the start command", func() {
						Expect(ioutil.WriteFile(filepath.Join(appRoot, "buildpack.yml"),
							[]byte(`---
go:
  targets: ["./path/to/first"]`),
							os.FileMode(0666))).To(Succeed())

						mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
						mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
						mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

						mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

						mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "./path/to/first").Do(func(_ ...interface{}) {
							Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
						})

						Expect(contributor.Contribute()).To(Succeed())
					})
				})
			})
		})

		when("The app is vendored", func() {
			it("runs `go install`, gets app name and contributes the start command", func() {
				vendorDir := filepath.Join(factory.Build.Application.Root, "vendor")
				Expect(os.MkdirAll(vendorDir, 0666)).To(Succeed())
				defer os.RemoveAll(vendorDir)

				mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

				mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
				mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
				mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

				mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "-mod=vendor").Do(func(_ ...interface{}) {
					Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
				})

				Expect(contributor.Contribute()).To(Succeed())
			})
		})
	})

	when("Contribute for tiny stack", func() {
		when("the tiny id is io.paketo.stacks.tiny", func() {
			it.Before(func() {
				factory.AddPlan(buildpackplan.Plan{Name: mod.Dependency})
				factory.Build.Stack = "io.paketo.stacks.tiny"
				goModLayer = factory.Build.Layers.Layer(mod.Dependency)
				goCacheLayer = factory.Build.Layers.Layer(mod.Cache)
				launchLayer = factory.Build.Layers.Layer(mod.Launch)
				launchPath = filepath.Join(launchLayer.Root, "bin", appName)

				contributor = mod.NewContributor(factory.Build, mockRunner)
			})

			it.After(func() {
				if os.Getenv("BP_GO_TARGETS") != "" {
					os.Unsetenv("BP_GO_TARGETS")
				}

				Expect(factory.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
					Processes: []layers.Process{
						{
							Type:    "web",
							Command: launchPath,
							Direct:  true,
						},
					},
				}))

				Expect(goModLayer).To(test.HaveLayerMetadata(false, true, false))
				Expect(launchLayer).To(test.HaveLayerMetadata(false, false, true))
			})

			when("The app is NOT vendored", func() {
				it("runs `go install`, gets app name and contributes the start command", func() {
					mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

					mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
					mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
					mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

					mockRunner.EXPECT().Run("go", appRoot, false, "mod", "download")

					mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry").Do(func(_ ...interface{}) {
						Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
					})

					Expect(contributor.Contribute()).To(Succeed())
				})
			})

			when("The app is vendored", func() {
				it("runs `go install`, gets app name and contributes the start command", func() {
					vendorDir := filepath.Join(factory.Build.Application.Root, "vendor")
					Expect(os.MkdirAll(vendorDir, 0666)).To(Succeed())
					defer os.RemoveAll(vendorDir)

					mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)

					mockRunner.EXPECT().SetEnv("GOPATH", goModLayer.Root)
					mockRunner.EXPECT().SetEnv("GOCACHE", goCacheLayer.Root)
					mockRunner.EXPECT().SetEnv("GOBIN", filepath.Join(launchLayer.Root, "bin"))

					mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry", "-mod=vendor").Do(func(_ ...interface{}) {
						Expect(helper.WriteFile(launchPath, os.ModePerm, "")).To(Succeed())
					})

					Expect(contributor.Contribute()).To(Succeed())
				})
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
