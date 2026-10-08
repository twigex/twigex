// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store/sqlstore"
	"github.com/twigex/twigex/tlog"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",
}

var createUserCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create user",
	Example: `user --email user@example.com --username userexample --password Password1 --firstname Name --lastname Last --role [user/admin]`,
	Run:     createUser,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().String("username", "", "Required. Username for the new user account.")
	createUserCmd.Flags().String("email", "", "Required. The email address for the new user account.")
	createUserCmd.Flags().String("password", "", "Required. The password for the new user account.")
	createUserCmd.Flags().String("firstname", "", "Required. The first name for the new user account.")
	createUserCmd.Flags().String("lastname", "", "Required. The last name for the new user account.")
	createUserCmd.Flags().String("role", "", "Required. Role for user [system_user/system_admin]")
	cobra.CheckErr(createUserCmd.MarkFlagRequired("username"))
	cobra.CheckErr(createUserCmd.MarkFlagRequired("email"))
	cobra.CheckErr(createUserCmd.MarkFlagRequired("password"))
	cobra.CheckErr(createUserCmd.MarkFlagRequired("firstname"))
	cobra.CheckErr(createUserCmd.MarkFlagRequired("lastname"))
	cobra.CheckErr(createUserCmd.MarkFlagRequired("role"))
}

func createUser(cmd *cobra.Command, args []string) {
	app, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	updateDatabaseIfNeeded()

	so, err := app.LoadStorages()
	if err != nil {
		tlog.Errorw("Failed to load storages", "error", err)
		os.Exit(1)
	}

	app.FileStorageObjects = so

	username, err := cmd.Flags().GetString("username")
	if err != nil || username == "" {
		tlog.Errorw("Failed to get username flag", "error", err)
		os.Exit(1)
	}

	username = model.NormalizeUsername(username)
	if !model.ValidUsername(username) {
		tlog.Errorw("Invalid username", "username", username)
		os.Exit(1)
	}

	if model.ReservedUsername(username) {
		tlog.Errorw("Reserved username", "username", username)
		os.Exit(1)
	}

	email, err := cmd.Flags().GetString("email")
	if err != nil || email == "" {
		tlog.Errorw("Failed to get email flag", "error", err)
		os.Exit(1)
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil || password == "" {
		tlog.Errorw("Failed to get password flag", "error", err)
		os.Exit(1)
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		tlog.Errorw("Failed to hash password", "error", err)
		os.Exit(1)
	}

	name, err := cmd.Flags().GetString("firstname")
	if err != nil || name == "" {
		tlog.Errorw("Failed to get firstname flag", "error", err)
		os.Exit(1)
	}

	lastName, err := cmd.Flags().GetString("lastname")
	if err != nil || lastName == "" {
		tlog.Errorw("Failed to get lastname flag", "error", err)
		os.Exit(1)
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil || role == "" {
		tlog.Errorw("Failed to get role flag", "error", err)
		os.Exit(1)
	}

	_, err = insertUser(&app.ConfigStore, model.NewUser{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Name:     name,
		LastName: lastName,
		Role:     role,
	})
	if err != nil {
		tlog.Errorw("Failed to create user", "username", username, "error", err)
		os.Exit(1)
	}
}

func insertUser(c *config.ConfigStore, newUser model.NewUser) (*model.User, error) {
	dbType := c.Config.SqlSettings.DriverName

	if *dbType == "mysql" {
		db, err := sqlstore.NewClient(c.Config.SqlSettings)
		if err != nil {
			return nil, err
		}

		defer db.Close()

		userRepo, err := sqlstore.NewUserRepository(db)
		if err != nil {
			return nil, err
		}

		u, err := userRepo.Create(model.NewUser{
			Username:      newUser.Username,
			Email:         newUser.Email,
			Password:      newUser.Password,
			Name:          newUser.Name,
			LastName:      newUser.LastName,
			Role:          newUser.Role,
			ClockDisplay:  "24h",
			StorageLimit:  0, // unlimited
			DeactivatedAt: 0,
		})
		if err != nil {
			return nil, err
		}

		return u, nil
	}

	return nil, nil
}
