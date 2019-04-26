// Code generated by MockGen. DO NOT EDIT.
// Source: mod.go

// Package mod_test is a generated GoMock package.
package mod_test

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRunner is a mock of Runner interface
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockRunner) Run(bin, dir string, quiet bool, args ...string) error {
	varargs := []interface{}{bin, dir, quiet}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Run", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockRunnerMockRecorder) Run(bin, dir, quiet interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{bin, dir, quiet}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockRunner)(nil).Run), varargs...)
}

// RunWithOutput mocks base method
func (m *MockRunner) RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error) {
	varargs := []interface{}{bin, dir, quiet}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RunWithOutput", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunWithOutput indicates an expected call of RunWithOutput
func (mr *MockRunnerMockRecorder) RunWithOutput(bin, dir, quiet interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{bin, dir, quiet}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunWithOutput", reflect.TypeOf((*MockRunner)(nil).RunWithOutput), varargs...)
}

// SetEnv mocks base method
func (m *MockRunner) SetEnv(variableName, path string) error {
	ret := m.ctrl.Call(m, "SetEnv", variableName, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetEnv indicates an expected call of SetEnv
func (mr *MockRunnerMockRecorder) SetEnv(variableName, path interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetEnv", reflect.TypeOf((*MockRunner)(nil).SetEnv), variableName, path)
}

// Rename mocks base method
func (m *MockRunner) Rename(existingPath, newPath string) error {
	ret := m.ctrl.Call(m, "Rename", existingPath, newPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rename indicates an expected call of Rename
func (mr *MockRunnerMockRecorder) Rename(existingPath, newPath interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rename", reflect.TypeOf((*MockRunner)(nil).Rename), existingPath, newPath)
}

// MockLogger is a mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Info mocks base method
func (m *MockLogger) Info(format string, args ...interface{}) {
	varargs := []interface{}{format}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info
func (mr *MockLoggerMockRecorder) Info(format interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{format}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), varargs...)
}

// MockPkgManager is a mock of PkgManager interface
type MockPkgManager struct {
	ctrl     *gomock.Controller
	recorder *MockPkgManagerMockRecorder
}

// MockPkgManagerMockRecorder is the mock recorder for MockPkgManager
type MockPkgManagerMockRecorder struct {
	mock *MockPkgManager
}

// NewMockPkgManager creates a new mock instance
func NewMockPkgManager(ctrl *gomock.Controller) *MockPkgManager {
	mock := &MockPkgManager{ctrl: ctrl}
	mock.recorder = &MockPkgManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPkgManager) EXPECT() *MockPkgManagerMockRecorder {
	return m.recorder
}

// Install mocks base method
func (m *MockPkgManager) Install(location, cacheDir string) error {
	ret := m.ctrl.Call(m, "Install", location, cacheDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// Install indicates an expected call of Install
func (mr *MockPkgManagerMockRecorder) Install(location, cacheDir interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Install", reflect.TypeOf((*MockPkgManager)(nil).Install), location, cacheDir)
}

// MockMetadataInterface is a mock of MetadataInterface interface
type MockMetadataInterface struct {
	ctrl     *gomock.Controller
	recorder *MockMetadataInterfaceMockRecorder
}

// MockMetadataInterfaceMockRecorder is the mock recorder for MockMetadataInterface
type MockMetadataInterfaceMockRecorder struct {
	mock *MockMetadataInterface
}

// NewMockMetadataInterface creates a new mock instance
func NewMockMetadataInterface(ctrl *gomock.Controller) *MockMetadataInterface {
	mock := &MockMetadataInterface{ctrl: ctrl}
	mock.recorder = &MockMetadataInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetadataInterface) EXPECT() *MockMetadataInterfaceMockRecorder {
	return m.recorder
}

// Identity mocks base method
func (m *MockMetadataInterface) Identity() (string, string) {
	ret := m.ctrl.Call(m, "Identity")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// Identity indicates an expected call of Identity
func (mr *MockMetadataInterfaceMockRecorder) Identity() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Identity", reflect.TypeOf((*MockMetadataInterface)(nil).Identity))
}