// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/diogox/bspc-go (interfaces: Client)

// Package feature is a generated GoMock package.
package feature

import (
	bspc "github.com/diogox/bspc-go"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockClient) Query(arg0 string, arg1 bspc.QueryResponseResolver) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Query indicates an expected call of Query
func (mr *MockClientMockRecorder) Query(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockClient)(nil).Query), arg0, arg1)
}

// SubscribeEvents mocks base method
func (m *MockClient) SubscribeEvents(arg0 bspc.EventType, arg1 ...bspc.EventType) (chan bspc.Event, chan error, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SubscribeEvents", varargs...)
	ret0, _ := ret[0].(chan bspc.Event)
	ret1, _ := ret[1].(chan error)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SubscribeEvents indicates an expected call of SubscribeEvents
func (mr *MockClientMockRecorder) SubscribeEvents(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeEvents", reflect.TypeOf((*MockClient)(nil).SubscribeEvents), varargs...)
}
