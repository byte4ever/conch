// Code generated by mockery v2.36.0. DO NOT EDIT.

package observer

import mock "github.com/stretchr/testify/mock"

// MockObserver is an autogenerated mock type for the Observer type
type MockObserver struct {
	mock.Mock
}

// OnNotify provides a mock function with given fields: i
func (_m *MockObserver) OnNotify(i interface{}) {
	_m.Called(i)
}

// NewMockObserver creates a new instance of MockObserver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockObserver(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockObserver {
	mock := &MockObserver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
