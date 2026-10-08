// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/tlog"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Settings management",
}

var emailCmd = &cobra.Command{
	Use:     "email",
	Short:   "Configure email server settings",
	Example: `email --username user@example.com --password password --server smtp.example.com --port 465`,
	Run:     addEmail,
}

func init() {
	rootCmd.AddCommand(settingsCmd)
	settingsCmd.AddCommand(emailCmd)

	emailCmd.Flags().String("username", "", "SMTP username or email")
	cobra.CheckErr(emailCmd.MarkFlagRequired("username"))

	emailCmd.Flags().String("password", "", "SMTP password")
	cobra.CheckErr(emailCmd.MarkFlagRequired("password"))

	emailCmd.Flags().String("server", "", "SMTP server address")
	cobra.CheckErr(emailCmd.MarkFlagRequired("server"))

	emailCmd.Flags().String("port", "", "SMTP server port")
	cobra.CheckErr(emailCmd.MarkFlagRequired("port"))

	emailCmd.Flags().String("security", "TLS", "Connection security: TLS, STARTTLS or PLAIN")
	emailCmd.Flags().Bool("auth", true, "Enable SMTP authentication")

	emailCmd.Flags().String("from-address", "", "From address for outgoing mail (defaults to username)")
	emailCmd.Flags().String("from-name", "Twigex", "From name for outgoing mail")
	emailCmd.Flags().String("reply-to", "", "Reply-To address for outgoing mail")
}

func addEmail(cmd *cobra.Command, args []string) {
	a, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize app", "error", err)
		os.Exit(1)
	}

	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	server, _ := cmd.Flags().GetString("server")
	port, _ := cmd.Flags().GetString("port")
	security, _ := cmd.Flags().GetString("security")
	auth, _ := cmd.Flags().GetBool("auth")
	fromAddress, _ := cmd.Flags().GetString("from-address")
	fromName, _ := cmd.Flags().GetString("from-name")
	replyTo, _ := cmd.Flags().GetString("reply-to")

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	encryptedPassword, err := crypto.Encrypt(encryptKey, password)
	if err != nil {
		tlog.Errorw("Failed to encrypt SMTP password", "error", err)
		os.Exit(1)
	}

	err = a.Store.SystemSettings.UpdateBatch(map[string]string{
		"email.smtp_host":         server,
		"email.smtp_port":         port,
		"email.smtp_username":     username,
		"email.smtp_password":     encryptedPassword,
		"email.smtp_security":     security,
		"email.smtp_auth":         strconv.FormatBool(auth),
		"email.smtp_from_address": fromAddress,
		"email.smtp_from_name":    fromName,
		"email.smtp_reply_to":     replyTo,
	})
	if err != nil {
		tlog.Errorw("Failed to save email settings",
			"server", server,
			"error", err,
		)
		os.Exit(1)
	}

	tlog.Infow("Email settings saved successfully",
		"server", server,
		"port", port,
		"username", username,
	)
}
