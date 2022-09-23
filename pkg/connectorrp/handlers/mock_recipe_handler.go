// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/project-radius/radius/pkg/connectorrp/handlers (interfaces: RecipeHandler)

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRecipeHandler is a mock of RecipeHandler interface.
type MockRecipeHandler struct {
	ctrl     *gomock.Controller
	recorder *MockRecipeHandlerMockRecorder
}

// MockRecipeHandlerMockRecorder is the mock recorder for MockRecipeHandler.
type MockRecipeHandlerMockRecorder struct {
	mock *MockRecipeHandler
}

// NewMockRecipeHandler creates a new mock instance.
func NewMockRecipeHandler(ctrl *gomock.Controller) *MockRecipeHandler {
	mock := &MockRecipeHandler{ctrl: ctrl}
	mock.recorder = &MockRecipeHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRecipeHandler) EXPECT() *MockRecipeHandlerMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockRecipeHandler) Delete(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockRecipeHandlerMockRecorder) Delete(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRecipeHandler)(nil).Delete), arg0, arg1, arg2)
}

// DeployRecipe mocks base method.
func (m *MockRecipeHandler) DeployRecipe(arg0 context.Context, arg1, arg2, arg3 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeployRecipe", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeployRecipe indicates an expected call of DeployRecipe.
func (mr *MockRecipeHandlerMockRecorder) DeployRecipe(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeployRecipe", reflect.TypeOf((*MockRecipeHandler)(nil).DeployRecipe), arg0, arg1, arg2, arg3)
}
