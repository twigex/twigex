// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Test helpers for store-layer integration tests.
//
// Connection: tests read TEST_MYSQL_DSN. If unset, TestMain marks the suite as
// "no DB" and every test that calls requireDB(t) will t.Skip. If set but the
// connection fails, TestMain exits non-zero so CI catches that rather than
// silently skipping.
//
// Isolation: one fresh schema per test binary run (named twigex_test_<nanos>),
// migrations applied once. Between tests, cleanTables(t) wipes the rows in the
// tables the test touches. This is the Mattermost-style pattern: fast, with
// per-test cleanup discipline.
//
// Schema is dropped on teardown. Crashed runs leave behind twigex_test_*
// schemas that the operator can drop manually.

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migrateMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/twigex/twigex/db"
)

var (
	testDB       *sql.DB
	testDBName   string
	testDBSkipOK = "TEST_MYSQL_DSN not set; skipping DB-backed tests"
)

func TestMain(m *testing.M) {
	rootDSN := os.Getenv("TEST_MYSQL_DSN")
	if rootDSN == "" {
		// Run anyway so non-DB tests in this package still execute. DB-backed
		// tests will call requireDB(t) and skip themselves with a clear message.
		os.Exit(m.Run())
	}

	if err := setupTestDB(rootDSN); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func setupTestDB(rootDSN string) error {
	cfg, err := mysql.ParseDSN(rootDSN)
	if err != nil {
		return fmt.Errorf("parse DSN: %w", err)
	}

	rootCfg := *cfg
	rootCfg.DBName = ""
	rootCfg.MultiStatements = true
	rootDB, err := sql.Open("mysql", rootCfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open root: %w", err)
	}
	defer rootDB.Close()

	schema := fmt.Sprintf("twigex_test_%d", time.Now().UnixNano())
	if _, err := rootDB.Exec("CREATE DATABASE `" + schema + "`"); err != nil {
		return fmt.Errorf("create schema %s: %w", schema, err)
	}
	testDBName = schema

	cfg.DBName = schema
	cfg.MultiStatements = true
	conn, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open schema: %w", err)
	}
	defer conn.Close()

	driver, err := migrateMysql.WithInstance(conn, &migrateMysql.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	source, err := iofs.New(db.FS, "mysql")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}
	mig, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	// Production runs only migrations with multiStatements, so the tests do too.
	cfg.MultiStatements = false
	testDB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("open test connection: %w", err)
	}
	return nil
}

func teardownTestDB() {
	if testDB != nil {
		testDB.Close()
		testDB = nil
	}
	if testDBName == "" {
		return
	}
	rootDSN := os.Getenv("TEST_MYSQL_DSN")
	cfg, err := mysql.ParseDSN(rootDSN)
	if err != nil {
		return
	}
	cfg.DBName = ""
	root, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return
	}
	defer root.Close()
	_, _ = root.Exec("DROP DATABASE IF EXISTS `" + testDBName + "`")
}

// requireDB returns the shared test DB. Tests that need the database call
// this; if TEST_MYSQL_DSN is unset, the test skips cleanly with a clear
// message instead of failing on a nil pointer.
func requireDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDB == nil {
		t.Skip(testDBSkipOK)
	}
	return testDB
}

// cleanTables empties the named tables. Used between tests in the same run
// since the schema is shared across the package's test binary.
//
// FK checks are toggled off to allow truncating in any order, even though
// the current schema doesn't declare FOREIGN KEY constraints; belt and
// suspenders if/when they're added.
func cleanTables(t *testing.T, tables ...string) {
	t.Helper()
	if testDB == nil {
		return
	}
	if _, err := testDB.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK checks: %v", err)
	}
	for _, table := range tables {
		if _, err := testDB.Exec("DELETE FROM `" + table + "`"); err != nil {
			t.Fatalf("clean %s: %v", table, err)
		}
	}
	if _, err := testDB.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("enable FK checks: %v", err)
	}
}
