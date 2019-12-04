// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/NiR-/webdf/pkg/registry (interfaces: KindHandler)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	builddef "github.com/NiR-/webdf/pkg/builddef"
	image "github.com/NiR-/webdf/pkg/image"
	pkgsolver "github.com/NiR-/webdf/pkg/pkgsolver"
	gomock "github.com/golang/mock/gomock"
	llb "github.com/moby/buildkit/client/llb"
	client "github.com/moby/buildkit/frontend/gateway/client"
	reflect "reflect"
)

// MockKindHandler is a mock of KindHandler interface
type MockKindHandler struct {
	ctrl     *gomock.Controller
	recorder *MockKindHandlerMockRecorder
}

// MockKindHandlerMockRecorder is the mock recorder for MockKindHandler
type MockKindHandlerMockRecorder struct {
	mock *MockKindHandler
}

// NewMockKindHandler creates a new mock instance
func NewMockKindHandler(ctrl *gomock.Controller) *MockKindHandler {
	mock := &MockKindHandler{ctrl: ctrl}
	mock.recorder = &MockKindHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockKindHandler) EXPECT() *MockKindHandlerMockRecorder {
	return m.recorder
}

// Build mocks base method
func (m *MockKindHandler) Build(arg0 context.Context, arg1 client.Client, arg2 builddef.BuildOpts) (llb.State, *image.Image, error) {
	ret := m.ctrl.Call(m, "Build", arg0, arg1, arg2)
	ret0, _ := ret[0].(llb.State)
	ret1, _ := ret[1].(*image.Image)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Build indicates an expected call of Build
func (mr *MockKindHandlerMockRecorder) Build(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockKindHandler)(nil).Build), arg0, arg1, arg2)
}

// DebugLLB mocks base method
func (m *MockKindHandler) DebugLLB(arg0 builddef.BuildOpts) (llb.State, error) {
	ret := m.ctrl.Call(m, "DebugLLB", arg0)
	ret0, _ := ret[0].(llb.State)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DebugLLB indicates an expected call of DebugLLB
func (mr *MockKindHandlerMockRecorder) DebugLLB(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DebugLLB", reflect.TypeOf((*MockKindHandler)(nil).DebugLLB), arg0)
}

// UpdateLocks mocks base method
func (m *MockKindHandler) UpdateLocks(arg0 *builddef.BuildDef, arg1 pkgsolver.PackageSolver) (builddef.Locks, error) {
	ret := m.ctrl.Call(m, "UpdateLocks", arg0, arg1)
	ret0, _ := ret[0].(builddef.Locks)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateLocks indicates an expected call of UpdateLocks
func (mr *MockKindHandlerMockRecorder) UpdateLocks(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLocks", reflect.TypeOf((*MockKindHandler)(nil).UpdateLocks), arg0, arg1)
}
