// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	sapcontrol "github.com/trento-project/trento/internal/sapsystem/sapcontrol"
)

// WebService is an autogenerated mock type for the WebService type
type WebService struct {
	mock.Mock
}

// GetInstanceProperties provides a mock function with given fields:
func (_m *WebService) GetInstanceProperties() (*sapcontrol.GetInstancePropertiesResponse, error) {
	ret := _m.Called()

	var r0 *sapcontrol.GetInstancePropertiesResponse
	if rf, ok := ret.Get(0).(func() *sapcontrol.GetInstancePropertiesResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sapcontrol.GetInstancePropertiesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProcessList provides a mock function with given fields:
func (_m *WebService) GetProcessList() (*sapcontrol.GetProcessListResponse, error) {
	ret := _m.Called()

	var r0 *sapcontrol.GetProcessListResponse
	if rf, ok := ret.Get(0).(func() *sapcontrol.GetProcessListResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sapcontrol.GetProcessListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSystemInstanceList provides a mock function with given fields:
func (_m *WebService) GetSystemInstanceList() (*sapcontrol.GetSystemInstanceListResponse, error) {
	ret := _m.Called()

	var r0 *sapcontrol.GetSystemInstanceListResponse
	if rf, ok := ret.Get(0).(func() *sapcontrol.GetSystemInstanceListResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sapcontrol.GetSystemInstanceListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
