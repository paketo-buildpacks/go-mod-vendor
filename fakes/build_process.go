package fakes

import "sync"

type BuildProcess struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path       string
			WorkingDir string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
	ShouldRunCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			WorkingDir string
		}
		Returns struct {
			Ok     bool
			Reason string
			Err    error
		}
		Stub func(string) (bool, string, error)
	}
}

func (f *BuildProcess) Execute(param1 string, param2 string) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Path = param1
	f.ExecuteCall.Receives.WorkingDir = param2
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2)
	}
	return f.ExecuteCall.Returns.Error
}
func (f *BuildProcess) ShouldRun(param1 string) (bool, string, error) {
	f.ShouldRunCall.mutex.Lock()
	defer f.ShouldRunCall.mutex.Unlock()
	f.ShouldRunCall.CallCount++
	f.ShouldRunCall.Receives.WorkingDir = param1
	if f.ShouldRunCall.Stub != nil {
		return f.ShouldRunCall.Stub(param1)
	}
	return f.ShouldRunCall.Returns.Ok, f.ShouldRunCall.Returns.Reason, f.ShouldRunCall.Returns.Err
}
