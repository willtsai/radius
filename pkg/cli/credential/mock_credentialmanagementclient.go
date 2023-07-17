// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/project-radius/radius/pkg/cli/credential (interfaces: CredentialManagementClient)

// Package credential is a generated GoMock package.
package credential

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v20220901privatepreview "github.com/project-radius/radius/pkg/ucp/api/v20220901privatepreview"
)

// MockCredentialManagementClient is a mock of CredentialManagementClient interface.
type MockCredentialManagementClient struct {
	ctrl     *gomock.Controller
	recorder *MockCredentialManagementClientMockRecorder
}

// MockCredentialManagementClientMockRecorder is the mock recorder for MockCredentialManagementClient.
type MockCredentialManagementClientMockRecorder struct {
	mock *MockCredentialManagementClient
}

// NewMockCredentialManagementClient creates a new mock instance.
func NewMockCredentialManagementClient(ctrl *gomock.Controller) *MockCredentialManagementClient {
	mock := &MockCredentialManagementClient{ctrl: ctrl}
	mock.recorder = &MockCredentialManagementClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCredentialManagementClient) EXPECT() *MockCredentialManagementClientMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockCredentialManagementClient) Delete(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockCredentialManagementClientMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCredentialManagementClient)(nil).Delete), arg0, arg1)
}

// Get mocks base method.
func (m *MockCredentialManagementClient) Get(arg0 context.Context, arg1 string) (ProviderCredentialConfiguration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(ProviderCredentialConfiguration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCredentialManagementClientMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCredentialManagementClient)(nil).Get), arg0, arg1)
}

// List mocks base method.
func (m *MockCredentialManagementClient) List(arg0 context.Context) ([]CloudProviderStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0)
	ret0, _ := ret[0].([]CloudProviderStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockCredentialManagementClientMockRecorder) List(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockCredentialManagementClient)(nil).List), arg0)
}

// PutAWS mocks base method.
func (m *MockCredentialManagementClient) PutAWS(arg0 context.Context, arg1 v20220901privatepreview.AWSCredentialResource) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutAWS", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutAWS indicates an expected call of PutAWS.
func (mr *MockCredentialManagementClientMockRecorder) PutAWS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutAWS", reflect.TypeOf((*MockCredentialManagementClient)(nil).PutAWS), arg0, arg1)
}

// PutAzure mocks base method.
func (m *MockCredentialManagementClient) PutAzure(arg0 context.Context, arg1 v20220901privatepreview.AzureCredentialResource) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutAzure", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutAzure indicates an expected call of PutAzure.
func (mr *MockCredentialManagementClientMockRecorder) PutAzure(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutAzure", reflect.TypeOf((*MockCredentialManagementClient)(nil).PutAzure), arg0, arg1)
}