// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/project-radius/radius/pkg/ucp/aws (interfaces: AWSCloudControlClient)

// Package aws is a generated GoMock package.
package aws

import (
	context "context"
	reflect "reflect"

	cloudcontrol "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	gomock "github.com/golang/mock/gomock"
)

// MockAWSCloudControlClient is a mock of AWSCloudControlClient interface.
type MockAWSCloudControlClient struct {
	ctrl     *gomock.Controller
	recorder *MockAWSCloudControlClientMockRecorder
}

// MockAWSCloudControlClientMockRecorder is the mock recorder for MockAWSCloudControlClient.
type MockAWSCloudControlClientMockRecorder struct {
	mock *MockAWSCloudControlClient
}

// NewMockAWSCloudControlClient creates a new mock instance.
func NewMockAWSCloudControlClient(ctrl *gomock.Controller) *MockAWSCloudControlClient {
	mock := &MockAWSCloudControlClient{ctrl: ctrl}
	mock.recorder = &MockAWSCloudControlClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAWSCloudControlClient) EXPECT() *MockAWSCloudControlClientMockRecorder {
	return m.recorder
}

// CancelResourceRequest mocks base method.
func (m *MockAWSCloudControlClient) CancelResourceRequest(arg0 context.Context, arg1 *cloudcontrol.CancelResourceRequestInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.CancelResourceRequestOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CancelResourceRequest", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.CancelResourceRequestOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CancelResourceRequest indicates an expected call of CancelResourceRequest.
func (mr *MockAWSCloudControlClientMockRecorder) CancelResourceRequest(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelResourceRequest", reflect.TypeOf((*MockAWSCloudControlClient)(nil).CancelResourceRequest), varargs...)
}

// CreateResource mocks base method.
func (m *MockAWSCloudControlClient) CreateResource(arg0 context.Context, arg1 *cloudcontrol.CreateResourceInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.CreateResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateResource", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.CreateResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateResource indicates an expected call of CreateResource.
func (mr *MockAWSCloudControlClientMockRecorder) CreateResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResource", reflect.TypeOf((*MockAWSCloudControlClient)(nil).CreateResource), varargs...)
}

// DeleteResource mocks base method.
func (m *MockAWSCloudControlClient) DeleteResource(arg0 context.Context, arg1 *cloudcontrol.DeleteResourceInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.DeleteResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteResource", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.DeleteResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteResource indicates an expected call of DeleteResource.
func (mr *MockAWSCloudControlClientMockRecorder) DeleteResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResource", reflect.TypeOf((*MockAWSCloudControlClient)(nil).DeleteResource), varargs...)
}

// GetResource mocks base method.
func (m *MockAWSCloudControlClient) GetResource(arg0 context.Context, arg1 *cloudcontrol.GetResourceInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.GetResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetResource", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.GetResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResource indicates an expected call of GetResource.
func (mr *MockAWSCloudControlClientMockRecorder) GetResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResource", reflect.TypeOf((*MockAWSCloudControlClient)(nil).GetResource), varargs...)
}

// GetResourceRequestStatus mocks base method.
func (m *MockAWSCloudControlClient) GetResourceRequestStatus(arg0 context.Context, arg1 *cloudcontrol.GetResourceRequestStatusInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.GetResourceRequestStatusOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetResourceRequestStatus", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.GetResourceRequestStatusOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResourceRequestStatus indicates an expected call of GetResourceRequestStatus.
func (mr *MockAWSCloudControlClientMockRecorder) GetResourceRequestStatus(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceRequestStatus", reflect.TypeOf((*MockAWSCloudControlClient)(nil).GetResourceRequestStatus), varargs...)
}

// ListResourceRequests mocks base method.
func (m *MockAWSCloudControlClient) ListResourceRequests(arg0 context.Context, arg1 *cloudcontrol.ListResourceRequestsInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.ListResourceRequestsOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListResourceRequests", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.ListResourceRequestsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListResourceRequests indicates an expected call of ListResourceRequests.
func (mr *MockAWSCloudControlClientMockRecorder) ListResourceRequests(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListResourceRequests", reflect.TypeOf((*MockAWSCloudControlClient)(nil).ListResourceRequests), varargs...)
}

// ListResources mocks base method.
func (m *MockAWSCloudControlClient) ListResources(arg0 context.Context, arg1 *cloudcontrol.ListResourcesInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.ListResourcesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListResources", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.ListResourcesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListResources indicates an expected call of ListResources.
func (mr *MockAWSCloudControlClientMockRecorder) ListResources(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListResources", reflect.TypeOf((*MockAWSCloudControlClient)(nil).ListResources), varargs...)
}

// UpdateResource mocks base method.
func (m *MockAWSCloudControlClient) UpdateResource(arg0 context.Context, arg1 *cloudcontrol.UpdateResourceInput, arg2 ...func(*cloudcontrol.Options)) (*cloudcontrol.UpdateResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateResource", varargs...)
	ret0, _ := ret[0].(*cloudcontrol.UpdateResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateResource indicates an expected call of UpdateResource.
func (mr *MockAWSCloudControlClientMockRecorder) UpdateResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateResource", reflect.TypeOf((*MockAWSCloudControlClient)(nil).UpdateResource), varargs...)
}