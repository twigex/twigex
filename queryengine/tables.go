// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
)

// Table is one table or view of a database, for choosing what a dataset
// reads.
type Table struct {
	// Schema is the MySQL database or the Postgres schema the table is in.
	Schema  string
	Name    string
	Columns []TableColumn
}

type TableColumn struct {
	Name string

	// ColumnType is the type as the database names it: "varchar(100)" or
	// "tinyint(1)" in MySQL, "character varying" or "integer" in Postgres.
	ColumnType string

	CanBeEmpty bool

	// SuggestedType is the field type a dataset would give the column, or
	// empty when none fits (binary data, JSON and so on).
	SuggestedType FieldType
}

// ListTables lists a database's tables and their columns, sorted by schema
// and name, each table's columns in their own order.
func ListTables(ctx context.Context, connection *sql.DB, databaseType DatabaseType) ([]Table, error) {
	database, isSupported := databaseFor(databaseType)
	if !isSupported {
		return nil, fmt.Errorf("query engine: unknown database type %q", databaseType)
	}

	rows, err := connection.QueryContext(ctx, database.listTablesSQL())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	tablePositions := make(map[string]int)

	for rows.Next() {
		var schema, tableName, columnName, columnType, isNullable string

		if err := rows.Scan(&schema, &tableName, &columnName, &columnType, &isNullable); err != nil {
			return nil, err
		}

		column := TableColumn{
			Name:          columnName,
			ColumnType:    columnType,
			CanBeEmpty:    isNullable == "YES",
			SuggestedType: SuggestFieldType(databaseType, columnType),
		}

		key := schema + "\x00" + tableName

		position, seen := tablePositions[key]
		if !seen {
			tables = append(tables, Table{Schema: schema, Name: tableName})
			position = len(tables) - 1
			tablePositions[key] = position
		}

		tables[position].Columns = append(tables[position].Columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortStableFunc(tables, func(first Table, second Table) int {
		if first.Schema != second.Schema {
			return strings.Compare(first.Schema, second.Schema)
		}

		return strings.Compare(first.Name, second.Name)
	})

	return tables, nil
}

// SuggestFieldType suggests the field type for a column of a database type,
// or empty when no field type fits. A time of day, without a date, is text:
// it cannot be grouped by day or month.
func SuggestFieldType(databaseType DatabaseType, columnType string) FieldType {
	database, isSupported := databaseFor(databaseType)
	if !isSupported {
		return ""
	}

	lowercase := strings.ToLower(strings.TrimSpace(columnType))

	return database.suggestFieldType(lowercase)
}
