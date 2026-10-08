// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type passwordResetRepository struct {
	Db *sql.DB
}

const passwordResetColumns = `id, user_id, token, created_at, expires_at`

func NewPasswordResetRepository(Db *sql.DB) (*passwordResetRepository, error) {
	repo := &passwordResetRepository{}

	repo.Db = Db
	return repo, nil
}

func (p *passwordResetRepository) Create(user model.User, token string) error {
	t := time.Now().Unix()
	expires := time.Now().Add(time.Minute * 30).Unix()

	_, err := p.Db.Exec(`INSERT INTO password_resets (`+passwordResetColumns+`) VALUES(?, ?, ?, ?, ?)`, model.NewID(), user.ID, token, t, expires)
	if err != nil {
		return err
	}

	return nil
}

func (p *passwordResetRepository) Delete(token string) error {
	_, err := p.Db.Exec(`DELETE FROM password_resets WHERE token=?`, token)
	if err != nil {
		return err
	}

	return nil
}

func (p *passwordResetRepository) Get(token string) (*model.PasswordReset, error) {
	row := p.Db.QueryRow(`SELECT `+passwordResetColumns+` FROM password_resets WHERE token=?`, token)

	pr := model.PasswordReset{}

	err := row.Scan(&pr.ID, &pr.UserID, &pr.Token, &pr.CreatedAt, &pr.ExpiresAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, err
	}

	return &pr, nil
}
