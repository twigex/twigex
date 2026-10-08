// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/db"
	"github.com/twigex/twigex/store"
	"github.com/twigex/twigex/tlog"
)

func migrateUp(c *config.ConfigStore) error {
	dbType := c.Config.SqlSettings.DriverName
	if *dbType != "mysql" {
		return nil
	}

	conn, err := store.SetupConnection(c.Config.SqlSettings, true)
	if err != nil {
		tlog.Errorw("Failed to set up database connection", "error", err)
		return err
	}

	defer conn.Close()

	if err := checkUpgradePath(context.Background(), conn); err != nil {
		return err
	}

	driver, err := mysql.WithInstance(conn, &mysql.Config{})
	if err != nil {
		tlog.Errorw("Failed to set up migration driver", "error", err)
		return err
	}

	source, err := iofs.New(db.FS, "mysql")
	if err != nil {
		tlog.Errorw("Failed to read embedded migrations", "error", err)
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		tlog.Errorw("Failed to set up migrations", "error", err)
		return err
	}

	defer m.Close()

	err = m.Up()
	switch err {
	case nil:
		tlog.Infow("Database migrations applied successfully")
		return nil
	case migrate.ErrNoChange:
		tlog.Infow("Database is up to date")
		return nil
	default:
		tlog.Errorw("Failed to run migrations", "error", err)
		return err
	}
}
