// Code generated by mockery v2.40.1. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockInteger is an autogenerated mock type for the Integer type
type MockInteger struct {
	mock.Mock
}

type MockInteger_Expecter struct {
	mock *mock.Mock
}

func (_m *MockInteger) EXPECT() *MockInteger_Expecter {
	return &MockInteger_Expecter{mock: &_m.Mock}
}

// NewMockInteger creates a new instance of MockInteger. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockInteger(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockInteger {
	mock := &MockInteger{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
