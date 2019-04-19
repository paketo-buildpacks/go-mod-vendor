package main

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/google/go-cmp/cmp"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {
	var factory *test.BuildFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
	})

	it("always passes", func() {
		code, err := runBuild(factory.Build)
		if err != nil {
			t.Error("Err in build : ", err)
		}

		if diff := cmp.Diff(code, build.SuccessStatusCode); diff != "" {
			t.Error("Problem : ", diff)
		}
	})
}
