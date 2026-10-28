// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"strings"
)

func (w *workspaceRepository) ConvertTaskDatesToDays() error {
	taskTables, err := w.taskTableNames()
	if err != nil {
		return err
	}

	pending, err := w.tablesWithTimestampDates()
	if err != nil {
		return err
	}

	for table, columns := range pending {
		if !taskTables[table] {
			continue
		}

		if err := w.convertTaskTableDates(table, columns); err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) taskTableNames() (map[string]bool, error) {
	rows, err := w.Db.Query(`SELECT name FROM workspace_tables WHERE linked = 0 AND single_select = 0`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	names := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		names[name] = true
	}

	return names, rows.Err()
}

func (w *workspaceRepository) tablesWithTimestampDates() (map[string][]string, error) {
	rows, err := w.Db.Query(`
		SELECT table_name, column_name FROM information_schema.columns
		WHERE table_schema = DATABASE()
		AND column_name IN ('start_date', 'due_date')
		AND data_type = 'bigint'`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tables := map[string][]string{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, err
		}

		tables[table] = append(tables[table], column)
	}

	return tables, rows.Err()
}

// Each old value is midnight in the setter's time zone, so rounding it to
// the nearest UTC midnight gives the day they picked. The day is stored as
// the number YYYYMMDD, which MODIFY ... DATE reads as that day. Converted
// values are small numbers that the UPDATEs skip, so a second run is safe.
func (w *workspaceRepository) convertTaskTableDates(tableName string, columns []string) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, column := range columns {
		col, err := quoteIdent(column)
		if err != nil {
			return err
		}

		steps := []string{
			"UPDATE " + table + " SET " + col + " = NULL WHERE " + col + " < 19000101",
			"UPDATE " + table + " SET " + col + " = FLOOR(" + col + " / 1000) WHERE " + col + " > 1000000000000",
			"UPDATE " + table + " SET " + col + " = DATE_FORMAT(DATE('1970-01-01') + INTERVAL ROUND(" + col + " / 86400) DAY, '%Y%m%d') WHERE " + col + " >= 100000000",
		}

		for _, step := range steps {
			if _, err := tx.Exec(step); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	modify := make([]string, 0, len(columns))
	for _, column := range columns {
		col, err := quoteIdent(column)
		if err != nil {
			return err
		}

		modify = append(modify, "MODIFY "+col+" DATE")
	}

	_, err = w.Db.Exec("ALTER TABLE " + table + " " + strings.Join(modify, ", "))
	return err
}
