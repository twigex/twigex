// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"errors"

	"github.com/twigex/twigex/model"
)

type uploadSessionRepository struct {
	Db *sql.DB
}

func NewUploadSessionRepository(Db *sql.DB) *uploadSessionRepository {

	return &uploadSessionRepository{
		Db: Db,
	}
}

func (r *uploadSessionRepository) Create(
	session model.UploadSession,
) (*model.UploadSession, error) {

	query := `
		INSERT INTO upload_sessions (
			id,
			upload_id,
			user_id,
			context_type,
			context_id,
			storage,
			file_name,
			mime_type,
			total_size,
			uploaded_size,
			uploaded_parts,
			status,
			parts_json,          
			backend_metadata,   
			created_at,
			updated_at,
			expires_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.Db.Exec(
		query,
		session.ID,
		session.UploadID,
		session.UserID,
		session.ContextType,
		session.ContextID,
		session.Storage,
		session.FileName,
		session.MimeType,
		session.TotalSize,
		session.UploadedSize,
		session.UploadedParts,
		session.Status,
		session.PartsJSON,       // Could be nil
		session.BackendMetadata, // Could be nil
		session.CreatedAt,
		session.UpdatedAt,
		session.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *uploadSessionRepository) Get(
	id string,
) (*model.UploadSession, error) {

	query := `
		SELECT
			id,
			upload_id,
			user_id,
			context_type,
			context_id,
			storage,
			file_name,
			mime_type,
			total_size,
			uploaded_size,
			uploaded_parts,
			status,
			error_reason,        
			parts_json,          
			backend_metadata,   
			created_at,
			updated_at,
			expires_at
		FROM upload_sessions
		WHERE id = ?
		LIMIT 1
	`

	var (
		session             model.UploadSession
		contextIDNullable   sql.NullString
		expiresAtNullable   sql.NullInt64
		errorReasonNullable sql.NullString
		partsJSONNullable   sql.NullString
		backendMetaNullable sql.NullString
	)

	err := r.Db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UploadID,
		&session.UserID,
		&session.ContextType,
		&contextIDNullable,
		&session.Storage,
		&session.FileName,
		&session.MimeType,
		&session.TotalSize,
		&session.UploadedSize,
		&session.UploadedParts,
		&session.Status,
		&errorReasonNullable,
		&partsJSONNullable,
		&backendMetaNullable,
		&session.CreatedAt,
		&session.UpdatedAt,
		&expiresAtNullable,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	if contextIDNullable.Valid {
		session.ContextID = contextIDNullable.String
	}

	if expiresAtNullable.Valid {
		session.ExpiresAt = expiresAtNullable.Int64
	}

	if errorReasonNullable.Valid {
		session.ErrorReason = &errorReasonNullable.String
	}

	if partsJSONNullable.Valid {
		session.PartsJSON = &partsJSONNullable.String
	}

	if backendMetaNullable.Valid {
		session.BackendMetadata = &backendMetaNullable.String
	}

	return &session, nil
}

func (r *uploadSessionRepository) UpdateParts(
	id string,
	partsJSON string,
	uploadedSize int64,
	uploadedParts int,
	now int64,
) error {

	query := `
		UPDATE upload_sessions
		SET parts_json = ?,
		    uploaded_size = ?,
		    uploaded_parts = ?,
		    updated_at = ?
		WHERE id = ?
	`

	// Handle NULL parts_json
	var partsJSONArg interface{}
	if partsJSON == "" {
		partsJSONArg = nil
	} else {
		partsJSONArg = partsJSON
	}

	res, err := r.Db.Exec(
		query,
		partsJSONArg,
		uploadedSize,
		uploadedParts,
		now,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *uploadSessionRepository) UpdateStatus(
	id string,
	status model.UploadSessionStatus,
	now int64,
) error {

	query := `
		UPDATE upload_sessions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	res, err := r.Db.Exec(query, status, now, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *uploadSessionRepository) Delete(id string) error {
	query := `DELETE FROM upload_sessions WHERE id = ?`

	res, err := r.Db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
