// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/internal/tmail"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
	"github.com/wneessen/go-mail"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var appNotifications = map[string]string{
	model.NOTIFICATION_FILE_CREATE:                 "notify_app_file_or_folder_created",
	model.NOTIFICATION_FILE_RENAME:                 "notify_app_file_or_folder_renamed",
	model.NOTIFICATION_FILE_SHARE:                  "notify_app_file_or_folder_shared",
	model.NOTIFICATION_FILE_DELETE:                 "notify_app_file_or_folder_deleted",
	model.NOTIFICATION_FILE_RESTORE:                "notify_app_file_or_folder_restored",
	model.NOTIFICATION_FILE_DOWNLOAD:               "notify_app_file_or_folder_downloaded",
	model.NOTIFICATION_FILE_PUBLIC_DOWNLOAD:        "notify_app_file_or_folder_public_downloaded",
	model.NOTIFICATION_PROJECT_TASK_CREATED:        "notify_app_task_created",
	model.NOTIFICATION_PROJECT_DELETED:             "notify_app_project_deleted",
	model.NOTIFICATION_PROJECT_INVITE:              "notify_app_project_invited",
	model.NOTIFICATION_PROJECT_TASK_DELETED:        "notify_app_task_deleted",
	model.NOTIFICATION_PROJECT_TASK_ASSIGNED:       "notify_app_task_assigned",
	model.NOTIFICATION_PROJECT_TASK_STATUS_CHANGED: "notify_app_assigned_task_status_changed",
	model.NOTIFICATION_TASK_COMMENT:                "notify_app_task_comment",
	model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED:   "notify_app_meeting_scheduled",
	model.NOTIFICATION_CHANNEL_MEETING_UPDATED:     "notify_app_meeting_updated",
	model.NOTIFICATION_CHANNEL_MEETING_CANCELLED:   "notify_app_meeting_cancelled",
	model.NOTIFICATION_CHANNEL_MENTION:             "notify_app_channel_mention",
}

var emailNotifications = map[string]string{
	model.NOTIFICATION_FILE_CREATE:                 "notify_email_file_or_folder_created",
	model.NOTIFICATION_FILE_RENAME:                 "notify_email_file_or_folder_renamed",
	model.NOTIFICATION_FILE_SHARE:                  "notify_email_file_or_folder_shared",
	model.NOTIFICATION_FILE_DELETE:                 "notify_email_file_or_folder_deleted",
	model.NOTIFICATION_FILE_RESTORE:                "notify_email_file_or_folder_restored",
	model.NOTIFICATION_FILE_DOWNLOAD:               "notify_email_file_or_folder_downloaded",
	model.NOTIFICATION_FILE_PUBLIC_DOWNLOAD:        "notify_email_file_or_folder_public_downloaded",
	model.NOTIFICATION_PROJECT_TASK_CREATED:        "notify_email_task_created",
	model.NOTIFICATION_PROJECT_DELETED:             "notify_email_project_deleted",
	model.NOTIFICATION_PROJECT_INVITE:              "notify_email_project_invited",
	model.NOTIFICATION_PROJECT_TASK_DELETED:        "notify_email_task_deleted",
	model.NOTIFICATION_PROJECT_TASK_ASSIGNED:       "notify_email_task_assigned",
	model.NOTIFICATION_PROJECT_TASK_STATUS_CHANGED: "notify_email_assigned_task_status_changed",
	model.NOTIFICATION_TASK_COMMENT:                "notify_email_task_comment",
	model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED:   "notify_email_meeting_scheduled",
	model.NOTIFICATION_CHANNEL_MEETING_UPDATED:     "notify_email_meeting_updated",
	model.NOTIFICATION_CHANNEL_MEETING_CANCELLED:   "notify_email_meeting_cancelled",
	model.NOTIFICATION_CHANNEL_MENTION:             "notify_email_channel_mention",
}

var pushNotificationTypes = map[string]struct{}{
	model.NOTIFICATION_CHANNEL_UPDATE:        {},
	model.NOTIFICATION_CHANNEL_POST:          {},
	model.NOTIFICATION_CHANNEL_POST_REACTION: {},
	model.NOTIFICATION_CHANNEL_POST_UPDATE:   {},
	model.NOTIFICATION_CHANNEL_POST_DELETE:   {},
	model.NOTIFICATION_CHANNEL_TYPING:        {},
	model.NOTIFICATION_CHANNEL_LINK_PREVIEW:  {},
	model.NOTIFICATION_CHANNEL_PRESENCE:      {},
}

type TypingMessage struct {
	Event string `json:"event"` // "typing" or "stop_typing"
	Data  struct {
		ChannelID string `json:"channel_id"`
	} `json:"data"`
}

type Hub struct {
	mu sync.RWMutex

	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	app        *App
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	user string
	send chan []byte
}

func newHub() *Hub {
	hub := Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	go hub.run()

	return &hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			_, registered := h.clients[client]
			if registered {
				delete(h.clients, client)
				close(client.send)
			}

			h.mu.Unlock()

			// Outside the lock: SetUserStatus broadcasts, which takes it again.
			if registered {
				h.app.SetUserStatus(client.user, model.StatusOffline, false)
			}
		}
	}
}

func (h *Hub) send(message []byte, match func(*Client) bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if match != nil && !match(client) {
			continue
		}

		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) connectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]bool, len(h.clients))
	users := make([]string, 0, len(h.clients))

	for client := range h.clients {
		if seen[client.user] {
			continue
		}

		seen[client.user] = true
		users = append(users, client.user)
	}

	return users
}

func (h *Hub) SendEvent(message model.WebsocketEvent) error {
	m, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.send(m, nil)

	return nil
}

func (c *Client) readMessage() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				tlog.Errorw("Unexpected WebSocket close",
					"user_id", c.user,
					"error", err,
				)
			}

			break
		}

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))

		var msg TypingMessage
		if err = json.Unmarshal(message, &msg); err != nil {
			tlog.Errorw("Failed to unmarshal WebSocket message",
				"user_id", c.user,
				"error", err,
			)
			continue
		}

		switch msg.Event {
		case model.NOTIFICATION_CHANNEL_TYPING:
			if err = c.hub.app.SendTypingEvent(msg.Data.ChannelID, c.user); err != nil {
				tlog.Errorw("Failed to send typing event",
					"user_id", c.user,
					"channel_id", msg.Data.ChannelID,
					"error", err,
				)
			}
		case model.NOTIFICATION_CHANNEL_PRESENCE:
			c.hub.app.SetUserStatus(c.user, model.StatusOnline, false)
		}
	}
}

func (c *Client) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if !ok {
				// The hub closed the channel. Error ignored because the
				// connection is closed below whether the close frame was sent
				// or not.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if _, err := w.Write(message); err != nil {
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					return
				}

				if _, err := w.Write(<-c.send); err != nil {
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (a *App) SendStatusEvent(status model.UserStatus) error {
	data := map[string]any{
		"status": status,
	}

	newEvent := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_PRESENCE,
		App:   model.AppChat,
		Data:  data,
	}

	err := a.Server.NotificationHub.SendEvent(newEvent)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) SendTypingEvent(channelID, userID string) error {
	ok, err := a.Store.Channels.IsMember(channelID, userID)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("not a member of channel %s", channelID)
	}

	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil {
		return err
	}

	newEvent := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_TYPING,
		App:   model.AppChat,
		Data: map[string]any{
			"channel_id": channelID,
			"user_id":    userID,
		},
	}

	for _, v := range members {
		if v.UserID == userID {
			continue
		}

		user, err := a.Store.User.Get(v.UserID)
		if err != nil {
			return err
		}

		if user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		go a.notifyUser(*user, newEvent)
	}

	return nil
}

// notificationAllowed reports whether the user wants the app (push) and email
// notification for a type. It honors saved preferences and falls back to the
// DefaultNotifications defaults for keys the user never set, so existing users
// keep receiving notifications until they explicitly opt out.
func (a *App) notificationAllowed(user model.User, notificationType string) (app bool, email bool) {
	notifications, err := a.Store.Preferences.GetNotifications(user.ID, model.PreferencesCategoryNotifications)
	if err != nil {
		tlog.Errorw("Failed to retrieve notification preferences", "user_id", user.ID, "error", err)
		return false, false
	}

	combined := append(generateMissingNotifications(user.ID, notifications), notifications...)

	appKey := appNotifications[notificationType]
	emailKey := emailNotifications[notificationType]
	for _, v := range combined {
		if appKey != "" && v.Name == appKey && v.Value == "true" {
			app = true
		}

		if emailKey != "" && v.Name == emailKey && v.Value == "true" {
			email = true
		}
	}

	return app, email
}

func (a *App) notifyUser(user model.User, message model.WebsocketEvent) {
	notificationType := message.Event

	if _, ok := pushNotificationTypes[notificationType]; ok {
		if err := a.sendPushNotifications(user, message); err != nil {
			tlog.Errorw("Failed to send push notification",
				"user_id", user.ID,
				"event", notificationType,
				"error", err,
			)
		}

		return
	}

	app, email := a.notificationAllowed(user, notificationType)
	if app {
		if err := a.sendPushNotifications(user, message); err != nil {
			tlog.Errorw("Failed to send push notification",
				"user_id", user.ID,
				"event", notificationType,
				"error", err,
			)
		}
	}

	if email {
		if err := a.sendEmailNotification(user, message); err != nil {
			tlog.Errorw("Failed to send email notification",
				"user_id", user.ID,
				"event", notificationType,
				"error", err,
			)
		}
	}
}

func (a *App) sendPushNotifications(user model.User, message model.WebsocketEvent) error {
	m, err := json.Marshal(message)
	if err != nil {
		return err
	}

	a.Server.NotificationHub.send(m, func(c *Client) bool { return c.user == user.ID })

	return nil
}

