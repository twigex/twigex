// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

type notificationRepository struct {
	Db *sql.DB
}

const notificationColumns = `id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at, deleted_at`

func NewNotificationRepository(Db *sql.DB) (*notificationRepository, error) {
	repo := &notificationRepository{}

	repo.Db = Db
	return repo, nil
}

func (n *notificationRepository) Create(sender model.User, receiver string, app string, notificationType string, item string, details map[string]interface{}) (*model.NotificationMessage, error) {
	t := time.Now().Unix()
	id := model.NewID()

	ps, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	_, err = n.Db.Exec(`INSERT INTO
		notifications(id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at, deleted_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, id, sender.ID, receiver, app, notificationType, item, ps, 0, t, t, t, 0)
	if err != nil {
		return nil, err
	}

	return &model.NotificationMessage{
		ID:               id,
		Sender:           sender.ID,
		Receiver:         receiver,
		App:              app,
		NotificationType: notificationType,
		ItemID:           item,
		Details:          details,
		ReadAt:           0,
		SentAt:           t,
		CreatedAt:        t,
		UpdatedAt:        t,
		DeletedAt:        0,
	}, nil
}

// CreateBulk inserts one identical notification per receiver in chunked
// multi-row statements (MySQL caps a statement at 65535 placeholders), returning
// the created messages in receiver order so the caller can deliver each.
func (n *notificationRepository) CreateBulk(sender model.User, receivers []string, app string, notificationType string, item string, details map[string]interface{}) ([]model.NotificationMessage, error) {
	if len(receivers) == 0 {
		return nil, nil
	}

	ps, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	t := time.Now().Unix()

	msgs := make([]model.NotificationMessage, 0, len(receivers))
	const cols = 12
	const perStmt = 1000 // 1000 * 12 = 12000 placeholders

	for start := 0; start < len(receivers); start += perStmt {
		end := start + perStmt
		if end > len(receivers) {
			end = len(receivers)
		}

		batch := receivers[start:end]

		placeholders := make([]string, 0, len(batch))
		args := make([]any, 0, len(batch)*cols)
		for _, r := range batch {
			id := model.NewID()
			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?,?,?)")
			args = append(args, id, sender.ID, r, app, notificationType, item, ps, 0, t, t, t, 0)
			msgs = append(msgs, model.NotificationMessage{
				ID:               id,
				Sender:           sender.ID,
				Receiver:         r,
				App:              app,
				NotificationType: notificationType,
				ItemID:           item,
				Details:          details,
				SentAt:           t,
				CreatedAt:        t,
				UpdatedAt:        t,
			})
		}

		_, err = n.Db.Exec(`INSERT INTO
			notifications(id, sender, receiver, app, notification_type, item_id, details, read_at, sent_at, created_at, updated_at, deleted_at)
			VALUES `+strings.Join(placeholders, ","), args...)
		if err != nil {
			return nil, err
		}
	}

	return msgs, nil
}

func (n *notificationRepository) Delete(id string, user string) error {
	_, err := n.Db.Exec(`DELETE FROM notifications WHERE id=? AND receiver=?`, id, user)
	if err != nil {
		return err
	}

	return nil
}

// notificationListLimit caps one response. Unread notifications fill it first
// at any age, so nothing unread is lost while the user is away; read ones from
// the last week fill whatever room is left.
const notificationListLimit = 500

func (n *notificationRepository) GetAllForReceiver(receiverID string) ([]model.NotificationMessage, error) {
	notifications, err := n.queryNotifications(
		`SELECT `+notificationColumns+` FROM notifications
		 WHERE receiver = ? AND read_at = 0
		 ORDER BY created_at DESC LIMIT ?`,
		receiverID, notificationListLimit)
	if err != nil {
		return nil, err
	}

	if room := notificationListLimit - len(notifications); room > 0 {
		read, err := n.queryNotifications(
			`SELECT `+notificationColumns+` FROM notifications
			 WHERE receiver = ? AND created_at > ? AND read_at <> 0
			 ORDER BY created_at DESC LIMIT ?`,
			receiverID, time.Now().Add(-7*24*time.Hour).Unix(), room)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, read...)
	}

	sort.SliceStable(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt > notifications[j].CreatedAt
	})

	return notifications, nil
}

func (n *notificationRepository) queryNotifications(query string, args ...any) ([]model.NotificationMessage, error) {
	results, err := n.Db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	notifications := make([]model.NotificationMessage, 0)

	for results.Next() {
		nm := model.NotificationMessage{}
		var details []byte

		err = results.Scan(&nm.ID, &nm.Sender, &nm.Receiver, &nm.App, &nm.NotificationType, &nm.ItemID,
			&details, &nm.ReadAt, &nm.SentAt, &nm.CreatedAt, &nm.UpdatedAt, &nm.DeletedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(details, &nm.Details)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, nm)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (n *notificationRepository) GetByID(id string) (*model.NotificationMessage, error) {
	result := n.Db.QueryRow(`SELECT `+notificationColumns+` FROM notifications WHERE id=?`, id)

	nm := model.NotificationMessage{}
	var details []byte

	err := result.Scan(&nm.ID, &nm.Sender, &nm.Receiver, &nm.App, &nm.NotificationType, &nm.ItemID,
		&details, &nm.ReadAt, &nm.SentAt, &nm.CreatedAt, &nm.UpdatedAt, &nm.DeletedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(details, &nm.Details)
	if err != nil {
		return nil, err
	}

	return &nm, nil
}

func (n *notificationRepository) MarkAsSent(messages []model.NotificationMessage) error {
	if len(messages) == 0 {
		return nil
	}

	args := make([]any, 0, len(messages)+1)
	args = append(args, time.Now().Unix())
	for _, m := range messages {
		args = append(args, m.ID)
	}

	_, err := n.Db.Exec(`UPDATE notifications SET sent_at=? WHERE id IN (`+sqlPlaceholders(len(messages))+`)`, args...)
	if err != nil {
		return err
	}

	return nil
}

func (n *notificationRepository) MarkAsRead(id []string, receiver string) error {
	if len(id) == 0 {
		return nil
	}

	args := make([]any, 0, len(id)+2)
	args = append(args, time.Now().Unix())
	for _, id := range id {
		args = append(args, id)
	}

	args = append(args, receiver)

	_, err := n.Db.Exec(`UPDATE notifications SET read_at=? WHERE id IN (`+sqlPlaceholders(len(id))+`) AND receiver=?`, args...)
	if err != nil {
		return err
	}

	return nil
}
