// Code generated by mockery v2.26.0. DO NOT EDIT.

package cbreaker

import mock "github.com/stretchr/testify/mock"

// MockOption is an autogenerated mock type for the Option type
type MockOption struct {
	mock.Mock
}

// apply provides a mock function with given fields: conf
func (_m *MockOption) apply(conf *config) error {
	ret := _m.Called(conf)

	var r0 error
	if rf, ok := ret.Get(0).(func(*config) error); ok {
		r0 = rf(conf)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockOption interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockOption creates a new instance of MockOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockOption(t mockConstructorTestingTNewMockOption) *MockOption {
	mock := &MockOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}