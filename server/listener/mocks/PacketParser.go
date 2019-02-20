// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import datatypes "server/datatypes"

import mock "github.com/stretchr/testify/mock"

// PacketParser is an autogenerated mock type for the PacketParser type
type PacketParser struct {
	mock.Mock
}

// ParseByte provides a mock function with given fields: value
func (_m *PacketParser) ParseByte(value byte) bool {
	ret := _m.Called(value)

	var r0 bool
	if rf, ok := ret.Get(0).(func(byte) bool); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ParsePacket provides a mock function with given fields:
func (_m *PacketParser) ParsePacket() []*datatypes.Datapoint {
	ret := _m.Called()

	var r0 []*datatypes.Datapoint
	if rf, ok := ret.Get(0).(func() []*datatypes.Datapoint); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*datatypes.Datapoint)
		}
	}

	return r0
}