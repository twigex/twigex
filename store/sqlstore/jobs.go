// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/twigex/twigex/model"
)

type jobsRepository struct {
	Db *sql.DB
}

const jobColumns = `id, type, status, user_id, payload, progress, error, created_at, started_at, finished_at, updated_at, acknowledged_at`

func NewJobsRepository(Db *sql.DB) *jobsRepository {
	return &jobsRepository{
		Db: Db,
	}
}

func (r *jobsRepository) Create(job model.Job) (*model.Job, error) {
	_, err := r.Db.Exec("INSERT INTO jobs ("+jobColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		job.ID,
		job.Type,
		job.Status,
		sql.NullString{String: job.UserID, Valid: job.UserID != ""},
		job.Payload,
		job.Progress,
		job.Error,
		job.CreatedAt,
		job.StartedAt,
		job.FinishedAt,
		job.UpdatedAt,
		job.AcknowledgedAt,
	)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// COALESCE on user_id so this keeps working once system jobs store NULL there.
func (r *jobsRepository) ClaimNextPending() (*model.Job, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row := tx.QueryRow(`
		SELECT
			id,
			type,
			status,
			COALESCE(user_id, ''),
			payload,
			progress,
			error,
			created_at,
			started_at,
			finished_at,
			updated_at,
			acknowledged_at
		FROM jobs
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, model.JobStatusPending)

	var job model.Job

	err = row.Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.UserID,
		&job.Payload,
		&job.Progress,
		&job.Error,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.UpdatedAt,
		&job.AcknowledgedAt,
	)

	if err == sql.ErrNoRows {
		_ = tx.Commit()
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()

	_, err = tx.Exec(`
		UPDATE jobs
		SET status = ?, started_at = ?, updated_at = ?
		WHERE id = ?
	`, model.JobStatusRunning, now, now, job.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	job.Status = model.JobStatusRunning
	job.StartedAt = now
	job.UpdatedAt = now

	return &job, nil
}

func (r *jobsRepository) UpdateStatus(id string, status string, result string) error {
	t := time.Now().Unix()

	finished := int64(0)
	if model.IsJobFinished(status) {
		finished = t
	}

	_, err := r.Db.Exec("UPDATE jobs SET status = ?, error = ?, updated_at = ?, finished_at = ? WHERE id = ?", status, result, t, finished, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *jobsRepository) IsCancelled(jobID string) (bool, error) {
	var status string
	err := r.Db.QueryRow(
		`SELECT status FROM jobs WHERE id = ?`,
		jobID,
	).Scan(&status)

	if err != nil {
		return false, err
	}

	return status == model.JobStatusCancelled, nil
}

func (r *jobsRepository) UpdatePayload(id string, payload json.RawMessage) error {
	_, err := r.Db.Exec(
		"UPDATE jobs SET payload = ?, updated_at = ? WHERE id = ?",
		[]byte(payload),
		time.Now().Unix(),
		id,
	)

	return err
}

func (r *jobsRepository) UpdateProgress(id string, progress int) error {
	t := time.Now().Unix()

	_, err := r.Db.Exec("UPDATE jobs SET progress = ?, updated_at = ? WHERE id = ?", progress, t, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *jobsRepository) GetRunningForUser(userID string) ([]model.Job, error) {
	rows, err := r.Db.Query(`
		SELECT `+jobColumns+`
		FROM jobs
		WHERE user_id = ?
		  AND (
			  status IN (?, ?)
			  OR (
				  status IN (?, ?, ?)
				  AND acknowledged_at IS NULL
				  AND finished_at >= UNIX_TIMESTAMP() - 60
			  )
		  )
		ORDER BY created_at DESC
	`,
		userID,
		model.JobStatusPending,
		model.JobStatusRunning,
		model.JobStatusCompleted,
		model.JobStatusFailed,
		model.JobStatusCancelled,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	jobs := make([]model.Job, 0)

	for rows.Next() {
		var job model.Job

		err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Status,
			&job.UserID,
			&job.Payload,
			&job.Progress,
			&job.Error,
			&job.CreatedAt,
			&job.StartedAt,
			&job.FinishedAt,
			&job.UpdatedAt,
			&job.AcknowledgedAt,
		)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *jobsRepository) GetForUserByID(userID, id string) (*model.Job, error) {
	var job model.Job

	err := r.Db.QueryRow(
		`SELECT `+jobColumns+` FROM jobs WHERE id = ? AND user_id = ?`,
		id,
		userID,
	).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.UserID,
		&job.Payload,
		&job.Progress,
		&job.Error,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.UpdatedAt,
		&job.AcknowledgedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &job, nil
}

func (r *jobsRepository) Get(id string) (*model.Job, error) {
	var job model.Job

	err := r.Db.QueryRow(
		`SELECT
			id,
			type,
			status,
			COALESCE(user_id, ''),
			payload,
			progress,
			error,
			created_at,
			started_at,
			finished_at,
			updated_at,
			acknowledged_at
		FROM jobs WHERE id = ?`,
		id,
	).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.UserID,
		&job.Payload,
		&job.Progress,
		&job.Error,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.UpdatedAt,
		&job.AcknowledgedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &job, nil
}

func (r *jobsRepository) GetLatestByType(jobType string) (*model.Job, error) {
	var job model.Job

	err := r.Db.QueryRow(
		`SELECT
			id,
			type,
			status,
			COALESCE(user_id, ''),
			payload,
			progress,
			error,
			created_at,
			started_at,
			finished_at,
			updated_at,
			acknowledged_at
		FROM jobs
		WHERE type = ?
		ORDER BY created_at DESC
		LIMIT 1`,
		jobType,
	).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.UserID,
		&job.Payload,
		&job.Progress,
		&job.Error,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.UpdatedAt,
		&job.AcknowledgedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *jobsRepository) CountActiveByType(jobType string) (int, error) {
	var count int

	err := r.Db.QueryRow(
		`SELECT COUNT(*) FROM jobs WHERE type = ? AND status IN (?, ?)`,
		jobType,
		model.JobStatusPending,
		model.JobStatusRunning,
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *jobsRepository) AcknowledgeForUser(userID, id string) error {
	t := time.Now().Unix()

	_, err := r.Db.Exec(
		`UPDATE jobs
		 SET acknowledged_at = ?
		 WHERE id = ?
		   AND user_id = ?
		   AND acknowledged_at IS NULL`,
		t,
		id,
		userID,
	)

	return err
}

// Keeps the newest unsuccessful run per type until a later run supersedes it.
// The derived table is required: MySQL rejects a subquery naming the delete target.
func (r *jobsRepository) DeleteOld(finishedBefore int64) error {
	_, err := r.Db.Exec(`
		DELETE FROM jobs
		WHERE finished_at > 0
		  AND finished_at < ?
		  AND id NOT IN (
			SELECT id FROM (
				SELECT j.id
				FROM jobs j
				WHERE j.finished_at > 0
				  AND j.status <> ?
				  AND NOT EXISTS (
					SELECT 1
					FROM jobs newer
					WHERE newer.type = j.type
					  AND newer.finished_at > 0
					  AND newer.created_at > j.created_at
				  )
			) AS keep_rows
		  )
	`,
		finishedBefore,
		model.JobStatusCompleted,
	)
	return err
}
