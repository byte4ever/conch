// Code generated by mockery v2.40.1. DO NOT EDIT.

package conch

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockReturnFun is an autogenerated mock type for the ReturnFun type
type MockReturnFun[R interface{}] struct {
	mock.Mock
}

type MockReturnFun_Expecter[R interface{}] struct {
	mock *mock.Mock
}

func (_m *MockReturnFun[R]) EXPECT() *MockReturnFun_Expecter[R] {
	return &MockReturnFun_Expecter[R]{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: _a0, _a1
func (_m *MockReturnFun[R]) Execute(_a0 context.Context, _a1 ValErrorPair[R]) {
	_m.Called(_a0, _a1)
}

// MockReturnFun_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockReturnFun_Execute_Call[R interface{}] struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 ValErrorPair[R]
func (_e *MockReturnFun_Expecter[R]) Execute(_a0 interface{}, _a1 interface{}) *MockReturnFun_Execute_Call[R] {
	return &MockReturnFun_Execute_Call[R]{Call: _e.mock.On("Execute", _a0, _a1)}
}

func (_c *MockReturnFun_Execute_Call[R]) Run(run func(_a0 context.Context, _a1 ValErrorPair[R])) *MockReturnFun_Execute_Call[R] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ValErrorPair[R]))
	})
	return _c
}

func (_c *MockReturnFun_Execute_Call[R]) Return() *MockReturnFun_Execute_Call[R] {
	_c.Call.Return()
	return _c
}

func (_c *MockReturnFun_Execute_Call[R]) RunAndReturn(run func(context.Context, ValErrorPair[R])) *MockReturnFun_Execute_Call[R] {
	_c.Call.Return(run)
	return _c
}

// NewMockReturnFun creates a new instance of MockReturnFun. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockReturnFun[R interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReturnFun[R] {
	mock := &MockReturnFun[R]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
