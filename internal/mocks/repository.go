// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/actatum/approved-ball-list/internal/balls"
)

// Ensure, that RepositoryMock does implement balls.Repository.
// If this is not the case, regenerate this file with moq.
var _ balls.Repository = &RepositoryMock{}

// RepositoryMock is a mock implementation of balls.Repository.
//
//	func TestSomethingThatUsesRepository(t *testing.T) {
//
//		// make and configure a mocked balls.Repository
//		mockedRepository := &RepositoryMock{
//			AddFunc: func(ctx context.Context, ballsMoqParam ...balls.Ball) ([]balls.Ball, error) {
//				panic("mock out the Add method")
//			},
//		}
//
//		// use mockedRepository in code that requires balls.Repository
//		// and then make assertions.
//
//	}
type RepositoryMock struct {
	// AddFunc mocks the Add method.
	AddFunc func(ctx context.Context, ballsMoqParam ...balls.Ball) ([]balls.Ball, error)

	// calls tracks calls to the methods.
	calls struct {
		// Add holds details about calls to the Add method.
		Add []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// BallsMoqParam is the ballsMoqParam argument value.
			BallsMoqParam []balls.Ball
		}
	}
	lockAdd sync.RWMutex
}

// Add calls AddFunc.
func (mock *RepositoryMock) Add(ctx context.Context, ballsMoqParam ...balls.Ball) ([]balls.Ball, error) {
	if mock.AddFunc == nil {
		panic("RepositoryMock.AddFunc: method is nil but Repository.Add was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		BallsMoqParam []balls.Ball
	}{
		Ctx:           ctx,
		BallsMoqParam: ballsMoqParam,
	}
	mock.lockAdd.Lock()
	mock.calls.Add = append(mock.calls.Add, callInfo)
	mock.lockAdd.Unlock()
	return mock.AddFunc(ctx, ballsMoqParam...)
}

// AddCalls gets all the calls that were made to Add.
// Check the length with:
//
//	len(mockedRepository.AddCalls())
func (mock *RepositoryMock) AddCalls() []struct {
	Ctx           context.Context
	BallsMoqParam []balls.Ball
} {
	var calls []struct {
		Ctx           context.Context
		BallsMoqParam []balls.Ball
	}
	mock.lockAdd.RLock()
	calls = mock.calls.Add
	mock.lockAdd.RUnlock()
	return calls
}
