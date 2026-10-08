// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/twigex/twigex/tlog"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Set up application",
	Run:   installApp,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installApp(cmd *cobra.Command, args []string) {
	setupCloud(cmd, args)
}

func setupCloud(cmd *cobra.Command, args []string) {
	c, err := openConfig()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	if err = migrateUp(c); err != nil {
		os.Exit(1)
	}
}
