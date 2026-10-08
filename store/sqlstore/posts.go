// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/twigex/twigex/model"
)

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}

	return s
}

type postRepository struct {
	Db *sql.DB
}

func NewPostRepository(Db *sql.DB) (*postRepository, error) {
	repo := &postRepository{}

	repo.Db = Db
	return repo, nil
}

// A failure to load attachments or reactions returns the posts alongside the
// error, so a caller can serve the page and report the gap rather than nothing.
func (p *postRepository) GetAllForChannel(ctx context.Context, channelID, postID, param string) (*model.PostResponse, error) {
	stmt := ""
	args := make([]any, 0)

	if postID != "" && param == "previous" {
		stmt = `SELECT ` + postColumns + ` FROM posts WHERE channel_id=? AND created_at < (SELECT created_at FROM posts WHERE id=?) ORDER BY created_at DESC LIMIT 30`
		args = append(args, channelID, postID)
	} else if postID != "" && param == "next" {
		stmt = `SELECT ` + postColumns + ` FROM posts WHERE channel_id=? AND created_at > (SELECT created_at FROM posts WHERE id=?) ORDER BY created_at ASC LIMIT 30`
		args = append(args, channelID, postID)
	} else if postID != "" && param == "history" {
		stmt = `
				(
					SELECT ` + postColumns + ` FROM posts
					WHERE channel_id = ?
					  AND created_at < (SELECT created_at FROM posts WHERE id = ?)
					ORDER BY created_at DESC
					LIMIT 30
				)
				UNION ALL
				(
					SELECT ` + postColumns + ` FROM posts
					WHERE id = ?
				)
				UNION ALL
				(
					SELECT ` + postColumns + ` FROM posts
					WHERE channel_id = ?
					  AND created_at > (SELECT created_at FROM posts WHERE id = ?)
					ORDER BY created_at ASC
					LIMIT 30
				)
				ORDER BY created_at DESC
			`
		args = append(args, channelID, postID, postID, channelID, postID)
	} else {
		stmt = `SELECT ` + postColumns + ` FROM posts WHERE channel_id=? ORDER BY created_at DESC LIMIT 30`
		args = append(args, channelID)
	}

	results, err := p.Db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	posts := make([]model.Post, 0)
	replyIDs := make([]string, 0)

	for results.Next() {
		post := model.Post{}
		if err := scanPost(results, &post); err != nil {
			return nil, err
		}

		if post.ReplyID != "" {
			replyIDs = append(replyIDs, post.ReplyID)
		}

		posts = append(posts, post)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	partial := func(err error) (*model.PostResponse, error) {
		return &model.PostResponse{Posts: posts, ReplyPosts: make([]model.Post, 0)}, err
	}

	replyPosts, err := p.GetByIDs(ctx, replyIDs)
	if err != nil {
		return partial(err)
	}

	ids := make([]string, 0, len(posts)+len(replyPosts))
	for _, post := range posts {
		ids = append(ids, post.ID)
	}

	for _, reply := range replyPosts {
		ids = append(ids, reply.ID)
	}

	attachments, err := p.GetAttachmentsForMany(ctx, ids)
	if err != nil {
		return partial(err)
	}

	attachmentsByPost := make(map[string][]model.PostFileAttachment, len(ids))
	for _, a := range attachments {
		attachmentsByPost[a.PostID] = append(attachmentsByPost[a.PostID], a)
	}

	postIDs := make([]string, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	reactions, err := p.GetReactionsForMany(ctx, postIDs)
	if err != nil {
		return partial(err)
	}

	reactionsByPost := make(map[string][]model.PostReaction, len(postIDs))
	for _, r := range reactions {
		reactionsByPost[r.PostID] = append(reactionsByPost[r.PostID], r)
	}

	repliesByID := make(map[string]model.Post, len(replyPosts))
	for _, reply := range replyPosts {
		reply.Metadata.Files = orEmptyAttachments(attachmentsByPost[reply.ID])
		repliesByID[reply.ID] = reply
	}

	for i := range posts {
		posts[i].Metadata.Files = orEmptyAttachments(attachmentsByPost[posts[i].ID])

		if r := reactionsByPost[posts[i].ID]; r != nil {
			posts[i].Reactions = r
		} else {
			posts[i].Reactions = make([]model.PostReaction, 0)
		}

		if posts[i].ReplyID != "" {
			if reply, ok := repliesByID[posts[i].ReplyID]; ok {
				posts[i].Reply = &reply
			}
		}
	}

	return &model.PostResponse{Posts: posts, ReplyPosts: replyPosts}, nil
}

func orEmptyAttachments(a []model.PostFileAttachment) []model.PostFileAttachment {
	if a == nil {
		return make([]model.PostFileAttachment, 0)
	}

	return a
}

const (
	postColumns = `id, user_id, channel_id, reply_id, message, type, created_at, updated_at, deleted_at`

	postAttachmentColumnNames = `id, channel_id, post_id, user_id, name, size, mime_type, width, height, ` +
		`created_at, updated_at, deleted_at, url, provider, kind, storage_id`

	// Reads coalesce storage_id, which is nullable, so the scan can take a string.
	postAttachmentColumns = `id, channel_id, post_id, user_id, name, size, mime_type, width, height,
	       created_at, updated_at, deleted_at, url, provider, kind, COALESCE(storage_id, '')`
)

func scanPost(rows *sql.Rows, post *model.Post) error {
	return rows.Scan(&post.ID, &post.UserID, &post.ChannelID, &post.ReplyID, &post.Message,
		&post.Type, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt)
}

// Order is not guaranteed; callers index the result by id.
func (p *postRepository) GetByIDs(ctx context.Context, postIDs []string) ([]model.Post, error) {
	if len(postIDs) == 0 {
		return make([]model.Post, 0), nil
	}

	args := make([]any, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	rows, err := p.Db.QueryContext(ctx,
		`SELECT `+postColumns+` FROM posts WHERE id IN (`+sqlPlaceholders(len(postIDs))+`)`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]model.Post, 0)
	for rows.Next() {
		post := model.Post{}
		if err := scanPost(rows, &post); err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (p *postRepository) GetAttachments(id string) ([]model.PostFileAttachment, error) {
	attachmentsResults, err := p.Db.Query(`
		SELECT `+postAttachmentColumns+`
		FROM post_attachments WHERE post_id=?`, id)
	if err != nil {
		return nil, err
	}

	defer attachmentsResults.Close()

	attachments := make([]model.PostFileAttachment, 0)
	for attachmentsResults.Next() {
		attachment := model.PostFileAttachment{}
		err := scanPostAttachment(attachmentsResults, &attachment)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	if err := attachmentsResults.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (p *postRepository) Get(ctx context.Context, postID string) (*model.Post, error) {
	posts, err := p.GetByIDs(ctx, []string{postID})
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, sql.ErrNoRows
	}

	post := posts[0]

	if post.ReplyID != "" {
		replies, err := p.GetByIDs(ctx, []string{post.ReplyID})
		if err != nil {
			return nil, err
		}

		if len(replies) == 0 {
			return nil, sql.ErrNoRows
		}

		post.Reply = &replies[0]
	}

	attachments, err := p.GetAttachmentsForMany(ctx, []string{post.ID})
	if err != nil {
		return nil, err
	}

	post.Metadata.Files = attachments

	reactions, err := p.GetReactionsForMany(ctx, []string{post.ID})
	if err != nil {
		return nil, err
	}

	post.Reactions = reactions

	return &post, nil
}

func (p *postRepository) Create(post model.Post) (*model.Post, error) {
	t := time.Now().UnixMilli()

	tx, err := p.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO posts ("+postColumns+") VALUES(?,?,?,?,?,?,?,?,?)", post.ID, post.UserID, post.ChannelID, post.ReplyID, post.Message, post.Type, post.CreatedAt, post.UpdatedAt, 0)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE channels SET msg_count=msg_count+1, last_post=? WHERE id=?", post.CreatedAt, post.ChannelID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE channel_members SET last_viewed_at=?, updated_at=?, mention_count=0, msg_count=(SELECT msg_count FROM channels WHERE id=?) WHERE user_id=? AND channel_id=?", t, t, post.ChannelID, post.UserID, post.ChannelID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	post.Reactions = make([]model.PostReaction, 0)

	return &post, nil
}

func (p *postRepository) Update(id string, message string, remove []string, attachments []model.PostFileAttachment) (bool, error) {
	t := time.Now().UnixMilli()

	tx, err := p.Db.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	_, err = tx.Exec("UPDATE posts SET message=?, updated_at=? WHERE id=?", message, t, id)
	if err != nil {
		return false, err
	}

	if len(remove) > 0 {
		for _, v := range remove {
			_, err := tx.Exec("DELETE FROM post_attachments WHERE id=?", v)
			if err != nil {
				return false, err
			}
		}
	}

	if len(attachments) > 0 {
		for _, v := range attachments {
			_, err := tx.Exec(`INSERT INTO post_attachments
				(id, channel_id, post_id, user_id, name, size, mime_type, width, height, created_at, updated_at, deleted_at, url, provider, kind, storage_id)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				v.ID, v.ChannelID, v.PostID, v.UserID, v.Name, v.Size, v.MimeType, v.Width, v.Height,
				t, t, 0, v.URL, v.Provider, v.Kind, nullableString(v.StorageID))
			if err != nil {
				return false, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

// The children go in the same transaction as the post, or a failure halfway
// leaves rows pointing at a post that no longer exists.
func (p *postRepository) Delete(postID string) error {
	tx, err := p.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, table := range []string{"post_reactions", "post_attachments"} {
		if _, err := tx.Exec("DELETE FROM "+table+" WHERE post_id=?", postID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("DELETE FROM posts WHERE id=?", postID); err != nil {
		return err
	}

	return tx.Commit()
}

func (p *postRepository) MarkAllRead(userID, channelID string) error {
	t := time.Now().UnixMilli()

	_, err := p.Db.Exec("UPDATE channel_members SET last_viewed_at=?, updated_at=?, mention_count=0, msg_count=(SELECT msg_count FROM channels WHERE id=?) WHERE user_id=? AND channel_id=?", t, t, channelID, userID, channelID)
	if err != nil {
		return err
	}

	return nil
}

func (p *postRepository) CreateReaction(postID, channelID, userID, reaction string) (*model.PostReaction, error) {
	t := time.Now().UnixMilli()
	id := model.NewID()

	_, err := p.Db.Exec(`
    INSERT INTO post_reactions
    (id, channel_id, post_id, user_id, reaction, created_at, updated_at, deleted_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE updated_at = ?
`, id, channelID, postID, userID, reaction, t, t, 0, t)
	if err != nil {
		return nil, err
	}

	return &model.PostReaction{
		ID:        id,
		ChannelID: channelID,
		PostID:    postID,
		UserID:    userID,
		Reaction:  reaction,
		CreatedAt: t,
		UpdatedAt: t,
		DeletedAt: t,
	}, nil
}

func (p *postRepository) DeleteReaction(postID, channelID, userID, reaction string) error {
	_, err := p.Db.Exec("DELETE FROM post_reactions WHERE post_id=? AND channel_id=? AND user_id=? AND reaction=?", postID, channelID, userID, reaction)
	if err != nil {
		return err
	}

	return nil
}

func (p *postRepository) CreateAttachment(files []model.PostFileAttachment) ([]model.PostFileAttachment, error) {
	if len(files) == 0 {
		return files, nil
	}

	t := time.Now().UnixMilli()

	args := make([]any, 0, len(files)*16)
	for _, v := range files {
		args = append(args,
			v.ID, v.ChannelID, v.PostID, v.UserID, v.Name, v.Size, v.MimeType,
			v.Width, v.Height, t, t, 0, v.URL, v.Provider, v.Kind,
			nullableString(v.StorageID))
	}

	_, err := p.Db.Exec(
		`INSERT INTO post_attachments (`+postAttachmentColumnNames+`)
		 VALUES `+rowPlaceholders(len(files), 16), args...)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (p *postRepository) GetAttachmentByID(id string) (*model.PostFileAttachment, error) {
	row := p.Db.QueryRow(`
		SELECT id, channel_id, post_id, user_id, name, size, mime_type, width, height,
		       created_at, updated_at, deleted_at, url, provider, kind,
		       COALESCE(storage_id, '')
		FROM post_attachments WHERE id=?`, id)

	attachment := model.PostFileAttachment{}
	err := row.Scan(&attachment.ID, &attachment.ChannelID, &attachment.PostID, &attachment.UserID, &attachment.Name,
		&attachment.Size, &attachment.MimeType,
		&attachment.Width, &attachment.Height, &attachment.CreatedAt, &attachment.UpdatedAt, &attachment.DeletedAt,
		&attachment.URL, &attachment.Provider, &attachment.Kind, &attachment.StorageID)

	if err != nil && err == sql.ErrNoRows {
		return nil, err
	}

	return &attachment, nil
}

func (p *postRepository) CreateLinkMetadata(metadata model.LinkMetadata) error {
	data, err := json.Marshal(metadata.Data)
	if err != nil {
		return err
	}

	_, err = p.Db.Exec("INSERT INTO linkmetadata (hash, url, type, data, storage_id, size, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?)",
		metadata.Hash, metadata.URL, metadata.Type, data, nullableString(metadata.StorageID), metadata.Size, metadata.CreatedAt, metadata.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (p *postRepository) UpdateLinkMetadata(metadata model.LinkMetadata) error {
	data, err := json.Marshal(metadata.Data)
	if err != nil {
		return err
	}

	_, err = p.Db.Exec("UPDATE linkmetadata SET url=?, type=?, data=?, storage_id=?, size=?, updated_at=? WHERE hash=?",
		metadata.URL, metadata.Type, data, nullableString(metadata.StorageID), metadata.Size, metadata.UpdatedAt, metadata.Hash)
	if err != nil {
		return err
	}

	return nil
}

func (p *postRepository) GetLinkMetadataByHash(hash int64) (*model.LinkMetadata, error) {
	row := p.Db.QueryRow("SELECT hash, url, type, data, storage_id, size, created_at, updated_at FROM linkmetadata WHERE hash=?", hash)

	data := make([]byte, 0)
	var storageID sql.NullString

	metadata := model.LinkMetadata{}
	err := row.Scan(&metadata.Hash, &metadata.URL, &metadata.Type, &data, &storageID, &metadata.Size, &metadata.CreatedAt, &metadata.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	metadata.StorageID = storageID.String

	err = json.Unmarshal(data, &metadata.Data)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (p *postRepository) GetBatchForChannel(ctx context.Context, channelID string, limit int) ([]string, error) {
	rows, err := p.Db.QueryContext(ctx, "SELECT id FROM posts WHERE channel_id=? LIMIT ?", channelID, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func scanPostAttachment(rows *sql.Rows, a *model.PostFileAttachment) error {
	return rows.Scan(&a.ID, &a.ChannelID, &a.PostID, &a.UserID, &a.Name,
		&a.Size, &a.MimeType, &a.Width, &a.Height, &a.CreatedAt,
		&a.UpdatedAt, &a.DeletedAt, &a.URL, &a.Provider, &a.Kind, &a.StorageID)
}

func (p *postRepository) GetReactionsForMany(ctx context.Context, postIDs []string) ([]model.PostReaction, error) {
	if len(postIDs) == 0 {
		return make([]model.PostReaction, 0), nil
	}

	args := make([]any, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	rows, err := p.Db.QueryContext(ctx, `
		SELECT id, channel_id, post_id, user_id, reaction, created_at, updated_at, deleted_at
		FROM post_reactions WHERE post_id IN (`+sqlPlaceholders(len(postIDs))+`)`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reactions := make([]model.PostReaction, 0)
	for rows.Next() {
		r := model.PostReaction{}
		if err := rows.Scan(&r.ID, &r.ChannelID, &r.PostID, &r.UserID, &r.Reaction,
			&r.CreatedAt, &r.UpdatedAt, &r.DeletedAt); err != nil {
			return nil, err
		}

		reactions = append(reactions, r)
	}

	return reactions, rows.Err()
}

// The rows have to be read before they are deleted: they are the only record of
// which objects to remove from storage.
func (p *postRepository) GetAttachmentsForMany(ctx context.Context, postIDs []string) ([]model.PostFileAttachment, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	args := make([]any, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	rows, err := p.Db.QueryContext(ctx, `
		SELECT `+postAttachmentColumns+`
		FROM post_attachments WHERE post_id IN (`+sqlPlaceholders(len(postIDs))+`)`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	attachments := make([]model.PostFileAttachment, 0)
	for rows.Next() {
		a := model.PostFileAttachment{}
		if err := scanPostAttachment(rows, &a); err != nil {
			return nil, err
		}

		attachments = append(attachments, a)
	}

	return attachments, rows.Err()
}

func (p *postRepository) DeleteBatch(ctx context.Context, postIDs []string) error {
	if len(postIDs) == 0 {
		return nil
	}

	args := make([]any, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	in := sqlPlaceholders(len(postIDs))

	tx, err := p.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, table := range []string{"post_reactions", "post_attachments"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE post_id IN ("+in+")", args...); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM posts WHERE id IN ("+in+")", args...); err != nil {
		return err
	}

	return tx.Commit()
}
