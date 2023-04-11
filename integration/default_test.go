package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose().WithNoColor()
		docker = occam.NewDocker()
	})

	context("when building a simple go mod app", func() {
		var (
			image     occam.Image
			container occam.Container

			name    string
			source  string
			sbomDir string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			sbomDir, err = os.MkdirTemp("", "sbom")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())

			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(os.RemoveAll(sbomDir)).To(Succeed())
		})

		it("builds successfully", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.GoDist.Online,
					settings.Buildpacks.GoModVendor.Online,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL": "DEBUG",
				}).
				WithSBOMOutputDir(sbomDir).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Checking module graph",
				"    Running 'go mod graph'",
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Executing build process",
				"    Running 'go mod vendor'",
				MatchRegexp(`      go: downloading github.com/BurntSushi/toml v.+`),
				MatchRegexp(`      go: downloading github.com/satori/go.uuid v.+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Generating SBOM for /workspace/go.mod",
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Writing SBOM in the following format(s):",
				"    application/vnd.cyclonedx+json",
				"    application/spdx+json",
				"    application/vnd.syft+json",
			))

			container, err = docker.Container.Run.
				WithCommand("ls -alR /workspace").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() fmt.Stringer {
				logs, _ = docker.Container.Logs.Execute(container.ID)
				return logs
			}).Should(SatisfyAll(
				ContainSubstring("go.sum"),
				ContainSubstring("vendor/github.com/BurntSushi"),
				ContainSubstring("vendor/github.com/satori"),
			))

			// check that all required SBOM files are present
			Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "sbom.cdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "sbom.spdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "sbom.syft.json")).To(BeARegularFile())

			// check an SBOM file to make sure it contains the expected modules
			contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "build", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "sbom.cdx.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`"name": "github.com/BurntSushi/toml"`))
			Expect(string(contents)).To(ContainSubstring(`"name": "github.com/satori/go.uuid"`))
		})
	})
}
