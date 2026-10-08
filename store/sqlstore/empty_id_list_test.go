// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

// These IN clauses were built with strings.Repeat(",?", len(ids)-1), which
// panics on a negative count, and two take their slice straight from a request
// body, so "ID": [] took the handler down.
func TestEmptyIDListDoesNotPanic(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications", "file_share", "files", "favorites", "users")
	ctx := context.Background()

	calls := map[string]func() error{
		// api/files.go decodes ID []string and passes it to app.Trash.
		"File.GetByIDs": func() error {
			_, err := (&fileRepository{Db: db}).GetByIDs(model.User{ID: "user-1"}, nil)
			return err
		},
		// api/notifications.go decodes ID []string and passes it to ReadNotification.
		"Notifications.MarkAsRead": func() error {
			return (&notificationRepository{Db: db}).MarkAsRead(nil, "user-1")
		},
		"Notifications.MarkAsSent": func() error {
			return (&notificationRepository{Db: db}).MarkAsSent(nil)
		},
		"File.GetFavouriteIDs": func() error {
			_, err := (&fileRepository{Db: db}).GetFavouriteIDs("user-1", nil)
			return err
		},
		"File.GetMetadataEntriesForFiles": func() error {
			_, err := (&fileRepository{Db: db}).GetMetadataEntriesForFiles(nil)
			return err
		},
		"User.GetByIDs": func() error {
			_, err := (&userRepository{Db: db}).GetByIDs(nil)
			return err
		},
		"User.GetActiveByIDs": func() error {
			_, err := (&userRepository{Db: db}).GetActiveByIDs(ctx, nil)
			return err
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panicked on an empty list: %v", r)
				}
			}()
			if err := call(); err != nil {
				t.Errorf("returned an error on an empty list: %v", err)
			}
		})
	}
}
