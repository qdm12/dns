// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/qdm12/dns/v2/pkg/doh (interfaces: Metrics,Logger)

// Package doh is a generated GoMock package.
package doh

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMetrics is a mock of Metrics interface.
type MockMetrics struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsMockRecorder
}

// MockMetricsMockRecorder is the mock recorder for MockMetrics.
type MockMetricsMockRecorder struct {
	mock *MockMetrics
}

// NewMockMetrics creates a new mock instance.
func NewMockMetrics(ctrl *gomock.Controller) *MockMetrics {
	mock := &MockMetrics{ctrl: ctrl}
	mock.recorder = &MockMetricsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetrics) EXPECT() *MockMetricsMockRecorder {
	return m.recorder
}

// DNSDialInc mocks base method.
func (m *MockMetrics) DNSDialInc(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DNSDialInc", arg0, arg1)
}

// DNSDialInc indicates an expected call of DNSDialInc.
func (mr *MockMetricsMockRecorder) DNSDialInc(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DNSDialInc", reflect.TypeOf((*MockMetrics)(nil).DNSDialInc), arg0, arg1)
}

// DoHDialInc mocks base method.
func (m *MockMetrics) DoHDialInc(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DoHDialInc", arg0)
}

// DoHDialInc indicates an expected call of DoHDialInc.
func (mr *MockMetricsMockRecorder) DoHDialInc(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoHDialInc", reflect.TypeOf((*MockMetrics)(nil).DoHDialInc), arg0)
}

// DoTDialInc mocks base method.
func (m *MockMetrics) DoTDialInc(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DoTDialInc", arg0, arg1, arg2)
}

// DoTDialInc indicates an expected call of DoTDialInc.
func (mr *MockMetricsMockRecorder) DoTDialInc(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoTDialInc", reflect.TypeOf((*MockMetrics)(nil).DoTDialInc), arg0, arg1, arg2)
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
