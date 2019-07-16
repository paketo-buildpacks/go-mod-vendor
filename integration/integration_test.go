package integration

import (
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
)

var (
	bpDir, goURI, goModURI string
)

func TestIntegration(t *testing.T) {
	var err error
	Expect := NewWithT(t).Expect
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())
	goModURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(goModURI)

	goURI, err = dagger.GetLatestBuildpack("go-compiler-cnb")
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(goURI)

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var Expect func(interface{}, ...interface{}) GomegaAssertion
	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	const (
		goFinding     = "go: finding github.com/"
		goDownloading = "go: downloading github.com/"
		goExtracting  = "go: extracting github.com/"
	)

	it("should build a working OCI image for a simple app", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), goURI, goModURI)
		Expect(err).ToNot(HaveOccurred())
		defer app.Destroy()

		Expect(app.Start()).To(Succeed())

		_, _, err = app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
	})

	when("the app is pushed twice", func() {
		it("does not reinstall go modules", func() {
			appDir := filepath.Join("testdata", "simple_app")
			app, err := dagger.PackBuild(appDir, goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.BuildLogs()).To(MatchRegexp(goFinding))
			Expect(app.BuildLogs()).To(MatchRegexp(goDownloading))
			Expect(app.BuildLogs()).To(MatchRegexp(goExtracting))

			_, imageID, _, err := app.Info()
			Expect(err).NotTo(HaveOccurred())

			app, err = dagger.PackBuildNamedImage(imageID, appDir, goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			repeatBuildLogs := app.BuildLogs()
			Expect(repeatBuildLogs).NotTo(MatchRegexp(goFinding))
			Expect(repeatBuildLogs).NotTo(MatchRegexp(goDownloading))
			Expect(repeatBuildLogs).NotTo(MatchRegexp(goExtracting))

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	when("the app is vendored", func() {
		it("builds an OCI image without downloading any extra packages", func() {
			appDir := filepath.Join("testdata", "vendored")
			app, err := dagger.PackBuild(appDir, goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.BuildLogs()).NotTo(MatchRegexp(goDownloading))

			Expect(app.Start()).To(Succeed())
			_, _, err = app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	when("the app target is not at the root of the directory", func() {
		it("should build a working OCI image for a simple app", func() {
			app, err := dagger.PackBuild(filepath.Join("testdata", "non_root_target"), goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})

}
