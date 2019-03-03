// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/c0va23/go-proxyprotocol (interfaces: Logger,HeaderParser)

// Package proxyprotocol_test is a generated GoMock package.
package proxyprotocol_test

import (
	bufio "bufio"
	go_proxyprotocol "github.com/c0va23/go-proxyprotocol"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLogger is a mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Printf mocks base method
func (m *MockLogger) Printf(arg0 string, arg1 ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Printf", varargs...)
}

// Printf indicates an expected call of Printf
func (mr *MockLoggerMockRecorder) Printf(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Printf", reflect.TypeOf((*MockLogger)(nil).Printf), varargs...)
}

// MockHeaderParser is a mock of HeaderParser interface
type MockHeaderParser struct {
	ctrl     *gomock.Controller
	recorder *MockHeaderParserMockRecorder
}

// MockHeaderParserMockRecorder is the mock recorder for MockHeaderParser
type MockHeaderParserMockRecorder struct {
	mock *MockHeaderParser
}

// NewMockHeaderParser creates a new mock instance
func NewMockHeaderParser(ctrl *gomock.Controller) *MockHeaderParser {
	mock := &MockHeaderParser{ctrl: ctrl}
	mock.recorder = &MockHeaderParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHeaderParser) EXPECT() *MockHeaderParserMockRecorder {
	return m.recorder
}

// Parse mocks base method
func (m *MockHeaderParser) Parse(arg0 *bufio.Reader) (*go_proxyprotocol.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Parse", arg0)
	ret0, _ := ret[0].(*go_proxyprotocol.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Parse indicates an expected call of Parse
func (mr *MockHeaderParserMockRecorder) Parse(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Parse", reflect.TypeOf((*MockHeaderParser)(nil).Parse), arg0)
}