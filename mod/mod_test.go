package mod_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/layers"

	"github.com/buildpack/libbuildpack/buildplan"

	"github.com/cloudfoundry/libcfbuildpack/test"

	"github.com/golang/mock/gomock"
	_ "github.com/onsi/ginkgo"
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
		mockCtrl   *gomock.Controller
		factory    *test.BuildFactory
		mockRunner *MockRunner
		appRoot    string
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

	when("NewContributor", func() {
		it("returns true if it exists in the build plan", func() {
			factory.AddBuildPlan(mod.Dependency, buildplan.Dependency{})

			_, willContribute, err := mod.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			_, willContribute, err := mod.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("Contribute", func() {
		it("runs `go install`, gets app name and contributes the start command", func() {
			factory.AddBuildPlan(mod.Dependency, buildplan.Dependency{})
			layer := factory.Build.Layers.Layer(mod.Dependency)
			proc := filepath.Join(layer.Root, "bin", appName)

			contributor, willCont, err := mod.NewContributor(factory.Build, mockRunner)
			Expect(willCont).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())

			mockRunner.EXPECT().SetEnv("GOPATH", layer.Root)
			mockRunner.EXPECT().Run("go", appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry").Return(nil)
			mockRunner.EXPECT().RunWithOutput("go", appRoot, false, "list", "-m").Return(appName, nil)
			Expect(contributor.Contribute()).To(Succeed())

			Expect(factory.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", proc}}}))

			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
		})
	})
}
