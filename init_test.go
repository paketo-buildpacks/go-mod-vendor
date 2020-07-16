package gomodvendor_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGoModVendor(t *testing.T) {
	suite := spec.New("go-mod-vendor", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("Mod Vendor", testModVendor)
	suite.Run(t)
}
