// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package filestore

import (
	"context"
	"errors"
	"io"
	"net/http"
)

const (
	Local = "local"
	S3    = "s3"
)

type FileBackend interface {
	StartMultipartUpload(ctx context.Context, path string) (string, error)
	UploadPart(ctx context.Context, path, uploadID string, partNumber int, reader io.Reader, size int64) (etag string, err error)
	CompleteMultipartUpload(ctx context.Context, path, uploadID string, parts map[int]string) error
	AbortMultipartUpload(ctx context.Context, path, uploadID string) error

	WriteFile(ctx context.Context, path string, fr io.Reader, size int64) error
	MoveFile(ctx context.Context, src, dst string) error
	ComposeFile(ctx context.Context, path string) error
	ServeFile(path string, w http.ResponseWriter, r *http.Request) error
	RemoveFile(ctx context.Context, path string) error
	CreateDirectory(path string) error
	ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
	FileExists(ctx context.Context, path string) (bool, error)
	FileSize(path string) (int64, error)
	FileETag(path string) (string, error)
	CopyFileFromReader(ctx context.Context, path string, reader io.Reader, size int64) error
}

type FileBackendSettings struct {
	DriverName        string
	Directory         string
	S3AccessKeyId     string
	S3SecretAccessKey string
	S3Bucket          string
	S3Endpoint        string
	S3SSL             bool
}

func NewFileBackend(settings FileBackendSettings) (FileBackend, error) {
	switch settings.DriverName {
	case S3:
		backend, err := NewS3FileBackend(settings)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		return backend, nil
	case Local:
		return &LocalFileBackend{}, nil
	}

	return nil, errors.New("no valid filestorage driver found")
}
