// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

// TopLevelTaskSQL matches the tasks of a task table that have no parent. A
// top-level task holds NULL, but the condition is written so that no index
// on parent_task_id can serve it: most tasks are top-level, and MySQL would
// read them all through that index and sort them rather than walk the index
// of the column a page is ordered by.
const TopLevelTaskSQL = "COALESCE(main.parent_task_id, '') = ''"

// SQLFilter is a WHERE fragment with ? placeholders and the values bound to
// them, in order. User input only ever travels in Args, never in SQL.
type SQLFilter struct {
	SQL  string
	Args []any
}

// And joins two filters with AND, keeping their arguments in order.
func (f SQLFilter) And(other SQLFilter) SQLFilter {
	switch {
	case other.SQL == "":
		return f
	case f.SQL == "":
		return other
	}

	args := make([]any, 0, len(f.Args)+len(other.Args))
	args = append(append(args, f.Args...), other.Args...)
	return SQLFilter{SQL: f.SQL + " AND " + other.SQL, Args: args}
}

// RowsAfter matches the rows that come after the row with value and id when
// rows are ordered by column and then main.id. A nil value is NULL, which
// sorts first. column is written into the SQL as given, so it must be a
// constant, never a name from a request.
func RowsAfter(column string, value *int64, id string) SQLFilter {
	if value == nil {
		return SQLFilter{
			SQL:  "(" + column + " IS NOT NULL OR main.id > ?)",
			Args: []any{id},
		}
	}

	return SQLFilter{
		SQL:  "(" + column + " > ? OR (" + column + " = ? AND main.id > ?))",
		Args: []any{*value, *value, id},
	}
}
