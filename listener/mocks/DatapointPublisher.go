// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	"telemetry-server/datatypes"

	mock "github.com/stretchr/testify/mock"
)

// DatapointPublisher is an autogenerated mock type for the DatapointPublisher type
type DatapointPublisher struct {
	mock.Mock
}

// Publish provides a mock function with given fields: point
func (_m *DatapointPublisher) Publish(point *datatypes.Datapoint) {
	_m.Called(point)
}

// Subscribe provides a mock function with given fields: c
func (_m *DatapointPublisher) Subscribe(c chan *datatypes.Datapoint) error {
	ret := _m.Called(c)

	var r0 error
	if rf, ok := ret.Get(0).(func(chan *datatypes.Datapoint) error); ok {
		r0 = rf(c)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: c
func (_m *DatapointPublisher) Unsubscribe(c chan *datatypes.Datapoint) error {
	ret := _m.Called(c)

	var r0 error
	if rf, ok := ret.Get(0).(func(chan *datatypes.Datapoint) error); ok {
		r0 = rf(c)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
