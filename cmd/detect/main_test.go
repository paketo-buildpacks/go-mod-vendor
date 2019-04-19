package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/go-mod-cnb/mod"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("there is a go.mod", func() {
		it("should add go-mod to the buildplan", func() {
			goModString := fmt.Sprintf("This is a go mod file")
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "go.mod"), goModString)

			plan := buildplan.BuildPlan{
				mod.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"build": true},
				},
			}
			runDetectAndExpectBuildplan(factory, plan)
		})
	})

	when("there is no go.mod", func() {
		it("should fail", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

}

func runDetectAndExpectBuildplan(factory *test.DetectFactory, buildplan buildplan.BuildPlan) {
	code, err := runDetect(factory.Detect)
	Expect(err).NotTo(HaveOccurred())

	Expect(code).To(Equal(detect.PassStatusCode))

	Expect(factory.Output).To(Equal(buildplan))
}
