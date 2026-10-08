// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Storage management",
}

var createStorageCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create storage",
	Example: `s3 --flags or local --flags`,
}

var deleteStorageCobraCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete storage",
	Example: `delete --id=""`,
	Run:     deleteStorageCmd,
}

var listStorageCobraCmd = &cobra.Command{
	Use:     "list",
	Short:   "List created storages",
	Example: `list`,
	Run:     listStorageCmd,
}

var createLocalCmd = &cobra.Command{
	Use:     "local",
	Short:   "Create local storage",
	Example: `local --name "My Storage" --description "Local files"`,
	Run:     createLocalStorageCmd,
}

var createS3Cmd = &cobra.Command{
	Use:     "s3",
	Short:   "Create S3 storage",
	Example: `s3 --name "S3" --endpoint "s3.amazonaws.com" --accesskey "key" --secretkey "secret" --bucket "mybucket" --ssl true`,
	Run:     createS3StorageCmd,
}

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(createStorageCmd)
	storageCmd.AddCommand(deleteStorageCobraCmd)
	storageCmd.AddCommand(listStorageCobraCmd)
	createStorageCmd.AddCommand(createLocalCmd)
	createStorageCmd.AddCommand(createS3Cmd)

	deleteStorageCobraCmd.Flags().String("id", "", "Storage ID to delete")
	cobra.CheckErr(deleteStorageCobraCmd.MarkFlagRequired("id"))

	createLocalCmd.Flags().String("name", "", "Required. Local storage name")
	cobra.CheckErr(createLocalCmd.MarkFlagRequired("name"))
	createLocalCmd.Flags().String("description", "", "Local storage description")

	createS3Cmd.Flags().String("name", "", "Required. S3 storage name")
	cobra.CheckErr(createS3Cmd.MarkFlagRequired("name"))
	createS3Cmd.Flags().String("description", "", "S3 storage description")
	createS3Cmd.Flags().String("endpoint", "", "Required. S3 endpoint URL")
	cobra.CheckErr(createS3Cmd.MarkFlagRequired("endpoint"))
	createS3Cmd.Flags().String("accesskey", "", "Required. S3 access key")
	cobra.CheckErr(createS3Cmd.MarkFlagRequired("accesskey"))
	createS3Cmd.Flags().String("secretkey", "", "Required. S3 secret key")
	cobra.CheckErr(createS3Cmd.MarkFlagRequired("secretkey"))
	createS3Cmd.Flags().String("bucket", "", "Required. S3 bucket name")
	cobra.CheckErr(createS3Cmd.MarkFlagRequired("bucket"))
	createS3Cmd.Flags().Bool("ssl", true, "Enable SSL (default: true)")
}

func createLocalStorageCmd(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	desc, _ := cmd.Flags().GetString("description")

	a, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize app", "error", err)
		os.Exit(1)
	}

	_, appErr := a.CreateStorage(cliAdminUser(), model.Storage{
		Label:       name,
		Description: desc,
		Type:        model.LocalStorage,
	})
	if appErr != nil {
		tlog.Errorw("Failed to create local storage", "error", appErr.Message)
		os.Exit(1)
	}
}

func createS3StorageCmd(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	desc, _ := cmd.Flags().GetString("description")
	endpoint, _ := cmd.Flags().GetString("endpoint")
	accesskey, _ := cmd.Flags().GetString("accesskey")
	secret, _ := cmd.Flags().GetString("secretkey")
	bucket, _ := cmd.Flags().GetString("bucket")
	ssl, _ := cmd.Flags().GetBool("ssl")

	a, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize app", "error", err)
		os.Exit(1)
	}

	_, appErr := a.CreateStorage(cliAdminUser(), model.Storage{
		Label:       name,
		Description: desc,
		Type:        model.S3Storage,
		Endpoint:    endpoint,
		AccessKey:   accesskey,
		SecretKey:   secret,
		Bucket:      bucket,
		SSL:         ssl,
	})
	if appErr != nil {
		tlog.Errorw("Failed to create S3 storage", "error", appErr.Message)
		os.Exit(1)
	}
}

func deleteStorageCmd(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")

	a, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize app", "error", err)
		os.Exit(1)
	}

	if appErr := a.DeleteStorage(context.Background(), cliAdminUser(), id); appErr != nil {
		tlog.Errorw("Failed to delete storage", "storage_id", id, "error", appErr.Message)
		os.Exit(1)
	}

	tlog.Infow("Storage deleted", "storage_id", id)
}

func listStorageCmd(cmd *cobra.Command, args []string) {
	a, err := openApp()
	if err != nil {
		tlog.Errorw("Failed to initialize app", "error", err)
		os.Exit(1)
	}

	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		tlog.Errorw("Failed to retrieve storages", "error", err)
		os.Exit(1)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintln(writer, "id\tlabel\ttype\tdirectory\tendpoint\tprimary")
	for _, s := range storages {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%v\n",
			s.ID, s.Label, s.DriverName(), s.Directory, s.Endpoint, s.Primary,
		)
	}

	writer.Flush()
}

func createLocalStorageFromEnv(a *app.App) error {
	tlog.Infow("Creating local storage from environment variables")

	storageName := os.Getenv("TWIGEX_STORAGE_NAME")
	if storageName == "" {
		storageName = "Local storage"
	}

	_, appErr := a.CreateStorage(cliAdminUser(), model.Storage{
		Label:       storageName,
		Description: "Local storage",
		Type:        model.LocalStorage,
		Primary:     true,
	})
	if appErr != nil {
		return errors.New(appErr.Message)
	}

	return nil
}

func createS3StorageFromEnv(a *app.App) error {
	tlog.Infow("Creating S3 storage from environment variables")

	storageName := os.Getenv("TWIGEX_STORAGE_NAME")
	if storageName == "" {
		storageName = "S3 storage"
	}

	storageEndpoint := os.Getenv("TWIGEX_STORAGE_ENDPOINT")
	if storageEndpoint == "" {
		return errors.New("TWIGEX_STORAGE_ENDPOINT is required for S3 storage")
	}

	storageAccess := os.Getenv("TWIGEX_STORAGE_ACCESS_KEY")
	if storageAccess == "" {
		return errors.New("TWIGEX_STORAGE_ACCESS_KEY is required for S3 storage")
	}

	storageSecret := os.Getenv("TWIGEX_STORAGE_SECRET_KEY")
	if storageSecret == "" {
		return errors.New("TWIGEX_STORAGE_SECRET_KEY is required for S3 storage")
	}

	storageBucket := os.Getenv("TWIGEX_STORAGE_BUCKET")
	if storageBucket == "" {
		return errors.New("TWIGEX_STORAGE_BUCKET is required for S3 storage")
	}

	storageSSL := os.Getenv("TWIGEX_STORAGE_SSL") != "false"

	_, appErr := a.CreateStorage(cliAdminUser(), model.Storage{
		Label:       storageName,
		Description: "S3 storage",
		Type:        model.S3Storage,
		Endpoint:    storageEndpoint,
		AccessKey:   storageAccess,
		SecretKey:   storageSecret,
		Bucket:      storageBucket,
		SSL:         storageSSL,
		Primary:     true,
	})
	if appErr != nil {
		return errors.New(appErr.Message)
	}

	return nil
}

// cliAdminUser returns a synthetic admin user for CLI commands that bypass session auth.
func cliAdminUser() model.User {
	return model.User{Role: model.SystemAdminRoleId}
}
