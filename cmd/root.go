// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/joho/godotenv"
)

var rootCmd = &cobra.Command{
	Use:   "twigex",
	Short: "Private cloud",
	Long:  `Private cloud for your business or home`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load .env file if it exists
		_ = godotenv.Load()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(cmd.Help())
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
