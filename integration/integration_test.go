package integration

import (
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		app *dagger.App
		err error
		packageBuildpack string
		goBuildpack      string
	)

	const (
		goFinding = "go: finding github.com/"
		goDownloading = "go: downloading github.com/"
		goExtracting = "go: extracting github.com/"
	)

	it.Before(func() {
		RegisterTestingT(t)

		packageBuildpack, err = dagger.PackageBuildpack()
		Expect(err).ToNot(HaveOccurred())

		goBuildpack, err = dagger.GetLatestBuildpack("go-cnb")
		Expect(err).ToNot(HaveOccurred())
	})

	it.After(func() {
		app.Destroy()
	})

	it("should build a working OCI image for a simple app", func() {
		app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app"), goBuildpack, packageBuildpack)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.Start()).To(Succeed())

		_, _, err = app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
	})

	when("the app is pushed twice", func() {
		it("does not reinstall go modules", func() {
			appDir := filepath.Join("testdata", "simple_app")
			app, err = dagger.PackBuild(appDir, goBuildpack, packageBuildpack)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.BuildLogs()).To(MatchRegexp(goFinding))
			Expect(app.BuildLogs()).To(MatchRegexp(goDownloading))
			Expect(app.BuildLogs()).To(MatchRegexp(goExtracting))

			_, imageID, _, err := app.Info()
			Expect(err).NotTo(HaveOccurred())

			app, err = dagger.PackBuildNamedImage(imageID, appDir, goBuildpack, packageBuildpack)
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
			app, err = dagger.PackBuild(appDir, goBuildpack, packageBuildpack)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.BuildLogs()).NotTo(MatchRegexp(goDownloading))

			Expect(app.Start()).To(Succeed())
			_, _, err = app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())
		})
	})
}
