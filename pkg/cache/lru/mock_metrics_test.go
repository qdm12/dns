// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/qdm12/dns/v2/pkg/cache/metrics (interfaces: Interface)

// Package lru is a generated GoMock package.
package lru

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMetrics is a mock of Interface interface.
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

// CacheExpiredInc mocks base method.
func (m *MockMetrics) CacheExpiredInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheExpiredInc")
}

// CacheExpiredInc indicates an expected call of CacheExpiredInc.
func (mr *MockMetricsMockRecorder) CacheExpiredInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheExpiredInc", reflect.TypeOf((*MockMetrics)(nil).CacheExpiredInc))
}

// CacheGetEmptyInc mocks base method.
func (m *MockMetrics) CacheGetEmptyInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheGetEmptyInc")
}

// CacheGetEmptyInc indicates an expected call of CacheGetEmptyInc.
func (mr *MockMetricsMockRecorder) CacheGetEmptyInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheGetEmptyInc", reflect.TypeOf((*MockMetrics)(nil).CacheGetEmptyInc))
}

// CacheHitInc mocks base method.
func (m *MockMetrics) CacheHitInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheHitInc")
}

// CacheHitInc indicates an expected call of CacheHitInc.
func (mr *MockMetricsMockRecorder) CacheHitInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheHitInc", reflect.TypeOf((*MockMetrics)(nil).CacheHitInc))
}

// CacheInsertEmptyInc mocks base method.
func (m *MockMetrics) CacheInsertEmptyInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheInsertEmptyInc")
}

// CacheInsertEmptyInc indicates an expected call of CacheInsertEmptyInc.
func (mr *MockMetricsMockRecorder) CacheInsertEmptyInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheInsertEmptyInc", reflect.TypeOf((*MockMetrics)(nil).CacheInsertEmptyInc))
}

// CacheInsertInc mocks base method.
func (m *MockMetrics) CacheInsertInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheInsertInc")
}

// CacheInsertInc indicates an expected call of CacheInsertInc.
func (mr *MockMetricsMockRecorder) CacheInsertInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheInsertInc", reflect.TypeOf((*MockMetrics)(nil).CacheInsertInc))
}

// CacheMaxEntriesSet mocks base method.
func (m *MockMetrics) CacheMaxEntriesSet(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheMaxEntriesSet", arg0)
}

// CacheMaxEntriesSet indicates an expected call of CacheMaxEntriesSet.
func (mr *MockMetricsMockRecorder) CacheMaxEntriesSet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheMaxEntriesSet", reflect.TypeOf((*MockMetrics)(nil).CacheMaxEntriesSet), arg0)
}

// CacheMissInc mocks base method.
func (m *MockMetrics) CacheMissInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheMissInc")
}

// CacheMissInc indicates an expected call of CacheMissInc.
func (mr *MockMetricsMockRecorder) CacheMissInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheMissInc", reflect.TypeOf((*MockMetrics)(nil).CacheMissInc))
}

// CacheMoveInc mocks base method.
func (m *MockMetrics) CacheMoveInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheMoveInc")
}

// CacheMoveInc indicates an expected call of CacheMoveInc.
func (mr *MockMetricsMockRecorder) CacheMoveInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheMoveInc", reflect.TypeOf((*MockMetrics)(nil).CacheMoveInc))
}

// CacheRemoveInc mocks base method.
func (m *MockMetrics) CacheRemoveInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheRemoveInc")
}

// CacheRemoveInc indicates an expected call of CacheRemoveInc.
func (mr *MockMetricsMockRecorder) CacheRemoveInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheRemoveInc", reflect.TypeOf((*MockMetrics)(nil).CacheRemoveInc))
}

// SetCacheType mocks base method.
func (m *MockMetrics) SetCacheType(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCacheType", arg0)
}

// SetCacheType indicates an expected call of SetCacheType.
func (mr *MockMetricsMockRecorder) SetCacheType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCacheType", reflect.TypeOf((*MockMetrics)(nil).SetCacheType), arg0)
}
