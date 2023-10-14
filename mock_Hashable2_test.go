// Code generated by mockery v2.26.0. DO NOT EDIT.

package conch

import mock "github.com/stretchr/testify/mock"

// MockHashable2 is an autogenerated mock type for the Hashable2 type
type MockHashable2 struct {
	mock.Mock
}

// Hash provides a mock function with given fields:
func (_m *MockHashable2) Hash() Key2 {
	ret := _m.Called()

	var r0 Key2
	if rf, ok := ret.Get(0).(func() Key2); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(Key2)
	}

	return r0
}

type mockConstructorTestingTNewMockHashable2 interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockHashable2 creates a new instance of MockHashable2. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockHashable2(t mockConstructorTestingTNewMockHashable2) *MockHashable2 {
	mock := &MockHashable2{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}