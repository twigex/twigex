// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type licenseRepository struct {
	Db *sql.DB
}

const licenseColumns = `id, created, active, bytes`

func NewLicenseRepository(Db *sql.DB) (*licenseRepository, error) {
	repo := &licenseRepository{}

	repo.Db = Db
	return repo, nil
}

func (l *licenseRepository) Create(licenseBytes string) error {
	time := time.Now().Unix()
	id := model.NewID()

	_, err := l.Db.Exec("INSERT INTO licenses ("+licenseColumns+") VALUES(?,?,?,?)", id, time, true, licenseBytes)
	if err != nil {
		return err
	}

	return nil
}

func (l *licenseRepository) DeleteActive() error {
	_, err := l.Db.Exec("DELETE FROM licenses WHERE active=?", true)
	if err != nil {
		return err
	}

	return nil
}

func (l *licenseRepository) GetAll() ([]model.ActiveLicense, error) {
	rows, err := l.Db.Query(
		"SELECT id, created, active, bytes FROM licenses WHERE active=? ORDER BY created",
		true,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	licenses := make([]model.ActiveLicense, 0)

	for rows.Next() {
		license := model.ActiveLicense{}

		if err := rows.Scan(
			&license.ID,
			&license.Created,
			&license.Active,
			&license.Bytes,
		); err != nil {
			return nil, err
		}

		licenses = append(licenses, license)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return licenses, nil
}
