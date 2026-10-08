// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type storageRepository struct {
	Db *sql.DB
}

func NewStorageRepository(Db *sql.DB) (*storageRepository, error) {
	return &storageRepository{Db: Db}, nil
}

func (s *storageRepository) Create(storage *model.Storage) (*model.Storage, error) {
	storage.ID = model.NewID()
	storage.CreatedAt = time.Now().Unix()

	_, err := s.Db.Exec(`
		INSERT INTO file_storage
			(id, label, description, type, created_at, directory, endpoint, access_key, secret_key, bucket, use_ssl, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		storage.ID,
		storage.Label,
		storage.Description,
		storage.Type,
		storage.CreatedAt,
		storage.Directory,
		storage.Endpoint,
		storage.AccessKey,
		storage.SecretKey,
		storage.Bucket,
		storage.SSL,
		storage.Primary,
	)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *storageRepository) GetAll() ([]model.Storage, error) {
	rows, err := s.Db.Query(`
		SELECT id, label, description, type, created_at,
		       COALESCE(directory, ''),
		       COALESCE(endpoint, ''),
		       COALESCE(access_key, ''),
		       COALESCE(secret_key, ''),
		       COALESCE(bucket, ''),
		       COALESCE(use_ssl, 1),
		       is_primary
		FROM file_storage`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	storages := make([]model.Storage, 0)
	for rows.Next() {
		var storage model.Storage
		err := rows.Scan(
			&storage.ID,
			&storage.Label,
			&storage.Description,
			&storage.Type,
			&storage.CreatedAt,
			&storage.Directory,
			&storage.Endpoint,
			&storage.AccessKey,
			&storage.SecretKey,
			&storage.Bucket,
			&storage.SSL,
			&storage.Primary,
		)
		if err != nil {
			return nil, err
		}

		storages = append(storages, storage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return storages, nil
}

func (s *storageRepository) GetByID(id string) (*model.Storage, error) {
	storage := &model.Storage{}
	err := s.Db.QueryRow(`
		SELECT id, label, description, type, created_at,
		       COALESCE(directory, ''),
		       COALESCE(endpoint, ''),
		       COALESCE(access_key, ''),
		       COALESCE(secret_key, ''),
		       COALESCE(bucket, ''),
		       COALESCE(use_ssl, 1),
		       is_primary
		FROM file_storage
		WHERE id = ?`, id).Scan(
		&storage.ID,
		&storage.Label,
		&storage.Description,
		&storage.Type,
		&storage.CreatedAt,
		&storage.Directory,
		&storage.Endpoint,
		&storage.AccessKey,
		&storage.SecretKey,
		&storage.Bucket,
		&storage.SSL,
		&storage.Primary,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *storageRepository) GetPrimary() (*model.Storage, error) {
	storage := &model.Storage{}
	err := s.Db.QueryRow(`
		SELECT id, label, description, type, created_at,
		       COALESCE(directory, ''),
		       COALESCE(endpoint, ''),
		       COALESCE(access_key, ''),
		       COALESCE(secret_key, ''),
		       COALESCE(bucket, ''),
		       COALESCE(use_ssl, 1),
		       is_primary
		FROM file_storage
		WHERE is_primary = 1
		LIMIT 1`).Scan(
		&storage.ID,
		&storage.Label,
		&storage.Description,
		&storage.Type,
		&storage.CreatedAt,
		&storage.Directory,
		&storage.Endpoint,
		&storage.AccessKey,
		&storage.SecretKey,
		&storage.Bucket,
		&storage.SSL,
		&storage.Primary,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *storageRepository) UpdatePrimary(id string) error {
	tx, err := s.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err = tx.Exec(`UPDATE file_storage SET is_primary = 0`); err != nil {
		return err
	}

	if _, err = tx.Exec(`UPDATE file_storage SET is_primary = 1 WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *storageRepository) Delete(id string) error {
	_, err := s.Db.Exec(`DELETE FROM file_storage WHERE id = ?`, id)
	return err
}

func (s *storageRepository) Update(id string, storage *model.Storage) error {
	_, err := s.Db.Exec(`
		UPDATE file_storage
		SET label=?, description=?, endpoint=?, access_key=?, secret_key=?, bucket=?, use_ssl=?
		WHERE id=?`,
		storage.Label,
		storage.Description,
		storage.Endpoint,
		storage.AccessKey,
		storage.SecretKey,
		storage.Bucket,
		storage.SSL,
		id,
	)
	return err
}

func (s *storageRepository) UpdateCredentials(id string, storage *model.Storage) error {
	_, err := s.Db.Exec(`
		UPDATE file_storage
		SET type=?, directory=?, endpoint=?, access_key=?, secret_key=?, bucket=?, use_ssl=?
		WHERE id=?`,
		storage.Type,
		storage.Directory,
		storage.Endpoint,
		storage.AccessKey,
		storage.SecretKey,
		storage.Bucket,
		storage.SSL,
		id,
	)
	return err
}
