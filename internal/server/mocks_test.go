// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/qdm12/dns/v2/internal/server (interfaces: Filter,Cache,Logger)

// Package server is a generated GoMock package.
package server

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dns "github.com/miekg/dns"
	update "github.com/qdm12/dns/v2/pkg/filter/update"
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

// Update mocks base method.
func (m *MockFilter) Update(arg0 update.Settings) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Update", arg0)
}

// Update indicates an expected call of Update.
func (mr *MockFilterMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockFilter)(nil).Update), arg0)
}

// MockCache is a mock of Cache interface.
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *MockCacheMockRecorder
}

// MockCacheMockRecorder is the mock recorder for MockCache.
type MockCacheMockRecorder struct {
	mock *MockCache
}

// NewMockCache creates a new mock instance.
func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &MockCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCache) EXPECT() *MockCacheMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockCache) Add(arg0, arg1 *dns.Msg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", arg0, arg1)
}

// Add indicates an expected call of Add.
func (mr *MockCacheMockRecorder) Add(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockCache)(nil).Add), arg0, arg1)
}

// Get mocks base method.
func (m *MockCache) Get(arg0 *dns.Msg) *dns.Msg {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*dns.Msg)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCacheMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCache)(nil).Get), arg0)
}

// Remove mocks base method.
func (m *MockCache) Remove(arg0 *dns.Msg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Remove", arg0)
}

// Remove indicates an expected call of Remove.
func (mr *MockCacheMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockCache)(nil).Remove), arg0)
}

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockLogger) Debug(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerMockRecorder) Debug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), arg0)
}

// Error mocks base method.
func (m *MockLogger) Error(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerMockRecorder) Error(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), arg0)
}

// Info mocks base method.
func (m *MockLogger) Info(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0)
}

// Info indicates an expected call of Info.
func (mr *MockLoggerMockRecorder) Info(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), arg0)
}

// Warn mocks base method.
func (m *MockLogger) Warn(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerMockRecorder) Warn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogger)(nil).Warn), arg0)
}
