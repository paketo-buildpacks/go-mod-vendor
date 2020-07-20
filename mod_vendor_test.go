package gomodvendor_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cloudfoundry/packit/pexec"
	gomodvendor "github.com/paketo-buildpacks/go-mod-vendor"
	"github.com/paketo-buildpacks/go-mod-vendor/fakes"
	"github.com/paketo-buildpacks/packit/chronos"
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
		logs         *bytes.Buffer

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

		logs = bytes.NewBuffer(nil)

		now := time.Now()
		times := []time.Time{now, now.Add(1 * time.Second)}

		clock := chronos.NewClock(func() time.Time {
			if len(times) == 0 {
				return time.Now()
			}

			t := times[0]
			times = times[1:]
			return t
		})

		modVendor = gomodvendor.NewModVendor(executable, gomodvendor.NewLogEmitter(logs), clock)
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

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))
			Expect(logs.String()).To(ContainSubstring("    Running 'go mod vendor'"))
			Expect(logs.String()).To(ContainSubstring("      Completed in 1s"))
		})

		context("failure cases", func() {
			context("the executable fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "build error stdout")
						fmt.Fprintln(execution.Stderr, "build error stderr")

						return errors.New("executable failed")
					}
				})

				it("returns an error", func() {
					err := modVendor.Execute(modCachePath, workingDir)
					Expect(err).To(MatchError(ContainSubstring("executable failed")))
					Expect(logs.String()).To(ContainSubstring("      Failed after 1s"))
					Expect(logs.String()).To(ContainSubstring("        build error stdout"))
					Expect(logs.String()).To(ContainSubstring("        build error stderr"))
				})
			})
		})
	})
}
