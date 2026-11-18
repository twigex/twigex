// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Result struct {
	Columns    []Column
	Rows       []map[string]any
	DateRanges []ResolvedDateRange
}

// RunQuery runs a built query on a database connection and returns its
// columns and rows. Cancelling ctx stops the query.
func RunQuery(ctx context.Context, connection *sql.DB, query BuiltQuery) (Result, error) {
	rows, err := connection.QueryContext(ctx, query.SQL, query.Values...)
	if err != nil {
		return Result{}, err
	}
	defer rows.Close()

	databaseColumns, err := rows.ColumnTypes()
	if err != nil {
		return Result{}, err
	}

	if len(databaseColumns) != len(query.Columns) {
		return Result{}, fmt.Errorf("query returned %d columns, expected %d", len(databaseColumns), len(query.Columns))
	}

	columns := slices.Clone(query.Columns)
	for position := range columns {
		resultColumnType := databaseColumns[position].DatabaseTypeName()
		columns[position].SentAsText = isSentAsText(columns[position], resultColumnType)
	}

	result := Result{
		Columns:    columns,
		Rows:       []map[string]any{},
		DateRanges: query.DateRanges,
	}

	for rows.Next() {
		row, err := readRow(rows, columns)
		if err != nil {
			return Result{}, err
		}

		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return Result{}, err
	}

	return result, nil
}

// isSentAsText says whether a number field's values are sent as text, because
// they can be longer than the 15 to 16 digits a browser keeps exactly. Measures
// are always numbers. The result does not say which database it came from,
// and each driver names its types in its own way, so every database is asked.
func isSentAsText(column Column, resultColumnType string) bool {
	isNumberDimension := column.Kind == FieldKindDimension && column.Type == FieldTypeNumber
	if !isNumberDimension {
		return false
	}

	for _, database := range supportedDatabases {
		if database.isLongNumberType(resultColumnType) {
			return true
		}
	}

	return false
}

func readRow(rows *sql.Rows, columns []Column) (map[string]any, error) {
	texts := make([]sql.NullString, len(columns))

	targets := make([]any, len(columns))
	for position := range texts {
		targets[position] = &texts[position]
	}

	if err := rows.Scan(targets...); err != nil {
		return nil, err
	}

	row := make(map[string]any, len(columns))

	for position, column := range columns {
		value, err := convertValue(column, texts[position])
		if err != nil {
			return nil, err
		}

		row[column.Key] = value
	}

	return row, nil
}

// convertValue turns a value read from the database into what the result
// sends: numbers as numbers, yes/no as true or false, empty as nil.
func convertValue(column Column, databaseValue sql.NullString) (any, error) {
	if !databaseValue.Valid {
		return nil, nil
	}

	text := databaseValue.String

	if column.SentAsText {
		return text, nil
	}

	switch column.Type {
	case FieldTypeNumber:
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %q is not a number", column.Key, text)
		}

		return number, nil

	case FieldTypeYesNo:
		return parseYesNo(column.Key, text)

	case FieldTypeTime:
		return parseDatabaseTime(column.Key, text)
	}

	return text, nil
}

// parseDatabaseTime reads a time the query already turned into UTC and writes
// it as results send dates: 2025-03-01T00:00:00Z.
func parseDatabaseTime(key string, text string) (string, error) {
	moment, _, err := parseDate(key, text)
	if err != nil {
		return "", fmt.Errorf("%s: %q is not a date", key, text)
	}

	return dateForResult(moment), nil
}

func parseYesNo(key string, text string) (bool, error) {
	isYes, isYesOrNo := readYesNo(text)
	if !isYesOrNo {
		return false, fmt.Errorf("%s: %q is not yes or no", key, text)
	}

	return isYes, nil
}

// readYesNo reads yes and no as databases and charts write them: 1 or 0,
// true or false, and Postgres's t or f.
func readYesNo(text string) (isYes bool, isYesOrNo bool) {
	lowercase := strings.ToLower(text)

	switch lowercase {
	case "1", "true", "t":
		return true, true
	case "0", "false", "f":
		return false, true
	}

	return false, false
}
