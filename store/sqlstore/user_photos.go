// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type userPhotoRepository struct {
	db *sql.DB
}

func NewUserPhotoRepository(db *sql.DB) *userPhotoRepository {
	return &userPhotoRepository{db: db}
}

func (r *userPhotoRepository) Get(ctx context.Context, userID string) (*model.UserPhoto, error) {
	var photo model.UserPhoto
	var storageID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, photo_id, storage_id
		FROM user_photos WHERE user_id = ?`, userID).Scan(
		&photo.UserID, &photo.PhotoID, &storageID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	photo.StorageID = storageID.String
	return &photo, nil
}

func (r *userPhotoRepository) Replace(ctx context.Context, photo model.UserPhoto) (*model.UserPhoto, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var old model.UserPhoto
	var oldStorageID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT user_id, photo_id, storage_id
		FROM user_photos WHERE user_id = ? FOR UPDATE`, photo.UserID).Scan(
		&old.UserID, &old.PhotoID, &oldStorageID,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	old.StorageID = oldStorageID.String

	if err == sql.ErrNoRows {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO user_photos (user_id, photo_id, storage_id)
			VALUES (?, ?, ?)`, photo.UserID, photo.PhotoID, photo.StorageID); err != nil {
			return nil, err
		}
	} else {
		if _, err = tx.ExecContext(ctx, `
			UPDATE user_photos SET photo_id = ?, storage_id = ?
			WHERE user_id = ?`, photo.PhotoID, photo.StorageID, photo.UserID); err != nil {
			return nil, err
		}
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET updated_at = ? WHERE id = ?`,
		time.Now().Unix(), photo.UserID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if old.UserID == "" {
		return nil, nil
	}

	return &old, nil
}

func (r *userPhotoRepository) Delete(ctx context.Context, userID, photoID string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	var currentID string
	err = tx.QueryRowContext(ctx, `SELECT photo_id FROM user_photos WHERE user_id = ? FOR UPDATE`,
		userID).Scan(&currentID)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if currentID != photoID {
		return false, nil
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM user_photos WHERE user_id = ?`, userID); err != nil {
		return false, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET updated_at = ? WHERE id = ?`,
		time.Now().Unix(), userID)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (r *userPhotoRepository) CountByStorage(ctx context.Context, storageID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_photos WHERE storage_id = ?`, storageID).Scan(&count)
	return count, err
}
