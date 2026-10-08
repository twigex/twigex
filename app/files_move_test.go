// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/twigex/twigex/internal/filestore"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeDiskMoveJobStore struct {
	store.JobsStore
	job *model.Job
}

func (f *fakeDiskMoveJobStore) Get(string) (*model.Job, error) {
	return f.job, nil
}

func (f *fakeDiskMoveJobStore) UpdateProgress(string, int) error {
	return nil
}

type fakeDiskMoveFileStore struct {
	store.FileStore
	files     map[string]*model.File
	commitErr error
	committed bool
}

func (f *fakeDiskMoveFileStore) Get(id string) (*model.File, error) {
	return f.files[id], nil
}

func (f *fakeDiskMoveFileStore) MoveToStorage(string, []string, []string, string, string) error {
	if f.commitErr != nil {
		return f.commitErr
	}

	f.committed = true
	return nil
}

type diskMove struct {
	app     *App
	jobs    *fakeDiskMoveJobStore
	files   *fakeDiskMoveFileStore
	payload model.MoveFilesJobPayload
	diskA   string
	diskB   string
}

func newDiskMove(t *testing.T, files ...*model.File) *diskMove {
	t.Helper()

	diskA := t.TempDir()
	diskB := t.TempDir()

	records := map[string]*model.File{
		"target": {
			ID:       "target",
			Storage:  "disk-b",
			IsFolder: true,
		},
	}

	roots := make([]model.MoveItem, 0, len(files))
	for _, f := range files {
		records[f.ID] = f
		roots = append(roots, model.MoveItem{ID: f.ID})
	}

	jobStore := &fakeDiskMoveJobStore{}
	fileStore := &fakeDiskMoveFileStore{files: records}

	a := &App{
		Store: store.Store{
			Jobs: jobStore,
			File: fileStore,
			Storage: &fakePhotoStorageStore{storages: map[string]*model.Storage{
				"disk-a": {ID: "disk-a", Directory: diskA},
				"disk-b": {ID: "disk-b", Directory: diskB},
			}},
		},
		FileStorageObjects: map[string]filestore.FileBackend{
			"disk-a": &filestore.LocalFileBackend{},
			"disk-b": &filestore.LocalFileBackend{},
		},
	}

	return &diskMove{
		app:   a,
		jobs:  jobStore,
		files: fileStore,
		payload: model.MoveFilesJobPayload{
			RootIDS:    roots,
			Target:     "target",
			TotalBytes: 1024,
		},
		diskA: diskA,
		diskB: diskB,
	}
}

func (m *diskMove) run(t *testing.T) error {
	t.Helper()

	payload, err := json.Marshal(m.payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	m.jobs.job = &model.Job{ID: "job-1", Payload: payload}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return m.app.MoveFilesAcrossDisks(ctx, "job-1")
}

func writeStored(t *testing.T, disk, name string, content []byte) {
	t.Helper()

	p := filepath.Join(disk, "files", name)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}

	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func stored(disk, name string) bool {
	_, err := os.Stat(filepath.Join(disk, "files", name))
	return err == nil
}

func imageFile(id string) *model.File {
	return &model.File{ID: id, Type: "image/png", Storage: "disk-a"}
}

func TestMoveFilesAcrossDisksMissingThumbnailDoesNotFailMove(t *testing.T) {
	m := newDiskMove(t, imageFile("with-thumb"), imageFile("no-thumb"))
	writeStored(t, m.diskA, "with-thumb", []byte("first"))
	writeStored(t, m.diskA, "thumbnails/with-thumb.jpg", []byte("thumb"))
	writeStored(t, m.diskA, "no-thumb", []byte("second"))

	if err := m.run(t); err != nil {
		t.Fatalf("move failed: %v", err)
	}

	if !m.files.committed {
		t.Fatal("expected the move to be committed")
	}

	for _, name := range []string{"with-thumb", "no-thumb", "thumbnails/with-thumb.jpg"} {
		if !stored(m.diskB, name) {
			t.Errorf("expected %s on the target disk", name)
		}
		if stored(m.diskA, name) {
			t.Errorf("expected %s removed from the source disk", name)
		}
	}
}

func TestMoveFilesAcrossDisksRollsBackWhenAFileFails(t *testing.T) {
	m := newDiskMove(t, imageFile("copied"), imageFile("missing"))
	writeStored(t, m.diskA, "copied", []byte("first"))
	writeStored(t, m.diskA, "thumbnails/copied.jpg", []byte("thumb"))

	if err := m.run(t); err == nil {
		t.Fatal("expected the move to fail")
	}

	if m.files.committed {
		t.Fatal("a failed move must not be committed")
	}

	for _, name := range []string{"copied", "thumbnails/copied.jpg"} {
		if stored(m.diskB, name) {
			t.Errorf("expected %s rolled back from the target disk", name)
		}
		if !stored(m.diskA, name) {
			t.Errorf("expected %s kept on the source disk", name)
		}
	}
}

func TestMoveFilesAcrossDisksRollsBackWhenCommitFails(t *testing.T) {
	m := newDiskMove(t, imageFile("photo"))
	m.files.commitErr = errors.New("database unavailable")
	writeStored(t, m.diskA, "photo", []byte("first"))
	writeStored(t, m.diskA, "thumbnails/photo.jpg", []byte("thumb"))

	if err := m.run(t); err == nil {
		t.Fatal("expected the move to fail")
	}

	for _, name := range []string{"photo", "thumbnails/photo.jpg"} {
		if stored(m.diskB, name) {
			t.Errorf("expected %s rolled back from the target disk", name)
		}
		if !stored(m.diskA, name) {
			t.Errorf("expected %s kept on the source disk", name)
		}
	}
}

func TestMoveFilesAcrossDisksWithoutTotalBytes(t *testing.T) {
	m := newDiskMove(t, imageFile("photo"))
	m.payload.TotalBytes = 0
	writeStored(t, m.diskA, "photo", []byte("not empty"))

	if err := m.run(t); err != nil {
		t.Fatalf("move failed: %v", err)
	}

	if !stored(m.diskB, "photo") {
		t.Error("expected the file on the target disk")
	}
}

func TestMoveFilesAcrossDisksMissingTarget(t *testing.T) {
	m := newDiskMove(t, imageFile("photo"))
	m.payload.Target = "deleted"
	writeStored(t, m.diskA, "photo", []byte("first"))

	if err := m.run(t); err == nil {
		t.Fatal("expected the move to fail")
	}

	if !stored(m.diskA, "photo") {
		t.Error("expected the file kept on the source disk")
	}
}

func TestOpenThumbnailRebuildsMissingThumbnail(t *testing.T) {
	m := newDiskMove(t)

	var img bytes.Buffer
	if err := png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 40, 20))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	writeStored(t, m.diskA, "photo", img.Bytes())

	fr, err := m.app.openThumbnail(context.Background(), *imageFile("photo"))
	if err != nil {
		t.Fatalf("openThumbnail: %v", err)
	}
	defer fr.Close()

	if _, err = io.ReadAll(fr); err != nil {
		t.Fatalf("read thumbnail: %v", err)
	}

	if !stored(m.diskA, "thumbnails/photo.jpg") {
		t.Error("expected the rebuilt thumbnail to be saved")
	}
}

func TestOpenThumbnailMissingForNonImage(t *testing.T) {
	m := newDiskMove(t)
	writeStored(t, m.diskA, "notes", []byte("text"))

	file := model.File{ID: "notes", Type: "text/plain", Storage: "disk-a"}

	if _, err := m.app.openThumbnail(context.Background(), file); err == nil {
		t.Fatal("expected no thumbnail for a text file")
	}
}
