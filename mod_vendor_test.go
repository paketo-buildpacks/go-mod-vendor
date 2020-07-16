package gomodvendor_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/go-mod-vendor/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testModVendor(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		modCachePath string
		workingDir   string
		environment  []string
		executable   *fakes.Executable

		modVendor gomodvendor.ModVendor
	)

	it.Before(func() {
		var err error
		modCachePath, err = ioutil.TempDir("", "mod-cache")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-directory")
		Expect(err).NotTo(HaveOccurred())

		environment = os.Environ()
		executable = &fakes.Executable{}

		modVendor = gomodvendor.NewModVendor(executable)
	})

	it.After(func() {
		Expect(os.RemoveAll(modCachePath)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Execute", func() {
		it("runs go mod vendor", func() {
			err := modVendor.Execute(modCachePath, workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"mod", "vendor"}))
			Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(environment, fmt.Sprintf("GOPATH=%s", modCachePath))))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		})

		context("failure cases", func() {
			context("the executable fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("executable failed")
				})

				it("returns an error", func() {
					err := modVendor.Execute(modCachePath, workingDir)
					Expect(err).To(MatchError(ContainSubstring("executable failed")))
				})
			})
		})
	})
}
