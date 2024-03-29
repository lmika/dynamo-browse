// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// UIService is an autogenerated mock type for the UIService type
type UIService struct {
	mock.Mock
}

type UIService_Expecter struct {
	mock *mock.Mock
}

func (_m *UIService) EXPECT() *UIService_Expecter {
	return &UIService_Expecter{mock: &_m.Mock}
}

// PrintMessage provides a mock function with given fields: ctx, msg
func (_m *UIService) PrintMessage(ctx context.Context, msg string) {
	_m.Called(ctx, msg)
}

// UIService_PrintMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PrintMessage'
type UIService_PrintMessage_Call struct {
	*mock.Call
}

// PrintMessage is a helper method to define mock.On call
//   - ctx context.Context
//   - msg string
func (_e *UIService_Expecter) PrintMessage(ctx interface{}, msg interface{}) *UIService_PrintMessage_Call {
	return &UIService_PrintMessage_Call{Call: _e.mock.On("PrintMessage", ctx, msg)}
}

func (_c *UIService_PrintMessage_Call) Run(run func(ctx context.Context, msg string)) *UIService_PrintMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *UIService_PrintMessage_Call) Return() *UIService_PrintMessage_Call {
	_c.Call.Return()
	return _c
}

func (_c *UIService_PrintMessage_Call) RunAndReturn(run func(context.Context, string)) *UIService_PrintMessage_Call {
	_c.Call.Return(run)
	return _c
}

// Prompt provides a mock function with given fields: ctx, msg
func (_m *UIService) Prompt(ctx context.Context, msg string) chan string {
	ret := _m.Called(ctx, msg)

	var r0 chan string
	if rf, ok := ret.Get(0).(func(context.Context, string) chan string); ok {
		r0 = rf(ctx, msg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan string)
		}
	}

	return r0
}

// UIService_Prompt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Prompt'
type UIService_Prompt_Call struct {
	*mock.Call
}

// Prompt is a helper method to define mock.On call
//   - ctx context.Context
//   - msg string
func (_e *UIService_Expecter) Prompt(ctx interface{}, msg interface{}) *UIService_Prompt_Call {
	return &UIService_Prompt_Call{Call: _e.mock.On("Prompt", ctx, msg)}
}

func (_c *UIService_Prompt_Call) Run(run func(ctx context.Context, msg string)) *UIService_Prompt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *UIService_Prompt_Call) Return(_a0 chan string) *UIService_Prompt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *UIService_Prompt_Call) RunAndReturn(run func(context.Context, string) chan string) *UIService_Prompt_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewUIService interface {
	mock.TestingT
	Cleanup(func())
}

// NewUIService creates a new instance of UIService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUIService(t mockConstructorTestingTNewUIService) *UIService {
	mock := &UIService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
