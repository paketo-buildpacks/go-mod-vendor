package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
)

type PlanRefinery struct {
	BillOfMaterialsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			BuildpackPlanEntry packit.BuildpackPlanEntry
			Error              error
		}
		Stub func(string) (packit.BuildpackPlanEntry, error)
	}
}

func (f *PlanRefinery) BillOfMaterials(param1 string) (packit.BuildpackPlanEntry, error) {
	f.BillOfMaterialsCall.Lock()
	defer f.BillOfMaterialsCall.Unlock()
	f.BillOfMaterialsCall.CallCount++
	f.BillOfMaterialsCall.Receives.WorkingDir = param1
	if f.BillOfMaterialsCall.Stub != nil {
		return f.BillOfMaterialsCall.Stub(param1)
	}
	return f.BillOfMaterialsCall.Returns.BuildpackPlanEntry, f.BillOfMaterialsCall.Returns.Error
}
