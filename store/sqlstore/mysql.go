// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

func NewClient(settings model.SqlSettings) (*sql.DB, error) {
	db, err := sql.Open(*settings.DriverName, *settings.DataSource)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(*settings.MaxOpenConns)
	db.SetMaxIdleConns(*settings.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(*settings.ConnMaxLifetimeMinutes) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(*settings.ConnMaxIdleTimeMinutes) * time.Minute)

	return db, nil
}
