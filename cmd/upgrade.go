// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/twigex/twigex/internal/upgrade"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store/sqlstore"
	"github.com/twigex/twigex/tlog"
)

func databaseState(ctx context.Context, db *sql.DB) (recorded string, schema upgrade.Schema, err error) {
	migrations, err := sqlstore.NewMigrationRepository(db)
	if err != nil {
		return "", schema, err
	}

	schema.Installed, err = migrations.TableExists(ctx, "schema_migrations")
	if err != nil || !schema.Installed {
		return "", schema, err
	}

	schema.Version, schema.Dirty, err = migrations.SchemaVersion(ctx)
	if err != nil {
		return "", schema, err
	}

	settings, err := migrations.TableExists(ctx, "system_settings")
	if err != nil || !settings {
		return "", schema, err
	}

	recorded, err = sqlstore.NewSystemSettingsRepository(db).Get(upgrade.VersionKey)

	return recorded, schema, err
}

func checkUpgradePath(ctx context.Context, db *sql.DB) error {
	recorded, schema, err := databaseState(ctx, db)
	if err != nil {
		tlog.Errorw("Failed to read the recorded app version", "error", err)
		return err
	}

	err = upgrade.Check(recorded, schema)
	if err == nil {
		return nil
	}

	if os.Getenv(upgrade.SkipEnv) != "" {
		tlog.Warnw("Upgrade version check overridden",
			"recorded", recorded,
			"binary", model.Version,
			"variable", upgrade.SkipEnv,
		)
		return nil
	}

	fmt.Fprintln(os.Stderr, err)

	return err
}

func printVersions(ctx context.Context, db *sql.DB) error {
	recorded, schema, err := databaseState(ctx, db)
	if err != nil {
		return err
	}

	fmt.Printf("binary:    %s\n", model.Version)
	switch {
	case !schema.Installed:
		fmt.Println("database:  empty, not installed yet")
	case recorded == "":
		fmt.Printf("database:  not recorded, assumed %s\n", model.LastUngatedVersion)
	default:
		fmt.Printf("database:  %s\n", recorded)
	}

	if schema.Installed {
		fmt.Printf("schema:    %d", schema.Version)
		if schema.Dirty {
			fmt.Print(" (dirty)")
		}

		fmt.Println()
	}

	fmt.Printf("floor:     %s\n", model.MinUpgradableVersion)

	return nil
}
