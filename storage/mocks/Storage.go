// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import datatypes "telemetry-server/datatypes"
import mock "github.com/stretchr/testify/mock"

import time "time"

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Storage) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMetric provides a mock function with given fields: metric
func (_m *Storage) DeleteMetric(metric string) error {
	ret := _m.Called(metric)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(metric)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Insert provides a mock function with given fields: points
func (_m *Storage) Insert(points []*datatypes.Datapoint) error {
	ret := _m.Called(points)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*datatypes.Datapoint) error); ok {
		r0 = rf(points)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Latest provides a mock function with given fields: metric
func (_m *Storage) Latest(metric string) (*datatypes.Datapoint, error) {
	ret := _m.Called(metric)

	var r0 *datatypes.Datapoint
	if rf, ok := ret.Get(0).(func(string) *datatypes.Datapoint); ok {
		r0 = rf(metric)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*datatypes.Datapoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(metric)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMetrics provides a mock function with given fields:
func (_m *Storage) ListMetrics() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
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

// SelectMetric provides a mock function with given fields: metric
func (_m *Storage) SelectMetric(metric string) ([]*datatypes.Datapoint, error) {
	ret := _m.Called(metric)

	var r0 []*datatypes.Datapoint
	if rf, ok := ret.Get(0).(func(string) []*datatypes.Datapoint); ok {
		r0 = rf(metric)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*datatypes.Datapoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(metric)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SelectMetricTimeRange provides a mock function with given fields: metric, start, end
func (_m *Storage) SelectMetricTimeRange(metric string, start time.Time, end time.Time) ([]*datatypes.Datapoint, error) {
	ret := _m.Called(metric, start, end)

	var r0 []*datatypes.Datapoint
	if rf, ok := ret.Get(0).(func(string, time.Time, time.Time) []*datatypes.Datapoint); ok {
		r0 = rf(metric, start, end)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*datatypes.Datapoint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, time.Time, time.Time) error); ok {
		r1 = rf(metric, start, end)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}