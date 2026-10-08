// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"github.com/twigex/twigex/tlog"
)

type appMigration struct {
	name string
	run  func(a *App) error
}

// appMigrations fix data that SQL alone cannot. Each runs once, in order, and
// its name is recorded in app_migrations, so a name must never change.
var appMigrations = []appMigration{}

func (a *App) DoAppMigrations() {
	for _, m := range appMigrations {
		a.runMigration(m.name, func() error {
			return m.run(a)
		})
	}
}

func (a *App) runMigration(name string, fn func() error) {
	migrated, err := a.Store.Migrations.IsMigrated(name)
	if err != nil {
		tlog.Errorw("Failed to check migration status",
			"migration", name,
			"error", err,
		)
		return
	}

	if migrated {
		return
	}

	if err = fn(); err != nil {
		tlog.Errorw("Failed to run migration",
			"migration", name,
			"error", err,
		)
		return
	}

	if err = a.Store.Migrations.Create(name); err != nil {
		tlog.Errorw("Failed to record migration",
			"migration", name,
			"error", err,
		)
	}
}
