package sqlite

import "context"

// BackupManager handles creating new backups and restoring old ones.
//
//go:generate moq -out ../mocks/backup_manager.go -pkg mocks -fmt goimports . BackupManager
type BackupManager interface {
	Backup(ctx context.Context, file string) error
	Restore(ctx context.Context, file string) error
	Close() error
}
