package fakes

import "sync"

type ChecksumCalculator struct {
	SumCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			StringSlice []string
		}
		Returns struct {
			Sha string
			Err error
		}
		Stub func(...string) (string, error)
	}
}

func (f *ChecksumCalculator) Sum(param1 ...string) (string, error) {
	f.SumCall.Lock()
	defer f.SumCall.Unlock()
	f.SumCall.CallCount++
	f.SumCall.Receives.StringSlice = param1
	if f.SumCall.Stub != nil {
		return f.SumCall.Stub(param1...)
	}
	return f.SumCall.Returns.Sha, f.SumCall.Returns.Err
}
