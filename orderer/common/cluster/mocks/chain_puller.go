// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	common "github.com/hyperledger/fabric-protos-go-apiv2/common"
	mock "github.com/stretchr/testify/mock"
)

// ChainPuller is an autogenerated mock type for the ChainPuller type
type ChainPuller struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *ChainPuller) Close() {
	_m.Called()
}

// HeightsByEndpoints provides a mock function with given fields:
func (_m *ChainPuller) HeightsByEndpoints() (map[string]uint64, string, error) {
	ret := _m.Called()

	var r0 map[string]uint64
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func() (map[string]uint64, string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() map[string]uint64); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]uint64)
		}
	}

	if rf, ok := ret.Get(1).(func() string); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// PullBlock provides a mock function with given fields: seq
func (_m *ChainPuller) PullBlock(seq uint64) *common.Block {
	ret := _m.Called(seq)

	var r0 *common.Block
	if rf, ok := ret.Get(0).(func(uint64) *common.Block); ok {
		r0 = rf(seq)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.Block)
		}
	}

	return r0
}

type mockConstructorTestingTNewChainPuller interface {
	mock.TestingT
	Cleanup(func())
}

// NewChainPuller creates a new instance of ChainPuller. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewChainPuller(t mockConstructorTestingTNewChainPuller) *ChainPuller {
	mock := &ChainPuller{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
