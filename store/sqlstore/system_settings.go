// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
)

type systemSettingsRepository struct {
	Db *sql.DB
}

func NewSystemSettingsRepository(Db *sql.DB) *systemSettingsRepository {
	return &systemSettingsRepository{Db: Db}
}

// Get retrieves a single system setting by name.
// Returns empty string and no error if the setting does not exist.
func (m *systemSettingsRepository) Get(name string) (string, error) {
	var value string
	err := m.Db.QueryRow("SELECT value FROM system_settings WHERE name=?", name).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return value, nil
}

// GetByPrefix retrieves all system settings whose name starts with the given prefix.
// Example: GetByPrefix("password.") returns all password policy settings.
func (m *systemSettingsRepository) GetByPrefix(prefix string) (map[string]string, error) {
	rows, err := m.Db.Query("SELECT name, value FROM system_settings WHERE name LIKE ?", prefix+"%")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}

		settings[name] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateBatch inserts or updates multiple system settings in a single transaction.
// Example: UpdateBatch(map[string]string{"password.min_length": "8", "password.require_uppercase": "true"})
func (m *systemSettingsRepository) UpdateBatch(settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO system_settings (name, value)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value)`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for name, value := range settings {
		if _, err := stmt.Exec(name, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}
