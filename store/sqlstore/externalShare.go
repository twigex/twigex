// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type shareRepository struct {
	Db *sql.DB
}

func NewShareRepository(Db *sql.DB) (*shareRepository, error) {
	repo := &shareRepository{}

	repo.Db = Db
	return repo, nil
}

const externalShareColumns = `id, owner, file_id, password_protected, password, source, parent,
	share_token, share_time, message, expiration, allow_download, allow_upload,
	downloaded, accessed, active, created_at, updated_at, allow_view, last_accessed_at, max_downloads, allow_edit`

func scanExternalShare(rows *sql.Rows, sh *model.ExternalShare) error {
	return rows.Scan(&sh.ID, &sh.Owner, &sh.FileID, &sh.PasswordProtected, &sh.Password, &sh.Source, &sh.Parent,
		&sh.ShareToken, &sh.ShareTime, &sh.Message, &sh.Expiration, &sh.AllowDownload, &sh.AllowUpload,
		&sh.Downloaded, &sh.Accessed, &sh.Active, &sh.CreatedAt, &sh.UpdatedAt,
		&sh.AllowView, &sh.LastAccessedAt, &sh.MaxDownloads, &sh.AllowEdit)
}

func (s *shareRepository) Create(ctx context.Context, userID string, shareToken string, share model.Request) (*model.SharedLinks, error) {
	created := time.Now().Unix()

	id := model.NewID()

	if share.Source == "" {
		share.Source = shareToken
	}

	_, err := s.Db.ExecContext(ctx, `INSERT INTO external_share
		(id, owner, file_id, password_protected, password, source, parent, share_token,
		share_time, message, expiration, allow_download, allow_upload, downloaded, accessed,
		active, created_at, updated_at, allow_view, last_accessed_at, max_downloads, allow_edit)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, userID, share.Item, share.PasswordProtected, share.Password, share.Source, share.Parent, shareToken,
		created, share.Message, share.Expiration, share.AllowDownload, share.AllowUpload, 0, 0, 1, created, created,
		share.AllowView, 0, share.MaxDownloads, share.AllowEdit)
	if err != nil {
		return nil, err
	}

	return &model.SharedLinks{
		ID:                shareToken,
		ShareTime:         int(created),
		Expiration:        share.Expiration,
		PasswordProtected: share.PasswordProtected,
		AllowView:         share.AllowView,
		AllowDownload:     share.AllowDownload,
		AllowUpload:       share.AllowUpload,
		AllowEdit:         share.AllowEdit,
		MaxDownloads:      share.MaxDownloads,
		Message:           share.Message,
	}, nil
}

func (s *shareRepository) GetByToken(ctx context.Context, token string) (*model.ExternalShare, error) {
	rows, err := s.Db.QueryContext(ctx, "SELECT "+externalShareColumns+" FROM external_share WHERE share_token=? AND active=?", token, true)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	sh := model.ExternalShare{}
	if err = scanExternalShare(rows, &sh); err != nil {
		return nil, err
	}

	return &sh, rows.Err()
}

// GetActiveForFile returns every active link on the file regardless of creator,
// for owners and managers who govern the file's whole external exposure.
func (s *shareRepository) GetActiveForFile(ctx context.Context, fileID string) ([]model.ExternalShare, error) {
	rows, err := s.Db.QueryContext(ctx, "SELECT "+externalShareColumns+" FROM external_share WHERE file_id=? AND active=?", fileID, true)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	arr := make([]model.ExternalShare, 0)
	for rows.Next() {
		sh := model.ExternalShare{}
		if err = scanExternalShare(rows, &sh); err != nil {
			return nil, err
		}

		arr = append(arr, sh)
	}

	return arr, rows.Err()
}

func (s *shareRepository) Update(ctx context.Context, owner string, source string, l model.LinkUpdate) error {
	_, err := s.Db.ExecContext(ctx, `UPDATE external_share SET password_protected=?, password=?,
		message=?, expiration=?, allow_view=?, allow_download=?, allow_upload=?, allow_edit=?, max_downloads=?, updated_at=?
		WHERE owner=? AND source=?`,
		l.PasswordProtected, l.Password, l.Message, l.Expiration, l.AllowView, l.AllowDownload, l.AllowUpload,
		l.AllowEdit, l.MaxDownloads, time.Now().Unix(), owner, source)
	return err
}

func (s *shareRepository) Delete(ctx context.Context, owner string, token string) error {
	_, err := s.Db.ExecContext(ctx, `DELETE FROM external_share WHERE owner=? AND source=?`, owner, token)
	return err
}

func (s *shareRepository) DeactivateForFile(ctx context.Context, fileID string) error {
	_, err := s.Db.ExecContext(ctx, `UPDATE external_share SET active=0, updated_at=? WHERE file_id=?`, time.Now().Unix(), fileID)
	return err
}

func (s *shareRepository) IncrementAccessed(ctx context.Context, token string) error {
	_, err := s.Db.ExecContext(ctx, `UPDATE external_share SET accessed=accessed+1, last_accessed_at=? WHERE share_token=?`, time.Now().Unix(), token)
	return err
}

func (s *shareRepository) IncrementDownloaded(ctx context.Context, token string) error {
	_, err := s.Db.ExecContext(ctx, `UPDATE external_share SET downloaded=downloaded+1, last_accessed_at=? WHERE share_token=?`, time.Now().Unix(), token)
	return err
}
