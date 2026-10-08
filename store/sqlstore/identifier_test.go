// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"strings"
	"testing"
)

func TestAnInjectedColumnNameNeverReachesTheTable(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_test_identifier", createTaskTable)
	mustExec(t, "INSERT INTO zz_test_identifier (id, name, assignee, deleted_at) VALUES ('t1', 'before', 'u1', 0)")

	w := &workspaceRepository{Db: testDB}
	for _, field := range []string{"name` = 'injected', `name", "assignee` = '', `name"} {
		if err := w.ClearAssigneeByUserID("zz_test_identifier", field, "u1"); err == nil {
			t.Errorf("clearing %q ran; want it refused", field)
		}
	}

	var name string
	if err := testDB.QueryRow("SELECT name FROM zz_test_identifier WHERE id = 't1'").Scan(&name); err != nil {
		t.Fatal(err)
	}

	if name != "before" {
		t.Errorf("name = %q, want it untouched", name)
	}
}

func TestQuoteIdentTakesOnlyNamesTheServerMakes(t *testing.T) {
	for _, name := range []string{
		"t3f9a1c2e8b7d4f6a0c1e2b3d4f5a6b7c",
		"c8b21e4f0a9d3c7e1b5a2f6d8c4e0b1a",
		"xKqTzWmPbR_t3f9a1c2e8b7d4f6a0c1e2b3d4f5a6b7c",
		"due_date",
		"priority_2",
		"7days",
	} {
		quoted, err := quoteIdent(name)
		if err != nil || quoted != "`"+name+"`" {
			t.Errorf("quoteIdent(%q) = %q, %v; want it quoted", name, quoted, err)
		}
	}

	for _, name := range []string{
		"",
		"x` TEXT, DROP COLUMN `created_by",
		"name; DROP TABLE users",
		"a b",
		"main.name",
		"Pārdošana",
		"name--",
		strings.Repeat("a", 65),
	} {
		if quoted, err := quoteIdent(name); err == nil {
			t.Errorf("quoteIdent(%q) = %q; want it refused", name, quoted)
		}
	}
}
