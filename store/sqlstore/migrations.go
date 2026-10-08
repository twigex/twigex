// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/twigex/twigex/model"
)

type migrationRepository struct {
	Db *sql.DB
}

const appMigrationColumns = `id, name, migrated_at`

func NewMigrationRepository(Db *sql.DB) (*migrationRepository, error) {
	repo := &migrationRepository{
		Db: Db,
	}

	return repo, nil
}

func (m *migrationRepository) Create(name string) error {
	id := model.NewID()
	t := time.Now().Unix()

	_, err := m.Db.Exec(`INSERT INTO app_migrations (`+appMigrationColumns+`) VALUES(?,?,?)`, id, name, t)
	if err != nil {
		return err
	}

	return nil
}

// SchemaVersion returns the migration last applied and whether it left the
// schema half applied.
func (m *migrationRepository) SchemaVersion(ctx context.Context) (int64, bool, error) {
	var version int64
	var dirty bool

	err := m.Db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations`).Scan(&version, &dirty)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

func (m *migrationRepository) TableExists(ctx context.Context, name string) (bool, error) {
	var count int
	err := m.Db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = ?`, name).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (m *migrationRepository) IsMigrated(migrationName string) (bool, error) {
	var name string
	err := m.Db.QueryRow(`SELECT name FROM app_migrations WHERE name = ?`, migrationName).Scan(&name)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if err == sql.ErrNoRows {
		return false, nil
	}

	return true, nil
}