func (a *App) sendEmailNotification(user model.User, message model.WebsocketEvent) error {
	rawMsg, ok := message.Data["message"]
	if !ok {
		return errors.New("missing message in websocket data")
	}

	jsonData, err := json.Marshal(rawMsg)
	if err != nil {
		return err
	}

	var notification model.NotificationMessage
	if err = json.Unmarshal(jsonData, &notification); err != nil {
		return err
	}

	return a.generateEmail(notification)
}

func (a *App) GetUserNotifications(receiverID string) ([]model.NotificationMessage, *model.AppError) {
	notifications, err := a.Store.Notifications.GetAllForReceiver(receiverID)
	if err != nil {
		tlog.Errorw("Failed to retrieve notifications",
			"receiver_id", receiverID,
			"error", err,
		)
		return nil, model.NewAppError("notification.retrieval_failed", http.StatusInternalServerError)
	}

	return notifications, nil
}

func (a *App) ResolveNotification(notificationID string, user model.User) (*map[string]any, *model.AppError) {
	notification, err := a.Store.Notifications.GetByID(notificationID)
	if err != nil {
		tlog.Errorw("Failed to retrieve notification",
			"notification_id", notificationID,
			"error", err,
		)
		return nil, model.NewAppError("notification.not_found", http.StatusInternalServerError)
	}

	if notification.Receiver != user.ID {
		return nil, model.NewAppError("notification.forbidden", http.StatusForbidden)
	}

	if notification.ReadAt == 0 {
		if err = a.Store.Notifications.MarkAsRead([]string{notificationID}, user.ID); err != nil {
			tlog.Errorw("Failed to mark notification as read",
				"notification_id", notificationID,
				"user_id", user.ID,
				"error", err,
			)
			return nil, model.NewAppError("notification.read_failed", http.StatusInternalServerError)
		}
	}

	data := map[string]any{
		"app": notification.App,
	}

	switch notification.App {
	case model.AppProjects:
		data["status"] = "ok"

	case model.AppChat:
		data["status"] = "ok"
		data["item"] = notification.ItemID

		// Meeting invites/updates deep-link to the meeting (which a non-member can open) rather than the channel.
		switch notification.NotificationType {
		case model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED, model.NOTIFICATION_CHANNEL_MEETING_UPDATED:
			if id, ok := notification.Details["meeting_id"].(string); ok && id != "" {
				data["redirect"] = "meeting"
				data["item"] = id
			}
		}

	case model.AppFiles:
		file, appErr := a.HasPermission(notification.ItemID, user)
		if appErr != nil {
			data["status"] = "no_access"
			return &data, nil
		}

		data["status"] = "ok"
		data["item"] = notification.ItemID

		if notification.NotificationType == model.NOTIFICATION_FILE_SHARE {
			data["redirect"] = "shared"
			return &data, nil
		}

		data["redirect"] = "files"

		if parentFile, appErr := a.HasPermission(file.Parent, user); appErr == nil {
			data["parent_id"] = parentFile.ID
			if parentFile.Type == "cloud#drive" {
				data["redirect"] = "shared"
			}
		} else {
			data["redirect"] = "shared"
		}
	}

	return &data, nil
}

