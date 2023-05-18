// Code generated by mockery v2.26.0. DO NOT EDIT.

package ratelimit

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockLimiter is an autogenerated mock type for the Limiter type
type MockLimiter struct {
	mock.Mock
}

// Take provides a mock function with given fields: ctx
func (_m *MockLimiter) Take(ctx context.Context) <-chan struct{} {
	ret := _m.Called(ctx)

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func(context.Context) <-chan struct{}); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

type mockConstructorTestingTNewMockLimiter interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockLimiter creates a new instance of MockLimiter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockLimiter(t mockConstructorTestingTNewMockLimiter) *MockLimiter {
	mock := &MockLimiter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}