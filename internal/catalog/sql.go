// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"context"
	"database/sql"
)

type sqlCatalog struct {
	db      *sql.DB
	dialect SqlDialect
}

func openSQL(ctx context.Context, c Connection, dialect SqlDialect) (*sqlCatalog, error) {
	db, err := c.Open()
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &sqlCatalog{
		db:      db,
		dialect: dialect,
	}, nil
}

func (s *sqlCatalog) Tables(ctx context.Context) ([]TableName, error) {
	results, err := s.db.QueryContext(ctx, s.dialect.Query)
	if err != nil {
		return nil, err
	}

	columnRows := make([]ColumnRow, 0)
	for results.Next() {
		columnRow := ColumnRow{}
		err = results.Scan(&columnRow.Schema, &columnRow.Table,
			&columnRow.Column, &columnRow.DataType, &columnRow.Nullable, &columnRow.Ordinal)
		if err != nil {
			return nil, err
		}

		columnRows = append(columnRows, columnRow)
	}

	// Group everything
	tables := groupByTable(columnRows)

	return tables, nil
}

func groupByTable(rows []ColumnRow) []TableName {
	grouped := make(map[string][]ColumnDetail, 0)

	for _, v := range rows {
		grouped[v.Table] = append(grouped[v.Table], ColumnDetail{
			Name: v.Column,
			Type: v.DataType,
		})
	}

	tableName := make([]TableName, 0)

	for k, v := range grouped {
		tableName = append(tableName, TableName{
			Name:    k,
			Columns: v,
		})
	}

	return tableName
}

func (s *sqlCatalog) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	return nil
}
