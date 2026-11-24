// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"maps"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

func TestEngineDatabaseTypes(t *testing.T) {
	connections := []model.Connection{
		{ID: "sales", Type: "mysql"},
		{ID: "warehouse", Type: "postgres"},
		{ID: "other", Type: "sqlite"},
	}

	got := EngineDatabaseTypes(connections)

	want := map[string]queryengine.DatabaseType{
		"sales":     queryengine.DatabaseMySQL,
		"warehouse": queryengine.DatabasePostgres,
	}

	if !maps.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestEngineConnectionSettings(t *testing.T) {
	connection := model.Connection{
		Type:     "postgres",
		Host:     "db.example.com",
		Port:     "5432",
		Database: "shop",
		Username: "reader",
		Password: "encrypted, not used",
		SSLMode:  model.SSLModeVerifyFull,
		Config:   model.ConnectionConfig{model.ConnectionCACertKey: "PEM"},
	}

	got := EngineConnectionSettings(connection, "secret")

	want := queryengine.ConnectionSettings{
		DatabaseType:  queryengine.DatabasePostgres,
		Host:          "db.example.com",
		Port:          "5432",
		Database:      "shop",
		User:          "reader",
		Password:      "secret",
		SSLMode:       queryengine.SSLModeVerifyFull,
		CACertificate: "PEM",
	}

	if got != want {
		t.Errorf("want %+v, got %+v", want, got)
	}
}

func TestEngineConnectionSettingsPassesAnUnknownTypeOn(t *testing.T) {
	got := EngineConnectionSettings(model.Connection{Type: "sqlite"}, "")

	if got.DatabaseType != "sqlite" {
		t.Errorf("want the engine to see sqlite and refuse it, got %q", got.DatabaseType)
	}
}

func TestBuilderTableName(t *testing.T) {
	cases := []struct {
		connectionType string
		table          queryengine.Table
		want           string
	}{
		{"mysql", queryengine.Table{Schema: "shop", Name: "orders"}, "orders"},
		{"postgres", queryengine.Table{Schema: "public", Name: "orders"}, "orders"},
		{"postgres", queryengine.Table{Schema: "archive", Name: "orders"}, "archive.orders"},
	}

	for _, testCase := range cases {
		got := BuilderTableName(testCase.connectionType, testCase.table)

		if got != testCase.want {
			t.Errorf("%s %s.%s: want %q, got %q", testCase.connectionType, testCase.table.Schema, testCase.table.Name, testCase.want, got)
		}
	}
}

func TestSuggestBuilderType(t *testing.T) {
	cases := []struct {
		connectionType string
		columnType     string
		want           string
	}{
		{"mysql", "tinyint(1)", "boolean"},
		{"mysql", "int(11)", "number"},
		{"mysql", "varchar(100)", "string"},
		{"mysql", "datetime", "time"},
		{"mysql", "json", ""},
		{"postgres", "character varying", "string"},
		{"postgres", "timestamp with time zone", "time"},
		{"postgres", "boolean", "boolean"},
		{"sqlite", "int", ""},
	}

	for _, testCase := range cases {
		got := SuggestBuilderType(testCase.connectionType, testCase.columnType)

		if got != testCase.want {
			t.Errorf("%s %q: want %q, got %q", testCase.connectionType, testCase.columnType, testCase.want, got)
		}
	}
}

func TestDatasetNameForTable(t *testing.T) {
	cases := map[string]string{
		"orders":          "orders",
		"archive.regions": "archive_regions",
	}

	for table, want := range cases {
		got := DatasetNameForTable(table)
		if got != want {
			t.Errorf("%s: want %q, got %q", table, want, got)
		}
	}
}