func (a *App) ReadNotification(status bool, id []string, user string) *model.AppError {
	if !status {
		return nil
	}

	if err := a.Store.Notifications.MarkAsRead(id, user); err != nil {
		tlog.Errorw("Failed to mark notifications as read",
			"user_id", user,
			"error", err,
		)
		return model.NewAppError("notification.read_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteNotification(id string, user string) *model.AppError {
	if err := a.Store.Notifications.Delete(id, user); err != nil {
		tlog.Errorw("Failed to delete notification",
			"notification_id", id,
			"user_id", user,
			"error", err,
		)
		return model.NewAppError("notification.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

var (
	mentionFencedCodeRe = regexp.MustCompile("(?s)```.*?```")
	mentionInlineCodeRe = regexp.MustCompile("`[^`]*`")
	mentionLinkRe       = regexp.MustCompile(`\[[^\]]*\]\([^)]*\)`)
	// The ID branch comes first so a malformed <@abc> falls through to the name
	// branch instead of being dropped.
	mentionHandleRe = regexp.MustCompile(`<@([0-9a-f]{32})>|@([a-zA-Z0-9](?:[a-zA-Z0-9._-]*[a-zA-Z0-9])?)`)
)

// Exactly one field is set. Comparable so callers can dedupe and diff it.
type mention struct {
	ID     string
	Handle string
}

// extractMentions returns the distinct mentions in a message, matching the
// renderer: whole-word handles, ignoring anything inside code spans, code
// fences, or markdown links so an @name there is not treated as a mention.
func extractMentions(message string) []mention {
	stripped := mentionFencedCodeRe.ReplaceAllString(message, " ")
	stripped = mentionInlineCodeRe.ReplaceAllString(stripped, " ")
	stripped = mentionLinkRe.ReplaceAllString(stripped, " ")

	seen := make(map[mention]struct{})
	mentions := make([]mention, 0)
	for _, m := range mentionHandleRe.FindAllStringSubmatch(stripped, -1) {
		ref := mention{ID: m[1], Handle: m[2]}
		if _, ok := seen[ref]; ok {
			continue
		}

		seen[ref] = struct{}{}
		mentions = append(mentions, ref)
	}

	return mentions
}

// PostMentions is the set of channel members a post mentions, resolved once so
// the same result drives the post's websocket payload (which decides the alert
// sound), the bell entries, and the mention badges, keeping them in agreement.
type PostMentions struct {
	UserIDs   []string // members whose @handle appears (may include the author)
	Broadcast bool     // message contains @all or @here
}

func (a *App) resolvePostMentions(post model.Post) PostMentions {
	return a.resolveHandles(post.ChannelID, extractMentions(post.Message))
}

func (a *App) resolveHandles(channelID string, mentions []mention) PostMentions {
	res := PostMentions{}
	if len(mentions) == 0 {
		return res
	}

	handles := make([]string, 0, len(mentions))
	ids := make([]string, 0, len(mentions))
	for _, m := range mentions {
		if m.ID != "" {
			ids = append(ids, m.ID)
			continue
		}

		if m.Handle == "all" || m.Handle == "here" {
			res.Broadcast = true
		}

		handles = append(handles, m.Handle)
	}

	if len(handles) == 0 && len(ids) == 0 {
		return res
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	// Both forms can name the same person, so dedupe on the resolved user.
	seen := make(map[string]struct{})
	add := func(members []model.User) {
		for _, m := range members {
			if _, ok := seen[m.ID]; ok {
				continue
			}

			seen[m.ID] = struct{}{}
			res.UserIDs = append(res.UserIDs, m.ID)
		}
	}

	if len(handles) > 0 {
		members, err := a.Store.Channels.GetMembersByUsernames(ctx, channelID, handles)
		if err != nil {
			tlog.Errorw("Failed to resolve mentioned channel members", "channel_id", channelID, "error", err)
		} else {
			add(members)
		}
	}

	if len(ids) > 0 {
		members, err := a.Store.Channels.GetMembersByIDs(ctx, channelID, ids)
		if err != nil {
			tlog.Errorw("Failed to resolve mentioned channel members by id", "channel_id", channelID, "error", err)
		} else {
			add(members)
		}
	}

	return res
}

// attachMentionUsers fills each post's Mentions map, following Reply chains.
// Users resolve globally rather than per channel, so a mention of someone who
// has since left it stays readable.
func (a *App) attachMentionUsers(posts ...*model.Post) {
	targets := make([]*model.Post, 0, len(posts))
	ids := make([]string, 0)
	seen := make(map[string]struct{})

	var collect func(p *model.Post)
	collect = func(p *model.Post) {
		if p == nil {
			return
		}

		targets = append(targets, p)
		for _, m := range extractMentions(p.Message) {
			if m.ID == "" {
				continue
			}

			if _, ok := seen[m.ID]; ok {
				continue
			}

			seen[m.ID] = struct{}{}
			ids = append(ids, m.ID)
		}

		collect(p.Reply)
	}

	for _, p := range posts {
		collect(p)
	}

	if len(ids) == 0 {
		return
	}

	users, err := a.Store.User.GetByIDs(ids)
	if err != nil {
		tlog.Errorw("Failed to resolve mentioned users for display", "error", err)
		return
	}

	resolved := make(map[string]model.MentionUser, len(users))
	for _, u := range users {
		resolved[u.ID] = model.MentionUser{Username: u.Username, Name: u.Name, LastName: u.LastName}
	}

	for _, p := range targets {
		for _, m := range extractMentions(p.Message) {
			u, ok := resolved[m.ID]
			if m.ID == "" || !ok {
				continue
			}

			if p.Mentions == nil {
				p.Mentions = make(map[string]model.MentionUser)
			}

			p.Mentions[m.ID] = u
		}
	}
}

func (a *App) attachResponseMentionUsers(res *model.PostResponse) {
	if res == nil {
		return
	}

	posts := make([]*model.Post, 0, len(res.Posts)+len(res.ReplyPosts))
	for i := range res.Posts {
		posts = append(posts, &res.Posts[i])
	}

	for i := range res.ReplyPosts {
		posts = append(posts, &res.ReplyPosts[i])
	}

	a.attachMentionUsers(posts...)
}

// addedMentions returns mentions present in current but not previous, so an
// edit only ever acts on newly added mentions and never re-notifies existing ones.
func addedMentions(previous, current string) []mention {
	existing := make(map[mention]struct{})
	for _, m := range extractMentions(previous) {
		existing[m] = struct{}{}
	}

	added := make([]mention, 0)
	for _, m := range extractMentions(current) {
		if _, ok := existing[m]; !ok {
			added = append(added, m)
		}
	}

	return added
}

// ApplyMentions gives each mentioned member (never the author) a bell entry and
// raises the mention badge: for @all/@here every member in one statement,
// otherwise only the targeted members.
func (a *App) ApplyMentions(post model.Post, author model.User, mentions PostMentions) {
	targeted := make([]string, 0, len(mentions.UserIDs))
	for _, id := range mentions.UserIDs {
		if id != author.ID {
			targeted = append(targeted, id)
		}
	}

	a.sendMentionBells(post, author, targeted)

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	if mentions.Broadcast {
		if err := a.Store.Channels.IncrementAllMentionCounts(ctx, post.ChannelID, author.ID); err != nil {
			tlog.Errorw("Failed to raise broadcast mention badge", "channel_id", post.ChannelID, "error", err)
		}

		return
	}

	if err := a.Store.Channels.IncrementMentionCount(ctx, post.ChannelID, targeted); err != nil {
		tlog.Errorw("Failed to raise mention badge", "channel_id", post.ChannelID, "error", err)
	}
}

func (a *App) sendMentionBells(post model.Post, author model.User, memberIDs []string) {
	if len(memberIDs) == 0 {
		return
	}

	channelName := ""
	if channel, err := a.Store.Channels.Get(post.ChannelID); err == nil && channel != nil {
		channelName = channel.DisplayName
	}

	for _, id := range memberIDs {
		data := map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_MENTION,
			"channel_id":       post.ChannelID,
			"post_id":          post.ID,
			"channelName":      channelName,
		}

		// item is the channel id so clicking the bell entry opens the channel.
		m, err := a.Store.Notifications.Create(author, id, model.AppChat, model.NOTIFICATION_CHANNEL_MENTION, post.ChannelID, data)
		if err != nil {
			tlog.Errorw("Failed to create mention notification", "user_id", id, "post_id", post.ID, "error", err)
			continue
		}

		data["message"] = m.ToMap()
		go a.notifyUser(model.User{ID: id}, model.WebsocketEvent{
			Event: model.NOTIFICATION_CHANNEL_MENTION,
			App:   model.AppChat,
			Data:  data,
		})
	}
}

func (a *App) createProjectNotification(workspaceID string, creator model.User, notificationType string, props map[string]any) {
	data := map[string]any{
		"app":              model.AppProjects,
		"workspace_id":     workspaceID,
		"notificationType": notificationType,
	}

	for key, value := range props {
		data[key] = value
	}

	if userID, ok := props["userID"].(string); ok && userID != "" {
		user, err := a.Store.User.Get(userID)
		if err != nil {
			tlog.Errorw("Failed to retrieve user for notification",
				"user_id", userID,
				"workspace_id", workspaceID,
				"notification_type", notificationType,
				"error", err,
			)
			return
		}

		if user == nil {
			return
		}

		if user.ID != creator.ID && a.SessionHasPermission(*user, model.ProjectSectionPermissions.PermissionViewProjects) {
			m, err := a.Store.Notifications.Create(creator, user.ID, model.AppProjects, notificationType, workspaceID, data)
			if err != nil {
				tlog.Errorw("Failed to create notification",
					"user_id", userID,
					"workspace_id", workspaceID,
					"notification_type", notificationType,
					"error", err,
				)
				return
			}

			data["message"] = m.ToMap()
			go a.notifyUser(*user, model.WebsocketEvent{
				Event: notificationType,
				App:   model.AppProjects,
				Data:  data,
			})
		}

		return
	}

	wsMem, err := a.Store.Workspace.GetMembers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace members",
			"workspace_id", workspaceID,
			"notification_type", notificationType,
			"error", err,
		)
		return
	}

	userIDs := make([]string, 0, len(wsMem))
	for _, member := range wsMem {
		userIDs = append(userIDs, member.UserID)
	}

	allUsers, err := a.Store.User.GetByIDs(userIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve users",
			"workspace_id", workspaceID,
			"notification_type", notificationType,
			"error", err,
		)
		return
	}

	for _, user := range allUsers {
		if user.ID == creator.ID {
			continue
		}

		if !a.SessionHasPermission(user, model.ProjectSectionPermissions.PermissionViewProjects) {
			continue
		}

		m, err := a.Store.Notifications.Create(creator, user.ID, model.AppProjects, notificationType, workspaceID, data)
		if err != nil {
			tlog.Errorw("Failed to create notification",
				"user_id", user.ID,
				"workspace_id", workspaceID,
				"notification_type", notificationType,
				"error", err,
			)
			continue
		}

		data["message"] = m.ToMap()
		go a.notifyUser(user, model.WebsocketEvent{
			Event: notificationType,
			App:   model.AppProjects,
			Data:  data,
		})
	}
}

func (a *App) CreateFileDownloadNotification(file model.File, creator model.User, notificationType string, props map[string]any) {
	owner, err := a.Store.User.Get(file.Owner)
	if err != nil {
		tlog.Errorw("Failed to retrieve file owner",
			"file_id", file.ID,
			"owner_id", file.Owner,
			"error", err,
		)
		return
	}

	if owner == nil {
		return
	}

	if owner.ID == creator.ID {
		return
	}

	if !a.SessionHasPermission(*owner, model.FilePermissions.PermissionViewFiles) {
		return
	}

	m, err := a.Store.Notifications.Create(creator, owner.ID, model.AppFiles, notificationType, file.ID, map[string]any{
		"filename":         file.DisplayName,
		"notificationType": notificationType,
	})
	if err != nil {
		tlog.Errorw("Failed to create file download notification",
			"file_id", file.ID,
			"owner_id", owner.ID,
			"error", err,
		)
		return
	}

	go a.notifyUser(*owner, model.WebsocketEvent{
		Event: notificationType,
		App:   model.AppFiles,
		Data: map[string]any{
			"filename":         file.DisplayName,
			"notificationType": notificationType,
			"message":          m.ToMap(),
		},
	})
}

func (a *App) CreateFileNotification(file model.File, creator model.User, notificationType string, props map[string]any) {
	all, err := a.Store.User.GetSharedUsers(file.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve shared users",
			"file_id", file.ID,
			"error", err,
		)
		return
	}

	og, err := a.Store.User.Get(file.Owner)
	if err != nil {
		tlog.Errorw("Failed to retrieve file owner",
			"file_id", file.ID,
			"owner_id", file.Owner,
			"error", err,
		)
		return
	}

	if og == nil {
		return
	}

	if creator.ID != og.ID {
		all = append(all, *og)
	}

	if notificationType == model.NOTIFICATION_FILE_PUBLIC_DOWNLOAD {
		data := map[string]any{
			"filename":         file.DisplayName,
			"notificationType": notificationType,
		}

		m, err := a.Store.Notifications.Create(creator, creator.ID, model.AppFiles, notificationType, file.ID, data)
		if err != nil {
			tlog.Errorw("Failed to create file notification",
				"file_id", file.ID,
				"user_id", creator.ID,
				"error", err,
			)
			return
		}

		data["message"] = m.ToMap()
		go a.notifyUser(creator, model.WebsocketEvent{
			Event: notificationType,
			App:   model.AppFiles,
			Data:  data,
		})

		return
	}

	a.dispatchFileNotifications(file, creator, notificationType, all)
}

// CreateFileShareNotification notifies only the recipients just granted access,
// not everyone who already had it, so sharing to a group does not re-notify
// unrelated users who were shared earlier.
func (a *App) CreateFileShareNotification(file model.File, creator model.User, recipientIDs []string) {
	seen := make(map[string]bool, len(recipientIDs))
	ids := make([]string, 0, len(recipientIDs))
	for _, id := range recipientIDs {
		if id == creator.ID || seen[id] {
			continue
		}

		seen[id] = true
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return
	}

	recipients, err := a.Store.User.GetActiveByIDs(context.Background(), ids)
	if err != nil {
		tlog.Errorw("Failed to load share notification recipients",
			"file_id", file.ID,
			"error", err,
		)
		return
	}

	a.dispatchFileNotifications(file, creator, model.NOTIFICATION_FILE_SHARE, recipients)
}

func (a *App) dispatchFileNotifications(file model.File, creator model.User, notificationType string, recipients []model.User) {
	// Permission depends only on the role string, so memoize it: a large group
	// fan-out then costs one role lookup per distinct role, not per user.
	canView := make(map[string]bool)
	viewAllowed := func(u model.User) bool {
		if allowed, ok := canView[u.Role]; ok {
			return allowed
		}

		allowed := a.SessionHasPermission(u, model.FilePermissions.PermissionViewFiles)
		canView[u.Role] = allowed
		return allowed
	}

	eligible := make([]model.User, 0, len(recipients))
	ids := make([]string, 0, len(recipients))
	for _, v := range recipients {
		if v.ID == creator.ID || !viewAllowed(v) {
			continue
		}

		eligible = append(eligible, v)
		ids = append(ids, v.ID)
	}

	if len(eligible) == 0 {
		return
	}

	data := map[string]any{
		"filename":         file.DisplayName,
		"notificationType": notificationType,
	}
	if notificationType == model.NOTIFICATION_FILE_RENAME {
		data["old_name"] = file.OldName
	}

	msgs, err := a.Store.Notifications.CreateBulk(creator, ids, model.AppFiles, notificationType, file.ID, data)
	if err != nil {
		tlog.Errorw("Failed to create file notifications",
			"file_id", file.ID,
			"error", err,
		)
		return
	}

	go func() {
		const maxConcurrent = 16
		sem := make(chan struct{}, maxConcurrent)
		for i := range eligible {
			payload := map[string]any{
				"filename":         file.DisplayName,
				"notificationType": notificationType,
				"message":          msgs[i].ToMap(),
			}
			if notificationType == model.NOTIFICATION_FILE_RENAME {
				payload["old_name"] = file.OldName
			}

			sem <- struct{}{}
			go func(u model.User, data map[string]any) {
				defer func() {
					<-sem
				}()
				a.notifyUser(u, model.WebsocketEvent{
					Event: notificationType,
					App:   model.AppFiles,
					Data:  data,
				})
			}(eligible[i], payload)
		}
	}()
}

func (a *App) CreateChannelUpdateNotification(channel model.Channel, creator model.User, notificationType string) {
	members, err := a.Store.Channels.GetMembers(channel.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel members",
			"channel_id", channel.ID,
			"error", err,
		)
		return
	}

	event := model.WebsocketEvent{
		Event: notificationType,
		App:   model.AppChat,
		Data:  map[string]any{"channel": channel},
	}

	for _, v := range members {
		if v.UserID == creator.ID {
			continue
		}

		user, err := a.Store.User.Get(v.UserID)
		if err != nil {
			tlog.Errorw("Failed to retrieve channel member",
				"channel_id", channel.ID,
				"user_id", v.UserID,
				"error", err,
			)
			return
		}

		if user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		go a.notifyUser(*user, event)
	}
}

