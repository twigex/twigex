// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"

	"github.com/twigex/twigex/model"
)

type preferencesRepository struct {
	Db *sql.DB
}

func NewPreferencesRepository(Db *sql.DB) (*preferencesRepository, error) {
	return &preferencesRepository{Db: Db}, nil
}

func (m *preferencesRepository) scanPreferences(rows *sql.Rows) ([]model.Preference, error) {
	defer rows.Close()
	preferences := make([]model.Preference, 0)
	for rows.Next() {
		p := model.Preference{}
		if err := rows.Scan(&p.UserID, &p.Category, &p.Name, &p.Value); err != nil {
			return nil, err
		}

		preferences = append(preferences, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return preferences, nil
}

func (m *preferencesRepository) GetForUser(userID, category string) ([]model.Preference, error) {
	rows, err := m.Db.Query("SELECT user_id, category, name, value FROM preferences WHERE user_id=? AND category=?", userID, category)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return m.scanPreferences(rows)
}

func (m *preferencesRepository) GetNotifications(userID, category string) ([]model.Preference, error) {
	rows, err := m.Db.Query("SELECT user_id, category, name, value FROM preferences WHERE user_id=? AND category=?", userID, category)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return m.scanPreferences(rows)
}

func (m *preferencesRepository) Update(preferences []model.Preference) error {
	if len(preferences) == 0 {
		return nil
	}

	args := make([]any, 0, len(preferences)*4)
	for _, p := range preferences {
		args = append(args, p.UserID, p.Category, p.Name, p.Value)
	}

	_, err := m.Db.Exec(`
		INSERT INTO preferences (user_id, category, name, value)
		VALUES `+rowPlaceholders(len(preferences), 4)+`
		ON DUPLICATE KEY UPDATE value = VALUES(value)`, args...)

	return err
}
