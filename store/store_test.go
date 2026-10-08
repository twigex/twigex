// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package store

import (
	"os"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/twigex/twigex/model"
)

func TestSetupConnectionAllowsMultipleStatementsOnlyWhereAsked(t *testing.T) {
	rootDSN := os.Getenv("TEST_MYSQL_DSN")
	if rootDSN == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping DB-backed tests")
	}
	cfg, err := mysql.ParseDSN(rootDSN)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	cfg.MultiStatements = false
	cfg.ParseTime = true
	dsn := cfg.FormatDSN()

	var settings model.SqlSettings
	settings.SetDefaults()
	settings.DataSource = model.NewString(dsn)

	migrations, err := SetupConnection(settings, true)
	if err != nil {
		t.Fatalf("migration connection: %v", err)
	}
	defer migrations.Close()
	if *settings.DataSource != dsn {
		t.Errorf("the shared data source became %q", *settings.DataSource)
	}
	if _, err := migrations.Exec("DO 1; DO 2"); err != nil {
		t.Errorf("migration connection rejected two statements: %v", err)
	}

	app, err := SetupConnection(settings, false)
	if err != nil {
		t.Fatalf("app connection: %v", err)
	}
	defer app.Close()
	if _, err := app.Exec("DO 1; DO 2"); err == nil {
		t.Error("app connection ran two statements in one query")
	}
}
