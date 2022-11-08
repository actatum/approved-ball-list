// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/actatum/approved-ball-list/internal/sqlite"
)

// Ensure, that BackupManagerMock does implement sqlite.BackupManager.
// If this is not the case, regenerate this file with moq.
var _ sqlite.BackupManager = &BackupManagerMock{}

// BackupManagerMock is a mock implementation of sqlite.BackupManager.
//
//	func TestSomethingThatUsesBackupManager(t *testing.T) {
//
//		// make and configure a mocked sqlite.BackupManager
//		mockedBackupManager := &BackupManagerMock{
//			BackupFunc: func(ctx context.Context, file string) error {
//				panic("mock out the Backup method")
//			},
//			CloseFunc: func() error {
//				panic("mock out the Close method")
//			},
//			RestoreFunc: func(ctx context.Context, file string) error {
//				panic("mock out the Restore method")
//			},
//		}
//
//		// use mockedBackupManager in code that requires sqlite.BackupManager
//		// and then make assertions.
//
//	}
type BackupManagerMock struct {
	// BackupFunc mocks the Backup method.
	BackupFunc func(ctx context.Context, file string) error

	// CloseFunc mocks the Close method.
	CloseFunc func() error

	// RestoreFunc mocks the Restore method.
	RestoreFunc func(ctx context.Context, file string) error

	// calls tracks calls to the methods.
	calls struct {
		// Backup holds details about calls to the Backup method.
		Backup []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// File is the file argument value.
			File string
		}
		// Close holds details about calls to the Close method.
		Close []struct {
		}
		// Restore holds details about calls to the Restore method.
		Restore []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// File is the file argument value.
			File string
		}
	}
	lockBackup  sync.RWMutex
	lockClose   sync.RWMutex
	lockRestore sync.RWMutex
}

// Backup calls BackupFunc.
func (mock *BackupManagerMock) Backup(ctx context.Context, file string) error {
	if mock.BackupFunc == nil {
		panic("BackupManagerMock.BackupFunc: method is nil but BackupManager.Backup was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		File string
	}{
		Ctx:  ctx,
		File: file,
	}
	mock.lockBackup.Lock()
	mock.calls.Backup = append(mock.calls.Backup, callInfo)
	mock.lockBackup.Unlock()
	return mock.BackupFunc(ctx, file)
}

// BackupCalls gets all the calls that were made to Backup.
// Check the length with:
//
//	len(mockedBackupManager.BackupCalls())
func (mock *BackupManagerMock) BackupCalls() []struct {
	Ctx  context.Context
	File string
} {
	var calls []struct {
		Ctx  context.Context
		File string
	}
	mock.lockBackup.RLock()
	calls = mock.calls.Backup
	mock.lockBackup.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *BackupManagerMock) Close() error {
	if mock.CloseFunc == nil {
		panic("BackupManagerMock.CloseFunc: method is nil but BackupManager.Close was just called")
	}
	callInfo := struct {
	}{}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc()
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//
//	len(mockedBackupManager.CloseCalls())
func (mock *BackupManagerMock) CloseCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// Restore calls RestoreFunc.
func (mock *BackupManagerMock) Restore(ctx context.Context, file string) error {
	if mock.RestoreFunc == nil {
		panic("BackupManagerMock.RestoreFunc: method is nil but BackupManager.Restore was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		File string
	}{
		Ctx:  ctx,
		File: file,
	}
	mock.lockRestore.Lock()
	mock.calls.Restore = append(mock.calls.Restore, callInfo)
	mock.lockRestore.Unlock()
	return mock.RestoreFunc(ctx, file)
}

// RestoreCalls gets all the calls that were made to Restore.
// Check the length with:
//
//	len(mockedBackupManager.RestoreCalls())
func (mock *BackupManagerMock) RestoreCalls() []struct {
	Ctx  context.Context
	File string
} {
	var calls []struct {
		Ctx  context.Context
		File string
	}
	mock.lockRestore.RLock()
	calls = mock.calls.Restore
	mock.lockRestore.RUnlock()
	return calls
}
