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
		packageBuildpack string
		goBuildpack      string
	)

	it.Before(func() {
		RegisterTestingT(t)

		var err error

		packageBuildpack, err = dagger.PackageBuildpack()
		Expect(err).ToNot(HaveOccurred())

		goBuildpack, err = dagger.GetLatestBuildpack("go-cnb")
		Expect(err).ToNot(HaveOccurred())
	})

	it("should build a working OCI image for a simple app", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), goBuildpack, packageBuildpack)
		Expect(err).ToNot(HaveOccurred())
		defer app.Destroy()

		Expect(app.Start()).To(Succeed())

		_, _, err = app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
	})

	when("the app is pushed twice", func() {
		it("does not reinstall go modules", func() {
			appDir := filepath.Join("testdata", "simple_app")
			app, err := dagger.PackBuild(appDir, goBuildpack, packageBuildpack)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.BuildLogs()).To(MatchRegexp("go: finding github.com/"))
			Expect(app.BuildLogs()).To(MatchRegexp("go: downloading github.com/"))
			Expect(app.BuildLogs()).To(MatchRegexp("go: extracting github.com/"))

			_, imageID, _, err := app.Info()
			Expect(err).NotTo(HaveOccurred())

			app, err = dagger.PackBuildNamedImage(imageID, appDir, goBuildpack, packageBuildpack)
			Expect(err).ToNot(HaveOccurred())

			repeatBuildLogs := app.BuildLogs()
			Expect(repeatBuildLogs).NotTo(MatchRegexp("go: finding github.com/"))
			Expect(repeatBuildLogs).NotTo(MatchRegexp("go: downloading github.com/"))
			Expect(repeatBuildLogs).NotTo(MatchRegexp("go: extracting github.com/"))

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
