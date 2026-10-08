// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"regexp"
)

var (
	PostTypeForward = "forward"
)

// MaxPostMessageRunes is the largest rune count that always fits the posts.message
// TEXT column: 65535 bytes / 4 bytes per utf8mb4 char (worst case) = 16383.
const MaxPostMessageRunes = 16383

var reactionRe = regexp.MustCompile(`^:[a-z0-9_+-]{1,64}:$`)

// IsValidReaction reports whether a reaction is one the picker could have
// produced. The renderer binds a reaction as HTML, so an arbitrary string
// stored here would reach another member's DOM.
func IsValidReaction(reaction string) bool {
	return reactionRe.MatchString(reaction)
}

type PostResponse struct {
	Posts      []Post `json:"posts"`
	ReplyPosts []Post `json:"reply_posts"`
}

type Post struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	ChannelID string         `json:"channel_id"`
	ReplyID   string         `json:"reply_id"`
	Reply     *Post          `json:"reply,omitempty"`
	Message   string         `json:"message"`
	Type      string         `json:"type"`
	Reactions []PostReaction `json:"reactions"`
	Metadata  PostMetadata   `json:"metadata"`
	// Derived per response, never stored, so a rename shows at once and an
	// erased profile leaves no copy behind.
	Mentions      map[string]MentionUser `json:"mentions,omitempty"`
	PendingPostID string                 `json:"pending_post_id,omitempty"`
	CreatedAt     int64                  `json:"created"` // timestamp in milliseconds
	UpdatedAt     int64                  `json:"updated"` // timestamp in milliseconds
	DeletedAt     int64                  `json:"deleted"` // timestamp in milliseconds
}

type MentionUser struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
}

type NewPost struct {
	Message       string             `json:"message"`
	Reply         string             `json:"reply"`
	Attachments   []PostFileMetadata `json:"attachments"`
	GIF           string             `json:"gif"`
	PendingPostID string             `json:"pending_post_id"`
}

const MaxPendingPostIDLength = 64

func (p NewPost) ValidPendingPostID() string {
	if len(p.PendingPostID) > MaxPendingPostIDLength {
		return ""
	}

	return p.PendingPostID
}

type PostFileMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Type      string `json:"type"`
	StorageID string `json:"storage_id,omitempty"`
}

type PostMetadata struct {
	Files []PostFileAttachment `json:"files,omitempty"`
	Links []*LinkData          `json:"links,omitempty"`
}

type PostFileAttachment struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Height    int64  `json:"height"`
	Width     int64  `json:"width"`
	MimeType  string `json:"mime_type"`
	URL       string `json:"url,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Kind      string `json:"kind,omitempty"`
	StorageID string `json:"storage_id,omitempty"`
	CreatedAt int64  `json:"created_at"` // timestamp in milliseconds
	UpdatedAt int64  `json:"updated_at"` // timestamp in milliseconds
	DeletedAt int64  `json:"deleted_at"` // timestamp in milliseconds
}

func (p *Post) HasSameChannel(channelID string) bool {
	return p.ChannelID == channelID
}

type PostReaction struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Reaction  string `json:"reaction"`
	CreatedAt int64  `json:"created_at"` // timestamp in milliseconds
	UpdatedAt int64  `json:"updated_at"` // timestamp in milliseconds
	DeletedAt int64  `json:"deleted_at"` // timestamp in milliseconds
}

type LinkMetadata struct {
	Hash      int64    `json:"hash"`
	URL       string   `json:"url"`
	Type      string   `json:"type"`
	Data      LinkData `json:"data"`
	StorageID string   `json:"storage_id,omitempty"`
	Size      int64    `json:"size"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

type LinkData struct {
	Title       string     `json:"title"`
	SiteName    string     `json:"site_name"`
	Description string     `json:"description"`
	FaviconURL  string     `json:"favicon"`
	Image       LinkImage  `json:"image"`
	Type        string     `json:"type"`
	Url         string     `json:"url"`
	Video       *LinkVideo `json:"video,omitempty"`
}

type LinkVideo struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
}

type LinkImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
