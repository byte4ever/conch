// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockHashable is an autogenerated mock type for the Hashable type
type MockHashable struct {
	mock.Mock
}

// Hash provides a mock function with given fields:
func (_m *MockHashable) Hash() Key {
	ret := _m.Called()

	var r0 Key
	if rf, ok := ret.Get(0).(func() Key); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(Key)
	}

	return r0
}

type mockConstructorTestingTNewMockHashable interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockHashable creates a new instance of MockHashable. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockHashable(t mockConstructorTestingTNewMockHashable) *MockHashable {
	mock := &MockHashable{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}