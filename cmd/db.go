// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/tlog"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management",
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply all pending SQL migrations",
	Run:   runDBMigrate,
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbMigrateCmd)
}

func runDBMigrate(cmd *cobra.Command, args []string) {
	c, err := openConfig()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	if err := migrateUp(c); err != nil {
		os.Exit(1)
	}

	tlog.Infow("SQL migrations applied; start the server to finish the upgrade and record the version")
}

func updateDatabaseIfNeeded() {
	config, err := config.NewConfig()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	if err := migrateUp(config); err != nil {
		os.Exit(1)
	}
}
