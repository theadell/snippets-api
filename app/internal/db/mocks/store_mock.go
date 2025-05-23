// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"snippets.adelh.dev/app/internal/db/sqlc"
)

// NewMockStore creates a new instance of MockStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStore {
	mock := &MockStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockStore is an autogenerated mock type for the Store type
type MockStore struct {
	mock.Mock
}

type MockStore_Expecter struct {
	mock *mock.Mock
}

func (_m *MockStore) EXPECT() *MockStore_Expecter {
	return &MockStore_Expecter{mock: &_m.Mock}
}

// Close provides a mock function for the type MockStore
func (_mock *MockStore) Close() error {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func() error); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockStore_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockStore_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockStore_Expecter) Close() *MockStore_Close_Call {
	return &MockStore_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockStore_Close_Call) Run(run func()) *MockStore_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockStore_Close_Call) Return(err error) *MockStore_Close_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockStore_Close_Call) RunAndReturn(run func() error) *MockStore_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Primary provides a mock function for the type MockStore
func (_mock *MockStore) Primary() sqlc.Querier {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Primary")
	}

	var r0 sqlc.Querier
	if returnFunc, ok := ret.Get(0).(func() sqlc.Querier); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sqlc.Querier)
		}
	}
	return r0
}

// MockStore_Primary_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Primary'
type MockStore_Primary_Call struct {
	*mock.Call
}

// Primary is a helper method to define mock.On call
func (_e *MockStore_Expecter) Primary() *MockStore_Primary_Call {
	return &MockStore_Primary_Call{Call: _e.mock.On("Primary")}
}

func (_c *MockStore_Primary_Call) Run(run func()) *MockStore_Primary_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockStore_Primary_Call) Return(querier sqlc.Querier) *MockStore_Primary_Call {
	_c.Call.Return(querier)
	return _c
}

func (_c *MockStore_Primary_Call) RunAndReturn(run func() sqlc.Querier) *MockStore_Primary_Call {
	_c.Call.Return(run)
	return _c
}

// Replica provides a mock function for the type MockStore
func (_mock *MockStore) Replica() sqlc.Querier {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Replica")
	}

	var r0 sqlc.Querier
	if returnFunc, ok := ret.Get(0).(func() sqlc.Querier); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sqlc.Querier)
		}
	}
	return r0
}

// MockStore_Replica_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Replica'
type MockStore_Replica_Call struct {
	*mock.Call
}

// Replica is a helper method to define mock.On call
func (_e *MockStore_Expecter) Replica() *MockStore_Replica_Call {
	return &MockStore_Replica_Call{Call: _e.mock.On("Replica")}
}

func (_c *MockStore_Replica_Call) Run(run func()) *MockStore_Replica_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockStore_Replica_Call) Return(querier sqlc.Querier) *MockStore_Replica_Call {
	_c.Call.Return(querier)
	return _c
}

func (_c *MockStore_Replica_Call) RunAndReturn(run func() sqlc.Querier) *MockStore_Replica_Call {
	_c.Call.Return(run)
	return _c
}

// WithTx provides a mock function for the type MockStore
func (_mock *MockStore) WithTx(ctx context.Context, fn func(sqlc.Querier) error) error {
	ret := _mock.Called(ctx, fn)

	if len(ret) == 0 {
		panic("no return value specified for WithTx")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, func(sqlc.Querier) error) error); ok {
		r0 = returnFunc(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockStore_WithTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithTx'
type MockStore_WithTx_Call struct {
	*mock.Call
}

// WithTx is a helper method to define mock.On call
//   - ctx
//   - fn
func (_e *MockStore_Expecter) WithTx(ctx interface{}, fn interface{}) *MockStore_WithTx_Call {
	return &MockStore_WithTx_Call{Call: _e.mock.On("WithTx", ctx, fn)}
}

func (_c *MockStore_WithTx_Call) Run(run func(ctx context.Context, fn func(sqlc.Querier) error)) *MockStore_WithTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(func(sqlc.Querier) error))
	})
	return _c
}

func (_c *MockStore_WithTx_Call) Return(err error) *MockStore_WithTx_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockStore_WithTx_Call) RunAndReturn(run func(ctx context.Context, fn func(sqlc.Querier) error) error) *MockStore_WithTx_Call {
	_c.Call.Return(run)
	return _c
}
