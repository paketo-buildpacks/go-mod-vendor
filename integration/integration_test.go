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

func BeforeSuite() {
	var err error
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())
	goModURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())

	goURI, err = dagger.GetLatestCommunityBuildpack("paketo-buildpacks", "go-compiler")
	Expect(err).ToNot(HaveOccurred())
}

func AfterSuite() {
	Expect(dagger.DeleteBuildpack(goModURI)).To(Succeed())
	Expect(dagger.DeleteBuildpack(goURI)).To(Succeed())
}

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	BeforeSuite()
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}), spec.Parallel())
	AfterSuite()
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect func(interface{}, ...interface{}) GomegaAssertion
		err    error
		app    *dagger.App
	)

	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	it.After(func() {
		if app != nil {
			Expect(app.Destroy()).To(Succeed())
		}
	})

	const (
		goFinding     = "go: finding github.com/"
		goDownloading = "go: downloading github.com/"
		goExtracting  = "go: extracting github.com/"
	)

	it("should build a working OCI image for a simple app", func() {
		app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app"), goURI, goModURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.Start()).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(MatchRegexp("PATH=.*/layers/paketo-buildpacks_go-mod/app-binary/bin:"))
	})

	when("the app is pushed twice", func() {
		var (
			firstApp  *dagger.App
			secondApp *dagger.App
		)

		it.After(func() {
			if firstApp != nil {
				Expect(firstApp.Destroy()).To(Succeed())
			}

			if secondApp != nil {
				Expect(secondApp.Destroy()).To(Succeed())
			}
		})

		it("does not reinstall go modules", func() {
			var err error
			appDir := filepath.Join("testdata", "simple_app")

			firstApp, err = dagger.PackBuild(appDir, goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			Expect(firstApp.BuildLogs()).To(MatchRegexp(goFinding), firstApp.BuildLogs())

			_, imageID, _, err := firstApp.Info()
			Expect(err).NotTo(HaveOccurred())

			secondApp, err = dagger.PackBuildNamedImage(imageID, appDir, goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			repeatBuildLogs := secondApp.BuildLogs()
			Expect(repeatBuildLogs).NotTo(MatchRegexp(goFinding))
			Expect(repeatBuildLogs).To(ContainSubstring(`Adding cache layer 'paketo-buildpacks/go-mod:go-cache'`))

			Expect(secondApp.Start()).To(Succeed())

			_, _, err = secondApp.HTTPGet("/")
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

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	when("the app has multiple targets", func() {
		it("should build a working OCI image for a simple app with all target binaries contributed", func() {
			app, err = dagger.PackBuild(filepath.Join("testdata", "multiple_targets"), goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.Start()).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("Hello From Helper!"))
		})
	})

	when("the app specifies ldflags", func() {
		it("should build the app with those build flags", func() {
			app, err := dagger.PackBuild(filepath.Join("testdata", "ldflags"), goURI, goModURI)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.Start()).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("main.version: v1.2.3"))
			Expect(body).To(ContainSubstring("main.sha: 7a82056"))
		})
	})

	when("the app does not build to a complete executable", func() {
		it("build should fail with a descriptive error", func() {
			_, err := dagger.PackBuild(filepath.Join("testdata", "nomain_app"), goURI, goModURI)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("`go install` failed to install executable(s) in /layers/paketo-buildpacks_go-mod/app-binary/bin"))
		})
	})
}
