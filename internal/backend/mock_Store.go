// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package backend

import (
	"context"

	mock "github.com/stretchr/testify/mock"
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

// Delete provides a mock function for the type MockStore
func (_mock *MockStore) Delete(ctx context.Context, key string) error {
	ret := _mock.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, key)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockStore_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockStore_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx
//   - key
func (_e *MockStore_Expecter) Delete(ctx interface{}, key interface{}) *MockStore_Delete_Call {
	return &MockStore_Delete_Call{Call: _e.mock.On("Delete", ctx, key)}
}

func (_c *MockStore_Delete_Call) Run(run func(ctx context.Context, key string)) *MockStore_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockStore_Delete_Call) Return(err error) *MockStore_Delete_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockStore_Delete_Call) RunAndReturn(run func(ctx context.Context, key string) error) *MockStore_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function for the type MockStore
func (_mock *MockStore) Get(ctx context.Context, key string) (string, error) {
	ret := _mock.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return returnFunc(ctx, key)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = returnFunc(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, key)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx
//   - key
func (_e *MockStore_Expecter) Get(ctx interface{}, key interface{}) *MockStore_Get_Call {
	return &MockStore_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *MockStore_Get_Call) Run(run func(ctx context.Context, key string)) *MockStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockStore_Get_Call) Return(s string, err error) *MockStore_Get_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockStore_Get_Call) RunAndReturn(run func(ctx context.Context, key string) (string, error)) *MockStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function for the type MockStore
func (_mock *MockStore) Set(ctx context.Context, key string, value string) error {
	ret := _mock.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = returnFunc(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockStore_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockStore_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx
//   - key
//   - value
func (_e *MockStore_Expecter) Set(ctx interface{}, key interface{}, value interface{}) *MockStore_Set_Call {
	return &MockStore_Set_Call{Call: _e.mock.On("Set", ctx, key, value)}
}

func (_c *MockStore_Set_Call) Run(run func(ctx context.Context, key string, value string)) *MockStore_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockStore_Set_Call) Return(err error) *MockStore_Set_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockStore_Set_Call) RunAndReturn(run func(ctx context.Context, key string, value string) error) *MockStore_Set_Call {
	_c.Call.Return(run)
	return _c
}
