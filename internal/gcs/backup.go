// Package gcs provides an implementation of the BackupManager interface
// using Google Cloud Storage as the backing store.
package gcs

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/actatum/approved-ball-list/internal/sqlite"
	"github.com/rs/zerolog"
)

const backupObjectName = "backup.gz"

// Ensure BackupManager implments the sqlite.BackupManager interface.
var _ sqlite.BackupManager = &BackupManager{}

// BackupManager provides functionality for backing up sqlite db file.
type BackupManager struct {
	client *storage.Client
	bucket *storage.BucketHandle
}

// NewBackupManager returns a new instance of BackupManager.
func NewBackupManager(bucket string) (*BackupManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &BackupManager{
		client: client,
		bucket: client.Bucket(bucket),
	}, nil
}

// Close shuts down the underlying gcs client.
func (bm *BackupManager) Close() error {
	return bm.client.Close()
}

// Backup creates a backup of the given data in the backup.gz file
// in the configured cloud storage bucket.
func (bm *BackupManager) Backup(ctx context.Context, file string) error {
	obj := bm.bucket.Object(backupObjectName)

	w := obj.NewWriter(ctx)
	gzipWriter := gzip.NewWriter(w)

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("os.ReadFile")
	}

	_, err = gzipWriter.Write(data)
	if err != nil {
		return fmt.Errorf("gzipWriter.Write: %w", err)
	}
	if err = gzipWriter.Close(); err != nil {
		return fmt.Errorf("gzipWriter.Close: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("w.Close: %w", err)
	}

	return nil
}

// Restore restores data from a backup stored in the backup.gz file in cloud storage.
func (bm *BackupManager) Restore(ctx context.Context, file string) (err error) {
	obj := bm.bucket.Object(backupObjectName)

	r, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("obj.NewReader: %w", err)
	}
	defer func() {
		e := r.Close()
		if e != nil {
			zerolog.Ctx(ctx).Info().Err(e).Msg("r.Close")
		}
	}()

	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %w", err)
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer func() {
		e := f.Close()
		if e != nil {
			zerolog.Ctx(ctx).Info().Err(e).Msg("f.Close")
		}
	}()

	_, err = io.Copy(f, gzipReader)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}
