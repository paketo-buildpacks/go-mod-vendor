package gomodvendor_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGoModVendor(t *testing.T) {
	suite := spec.New("go-mod-vendor", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite.Run(t)
}