func (a *App) CreatePostReactionNotification(ctx context.Context, postID string, creator model.User, notificationType string) {
	post, err := a.Store.Posts.Get(ctx, postID)
	if err != nil {
		tlog.Errorw("Failed to retrieve post",
			"post_id", postID,
			"error", err,
		)
		return
	}

	members, err := a.Store.Channels.GetMembers(post.ChannelID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel members",
			"channel_id", post.ChannelID,
			"error", err,
		)
		return
	}

	a.attachMentionUsers(post)

	event := model.WebsocketEvent{
		Event: notificationType,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"post":             post,
			"notificationType": notificationType,
			"sender_username":  creator.Username,
		},
	}

	for _, v := range members {
		if v.UserID == creator.ID {
			continue
		}

		user, err := a.Store.User.Get(v.UserID)
		if err != nil {
			tlog.Errorw("Failed to retrieve channel member",
				"channel_id", post.ChannelID,
				"user_id", v.UserID,
				"error", err,
			)
			return
		}

		if user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		go a.notifyUser(*user, event)
	}
}

func (a *App) CreatePostNotification(ctx context.Context, post model.Post, creator model.User, notificationType string, mentions PostMentions) {
	members, err := a.Store.Channels.GetMembers(post.ChannelID)
	if err != nil {
		tlog.Errorw("Failed to retrieve channel members",
			"channel_id", post.ChannelID,
			"error", err,
		)
		return
	}

	a.attachMentionUsers(&post)

	data := map[string]any{
		"app":              model.AppChat,
		"post":             post,
		"notificationType": notificationType,
		"mentions":         mentions.UserIDs,
		"mention_all":      mentions.Broadcast,
		"sender_username":  creator.Username,
	}

	if post.ReplyID != "" {
		replyPost, err := a.Store.Posts.Get(ctx, post.ReplyID)
		if err != nil {
			tlog.Errorw("Failed to retrieve reply post",
				"post_id", post.ID,
				"reply_post_id", post.ReplyID,
				"error", err,
			)
			return
		}

		if !replyPost.HasSameChannel(post.ChannelID) {
			tlog.Errorw("Reply post belongs to a different channel",
				"post_id", post.ID,
				"reply_post_id", post.ReplyID,
				"channel_id", post.ChannelID,
			)
			return
		}

		a.attachMentionUsers(replyPost)
		data["reply_post"] = replyPost
	}

	event := model.WebsocketEvent{
		Event: notificationType,
		App:   model.AppChat,
		Data:  data,
	}

	for _, v := range members {
		user, err := a.Store.User.Get(v.UserID)
		if err != nil {
			tlog.Errorw("Failed to retrieve channel member",
				"channel_id", post.ChannelID,
				"user_id", v.UserID,
				"error", err,
			)
			return
		}

		if user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		go a.notifyUser(*user, event)
	}
}

func (a *App) UpdatePostLinkMetadata(channelID, postID string, metadata model.LinkMetadata) error {
	hydrateLinkPreviewImage(&metadata)

	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil {
		return err
	}

	event := model.WebsocketEvent{
		Event: "link_preview_update",
		App:   model.AppChat,
		Data: map[string]any{
			"app":           model.AppChat,
			"channel_id":    channelID,
			"post_id":       postID,
			"link_metadata": metadata.Data,
		},
	}

	for _, v := range members {
		user, err := a.Store.User.Get(v.UserID)
		if err != nil {
			return err
		}

		if user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		go a.notifyUser(*user, event)
	}

	return nil
}

func (a *App) newClient(conn *websocket.Conn, userID string) *Client {
	client := &Client{hub: a.Server.NotificationHub, conn: conn, user: userID, send: make(chan []byte, 256)}

	return client
}

func registerClient(client *Client) {
	client.hub.register <- client

	go client.writeMessage()
	go client.readMessage()
}

func (a *App) NewWebsocketConnection(conn *websocket.Conn, user model.User) *model.AppError {
	client := a.newClient(conn, user.ID)

	registerClient(client)

	return nil
}

func GetDate(timestamp int64) string {
	timeT := time.Unix(int64(timestamp), 0)

	y, m, d := timeT.Date()
	fullDate := fmt.Sprintf("%d:%d, %d %s %d", timeT.Hour(), timeT.Minute(), y, m.String(), d)

	return fullDate
}

