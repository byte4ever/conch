// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockComplex is an autogenerated mock type for the Complex type
type MockComplex struct {
	mock.Mock
}

type mockConstructorTestingTNewMockComplex interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockComplex creates a new instance of MockComplex. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockComplex(t mockConstructorTestingTNewMockComplex) *MockComplex {
	mock := &MockComplex{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}