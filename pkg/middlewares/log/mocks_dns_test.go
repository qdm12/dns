// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/miekg/dns (interfaces: ResponseWriter)

// Package log is a generated GoMock package.
package log

import (
	net "net"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dns "github.com/miekg/dns"
)

// MockResponseWriter is a mock of ResponseWriter interface.
type MockResponseWriter struct {
	ctrl     *gomock.Controller
	recorder *MockResponseWriterMockRecorder
}

// MockResponseWriterMockRecorder is the mock recorder for MockResponseWriter.
type MockResponseWriterMockRecorder struct {
	mock *MockResponseWriter
}

// NewMockResponseWriter creates a new mock instance.
func NewMockResponseWriter(ctrl *gomock.Controller) *MockResponseWriter {
	mock := &MockResponseWriter{ctrl: ctrl}
	mock.recorder = &MockResponseWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResponseWriter) EXPECT() *MockResponseWriterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockResponseWriter) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockResponseWriterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockResponseWriter)(nil).Close))
}

// Hijack mocks base method.
func (m *MockResponseWriter) Hijack() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Hijack")
}

// Hijack indicates an expected call of Hijack.
func (mr *MockResponseWriterMockRecorder) Hijack() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hijack", reflect.TypeOf((*MockResponseWriter)(nil).Hijack))
}

// LocalAddr mocks base method.
func (m *MockResponseWriter) LocalAddr() net.Addr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LocalAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

// LocalAddr indicates an expected call of LocalAddr.
func (mr *MockResponseWriterMockRecorder) LocalAddr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalAddr", reflect.TypeOf((*MockResponseWriter)(nil).LocalAddr))
}

// RemoteAddr mocks base method.
func (m *MockResponseWriter) RemoteAddr() net.Addr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteAddr")
	ret0, _ := ret[0].(net.Addr)
	return ret0
}

// RemoteAddr indicates an expected call of RemoteAddr.
func (mr *MockResponseWriterMockRecorder) RemoteAddr() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteAddr", reflect.TypeOf((*MockResponseWriter)(nil).RemoteAddr))
}

// TsigStatus mocks base method.
func (m *MockResponseWriter) TsigStatus() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TsigStatus")
	ret0, _ := ret[0].(error)
	return ret0
}

// TsigStatus indicates an expected call of TsigStatus.
func (mr *MockResponseWriterMockRecorder) TsigStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TsigStatus", reflect.TypeOf((*MockResponseWriter)(nil).TsigStatus))
}

// TsigTimersOnly mocks base method.
func (m *MockResponseWriter) TsigTimersOnly(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TsigTimersOnly", arg0)
}

// TsigTimersOnly indicates an expected call of TsigTimersOnly.
func (mr *MockResponseWriterMockRecorder) TsigTimersOnly(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TsigTimersOnly", reflect.TypeOf((*MockResponseWriter)(nil).TsigTimersOnly), arg0)
}

// Write mocks base method.
func (m *MockResponseWriter) Write(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockResponseWriterMockRecorder) Write(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockResponseWriter)(nil).Write), arg0)
}

// WriteMsg mocks base method.
func (m *MockResponseWriter) WriteMsg(arg0 *dns.Msg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteMsg indicates an expected call of WriteMsg.
func (mr *MockResponseWriterMockRecorder) WriteMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteMsg", reflect.TypeOf((*MockResponseWriter)(nil).WriteMsg), arg0)
}
