// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package filestore

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

type S3FileBackend struct {
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	bucket    string
	client    *minio.Client
	core      *minio.Core
}

func NewS3FileBackend(settings FileBackendSettings) (*S3FileBackend, error) {
	backend := &S3FileBackend{
		endpoint:  settings.S3Endpoint,
		accessKey: settings.S3AccessKeyId,
		secretKey: settings.S3SecretAccessKey,
		secure:    settings.S3SSL,
		bucket:    settings.S3Bucket,
	}
	cli, err := backend.s3New()
	if err != nil {
		return nil, err
	}

	backend.client = cli

	core, err := minio.NewCore(settings.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.S3AccessKeyId, settings.S3SecretAccessKey, ""),
		Secure: settings.S3SSL,
	})
	if err != nil {
		return nil, err
	}

	backend.core = core

	return backend, nil
}

func (b *S3FileBackend) s3New() (*minio.Client, error) {
	minioClient, err := minio.New(b.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(b.accessKey, b.secretKey, ""),
		Secure: b.secure,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

func (b *S3FileBackend) StartMultipartUpload(ctx context.Context, path string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return "", errors.New("empty upload path")
	}

	if strings.Contains(path, "..") {
		return "", errors.New("invalid path")
	}

	uploadID, err := b.core.NewMultipartUpload(ctx, b.bucket, path, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

func (b *S3FileBackend) UploadPart(ctx context.Context, path, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return "", errors.New("empty upload path")
	}

	if strings.Contains(path, "..") {
		return "", errors.New("invalid path")
	}

	if uploadID == "" {
		return "", errors.New("empty upload id")
	}

	if partNumber < 1 || partNumber > 10000 { // S3 limit: 1-10,000 parts
		return "", errors.New("part number must be between 1 and 10000")
	}

	if size <= 0 {
		return "", errors.New("invalid part size")
	}

	partInfo, err := b.core.PutObjectPart(ctx, b.bucket, path, uploadID, partNumber, reader, size, "", "", nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to upload part %d", partNumber)
	}

	// The ETag is required when completing the multipart upload
	return partInfo.ETag, nil
}

func (b *S3FileBackend) AbortMultipartUpload(ctx context.Context, path, uploadID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	if uploadID == "" {
		return errors.New("empty upload id")
	}

	err := b.core.AbortMultipartUpload(ctx, b.bucket, path, uploadID)
	if err != nil {
		return errors.Wrap(err, "failed to abort S3 multipart upload")
	}

	return nil
}

func (b *S3FileBackend) CompleteMultipartUpload(ctx context.Context, path, uploadID string, parts map[int]string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	if uploadID == "" {
		return errors.New("empty upload id")
	}

	if len(parts) == 0 {
		return errors.New("no parts to complete")
	}

	completeParts := make([]minio.CompletePart, 0, len(parts))

	for partNumber, etag := range parts {
		completeParts = append(completeParts, minio.CompletePart{
			PartNumber: partNumber,
			ETag:       etag,
		})
	}

	sort.Slice(completeParts, func(i, j int) bool {
		return completeParts[i].PartNumber < completeParts[j].PartNumber
	})

	_, err := b.core.CompleteMultipartUpload(ctx, b.bucket, path, uploadID, completeParts, minio.PutObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to complete multipart upload")
	}

	return nil
}

func (b *S3FileBackend) CopyFileFromReader(ctx context.Context, path string, reader io.Reader, size int64) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	_, err := b.client.PutObject(
		ctx,
		b.bucket,
		path,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
			PartSize:    64 * 1024 * 1024, // 64 MB
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (b *S3FileBackend) CreateDirectory(path string) error {
	// No need to create directory for S3
	return nil
}

func (b *S3FileBackend) FileSize(path string) (int64, error) {
	if path == "" {
		return 0, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return 0, errors.New("invalid path")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objInfo, err := b.client.StatObject(
		ctx,
		b.bucket,
		path,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return 0, err
	}

	return objInfo.Size, nil
}

func (b *S3FileBackend) FileETag(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return "", errors.New("invalid path")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objInfo, err := b.client.StatObject(
		ctx,
		b.bucket,
		path,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return "", err
	}

	return objInfo.ETag, nil
}

func (b *S3FileBackend) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return nil, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return nil, errors.New("invalid path")
	}

	obj, err := b.client.GetObject(
		ctx,
		b.bucket,
		path,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}

	return obj, nil // caller MUST Close()
}

func (b *S3FileBackend) FileExists(ctx context.Context, path string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return false, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return false, errors.New("invalid path")
	}

	_, err := b.client.StatObject(
		ctx,
		b.bucket,
		path,
		minio.StatObjectOptions{},
	)
	if err == nil {
		return true, nil
	}

	// Not found is NOT an error
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		return false, nil
	}

	// Real error
	return false, err
}

func (b *S3FileBackend) WriteFile(ctx context.Context, path string, r io.Reader, size int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	if size <= 0 {
		return errors.New("invalid file size")
	}

	_, err := b.client.PutObject(
		ctx,
		b.bucket,
		path,
		r,
		size,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	return err
}

func (b *S3FileBackend) MoveFile(ctx context.Context, src, dst string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if src == "" || dst == "" {
		return errors.New("empty path")
	}

	if strings.Contains(src, "..") || strings.Contains(dst, "..") {
		return errors.New("invalid path")
	}

	_, err := b.client.CopyObject(ctx,
		minio.CopyDestOptions{Bucket: b.bucket, Object: dst},
		minio.CopySrcOptions{Bucket: b.bucket, Object: src},
	)
	if err != nil {
		return errors.Wrap(err, "failed to copy object during move")
	}

	return b.client.RemoveObject(ctx, b.bucket, src, minio.RemoveObjectOptions{})
}

func (b *S3FileBackend) ComposeFile(ctx context.Context, path string) error {
	// Not applicable for S3. Multipart uploads use StartMultipartUpload/UploadPart/CompleteMultipartUpload
	return nil
}

func (b *S3FileBackend) RemoveFile(ctx context.Context, filePath string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if filePath == "" {
		return errors.New("empty path")
	}

	if strings.Contains(filePath, "..") {
		return errors.New("invalid path")
	}

	return b.client.RemoveObject(ctx, b.bucket, filePath, minio.RemoveObjectOptions{})
}

func (b *S3FileBackend) ServeFile(
	filePath string,
	w http.ResponseWriter,
	r *http.Request,
) error {
	if filePath == "" {
		return errors.New("empty path")
	}

	if strings.Contains(filePath, "..") {
		return errors.New("invalid path")
	}

	obj, err := b.client.GetObject(
		r.Context(),
		b.bucket,
		filePath,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return err
	}

	defer obj.Close()

	// Stat first so ServeContent's seeks resolve against cached object info
	// rather than costing a round trip each.
	stat, err := obj.Stat()
	if err != nil {
		return err
	}

	if _, set := w.Header()["Content-Type"]; !set && stat.ContentType != "" {
		w.Header().Set("Content-Type", stat.ContentType)
	}

	http.ServeContent(w, r, filePath, stat.LastModified, obj)

	return nil
}
