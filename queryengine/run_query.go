// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
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
		columns[position].KeptAsText = isKeptAsText(columns[position], resultColumnType)
	}

	var textRows [][]sql.NullString

	for rows.Next() {
		texts, err := readTexts(rows, len(columns))
		if err != nil {
			return Result{}, err
		}

		textRows = append(textRows, texts)
	}

	if err := rows.Err(); err != nil {
		return Result{}, err
	}

	keepTooLongMeasuresAsText(columns, textRows)

	result := Result{
		Columns:    columns,
		Rows:       make([]map[string]any, 0, len(textRows)),
		DateRanges: query.DateRanges,
	}

	for _, texts := range textRows {
		row, err := rowFromTexts(texts, columns)
		if err != nil {
			return Result{}, err
		}

		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

// keepTooLongMeasuresAsText marks a count, sum, minimum or maximum column as
// kept as text when any of its values would lose digits as a float64, such as
// the largest of BIGINT ids. These add up or pick stored values, so every
// digit is real. An average or a formula divides, and its last digits never
// fit a float64 and do not matter, so it stays a number.
func keepTooLongMeasuresAsText(columns []Column, textRows [][]sql.NullString) {
	for position := range columns {
		column := &columns[position]

		if !digitsMatter(column.MeasureType) {
			continue
		}

		for _, texts := range textRows {
			value := texts[position]

			if value.Valid && !isExactAsNumber(value.String) {
				column.KeptAsText = true
				break
			}
		}
	}
}

func digitsMatter(measureType MeasureType) bool {
	switch measureType {
	case MeasureCount, MeasureCountDistinct, MeasureSum, MeasureMinimum, MeasureMaximum:
		return true
	}

	return false
}

// isExactAsNumber says whether a number written as text becomes a float64
// without losing digits: the shortest text of that float64 is the same
// number. "151.50" gives 151.5 and is exact; "9223372036854775807" gives
// 9223372036854775808 and is not. Text that is not a number counts as exact,
// so the conversion reports it.
func isExactAsNumber(text string) bool {
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return true
	}

	written, isNumber := new(big.Rat).SetString(text)
	if !isNumber {
		return true
	}

	shortest := strconv.FormatFloat(number, 'g', -1, 64)
	converted, _ := new(big.Rat).SetString(shortest)

	return written.Cmp(converted) == 0
}

// isKeptAsText says whether a number dimension's values are kept as text,
// because its database type can hold more than the 15 to 16 digits a float64
// holds exactly. Measures are decided by their values instead
// (keepTooLongMeasuresAsText). The result does not say which database it came
// from, and each driver names its types in its own way, so every database is
// asked.
func isKeptAsText(column Column, resultColumnType string) bool {
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

// readTexts reads one row, every value as the database's text.
func readTexts(rows *sql.Rows, columnCount int) ([]sql.NullString, error) {
	texts := make([]sql.NullString, columnCount)

	targets := make([]any, columnCount)
	for position := range texts {
		targets[position] = &texts[position]
	}

	if err := rows.Scan(targets...); err != nil {
		return nil, err
	}

	return texts, nil
}

func rowFromTexts(texts []sql.NullString, columns []Column) (map[string]any, error) {
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

	if column.KeptAsText {
		// Postgres's NUMERIC has NaN and Infinity too, and they are no more
		// a number here than float8's are (below).
		if isNotANumber(text) {
			return nil, nil
		}

		return text, nil
	}

	switch column.Type {
	case FieldTypeNumber:
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %q is not a number", column.Key, text)
		}

		// Postgres's NaN and Infinity have no JSON number to be sent as.
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return nil, nil
		}

		return number, nil

	case FieldTypeYesNo:
		return parseYesNo(column.Key, text)

	case FieldTypeTime:
		// Postgres's infinity is no date, and is sent as an empty one.
		if text == "infinity" || text == "-infinity" {
			return nil, nil
		}

		return parseDatabaseTime(column.Key, text)
	}

	return text, nil
}

func isNotANumber(text string) bool {
	switch text {
	case "NaN", "Infinity", "-Infinity":
		return true
	}

	return false
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

// readYesNo reads yes and no as databases and requests write them: 1 or 0,
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
