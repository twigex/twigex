// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func seedNotification(t *testing.T, id, receiver string) {
	t.Helper()
	now := time.Now().Unix()
	mustExec(t,
		`INSERT INTO notifications (id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at)
		 VALUES (?, 'sender', ?, 'chat', 'mention', 'item', '{}', 0, 0, ?, ?)`,
		id, receiver, now, now)
}

func notificationTimes(t *testing.T, repo *notificationRepository, id string) (readAt, sentAt int64) {
	t.Helper()
	n, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID %s: %v", id, err)
	}
	return n.ReadAt, n.SentAt
}

func TestMarkAsReadOnlyMarksReceiversNotifications(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications")
	repo := &notificationRepository{Db: db}

	seedNotification(t, "n-alice-1", "alice")
	seedNotification(t, "n-alice-2", "alice")
	seedNotification(t, "n-bob", "bob")

	if err := repo.MarkAsRead([]string{"n-alice-1", "n-alice-2", "n-bob"}, "alice"); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}

	for _, id := range []string{"n-alice-1", "n-alice-2"} {
		if readAt, _ := notificationTimes(t, repo, id); readAt == 0 {
			t.Errorf("%s not marked read", id)
		}
	}
	if readAt, _ := notificationTimes(t, repo, "n-bob"); readAt != 0 {
		t.Errorf("n-bob marked read by another receiver")
	}
}

func TestMarkAsReadTreatsQuotedReceiverAsData(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications")
	repo := &notificationRepository{Db: db}

	seedNotification(t, "n-bob", "bob")

	if err := repo.MarkAsRead([]string{"n-bob"}, `x" OR "1"="1`); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}

	if readAt, _ := notificationTimes(t, repo, "n-bob"); readAt != 0 {
		t.Errorf("n-bob marked read through a quoted receiver")
	}
}

func TestMarkAsSentMarksOnlyGivenNotifications(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications")
	repo := &notificationRepository{Db: db}

	seedNotification(t, "n-1", "alice")
	seedNotification(t, "n-2", "alice")

	if err := repo.MarkAsSent([]model.NotificationMessage{{ID: "n-1"}}); err != nil {
		t.Fatalf("MarkAsSent: %v", err)
	}

	if _, sentAt := notificationTimes(t, repo, "n-1"); sentAt == 0 {
		t.Errorf("n-1 not marked sent")
	}
	if _, sentAt := notificationTimes(t, repo, "n-2"); sentAt != 0 {
		t.Errorf("n-2 marked sent")
	}
}

func seedNotificationAt(t *testing.T, id, receiver string, createdAt, readAt int64) {
	t.Helper()
	mustExec(t,
		`INSERT INTO notifications (id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at)
		 VALUES (?, 'sender', ?, 'chat', 'mention', 'item', '{}', ?, 0, ?, ?)`,
		id, receiver, readAt, createdAt, createdAt)
}

func notificationIDs(list []model.NotificationMessage) string {
	ids := make([]string, len(list))
	for i, n := range list {
		ids[i] = n.ID
	}
	return strings.Join(ids, ",")
}

func TestGetAllForReceiverKeepsOldUnread(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications")
	repo := &notificationRepository{Db: db}
	day := int64(24 * 60 * 60)
	now := time.Now().Unix()

	seedNotificationAt(t, "unread-20d", "alice", now-20*day, 0)
	seedNotificationAt(t, "read-20d", "alice", now-20*day, now-19*day)
	seedNotificationAt(t, "read-2d", "alice", now-2*day, now-day)
	seedNotificationAt(t, "unread-1d", "alice", now-day, 0)
	seedNotificationAt(t, "bob-unread", "bob", now, 0)

	got, err := repo.GetAllForReceiver("alice")
	if err != nil {
		t.Fatalf("GetAllForReceiver: %v", err)
	}
	if ids := notificationIDs(got); ids != "unread-1d,read-2d,unread-20d" {
		t.Fatalf("ids = %s, want unread-1d,read-2d,unread-20d", ids)
	}
}

func TestGetAllForReceiverCapsWithUnreadFirst(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "notifications")
	repo := &notificationRepository{Db: db}
	now := time.Now().Unix()

	unread := notificationListLimit + 5
	values := make([]string, 0, unread)
	args := make([]any, 0, unread*2)
	for i := 0; i < unread; i++ {
		values = append(values, "(?, 'sender', 'alice', 'chat', 'mention', 'item', '{}', 0, 0, ?, 0)")
		args = append(args, fmt.Sprintf("unread-%04d", i), now-int64(unread-i))
	}
	mustExec(t, `INSERT INTO notifications (id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at)
		VALUES `+strings.Join(values, ","), args...)
	seedNotificationAt(t, "read-recent", "alice", now, now)

	got, err := repo.GetAllForReceiver("alice")
	if err != nil {
		t.Fatalf("GetAllForReceiver: %v", err)
	}
	if len(got) != notificationListLimit {
		t.Fatalf("got %d notifications, want %d", len(got), notificationListLimit)
	}
	if got[0].ID != fmt.Sprintf("unread-%04d", unread-1) || got[len(got)-1].ID != "unread-0005" {
		t.Fatalf("got %s..%s, want the newest %d unread", got[0].ID, got[len(got)-1].ID, notificationListLimit)
	}
}