func (a *App) generateEmail(message model.NotificationMessage) error {
	if !a.emailEnabled() {
		return nil
	}

	user, err := a.Store.User.Get(message.Receiver)
	if err != nil {
		tlog.Errorw("Failed to retrieve notification receiver",
			"receiver_id", message.Receiver,
			"notification_type", message.NotificationType,
			"error", err,
		)
		return err
	}

	if user == nil {
		return nil
	}

	sender, err := a.Store.User.Get(message.Sender)
	if err != nil {
		tlog.Errorw("Failed to retrieve notification sender",
			"sender_id", message.Sender,
			"notification_type", message.NotificationType,
			"error", err,
		)
		return err
	}

	settings, err := a.Store.Preferences.GetNotifications(user.ID, "user")
	if err != nil {
		tlog.Errorw("Failed to retrieve notification settings",
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	for _, s := range settings {
		if s.Name == "send_email_notifications" && s.Value == "never" {
			return nil
		}
	}

	var link string

	switch message.App {
	case model.AppFiles:
		file, appErr := a.HasPermission(message.ItemID, *user)
		if appErr != nil {
			return fmt.Errorf("user has no access to file %s", message.ItemID)
		}

		siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
		if message.NotificationType == model.NOTIFICATION_FILE_SHARE {
			link = siteURL + "/files/shared?file=" + file.ID
		} else if parentFile, appErr := a.HasPermission(file.Parent, *user); appErr == nil {
			if parentFile.Type == "cloud#drive" {
				link = siteURL + "/files/shared?file=" + file.ID
			} else {
				link = siteURL + "/files/" + parentFile.ID + "?file=" + file.ID
			}
		}

	case model.AppProjects:
		baseURL := *a.ConfigStore.Config.ServerSettings.SiteURL
		workspaceID := message.ItemID
		tableID, _ := message.Details["tableID"].(string)
		viewID, _ := message.Details["viewID"].(string)
		taskID, _ := message.Details["taskID"].(string)

		if workspaceID != "" && tableID != "" && viewID != "" {
			link = fmt.Sprintf("%s/projects/%s/grid/%s/view/%s", baseURL, workspaceID, tableID, viewID)
			if taskID != "" {
				link = fmt.Sprintf("%s?task=%s", link, taskID)
			}
		} else {
			link = baseURL + "/projects/assigned-to-me"
		}

	case model.AppChat:
		if message.NotificationType == model.NOTIFICATION_CHANNEL_MENTION {
			link = *a.ConfigStore.Config.ServerSettings.SiteURL + "/chat/" + message.ItemID
		}
	}

	locale := a.emailLocale(user.ID)
	nt, err := template.New("notification.html").Funcs(template.FuncMap{
		"getDate": GetDate,
		"T":       func(id string) string { return i18n.T(locale, id) },
	}).ParseFiles("templates/notification.html", "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse email template",
			"error", err,
		)
		return err
	}

	var body bytes.Buffer
	err = nt.Execute(&body, struct {
		Logo              string
		NotificationTitle string
		NotificationText  string
		Message           string
		Action            string
		Sender            string
		CompanyName       string
		ButtonColor       string
		CreatedAt         int64
		MessageObject     model.NotificationMessage
		Link              string
	}{
		Logo:              *a.ConfigStore.Config.ServerSettings.SiteURL + "/logo.png",
		NotificationTitle: message.Subject,
		NotificationText:  i18n.T(locale, "email.notification.intro"),
		Message:           constructMessage(locale, *sender, message),
		Action:            i18n.T(locale, "email.notification.action_open"),
		Sender:            sender.Name + " " + sender.LastName,
		CompanyName:       "Twigex",
		ButtonColor:       "#7E57C2",
		CreatedAt:         message.CreatedAt,
		MessageObject:     message,
		Link:              link,
	})
	if err != nil {
		tlog.Errorw("Failed to execute email template",
			"receiver_id", user.ID,
			"error", err,
		)
		return err
	}

	msg, err := a.newEmail(user.Email, message.Subject)
	if err != nil {
		tlog.Errorw("Failed to compose email",
			"receiver_id", user.ID,
			"error", err,
		)
		return err
	}

	msg.SetBodyString(mail.TypeTextHTML, body.String())

	if err = tmail.Send(a.getSMTPConfig(), msg); err != nil {
		tlog.Errorw("Failed to send email",
			"receiver_id", user.ID,
			"error", err,
		)
		return err
	}

	if err = a.Store.Notifications.MarkAsSent([]model.NotificationMessage{message}); err != nil {
		tlog.Errorw("Failed to mark notification as sent",
			"receiver_id", user.ID,
			"error", err,
		)
		return err
	}

	return nil
}

func (a *App) SendFinishRegistrationEmail(user model.User, token string) {
	if !a.emailEnabled() {
		return
	}

	locale := a.emailLocale(user.ID)
	nt, err := template.New("password_change.html").Funcs(emailFuncMap(locale)).ParseFiles("templates/password_change.html", "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse registration email template",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	var body bytes.Buffer
	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	err = nt.Execute(&body, struct {
		Logo        string
		Title       string
		Text        string
		ActionLink  string
		CompanyName string
		ButtonColor string
	}{
		Logo:        siteURL + "/logo.png",
		Title:       i18n.T(locale, "email.registration.title"),
		Text:        i18n.T(locale, "email.registration.text"),
		ActionLink:  siteURL + "/password/reset/" + token,
		CompanyName: "Twigex",
		ButtonColor: "#7E57C2",
	})
	if err != nil {
		tlog.Errorw("Failed to execute registration email template",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	msg, err := a.newEmail(user.Email, i18n.T(locale, "email.registration.subject"))
	if err != nil {
		tlog.Errorw("Failed to compose registration email",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	msg.SetBodyString(mail.TypeTextHTML, body.String())

	if err = tmail.Send(a.getSMTPConfig(), msg); err != nil {
		tlog.Errorw("Failed to send registration email",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	tlog.Infow("Registration email sent",
		"user_id", user.ID,
	)
}

func (a *App) SendPasswordResetEmail(user model.User, token string) {
	if !a.emailEnabled() {
		return
	}

	locale := a.emailLocale(user.ID)
	nt, err := template.New("password_change.html").Funcs(emailFuncMap(locale)).ParseFiles("templates/password_change.html", "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse password reset email template",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	var body bytes.Buffer
	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	if err = nt.Execute(&body, struct {
		Logo        string
		Title       string
		Text        string
		ActionLink  string
		CompanyName string
		ButtonColor string
	}{
		Logo:        siteURL + "/logo.png",
		Title:       i18n.T(locale, "email.password_reset.title"),
		Text:        i18n.T(locale, "email.password_reset.text"),
		ActionLink:  siteURL + "/password/reset/" + token,
		CompanyName: "Twigex",
		ButtonColor: "#7E57C2",
	}); err != nil {
		tlog.Errorw("Failed to execute password reset email template",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	msg, err := a.newEmail(user.Email, i18n.T(locale, "email.password_reset.subject"))
	if err != nil {
		tlog.Errorw("Failed to compose password reset email",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	msg.SetBodyString(mail.TypeTextHTML, body.String())

	if err = tmail.Send(a.getSMTPConfig(), msg); err != nil {
		tlog.Errorw("Failed to send password reset email",
			"user_id", user.ID,
			"error", err,
		)
		return
	}

	tlog.Infow("Password reset email sent",
		"user_id", user.ID,
	)
}

// Renders in the meeting's own timezone (the organizer's IANA zone), falling back to UTC if the zone is empty or can't be loaded.
func formatMeetingTime(unix int64, tz string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		loc = time.UTC
	}

	return time.Unix(unix, 0).In(loc).Format("Mon, 02 Jan 2006 15:04 MST")
}

func (a *App) SendVideoGuestInviteEmail(link model.ChannelGuestLink, meeting model.ChannelMeeting, invitedBy model.User, updated bool) {
	if !a.emailEnabled() {
		return
	}

	scheduled := meeting.ScheduledAt != nil
	locale := a.emailLocale("")

	templateName := "video_guest_invite.html"
	prefix := "email.video_invite"
	if scheduled {
		templateName = "video_meeting_scheduled.html"
		prefix = "email.meeting_invite"
		if updated {
			prefix = "email.meeting_updated"
		}
	}

	subject := i18n.T(locale, prefix+".subject")
	title := i18n.T(locale, prefix+".title")
	text := i18n.T(locale, prefix+".text")

	nt, err := template.New(templateName).Funcs(emailFuncMap(locale)).ParseFiles("templates/"+templateName, "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse video guest invite email template",
			"link_id", link.ID,
			"error", err,
		)
		return
	}

	scheduledFor := ""
	if scheduled {
		scheduledFor = formatMeetingTime(*meeting.ScheduledAt, meeting.Timezone)
	}

	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	actionLink := siteURL + "/guest/video/" + link.ID
	inviterName := strings.TrimSpace(invitedBy.Name + " " + invitedBy.LastName)

	scheduledLine := ""
	if scheduled {
		scheduledLine = i18n.Tf(locale, "email.meeting.scheduled_for", map[string]any{"Time": scheduledFor})
	}

	var htmlBody bytes.Buffer
	if err = nt.Execute(&htmlBody, struct {
		Logo          string
		Title         string
		Text          string
		ActionLink    string
		ScheduledFor  string
		InviteLine    template.HTML
		ScheduledLine string
		CompanyName   string
		ButtonColor   string
	}{
		Logo:          siteURL + "/logo.png",
		Title:         title,
		Text:          text,
		ActionLink:    actionLink,
		ScheduledFor:  scheduledFor,
		InviteLine:    template.HTML(meetingLine(locale, "email.meeting.invited_line", inviterName, meeting.Title, "", true)),
		ScheduledLine: scheduledLine,
		CompanyName:   "Twigex",
		ButtonColor:   "#7E57C2",
	}); err != nil {
		tlog.Errorw("Failed to execute video guest invite email template",
			"link_id", link.ID,
			"error", err,
		)
		return
	}

	cfg := a.getSMTPConfig()

	ics := ""
	if scheduled {
		ics = a.buildMeetingICS(meeting, meeting.Title, text+" Join: "+actionLink,
			actionLink, inviterName, invitedBy.Email, link.InvitedEmail, "REQUEST")
	}

	plain := meetingLine(locale, "email.meeting.invited_line", inviterName, meeting.Title, "", false) +
		"\r\n\r\n" + i18n.T(locale, "email.label.join") + ": " + actionLink
	if scheduled {
		plain += "\r\n" + scheduledLine
	}

	msg, err := a.composeMeetingMsg(link.InvitedEmail, subject, plain, htmlBody.String(), ics, "REQUEST")
	if err != nil {
		tlog.Errorw("Failed to compose video guest invite email",
			"link_id", link.ID,
			"error", err,
		)
		return
	}

	if err = tmail.Send(cfg, msg); err != nil {
		tlog.Errorw("Failed to send video guest invite email",
			"link_id", link.ID,
			"error", err,
		)
		return
	}

	tlog.Infow("Video guest invite email sent",
		"link_id", link.ID,
	)
}

// Skips the host and explicit invitees (notified directly) so a user in both an
// invited group and the invitee list isn't notified twice.
func (a *App) meetingGroupRecipients(meeting model.ChannelMeeting) []model.User {
	ids, err := a.Store.Channels.GetMeetingGroupMemberIDs(context.Background(), meeting.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve meeting group members", "meeting_id", meeting.ID, "error", err)
		return nil
	}

	seen := map[string]bool{meeting.HostID: true}
	for _, id := range meeting.Invitees {
		seen[id] = true
	}

	recipients := make([]model.User, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}

		seen[id] = true
		user, err := a.Store.User.Get(id)
		if err != nil || user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		recipients = append(recipients, *user)
	}

	return recipients
}

func (a *App) NotifyMeetingScheduled(creator model.User, channel model.Channel, meeting model.ChannelMeeting) {
	for _, inviteeID := range meeting.Invitees {
		if inviteeID == creator.ID {
			continue
		}

		user, err := a.Store.User.Get(inviteeID)
		if err != nil || user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		a.notifyMeetingMember(creator, *user, model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED, channel, meeting)
		a.SendMeetingScheduledMemberEmail(*user, channel, meeting, creator)
	}

	for _, user := range a.meetingGroupRecipients(meeting) {
		if user.ID == creator.ID {
			continue
		}

		a.notifyMeetingMember(creator, user, model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED, channel, meeting)
		a.SendMeetingScheduledMemberEmail(user, channel, meeting, creator)
	}
}

func (a *App) NotifyMeetingInvited(actor model.User, channel model.Channel, meeting model.ChannelMeeting, userIDs []string) {
	for _, id := range userIDs {
		if id == actor.ID {
			continue
		}

		user, err := a.Store.User.Get(id)
		if err != nil || user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		a.notifyMeetingMember(actor, *user, model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED, channel, meeting)
		a.SendMeetingScheduledMemberEmail(*user, channel, meeting, actor)
	}
}

// NotifyMeetingGroupsInvited notifies the members of a meeting's invited groups,
// mirroring NotifyMeetingInvited for individually invited members.
func (a *App) NotifyMeetingGroupsInvited(actor model.User, channel model.Channel, meeting model.ChannelMeeting) {
	for _, user := range a.meetingGroupRecipients(meeting) {
		if user.ID == actor.ID {
			continue
		}

		a.notifyMeetingMember(actor, user, model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED, channel, meeting)
		a.SendMeetingScheduledMemberEmail(user, channel, meeting, actor)
	}
}

func (a *App) notifyMeetingMember(actor, recipient model.User, notificationType string, channel model.Channel, meeting model.ChannelMeeting) {
	if appAllowed, _ := a.notificationAllowed(recipient, notificationType); !appAllowed {
		return
	}

	var scheduledAt int64
	if meeting.ScheduledAt != nil {
		scheduledAt = *meeting.ScheduledAt
	}

	details := map[string]any{
		"app":              model.AppChat,
		"notificationType": notificationType,
		"channelName":      channel.DisplayName,
		"meetingTitle":     meeting.Title,
		"channel_id":       channel.ID,
		"meeting_id":       meeting.ID,
		"scheduled_at":     scheduledAt,
	}

	m, err := a.Store.Notifications.Create(actor, recipient.ID, model.AppChat, notificationType, channel.ID, details)
	if err != nil {
		tlog.Errorw("Failed to create meeting notification", "user_id", recipient.ID, "type", notificationType, "error", err)
		return
	}

	event := model.WebsocketEvent{
		Event: notificationType,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": notificationType,
			"channel_id":       channel.ID,
			"message":          m.ToMap(),
		},
	}

	if err := a.sendPushNotifications(recipient, event); err != nil {
		tlog.Errorw("Failed to push meeting notification", "user_id", recipient.ID, "type", notificationType, "error", err)
	}
}

func (a *App) SendMeetingScheduledMemberEmail(user model.User, channel model.Channel, meeting model.ChannelMeeting, creator model.User) {
	if _, emailAllowed := a.notificationAllowed(user, model.NOTIFICATION_CHANNEL_MEETING_SCHEDULED); !emailAllowed {
		return
	}

	a.sendMeetingMemberEmail(user, channel, meeting, creator, "email.meeting_member_scheduled")
}

// Carries an updated calendar invite (same UID, higher SEQUENCE) so calendars move the existing event.
func (a *App) SendMeetingUpdatedMemberEmail(user model.User, channel model.Channel, meeting model.ChannelMeeting, creator model.User) {
	if _, emailAllowed := a.notificationAllowed(user, model.NOTIFICATION_CHANNEL_MEETING_UPDATED); !emailAllowed {
		return
	}

	a.sendMeetingMemberEmail(user, channel, meeting, creator, "email.meeting_member_updated")
}

func (a *App) sendMeetingMemberEmail(user model.User, channel model.Channel, meeting model.ChannelMeeting, creator model.User, prefix string) {
	if !a.emailEnabled() {
		return
	}

	locale := a.emailLocale(user.ID)
	title := i18n.T(locale, prefix+".title")
	text := i18n.T(locale, prefix+".text")
	intro := i18n.Tf(locale, prefix+".intro", map[string]any{"Channel": channel.DisplayName})

	nt, err := template.New("video_meeting_scheduled.html").Funcs(emailFuncMap(locale)).ParseFiles("templates/video_meeting_scheduled.html", "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse scheduled meeting email template", "user_id", user.ID, "error", err)
		return
	}

	scheduledFor := ""
	if meeting.ScheduledAt != nil {
		scheduledFor = formatMeetingTime(*meeting.ScheduledAt, meeting.Timezone)
	}

	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	// System users join via the standalone meeting page, which works whether or not they're a channel member.
	actionLink := siteURL + "/meeting/" + meeting.ID
	creatorName := strings.TrimSpace(creator.Name + " " + creator.LastName)

	var htmlBody bytes.Buffer
	if err = nt.Execute(&htmlBody, struct {
		Logo         string
		Title        string
		Text         string
		ActionLink   string
		ScheduledFor string
		InviteLine   template.HTML
		CompanyName  string
		ButtonColor  string
	}{
		Logo:         siteURL + "/logo.png",
		Title:        title,
		Text:         text,
		ActionLink:   actionLink,
		ScheduledFor: scheduledFor,
		InviteLine:   template.HTML(meetingLine(locale, "email.meeting.invited_line", creatorName, meeting.Title, channel.DisplayName, true)),
		CompanyName:  "Twigex",
		ButtonColor:  "#7E57C2",
	}); err != nil {
		tlog.Errorw("Failed to execute scheduled meeting email template", "user_id", user.ID, "error", err)
		return
	}

	cfg := a.getSMTPConfig()
	ics := a.buildMeetingICS(meeting, meeting.Title, "Meeting in "+channel.DisplayName+". Open: "+actionLink,
		actionLink, creatorName, creator.Email, user.Email, "REQUEST")

	plain := fmt.Sprintf("%s.\r\n\r\n%s\r\n%s: %s\r\n%s: %s",
		intro, meeting.Title,
		i18n.T(locale, "email.label.when"), scheduledFor,
		i18n.T(locale, "email.label.open"), actionLink)

	msg, err := a.composeMeetingMsg(user.Email, intro, plain, htmlBody.String(), ics, "REQUEST")
	if err != nil {
		tlog.Errorw("Failed to compose scheduled meeting email", "user_id", user.ID, "error", err)
		return
	}

	if err = tmail.Send(cfg, msg); err != nil {
		tlog.Errorw("Failed to send scheduled meeting email", "user_id", user.ID, "error", err)
		return
	}

	tlog.Infow("Scheduled meeting email sent", "user_id", user.ID)
}

// guestEmails are the meeting's guest-link addresses, collected before the links were revoked.
func (a *App) NotifyMeetingCancelled(host model.User, channel model.Channel, meeting model.ChannelMeeting, guestEmails []string) {
	for _, inviteeID := range meeting.Invitees {
		if inviteeID == host.ID {
			continue
		}

		user, err := a.Store.User.Get(inviteeID)
		if err != nil || user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		a.notifyMeetingMember(host, *user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED, channel, meeting)
		if _, emailAllowed := a.notificationAllowed(*user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED); emailAllowed {
			a.sendMeetingCancelledEmail(user.Email, channel.DisplayName, meeting, host, a.emailLocale(user.ID), true)
		}
	}

	for _, user := range a.meetingGroupRecipients(meeting) {
		if user.ID == host.ID {
			continue
		}

		a.notifyMeetingMember(host, user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED, channel, meeting)
		if _, emailAllowed := a.notificationAllowed(user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED); emailAllowed {
			a.sendMeetingCancelledEmail(user.Email, channel.DisplayName, meeting, host, a.emailLocale(user.ID), true)
		}
	}

	for _, email := range guestEmails {
		if strings.TrimSpace(email) == "" {
			continue
		}

		a.sendMeetingCancelledEmail(email, channel.DisplayName, meeting, host, a.emailLocale(""), false)
	}
}

// Pushes a transient websocket signal (no stored notification) so clients flip the meeting from scheduled to live.
func (a *App) BroadcastMeetingStarted(meeting model.ChannelMeeting, channelID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_MEETING_STARTED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_MEETING_STARTED,
			"channel_id":       channelID,
			"meeting_id":       meeting.ID,
		},
	}

	for _, uid := range a.meetingAudience(meeting, channelID) {
		if err := a.sendPushNotifications(model.User{ID: uid}, event); err != nil {
			tlog.Errorw("Failed to push meeting started", "user_id", uid, "meeting_id", meeting.ID, "error", err)
		}
	}
}

// Pushes a transient websocket signal so every audience member's UI reflects the end, not just whoever ended it.
func (a *App) BroadcastMeetingEnded(meeting model.ChannelMeeting, channelID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_MEETING_ENDED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_MEETING_ENDED,
			"channel_id":       channelID,
			"meeting_id":       meeting.ID,
		},
	}

	for _, uid := range a.meetingAudience(meeting, channelID) {
		if err := a.sendPushNotifications(model.User{ID: uid}, event); err != nil {
			tlog.Errorw("Failed to push meeting ended", "user_id", uid, "meeting_id", meeting.ID, "error", err)
		}
	}
}

// RefreshMeetingLists tells everyone who can see a scheduled meeting to reload their list.
// Notifications only reach invitees who want them, so an open meeting reached nobody else.
// formerAudience covers people who lost access in the change.
func (a *App) RefreshMeetingLists(meeting model.ChannelMeeting, channelID string, formerAudience []string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_MEETINGS_REFRESH,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_MEETINGS_REFRESH,
			"channel_id":       channelID,
			"meeting_id":       meeting.ID,
		},
	}

	recipients := DedupeIDs(append(a.meetingAudience(meeting, channelID), formerAudience...))
	for _, uid := range recipients {
		if err := a.sendPushNotifications(model.User{ID: uid}, event); err != nil {
			tlog.Errorw("Failed to push meeting list refresh", "user_id", uid, "meeting_id", meeting.ID, "error", err)
		}
	}
}

func (a *App) BroadcastMeetingPresence(meeting model.ChannelMeeting, channelID string, count int) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_MEETING_PRESENCE,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_MEETING_PRESENCE,
			"channel_id":       channelID,
			"meeting_id":       meeting.ID,
			"participants":     count,
		},
	}

	for _, uid := range a.meetingAudience(meeting, channelID) {
		if err := a.sendPushNotifications(model.User{ID: uid}, event); err != nil {
			tlog.Errorw("Failed to push meeting presence", "user_id", uid, "meeting_id", meeting.ID, "error", err)
		}
	}
}

