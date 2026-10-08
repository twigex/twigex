// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twigex/twigex/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generates configuration for the Twigex server",
	Run:   generateConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func generateConfig(cmd *cobra.Command, args []string) {
	_, err := config.NewConfig()
	cobra.CheckErr(err)
}
