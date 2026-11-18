// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"slices"
	"testing"
)

func TestSuggestFieldTypeForMySQL(t *testing.T) {
	cases := map[string]FieldType{
		"tinyint(1)":            FieldTypeYesNo,
		"tinyint(1) unsigned":   FieldTypeYesNo,
		"tinyint(4)":            FieldTypeNumber,
		"tinyint":               FieldTypeNumber,
		"int(11)":               FieldTypeNumber,
		"int unsigned":          FieldTypeNumber,
		"bigint(20)":            FieldTypeNumber,
		"decimal(12,2)":         FieldTypeNumber,
		"double":                FieldTypeNumber,
		"year":                  FieldTypeNumber,
		"varchar(100)":          FieldTypeText,
		"longtext":              FieldTypeText,
		"enum('a','b')":         FieldTypeText,
		"time":                  FieldTypeText,
		"date":                  FieldTypeTime,
		"datetime":              FieldTypeTime,
		"timestamp":             FieldTypeTime,
		"DATETIME":              FieldTypeTime,
		"json":                  "",
		"blob":                  "",
		"varbinary(16)":         "",
		"bit(1)":                "",
		"geometry":              "",
		"something new (later)": "",
	}

	for columnType, want := range cases {
		if got := SuggestFieldType(DatabaseMySQL, columnType); got != want {
			t.Errorf("%q: want %q, got %q", columnType, want, got)
		}
	}
}

func TestSuggestFieldTypeForPostgres(t *testing.T) {
	cases := map[string]FieldType{
		"boolean":                     FieldTypeYesNo,
		"integer":                     FieldTypeNumber,
		"bigint":                      FieldTypeNumber,
		"numeric":                     FieldTypeNumber,
		"real":                        FieldTypeNumber,
		"double precision":            FieldTypeNumber,
		"character varying":           FieldTypeText,
		"text":                        FieldTypeText,
		"uuid":                        FieldTypeText,
		"time without time zone":      FieldTypeText,
		"date":                        FieldTypeTime,
		"timestamp without time zone": FieldTypeTime,
		"timestamp with time zone":    FieldTypeTime,
		"money":                       "",
		"jsonb":                       "",
		"bytea":                       "",
		"ARRAY":                       "",
		"USER-DEFINED":                "",
		"interval":                    "",
	}

	for columnType, want := range cases {
		if got := SuggestFieldType(DatabasePostgres, columnType); got != want {
			t.Errorf("%q: want %q, got %q", columnType, want, got)
		}
	}
}

func TestListTables(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		tables, err := ListTables(context.Background(), database.connection, database.databaseType)
		if err != nil {
			t.Fatalf("ListTables: %v", err)
		}

		wantSchema := database.settings.Database
		if database.databaseType == DatabasePostgres {
			wantSchema = "public"
		}

		var gotNames []string
		for _, table := range tables {
			if table.Schema == wantSchema {
				gotNames = append(gotNames, table.Name)
			}
		}

		wantNames := []string{"channels", "events", "managers", "regions", "sales", "store_sales", "stores"}
		if !slices.Equal(gotNames, wantNames) {
			t.Errorf("want tables %v in %s, got %v", wantNames, wantSchema, gotNames)
		}

		sales := findTable(tables, wantSchema, "sales")

		wantColumns := []TableColumn{
			{Name: "id", CanBeEmpty: false, SuggestedType: FieldTypeNumber},
			{Name: "country", CanBeEmpty: true, SuggestedType: FieldTypeText},
			{Name: "channel", CanBeEmpty: false, SuggestedType: FieldTypeText},
			{Name: "amount", CanBeEmpty: false, SuggestedType: FieldTypeNumber},
			{Name: "cost", CanBeEmpty: false, SuggestedType: FieldTypeNumber},
			{Name: "quantity", CanBeEmpty: false, SuggestedType: FieldTypeNumber},
			{Name: "is_paid", CanBeEmpty: true, SuggestedType: FieldTypeYesNo},
			{Name: "big_code", CanBeEmpty: false, SuggestedType: FieldTypeNumber},
			{Name: "note", CanBeEmpty: true, SuggestedType: FieldTypeText},
			{Name: "rating", CanBeEmpty: true, SuggestedType: FieldTypeNumber},
		}

		if len(sales.Columns) != len(wantColumns) {
			t.Fatalf("want %d sales columns, got %v", len(wantColumns), sales.Columns)
		}

		for position, want := range wantColumns {
			got := sales.Columns[position]
			got.ColumnType = ""

			if got != want {
				t.Errorf("column %d: want %+v, got %+v (type %q)", position, want, got, sales.Columns[position].ColumnType)
			}
		}

		events := findTable(tables, wantSchema, "events")
		if events.Columns[1].SuggestedType != FieldTypeTime {
			t.Errorf("want happened_at as a date, got %+v", events.Columns[1])
		}
	})
}

// Postgres has the same table name in two schemas; they stay two tables.
func TestListTablesKeepsPostgresSchemasApart(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		if database.databaseType != DatabasePostgres {
			t.Skip("only Postgres has schemas inside one database")
		}

		tables, err := ListTables(context.Background(), database.connection, database.databaseType)
		if err != nil {
			t.Fatalf("ListTables: %v", err)
		}

		archived := findTable(tables, "archive", "regions")
		current := findTable(tables, "public", "regions")

		var archivedNames, currentNames []string
		for _, column := range archived.Columns {
			archivedNames = append(archivedNames, column.Name)
		}

		for _, column := range current.Columns {
			currentNames = append(currentNames, column.Name)
		}

		if !slices.Equal(archivedNames, []string{"code", "closed_on"}) {
			t.Errorf("archive.regions: want code, closed_on, got %v", archivedNames)
		}

		if !slices.Equal(currentNames, []string{"id", "name"}) {
			t.Errorf("public.regions: want id, name, got %v", currentNames)
		}

		if tables[0].Schema != "archive" {
			t.Errorf("want schemas sorted, archive first, got %q", tables[0].Schema)
		}
	})
}

func findTable(tables []Table, schema string, name string) Table {
	for _, table := range tables {
		if table.Schema == schema && table.Name == name {
			return table
		}
	}

	return Table{}
}