// A ring carries no meeting id because no meeting exists yet; the room is
// created only when the callee answers.
func (a *App) RingDirectCall(caller model.User, channelID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_INCOMING_CALL,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_INCOMING_CALL,
			"channel_id":       channelID,
			"caller_id":        caller.ID,
			"caller_name":      strings.TrimSpace(caller.Name + " " + caller.LastName),
		},
	}

	a.pushToChannelMembers(channelID, event, caller.ID)
}

// Only the waiting caller is notified; the answerer already has the room from
// its own response.
func (a *App) NotifyCallConnected(channelID string, meeting model.ChannelMeeting, callerID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_CALL_CONNECTED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_CALL_CONNECTED,
			"channel_id":       channelID,
			"meeting_id":       meeting.ID,
			"title":            meeting.Title,
		},
	}

	if err := a.sendPushNotifications(model.User{ID: callerID}, event); err != nil {
		tlog.Errorw("Failed to notify call connected", "user_id", callerID, "error", err)
	}
}

// NotifyCallHandled tells the callee's other sessions to stop ringing. The
// caller is excluded because they get connected or declined instead.
func (a *App) NotifyCallHandled(channelID, callerID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_CALL_HANDLED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_CALL_HANDLED,
			"channel_id":       channelID,
		},
	}

	a.pushToChannelMembers(channelID, event, callerID)
}

