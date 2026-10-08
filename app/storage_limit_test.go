// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeStorageFileStore struct {
	store.FileStore
	used int64
}

func (f *fakeStorageFileStore) GetStorageUsed(string) (int64, error) {
	return f.used, nil
}

func storageApp(used int64) *App {
	return &App{
		Server: Server{},
		Store:  store.Store{File: &fakeStorageFileStore{used: used}},
	}
}

func TestIsFreeSpaceTreatsZeroAsUnlimited(t *testing.T) {
	a := storageApp(1 << 40)

	free, err := a.isFreeSpace(model.User{StorageLimit: 0}, 1<<30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !free {
		t.Error("a zero limit should not block an upload")
	}
}

func TestIsFreeSpaceNeverErrorsOnTheLimitItself(t *testing.T) {
	for _, limit := range []int64{0, 1, 1 << 40} {
		if _, err := storageApp(0).isFreeSpace(model.User{StorageLimit: limit}, 1); err != nil {
			t.Errorf("limit %d: unexpected error: %v", limit, err)
		}
	}
}

func TestIsFreeSpaceComparesBytes(t *testing.T) {
	const limit = 100

	for _, c := range []struct {
		name           string
		used, incoming int64
		want           bool
	}{
		{"well under", 10, 10, true},
		{"exactly fits", 90, 10, true},
		{"one byte over", 91, 10, false},
		{"already full", 100, 1, false},
		{"single byte into an empty drive", 0, 1, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			free, err := storageApp(c.used).isFreeSpace(
				model.User{StorageLimit: limit}, c.incoming,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if free != c.want {
				t.Errorf("used=%d incoming=%d limit=%d: got %v, want %v",
					c.used, c.incoming, limit, free, c.want)
			}
		})
	}
}

func TestIsFreeSpaceHandlesLimitsBelowAGigabyte(t *testing.T) {
	const tenMB = 10 * 1024 * 1024

	free, err := storageApp(0).isFreeSpace(model.User{StorageLimit: tenMB}, tenMB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !free {
		t.Error("a file exactly filling a 10 MB limit should be allowed")
	}

	free, _ = storageApp(0).isFreeSpace(model.User{StorageLimit: tenMB}, tenMB+1)
	if free {
		t.Error("a file one byte over a 10 MB limit should be refused")
	}
}

type fakeUsageByOwner struct {
	store.FileStore
	used map[string]int64
}

func (f *fakeUsageByOwner) GetStorageUsed(owner string) (int64, error) {
	return f.used[owner], nil
}

type fakeLimitUsers struct {
	store.UserStore
	users map[string]model.User
}

func (f *fakeLimitUsers) Get(id string) (*model.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, nil
	}
	return &u, nil
}

func TestSharedFolderUploadCountsAgainstFolderOwner(t *testing.T) {
	const mb = 1 << 20
	folderOwner := model.User{ID: "owner", StorageLimit: 10 * mb}
	uploader := model.User{ID: "guest", StorageLimit: 100 * mb}
	a := &App{Store: store.Store{
		File: &fakeUsageByOwner{used: map[string]int64{"owner": 9 * mb, "guest": 0}},
		User: &fakeLimitUsers{users: map[string]model.User{"owner": folderOwner}},
	}}

	shared := model.File{ID: "shared-folder", Owner: "owner", Type: "folder"}
	if got := newFileOwner(shared, uploader); got != "owner" {
		t.Fatalf("owner of a file in a shared folder = %q, want the folder owner", got)
	}
	if got := newFileOwner(model.File{ID: "drive", Owner: "", Type: "cloud#drive"}, uploader); got != "guest" {
		t.Fatalf("owner of a file at the drive root = %q, want the uploader", got)
	}

	free, err := a.ownerHasSpace(newFileOwner(shared, uploader), uploader, 2*mb)
	if err != nil {
		t.Fatalf("ownerHasSpace: %v", err)
	}
	if free {
		t.Fatal("2 MB into a folder whose owner has 1 MB left was allowed; the uploader's own limit was checked instead")
	}

	free, err = a.ownerHasSpace(newFileOwner(shared, uploader), uploader, mb)
	if err != nil || !free {
		t.Fatalf("1 MB into the owner's last 1 MB = %v, %v; want allowed", free, err)
	}
}
