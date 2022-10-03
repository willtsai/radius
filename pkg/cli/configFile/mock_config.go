// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/project-radius/radius/pkg/cli/configFile (interfaces: Interface)

// Package configFile is a generated GoMock package.
package configFile

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	output "github.com/project-radius/radius/pkg/cli/output"
	workspaces "github.com/project-radius/radius/pkg/cli/workspaces"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// EditWorkspacesByName mocks base method.
func (m *MockInterface) EditWorkspacesByName(arg0 context.Context, arg1, arg2, arg3 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EditWorkspacesByName", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// EditWorkspacesByName indicates an expected call of EditWorkspacesByName.
func (mr *MockInterfaceMockRecorder) EditWorkspacesByName(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EditWorkspacesByName", reflect.TypeOf((*MockInterface)(nil).EditWorkspacesByName), arg0, arg1, arg2, arg3)
}

// ShowWorkspace mocks base method.
func (m *MockInterface) ShowWorkspace(arg0 output.Interface, arg1 string, arg2 *workspaces.Workspace) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShowWorkspace", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ShowWorkspace indicates an expected call of ShowWorkspace.
func (mr *MockInterfaceMockRecorder) ShowWorkspace(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShowWorkspace", reflect.TypeOf((*MockInterface)(nil).ShowWorkspace), arg0, arg1, arg2)
}

// UpdateWorkspaces mocks base method.
func (m *MockInterface) UpdateWorkspaces(arg0 context.Context, arg1 string, arg2 *workspaces.Workspace) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateWorkspaces", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateWorkspaces indicates an expected call of UpdateWorkspaces.
func (mr *MockInterfaceMockRecorder) UpdateWorkspaces(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateWorkspaces", reflect.TypeOf((*MockInterface)(nil).UpdateWorkspaces), arg0, arg1, arg2)
}