func (a *App) NotifyCallDeclined(decliner model.User, channelID, callerID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_CALL_DECLINED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_CALL_DECLINED,
			"channel_id":       channelID,
			"decliner_name":    strings.TrimSpace(decliner.Name + " " + decliner.LastName),
		},
	}

	if err := a.sendPushNotifications(model.User{ID: callerID}, event); err != nil {
		tlog.Errorw("Failed to notify call declined", "user_id", callerID, "error", err)
	}
}

// Pass excludeUserID to skip the member who triggered the cancel (they cleared
// their own dialog already); "" notifies both, e.g. on timeout.
func (a *App) NotifyCallCancelled(channelID, excludeUserID string) {
	event := model.WebsocketEvent{
		Event: model.NOTIFICATION_CHANNEL_CALL_CANCELLED,
		App:   model.AppChat,
		Data: map[string]any{
			"app":              model.AppChat,
			"notificationType": model.NOTIFICATION_CHANNEL_CALL_CANCELLED,
			"channel_id":       channelID,
		},
	}

	a.pushToChannelMembers(channelID, event, excludeUserID)
}

func (a *App) pushToChannelMembers(channelID string, event model.WebsocketEvent, excludeUserID string) {
	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil || members == nil {
		tlog.Errorw("Failed to load channel members for call event", "channel_id", channelID, "error", err)
		return
	}

	for _, m := range members {
		if m.UserID == excludeUserID {
			continue
		}

		if err := a.sendPushNotifications(model.User{ID: m.UserID}, event); err != nil {
			tlog.Errorw("Failed to push call event", "user_id", m.UserID, "error", err)
		}
	}
}

// Audience is host plus invitees and invited-group members for an invite-only
// meeting, or all channel members for an open one.
func (a *App) meetingAudience(meeting model.ChannelMeeting, channelID string) []string {
	if meeting.InviteOnly {
		ids := append([]string{meeting.HostID}, meeting.Invitees...)
		groupMembers, err := a.Store.Channels.GetMeetingGroupMemberIDs(context.Background(), meeting.ID)
		if err != nil {
			tlog.Errorw("Failed to resolve meeting group members", "meeting_id", meeting.ID, "error", err)
		} else {
			ids = append(ids, groupMembers...)
		}

		return DedupeIDs(ids)
	}

	members, err := a.Store.Channels.GetMembers(channelID)
	if err != nil || members == nil {
		return []string{meeting.HostID}
	}

	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}

	return ids
}

// Sends a cancellation to members dropped from the allowlist so the event leaves their calendar.
func (a *App) NotifyMeetingUpdated(host model.User, channel model.Channel, meeting model.ChannelMeeting, removedMemberIDs []string) {
	for _, inviteeID := range meeting.Invitees {
		if inviteeID == host.ID {
			continue
		}

		user, err := a.Store.User.Get(inviteeID)
		if err != nil || user == nil {
			continue
		}

		if !a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			continue
		}

		a.notifyMeetingMember(host, *user, model.NOTIFICATION_CHANNEL_MEETING_UPDATED, channel, meeting)
		a.SendMeetingUpdatedMemberEmail(*user, channel, meeting, host)
	}

	for _, user := range a.meetingGroupRecipients(meeting) {
		if user.ID == host.ID {
			continue
		}

		a.notifyMeetingMember(host, user, model.NOTIFICATION_CHANNEL_MEETING_UPDATED, channel, meeting)
		a.SendMeetingUpdatedMemberEmail(user, channel, meeting, host)
	}

	for _, inviteeID := range removedMemberIDs {
		if inviteeID == host.ID {
			continue
		}

		user, err := a.Store.User.Get(inviteeID)
		if err != nil || user == nil {
			continue
		}

		if a.SessionHasPermission(*user, model.ChatPermissions.PermissionViewChannels) {
			a.notifyMeetingMember(host, *user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED, channel, meeting)
		}

		if _, emailAllowed := a.notificationAllowed(*user, model.NOTIFICATION_CHANNEL_MEETING_CANCELLED); emailAllowed {
			a.sendMeetingCancelledEmail(user.Email, channel.DisplayName, meeting, host, a.emailLocale(user.ID), true)
		}
	}

	// Runs detached from the request, so use a background context.
	links, err := a.Store.Channels.GetActiveMeetingGuestLinks(context.Background(), meeting.ID, time.Now().Unix())
	if err != nil {
		tlog.Errorw("Failed to list guest links for meeting update", "meeting_id", meeting.ID, "error", err)
		return
	}

	for _, link := range links {
		a.SendVideoGuestInviteEmail(link, meeting, host, true)
	}
}

// Sends a METHOD:CANCEL ICS (same UID) so calendars that auto-added the event remove it.
func (a *App) sendMeetingCancelledEmail(to, channelName string, meeting model.ChannelMeeting, host model.User, locale string, showChannel bool) {
	if !a.emailEnabled() {
		return
	}

	nt, err := template.New("video_meeting_cancelled.html").Funcs(emailFuncMap(locale)).ParseFiles("templates/video_meeting_cancelled.html", "templates/email_footer.html")
	if err != nil {
		tlog.Errorw("Failed to parse cancellation email template", "to", to, "error", err)
		return
	}

	lineChannel := ""
	if showChannel {
		lineChannel = channelName
	}

	scheduledFor := ""
	if meeting.ScheduledAt != nil {
		scheduledFor = formatMeetingTime(*meeting.ScheduledAt, meeting.Timezone)
	}

	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	hostName := strings.TrimSpace(host.Name + " " + host.LastName)

	wasScheduledLine := ""
	if scheduledFor != "" {
		wasScheduledLine = i18n.Tf(locale, "email.meeting.was_scheduled_for", map[string]any{"Time": scheduledFor})
	}

	var htmlBody bytes.Buffer
	if err = nt.Execute(&htmlBody, struct {
		Logo             string
		Title            string
		Text             string
		CancelLine       template.HTML
		WasScheduledLine string
		CompanyName      string
		ButtonColor      string
	}{
		Logo:             siteURL + "/logo.png",
		Title:            i18n.T(locale, "email.meeting_cancelled.title"),
		Text:             i18n.T(locale, "email.meeting_cancelled.text"),
		CancelLine:       template.HTML(meetingLine(locale, "email.meeting.cancelled_line", hostName, meeting.Title, lineChannel, true)),
		WasScheduledLine: wasScheduledLine,
		CompanyName:      "Twigex",
		ButtonColor:      "#7E57C2",
	}); err != nil {
		tlog.Errorw("Failed to execute cancellation email template", "to", to, "error", err)
		return
	}

	cfg := a.getSMTPConfig()
	ics := a.buildMeetingICS(meeting, meeting.Title, "Cancelled", "", hostName, host.Email, to, "CANCEL")

	plain := meetingLine(locale, "email.meeting.cancelled_line", hostName, meeting.Title, lineChannel, false)
	if wasScheduledLine != "" {
		plain += "\r\n" + wasScheduledLine
	}

	subject := i18n.Tf(locale, "email.meeting_cancelled.subject", map[string]any{"Title": meeting.Title})
	msg, err := a.composeMeetingMsg(to, subject, plain, htmlBody.String(), ics, "CANCEL")
	if err != nil {
		tlog.Errorw("Failed to compose cancellation email", "to", to, "error", err)
		return
	}

	if err = tmail.Send(cfg, msg); err != nil {
		tlog.Errorw("Failed to send cancellation email", "to", to, "error", err)
		return
	}
}

func icsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// method is "REQUEST" (invite) or "CANCEL" (removal, same UID and higher SEQUENCE). Returns "" for meetings without a scheduled time.
func (a *App) buildMeetingICS(meeting model.ChannelMeeting, summary, description, location, organizerName, organizerEmail, attendeeEmail, method string) string {
	if meeting.ScheduledAt == nil {
		return ""
	}

	const layout = "20060102T150405Z"
	start := time.Unix(*meeting.ScheduledAt, 0).UTC()
	duration := meeting.DurationMinutes
	if duration <= 0 {
		duration = 60
	}

	end := start.Add(time.Duration(duration) * time.Minute)

	status := "CONFIRMED"
	if method == "CANCEL" {
		status = "CANCELLED"
	}

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Twigex//Meetings//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	fmt.Fprintf(&b, "METHOD:%s\r\n", method)
	b.WriteString("BEGIN:VEVENT\r\n")
	fmt.Fprintf(&b, "UID:%s@twigex\r\n", meeting.ID)
	fmt.Fprintf(&b, "DTSTAMP:%s\r\n", time.Now().UTC().Format(layout))
	fmt.Fprintf(&b, "DTSTART:%s\r\n", start.Format(layout))
	fmt.Fprintf(&b, "DTEND:%s\r\n", end.Format(layout))
	fmt.Fprintf(&b, "SUMMARY:%s\r\n", icsEscape(summary))
	fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", icsEscape(description))
	fmt.Fprintf(&b, "LOCATION:%s\r\n", icsEscape(location))
	fmt.Fprintf(&b, "ORGANIZER;CN=%s:mailto:%s\r\n", icsEscape(organizerName), organizerEmail)
	fmt.Fprintf(&b, "ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE;CN=%s:mailto:%s\r\n", attendeeEmail, attendeeEmail)
	fmt.Fprintf(&b, "SEQUENCE:%d\r\n", meeting.Sequence)
	fmt.Fprintf(&b, "STATUS:%s\r\n", status)
	b.WriteString("END:VEVENT\r\n")
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func (a *App) composeMeetingMsg(to, subject, plainText, html, ics, icsMethod string) (*mail.Msg, error) {
	m, err := a.newEmail(to, subject)
	if err != nil {
		return nil, err
	}

	m.SetBodyString(mail.TypeTextPlain, plainText)
	m.AddAlternativeString(mail.TypeTextHTML, html)

	if ics != "" {
		m.AddAlternativeString(mail.ContentType("text/calendar; charset=\"UTF-8\"; method="+icsMethod), ics)
		if err := m.AttachReader("invite.ics", strings.NewReader(ics), mail.WithFileContentType("application/ics")); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (a *App) newEmail(to, subject string) (*mail.Msg, error) {
	es := a.ConfigStore.Config.EmailSettings

	fromAddress := *es.SMTPUsername
	if es.FromAddress != nil && *es.FromAddress != "" {
		fromAddress = *es.FromAddress
	}

	fromName := "Twigex"
	if es.FromName != nil && *es.FromName != "" {
		fromName = *es.FromName
	}

	m := mail.NewMsg()
	if err := m.FromFormat(fromName, fromAddress); err != nil {
		return nil, err
	}

	if u := *es.SMTPUsername; u != "" {
		if err := m.EnvelopeFrom(u); err != nil {
			return nil, err
		}
	}

	if err := m.To(to); err != nil {
		return nil, err
	}

	if es.ReplyToAddress != nil && *es.ReplyToAddress != "" {
		if err := m.ReplyTo(*es.ReplyToAddress); err != nil {
			return nil, err
		}
	}

	m.Subject(subject)
	m.SetMessageID()
	return m, nil
}

func (a *App) emailLocale(userID string) string {
	def := *a.ConfigStore.Config.ServerSettings.DefaultLocale
	if userID == "" || !*a.ConfigStore.Config.ServerSettings.AllowUserLanguageOverride {
		return def
	}

	prefs, err := a.Store.Preferences.GetForUser(userID, model.PreferencesCategoryDisplay)
	if err != nil || prefs == nil {
		return def
	}

	for _, p := range prefs {
		if p.Name == "language" && p.Value != "" {
			return p.Value
		}
	}

	return def
}

func emailFuncMap(locale string) template.FuncMap {
	return template.FuncMap{
		"T": func(id string) string { return i18n.T(locale, id) },
	}
}

func boldHTML(s string) string {
	return "<strong>" + template.HTMLEscapeString(s) + "</strong>"
}

// meetingLine renders the "<inviter> <verb> <meeting> [in <channel>]" sentence.
// An empty channel selects the _no_channel variant, so guests are never shown
// internal channel names. When bold is set the dynamic values are wrapped for HTML.
func meetingLine(locale, base, inviter, meeting, channel string, bold bool) string {
	wrap := func(s string) string {
		if bold {
			return boldHTML(s)
		}

		return s
	}

	data := map[string]any{"Inviter": wrap(inviter), "Meeting": wrap(meeting)}
	key := base + "_no_channel"
	if channel != "" {
		data["Channel"] = wrap(channel)
		key = base
	}

	return i18n.Tf(locale, key, data)
}

func (a *App) emailEnabled() bool {
	es := a.ConfigStore.Config.EmailSettings
	if es.EnableEmail != nil {
		return *es.EnableEmail
	}

	return es.SMTPServer != nil && *es.SMTPServer != ""
}

func (a *App) getSMTPConfig() *tmail.SMTPConfig {
	return &tmail.SMTPConfig{
		ConnectionSecurity: *a.ConfigStore.Config.EmailSettings.ConnectionSecurity,
		Server:             *a.ConfigStore.Config.EmailSettings.SMTPServer,
		Port:               *a.ConfigStore.Config.EmailSettings.SMTPPort,
		Username:           *a.ConfigStore.Config.EmailSettings.SMTPUsername,
		Password:           *a.ConfigStore.Config.EmailSettings.SMTPPassword,
		EnableSMTPAuth:     *a.ConfigStore.Config.EmailSettings.EnableSMTPAuth,
	}
}

// smtpConfigFromSettings builds a config from a settings request so an admin can
// test unsaved form values. Missing fields fall back to the current config, and a
// masked password ("********") resolves to the stored one since the GUI never sends it.
func (a *App) smtpConfigFromSettings(settings model.EmailSettings) *tmail.SMTPConfig {
	es := a.ConfigStore.Config.EmailSettings

	str := func(req, cur *string) string {
		if req != nil {
			return *req
		}

		return *cur
	}

	password := str(settings.SMTPPassword, es.SMTPPassword)
	if password == "********" {
		password = *es.SMTPPassword
	}

	enableAuth := *es.EnableSMTPAuth
	if settings.EnableSMTPAuth != nil {
		enableAuth = *settings.EnableSMTPAuth
	}

	return &tmail.SMTPConfig{
		ConnectionSecurity: str(settings.ConnectionSecurity, es.ConnectionSecurity),
		Server:             str(settings.SMTPServer, es.SMTPServer),
		Port:               str(settings.SMTPPort, es.SMTPPort),
		Username:           str(settings.SMTPUsername, es.SMTPUsername),
		Password:           password,
		EnableSMTPAuth:     enableAuth,
	}
}

func constructMessage(locale string, sender model.User, message model.NotificationMessage) string {
	filename, _ := message.Details["filename"].(string)
	oldName := ""
	if model.NOTIFICATION_FILE_RENAME == message.NotificationType {
		oldName, _ = message.Details["old_name"].(string)
	}

	switch message.NotificationType {
	case model.NOTIFICATION_FILE_CREATE:
		return fmt.Sprintf(`%s %s created "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_SHARE:
		return fmt.Sprintf(`%s %s shared "%s" with you.`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_UPLOAD:
		return fmt.Sprintf(`%s %s uploaded "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_RENAME:
		return fmt.Sprintf(`%s %s renamed file "%s" to "%s".`, sender.Name, sender.LastName, oldName, filename)
	case model.NOTIFICATION_FILE_DELETE:
		return fmt.Sprintf(`%s %s deleted "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_RESTORE:
		return fmt.Sprintf(`%s %s restored "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_COMMENT:
		return fmt.Sprintf(`%s %s commented on "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_DOWNLOAD:
		return fmt.Sprintf(`%s %s downloaded "%s".`, sender.Name, sender.LastName, filename)
	case model.NOTIFICATION_FILE_PUBLIC_DOWNLOAD:
		return fmt.Sprintf(`"%s" was downloaded through a public link.`, filename)

	case model.NOTIFICATION_CHANNEL_MENTION:
		channelName, _ := message.Details["channelName"].(string)
		return i18n.Tf(locale, "email.notification.mention", map[string]any{
			"sender":  sender.Name + " " + sender.LastName,
			"channel": channelName,
		})

	case model.NOTIFICATION_PROJECT_INVITE:
		workspaceName, _ := message.Details["workspaceName"].(string)
		if workspaceName != "" {
			return fmt.Sprintf(`%s %s invited you to project "%s".`, sender.Name, sender.LastName, workspaceName)
		}

		return fmt.Sprintf(`%s %s invited you to a project.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_PROJECT_DELETED:
		workspaceName, _ := message.Details["workspaceName"].(string)
		if workspaceName != "" {
			return fmt.Sprintf(`%s %s deleted project "%s" that you were a member of.`, sender.Name, sender.LastName, workspaceName)
		}

		return fmt.Sprintf(`%s %s deleted a project you were a member of.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_PROJECT_TASK_CREATED:
		name, _ := message.Details["name"].(string)
		if name != "" {
			return fmt.Sprintf(`%s %s created task "%s".`, sender.Name, sender.LastName, name)
		}

		taskName, _ := message.Details["taskName"].(string)
		if taskName != "" {
			return fmt.Sprintf(`%s %s created task "%s".`, sender.Name, sender.LastName, taskName)
		}

		return fmt.Sprintf(`%s %s created a new task.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_PROJECT_TASK_DELETED:
		taskName, _ := message.Details["taskName"].(string)
		if taskName != "" {
			return fmt.Sprintf(`%s %s deleted task "%s".`, sender.Name, sender.LastName, taskName)
		}

		return fmt.Sprintf(`%s %s deleted a task.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_PROJECT_TASK_ASSIGNED:
		taskName, _ := message.Details["taskName"].(string)
		if taskName != "" {
			return fmt.Sprintf(`%s %s assigned you to task "%s".`, sender.Name, sender.LastName, taskName)
		}

		return fmt.Sprintf(`%s %s assigned you to a task.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_PROJECT_TASK_STATUS_CHANGED:
		taskName, _ := message.Details["taskName"].(string)
		statusName, _ := message.Details["statusName"].(string)
		if taskName != "" && statusName != "" {
			return fmt.Sprintf(`%s %s changed status of task "%s" to "%s".`, sender.Name, sender.LastName, taskName, statusName)
		} else if taskName != "" {
			return fmt.Sprintf(`%s %s updated task "%s".`, sender.Name, sender.LastName, taskName)
		}

		return fmt.Sprintf(`%s %s changed a task status.`, sender.Name, sender.LastName)

	case model.NOTIFICATION_TASK_COMMENT:
		taskName, _ := message.Details["taskName"].(string)
		if taskName != "" {
			return fmt.Sprintf(`%s %s commented on task "%s".`, sender.Name, sender.LastName, taskName)
		}

		return fmt.Sprintf(`%s %s commented on a task.`, sender.Name, sender.LastName)

	default:
		return "Unknown notification type."
	}
}
