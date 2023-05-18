// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockReorderOption is an autogenerated mock type for the ReorderOption type
type MockReorderOption struct {
	mock.Mock
}

// apply provides a mock function with given fields: _a0
func (_m *MockReorderOption) apply(_a0 *reorderOptions) {
	_m.Called(_a0)
}

type mockConstructorTestingTNewMockReorderOption interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockReorderOption creates a new instance of MockReorderOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockReorderOption(t mockConstructorTestingTNewMockReorderOption) *MockReorderOption {
	mock := &MockReorderOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}