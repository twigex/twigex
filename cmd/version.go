// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
	"github.com/twigex/twigex/tlog"
)

// startCmd represents the start command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns server version",
	Run:   version,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().Bool("db", false, "Also print the version recorded in the database")
}

func version(cmd *cobra.Command, args []string) {
	withDB, _ := cmd.Flags().GetBool("db")
	if !withDB {
		fmt.Println(model.Version)
		return
	}

	c, err := openConfig()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	db, err := store.SetupConnection(c.Config.SqlSettings, false)
	if err != nil {
		tlog.Errorw("Failed to set up database connection", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	if err := printVersions(context.Background(), db); err != nil {
		tlog.Errorw("Failed to read the recorded app version", "error", err)
		os.Exit(1)
	}
}
