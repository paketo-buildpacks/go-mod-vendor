package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
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
			image      occam.Image
			container1 occam.Container
			container2 occam.Container
			container3 occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container1.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container3.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
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
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Generating SBOM",
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			container1, err = docker.Container.Run.
				WithCommand("ls -alR /workspace").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() fmt.Stringer {
				logs, _ = docker.Container.Logs.Execute(container1.ID)
				return logs
			}).Should(SatisfyAll(
				ContainSubstring("go.sum"),
				ContainSubstring("vendor/github.com/BurntSushi"),
				ContainSubstring("vendor/github.com/satori"),
			))

			sbomFile := "/layers/sbom/launch/paketo-buildpacks_go-mod-vendor/mod-cache/sbom.cdx.json"
			container2, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("cat %s", sbomFile)).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() fmt.Stringer {
				logs, _ = docker.Container.Logs.Execute(container2.ID)
				return logs
			}).Should(SatisfyAll(
				ContainSubstring(`"bomFormat": "CycloneDX"`),
			))

			// check that all expected SBOM files are present
			sbomDir := "/layers/sbom/launch/paketo-buildpacks_go-mod-vendor/mod-cache/"
			container3, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("ls -al %s", sbomDir)).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() fmt.Stringer {
				logs, _ = docker.Container.Logs.Execute(container3.ID)
				return logs
			}).Should(And(
				ContainSubstring("sbom.cdx.json"),
				ContainSubstring("sbom.spdx.json"),
				ContainSubstring("sbom.syft.json"),
			))
		})
	})
}
