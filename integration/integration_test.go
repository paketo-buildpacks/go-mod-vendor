package integration

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

			buildLogs := &bytes.Buffer{}

			// TODO: Move this to dagger

			_, imageID, _, err := app.Info()
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("pack", "build", imageID, "--builder", "cfbuildpacks/cflinuxfs3-cnb-test-builder", "--buildpack", goBuildpack, "--buildpack", packageBuildpack)
			cmd.Dir = appDir
			cmd.Stdout = io.MultiWriter(os.Stdout, buildLogs)
			cmd.Stderr = io.MultiWriter(os.Stderr, buildLogs)
			Expect(cmd.Run()).To(Succeed())

			const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

			re := regexp.MustCompile(ansi)
			strippedLogs := re.ReplaceAllString(buildLogs.String(), "")

			Expect(strippedLogs).NotTo(MatchRegexp("go: finding github.com/"))
			Expect(strippedLogs).NotTo(MatchRegexp("go: downloading github.com/"))
			Expect(strippedLogs).NotTo(MatchRegexp("go: extracting github.com/"))

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
