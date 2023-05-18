// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockReturnFun is an autogenerated mock type for the ReturnFun type
type MockReturnFun[R interface{}] struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0, _a1
func (_m *MockReturnFun[R]) Execute(_a0 context.Context, _a1 ValErrorPair[R]) {
	_m.Called(_a0, _a1)
}

type mockConstructorTestingTNewMockReturnFun interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockReturnFun creates a new instance of MockReturnFun. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockReturnFun[R interface{}](t mockConstructorTestingTNewMockReturnFun) *MockReturnFun[R] {
	mock := &MockReturnFun[R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
