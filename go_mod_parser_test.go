package gomodvendor_test

import (
	"os"
	"testing"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGoModParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser gomodvendor.GoModParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "go.mod")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`module github.com/some-org/some-repo

go 1.15

require (
	github.com/some/dependency v0.3.1
	github.com/some-other/dependency v0.0.4
)
`)
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		parser = gomodvendor.NewGoModParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("ParseVersion", func() {
		it("parses the go version from a go.mod file", func() {
			version, err := parser.ParseVersion(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(">= 1.15"))
		})

		context("failure cases", func() {
			context("when the go.mod cannot be opened", func() {
				it.Before(func() {
					Expect(os.Chmod(path, 0000)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := parser.ParseVersion(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse go.mod:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
