// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockUnsigned is an autogenerated mock type for the Unsigned type
type MockUnsigned struct {
	mock.Mock
}

type mockConstructorTestingTNewMockUnsigned interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockUnsigned creates a new instance of MockUnsigned. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockUnsigned(t mockConstructorTestingTNewMockUnsigned) *MockUnsigned {
	mock := &MockUnsigned{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
