// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/qdm12/dns/pkg/filter (interfaces: Filter)

// Package mock_filter is a generated GoMock package.
package mock_filter

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dns "github.com/miekg/dns"
)

// MockFilter is a mock of Filter interface.
type MockFilter struct {
	ctrl     *gomock.Controller
	recorder *MockFilterMockRecorder
}

// MockFilterMockRecorder is the mock recorder for MockFilter.
type MockFilterMockRecorder struct {
	mock *MockFilter
}

// NewMockFilter creates a new mock instance.
func NewMockFilter(ctrl *gomock.Controller) *MockFilter {
	mock := &MockFilter{ctrl: ctrl}
	mock.recorder = &MockFilterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFilter) EXPECT() *MockFilterMockRecorder {
	return m.recorder
}

// FilterRequest mocks base method.
func (m *MockFilter) FilterRequest(arg0 *dns.Msg) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FilterRequest", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// FilterRequest indicates an expected call of FilterRequest.
func (mr *MockFilterMockRecorder) FilterRequest(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FilterRequest", reflect.TypeOf((*MockFilter)(nil).FilterRequest), arg0)
}

// FilterResponse mocks base method.
func (m *MockFilter) FilterResponse(arg0 *dns.Msg) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FilterResponse", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// FilterResponse indicates an expected call of FilterResponse.
func (mr *MockFilterMockRecorder) FilterResponse(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FilterResponse", reflect.TypeOf((*MockFilter)(nil).FilterResponse), arg0)
}