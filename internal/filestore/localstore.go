// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package filestore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/twigex/twigex/tlog"
)

type LocalFileBackend struct{}

// multipartTempDir derives the parts temp directory from the final file path.
// path: data/storage/<id>/files/<fileID> → data/storage/<id>/tmp/files/<fileID>
func multipartTempDir(path string) string {
	return filepath.Join(
		filepath.Dir(filepath.Dir(path)),
		"tmp",
		filepath.Base(filepath.Dir(path)),
		filepath.Base(path),
	)
}

func (b *LocalFileBackend) StartMultipartUpload(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	tempDir := multipartTempDir(path)

	os.RemoveAll(tempDir)

	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", errors.Wrap(err, "failed to create upload directory")
	}

	return filepath.Base(path), nil
}

func (b *LocalFileBackend) UploadPart(ctx context.Context, path, uploadID string, partNumber int, reader io.Reader, size int64) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	if partNumber < 1 {
		return "", errors.New("part number must be >= 1")
	}

	if size <= 0 {
		return "", errors.New("invalid part size")
	}

	tempDir := multipartTempDir(path)
	partFilename := filepath.Join(tempDir, fmt.Sprintf("part-%05d", partNumber))

	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", errors.Wrap(err, "failed to create directory")
	}

	file, err := os.OpenFile(partFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create part file %d", partNumber)
	}

	defer file.Close()

	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(partFilename)
		return "", errors.Wrapf(err, "failed to write part %d", partNumber)
	}

	if written != size {
		os.Remove(partFilename)
		return "", errors.Errorf("part %d size mismatch: expected %d, got %d", partNumber, size, written)
	}

	return fmt.Sprintf("%s-part-%d-%d", path, partNumber, written), nil
}

func (b *LocalFileBackend) CompleteMultipartUpload(ctx context.Context, path, uploadID string, parts map[int]string) error {
	if path == "" {
		return errors.New("empty path")
	}

	if len(parts) == 0 {
		return errors.New("no parts to complete")
	}

	tempDir := multipartTempDir(path)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.Wrap(err, "failed to create final directory")
	}

	partNumbers := make([]int, 0, len(parts))
	for partNumber := range parts {
		partNumbers = append(partNumbers, partNumber)
	}

	sort.Ints(partNumbers)

	for _, partNumber := range partNumbers {
		partFile := filepath.Join(tempDir, fmt.Sprintf("part-%05d", partNumber))
		if _, err := os.Stat(partFile); os.IsNotExist(err) {
			return errors.Errorf("part %d file not found", partNumber)
		}
	}

	finalFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return errors.Wrap(err, "failed to create final file")
	}

	defer finalFile.Close()

	for _, partNumber := range partNumbers {
		partFile := filepath.Join(tempDir, fmt.Sprintf("part-%05d", partNumber))

		partData, err := os.Open(partFile)
		if err != nil {
			return errors.Wrapf(err, "failed to open part %d", partNumber)
		}

		_, err = io.Copy(finalFile, partData)
		partData.Close()

		if err != nil {
			return errors.Wrapf(err, "failed to append part %d", partNumber)
		}

		if err := os.Remove(partFile); err != nil {
			tlog.Warnw("Failed to delete part file", "path", partFile, "error", err)
		}
	}

	if err := os.Remove(tempDir); err != nil {
		tlog.Warnw("Failed to remove temp directory", "path", tempDir, "error", err)
	}

	return nil
}

func (b *LocalFileBackend) AbortMultipartUpload(ctx context.Context, path, uploadID string) error {
	if path == "" {
		return errors.New("empty path")
	}

	tempDir := multipartTempDir(path)

	err := os.RemoveAll(tempDir)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to abort multipart upload")
	}

	return nil
}

func (l *LocalFileBackend) CopyFileFromReader(ctx context.Context, path string, reader io.Reader, size int64) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	tempPath := path + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := file.Sync(); err != nil {
		os.Remove(tempPath)
		return err
	}

	return os.Rename(tempPath, path)
}

func (l *LocalFileBackend) CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	return nil
}

func (l *LocalFileBackend) FileSize(path string) (int64, error) {
	if path == "" {
		return 0, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return 0, errors.New("invalid path")
	}

	fs, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return fs.Size(), nil
}

// FileETag identifies a file by modification time and size, local storage
// having no content hash to report.
func (l *LocalFileBackend) FileETag(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return "", errors.New("invalid path")
	}

	fs, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x-%x", fs.ModTime().UnixNano(), fs.Size()), nil
}

func (l *LocalFileBackend) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return nil, errors.New("invalid path")
	}

	return os.Open(path)
}

func (l *LocalFileBackend) FileExists(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return false, errors.New("invalid path")
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (l *LocalFileBackend) MoveFile(ctx context.Context, src, dst string) error {
	if src == "" || dst == "" {
		return errors.New("empty path")
	}

	if strings.Contains(src, "..") || strings.Contains(dst, "..") {
		return errors.New("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	return os.Rename(src, dst)
}

func (l *LocalFileBackend) WriteFile(ctx context.Context, path string, r io.Reader, size int64) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	return f.Sync()
}

func (l *LocalFileBackend) ComposeFile(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	tmpDir := filepath.Join(filepath.Dir(path), "tmp", filepath.Base(path))
	tmpOutPath := tmpDir + ".out"

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	partPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			partPaths = append(partPaths, filepath.Join(tmpDir, e.Name()))
		}
	}

	if len(partPaths) == 0 {
		return errors.New("no upload parts found")
	}

	sort.Strings(partPaths)

	out, err := os.Create(tmpOutPath)
	if err != nil {
		return err
	}

	defer out.Close()

	for _, part := range partPaths {
		f, err := os.Open(part)
		if err != nil {
			return err
		}

		_, err = io.Copy(out, f)
		f.Close()
		if err != nil {
			return err
		}
	}

	if err := out.Sync(); err != nil {
		return err
	}

	if err := os.Rename(tmpOutPath, path); err != nil {
		return err
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		tlog.Warnw("Failed to remove temp directory", "path", tmpDir, "error", err)
	}

	return nil
}

func (l *LocalFileBackend) RemoveFile(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (l *LocalFileBackend) ServeFile(path string, w http.ResponseWriter, r *http.Request) error {
	if path == "" {
		return errors.New("empty path")
	}

	if strings.Contains(path, "..") {
		return errors.New("invalid path")
	}

	http.ServeFile(w, r, path)

	return nil
}
