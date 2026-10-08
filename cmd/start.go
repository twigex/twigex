// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"errors"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/api"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the Twigex server",
	Run:   startServer,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().String("port", "", "Port number")
	startCmd.Flags().String("cert", "", "Path to certificate")
	startCmd.Flags().String("key", "", "Required if cert flag is set. Path to key")
	startCmd.Flags().Bool("dev", false, "Run server in development mode")
}

func startServer(cmd *cobra.Command, args []string) {
	isDev, err := cmd.Flags().GetBool("dev")
	if err != nil {
		tlog.Errorw("Failed to get dev flag", "error", err)
		os.Exit(1)
	}

	if isDev {
		config.IsDev = true
	}

	updateDatabaseIfNeeded()

	a := app.NewApp()

	a.EnsureSystemRoles()

	a.DoAppMigrations()

	a.RecordVersion()

	if err = syncStorage(a); err != nil {
		tlog.Errorw("Failed to sync storage", "error", err)
		os.Exit(1)
	}

	if err = createUserFromEnv(a); err != nil {
		tlog.Errorw("Failed to create user from env", "error", err)
		os.Exit(1)
	}

	port, err := cmd.Flags().GetString("port")
	if err != nil {
		tlog.Errorw("Failed to get port flag", "error", err)
		os.Exit(1)
	}

	if port == "" {
		port = *a.ConfigStore.Config.ServerSettings.Port
	}

	port = ":" + port

	cert, err := cmd.Flags().GetString("cert")
	if err != nil {
		tlog.Errorw("Failed to get cert flag", "error", err)
		os.Exit(1)
	}

	key, err := cmd.Flags().GetString("key")
	if err != nil {
		tlog.Errorw("Failed to get key flag", "error", err)
		os.Exit(1)
	}

	api.InitAPI(a)

	if err = a.InitJobs(); err != nil {
		tlog.Errorw("Failed to initialise jobs", "error", err)
	}

	if err = a.LoadLicense(); err != nil {
		a.Server.License = nil
		tlog.Errorw("Failed to load license", "error", err)
	}

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	go func() {
		<-channel
		tlog.Info("request for server to shutdown")
		if err := a.Server.Shutdown(); err != nil {
			tlog.Errorw("Server shutdown returned an error", "error", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	tlog.Infow("Server listening on port",
		"port", port,
		"version", model.Version,
		"edition", model.Edition,
		"enterprise_revision", model.EnterpriseRevision,
		"enterprise_hash", model.EnterpriseHash,
	)
	if cert != "" && key != "" {
		if err = a.StartTLS(port, cert, key); err != nil {
			tlog.Errorw("Failed to start TLS server", "error", err)
			os.Exit(1)
		}
	} else {
		if err = a.Start(port); err != nil {
			tlog.Errorw("Failed to start server", "error", err)
			os.Exit(1)
		}
	}
}

func syncStorage(a *app.App) error {
	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		return err
	}

	// Storages already exist in DB, nothing to do
	if len(storages) > 0 {
		so, err := a.LoadStorages()
		if err != nil {
			return err
		}

		a.FileStorageObjects = so
		return nil
	}

	// No storages, create from env vars
	tlog.Infow("No storage configured — creating from environment variables")

	storageType := os.Getenv("TWIGEX_STORAGE_TYPE")
	switch strings.ToLower(storageType) {
	case "s3":
		if err := createS3StorageFromEnv(a); err != nil {
			return err
		}
	default:
		if err := createLocalStorageFromEnv(a); err != nil {
			return err
		}
	}

	so, err := a.LoadStorages()
	if err != nil {
		return err
	}

	a.FileStorageObjects = so
	return nil
}

func createUserFromEnv(app *app.App) error {
	users, err := app.Store.User.GetAll()
	if err != nil {
		return err
	}

	if len(users) > 0 {
		return nil
	}

	name := os.Getenv("TWIGEX_ADMIN_NAME")
	lastname := os.Getenv("TWIGEX_ADMIN_LASTNAME")
	username := os.Getenv("TWIGEX_ADMIN_USERNAME")
	email := os.Getenv("TWIGEX_ADMIN_EMAIL")
	password := os.Getenv("TWIGEX_ADMIN_PASSWORD")

	if email == "" && password == "" {
		tlog.Infow("No admin credentials set — skipping admin creation",
			"hint", "set TWIGEX_ADMIN_EMAIL and TWIGEX_ADMIN_PASSWORD to auto-create admin",
		)
		return nil
	}

	if name == "" || lastname == "" || username == "" || email == "" || password == "" {
		return errors.New(
			"incomplete admin credentials — set all of: " +
				"TWIGEX_ADMIN_NAME, TWIGEX_ADMIN_LASTNAME, " +
				"TWIGEX_ADMIN_USERNAME, TWIGEX_ADMIN_EMAIL, " +
				"TWIGEX_ADMIN_PASSWORD",
		)
	}

	username = model.NormalizeUsername(username)
	if !model.ValidUsername(username) {
		return errors.New("TWIGEX_ADMIN_USERNAME may contain only lowercase letters, " +
			"numbers, dots, hyphens and underscores, and must start with a letter or number")
	}

	if model.ReservedUsername(username) {
		return errors.New("TWIGEX_ADMIN_USERNAME is reserved")
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = insertUser(&app.ConfigStore, model.NewUser{
		Name:     name,
		LastName: lastname,
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Role:     model.SystemAdminRoleId,
	})
	if err != nil {
		tlog.Errorw("Failed to create admin user from environment variables", "error", err)
		return err
	}

	tlog.Infow("Admin user created", "username", username, "email", email)
	return nil
}
