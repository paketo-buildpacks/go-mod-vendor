package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"

	"github.com/golang/mock/gomock"

	"github.com/cloudfoundry/go-mod-cnb/mod"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=main.go -destination=build_mocks.go -package=main

func TestUnitBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {
	var factory *test.BuildFactory
	var mockCtrl *gomock.Controller
	var mockGoModContributor *MockGoModContributor

	it.Before(func() {
		RegisterTestingT(t)
		mockCtrl = gomock.NewController(t)
		factory = test.NewBuildFactory(t)
		mockGoModContributor = NewMockGoModContributor(mockCtrl)
	})

	it("passes if it exists in the build plan", func() {
		factory.AddPlan(buildpackplan.Plan{
			Name: mod.Dependency,
		})
		mockGoModContributor.EXPECT().Contribute()
		mockGoModContributor.EXPECT().Cleanup()

		code, err := runBuild(factory.Build, mockGoModContributor)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(build.SuccessStatusCode))
	})

	it("fails false if a build plan does not exist", func() {
		code, err := runBuild(factory.Build, mockGoModContributor)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).NotTo(Equal(build.SuccessStatusCode))
	})

	it("fails false if go-mod is not in the build plan", func() {
		factory.AddPlan(buildpackplan.Plan{
			Name: "foo",
		})

		code, err := runBuild(factory.Build, mockGoModContributor)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).NotTo(Equal(build.SuccessStatusCode))
	})
}
