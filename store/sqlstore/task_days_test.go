// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestConvertTaskDatesToDaysKeepsTheDayEachDateWasSetFor(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_tables")

	const tasks, options = "zz_days_tasks", "zz_days_options"
	for _, table := range []string{tasks, options} {
		mustExec(t, "DROP TABLE IF EXISTS `"+table+"`")
		t.Cleanup(func() { _, _ = testDB.Exec("DROP TABLE IF EXISTS `" + table + "`") })
		mustExec(t, "CREATE TABLE `"+table+"` (id VARCHAR(36) PRIMARY KEY, start_date BIGINT, due_date BIGINT, INDEX (due_date))")
	}

	mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES ('t1', 'w1', ?, '', false, false, '', false, '', '', 0), ('t2', 'w1', ?, '', false, true, '', false, '', '', 0)`, tasks, options)

	for _, r := range []struct {
		id         string
		start, due any
	}{
		{"riga", 1793052000, 1793138400},
		{"los-angeles", nil, 1793170800},
		{"almaty", nil, 1793127600},
		{"milliseconds", nil, int64(1793138400000)},
		{"none", 0, nil},
		{"leap-day", nil, 1835388000},
	} {
		mustExec(t, "INSERT INTO `"+tasks+"` (id, start_date, due_date) VALUES (?, ?, ?)", r.id, r.start, r.due)
	}
	mustExec(t, "INSERT INTO `"+options+"` (id, due_date) VALUES ('kept', 1793138400)")

	w := &workspaceRepository{Db: testDB}
	for range 2 {
		if err := w.ConvertTaskDatesToDays(); err != nil {
			t.Fatal(err)
		}
	}

	for id, want := range map[string][2]string{
		"riga":         {"2026-10-27", "2026-10-28"},
		"los-angeles":  {"", "2026-10-28"},
		"almaty":       {"", "2026-10-28"},
		"milliseconds": {"", "2026-10-28"},
		"none":         {"", ""},
		"leap-day":     {"", "2028-02-29"},
	} {
		var start, due string
		err := testDB.QueryRow("SELECT COALESCE(CAST(start_date AS CHAR), ''), COALESCE(CAST(due_date AS CHAR), '') FROM `"+tasks+"` WHERE id = ?", id).Scan(&start, &due)
		if err != nil {
			t.Fatal(err)
		}

		if start != want[0] || due != want[1] {
			t.Errorf("%s = %q, %q; want %q, %q", id, start, due, want[0], want[1])
		}
	}

	var kept string
	if err := testDB.QueryRow("SELECT CAST(due_date AS CHAR) FROM `" + options + "`").Scan(&kept); err != nil || kept != "1793138400" {
		t.Errorf("a table that is not a task table holds %q, %v; want it left alone", kept, err)
	}

	var indexes int
	if err := testDB.QueryRow("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND column_name = 'due_date'", tasks).Scan(&indexes); err != nil || indexes != 1 {
		t.Errorf("due_date has %d indexes, %v; want its index kept", indexes, err)
	}
}

func TestDatePresetsAreDaysOfTheUsersToday(t *testing.T) {
	for _, zone := range []string{"Europe/Riga", "America/Los_Angeles", "Pacific/Auckland"} {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			t.Skip("no time zone data")
		}

		today := time.Now().In(loc).Format(time.DateOnly)
		start, end, ok := resolveDateOption("Today", loc)
		tomorrow, _ := time.Parse(time.DateOnly, today)
		if !ok || start != today || end != tomorrow.AddDate(0, 0, 1).Format(time.DateOnly) {
			t.Errorf("Today in %s = %s to %s, want %s and the day after", zone, start, end, today)
		}
	}

	start, end, ok := resolveDateOption("Next quarter", time.UTC)
	first, err := time.Parse(time.DateOnly, start)
	if !ok || err != nil || first.Day() != 1 || (first.Month()-1)%3 != 0 || end != first.AddDate(0, 3, 0).Format(time.DateOnly) {
		t.Errorf("Next quarter = %s to %s, want a whole quarter", start, end)
	}
}

func TestDateFiltersOnDaysNeverCompareWithText(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET start_date = NULL WHERE id = 't2'")

	for _, tc := range []struct {
		name   string
		filter model.Filter
		want   []string
	}{
		{"is set", model.Filter{Field: "start_date", Operator: "is_set"}, []string{"t1", "t3", "t4"}},
		{"is not set", model.Filter{Field: "start_date", Operator: "is_not_set"}, []string{"t2"}},
		{"an exact day", model.Filter{Field: "start_date", Operator: "is", Value: "Exact date", Date: "2026-01-10"}, []string{"t1", "t3", "t4"}},
		{"a preset the server does not know", model.Filter{Field: "start_date", Operator: "is", Value: "Someday"}, []string{}},
		{"not an exact day", model.Filter{Field: "start_date", Operator: "is_not", Value: "Exact date", Date: "2026-01-10"}, []string{"t2"}},
		{"not today", model.Filter{Field: "start_date", Operator: "is_not", Value: "Today"}, []string{"t1", "t2", "t3", "t4"}},
		{"not a preset the server does not know", model.Filter{Field: "start_date", Operator: "is_not", Value: "Someday"}, []string{"t1", "t2", "t3", "t4"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
				w.BuildFilterSQLInline(flat(tc.filter), filterHeaders, nil, filterTableNames))
			if err != nil {
				t.Fatal(err)
			}

			if got := rowIDs(rows); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNullConditionMatchesOnlyItsOwnOperators(t *testing.T) {
	for _, tc := range []struct {
		op   string
		want string
	}{
		{"is_set", "d IS NOT NULL"},
		{"is_not_set", "d IS NULL"},
		{"is", ""},
		{"", ""},
	} {
		got := nullCondition("d", tc.op)

		if got.SQL != tc.want {
			t.Errorf("nullCondition(%q) = %q, want %q", tc.op, got.SQL, tc.want)
		}
	}
}

func TestUpdatingADateWritesTheDayAndHandsItBack(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_tables")

	const table = "zz_days_update"
	createDynamicTable(t, table, createTaskTable)
	mustExec(t, `INSERT INTO workspace_tables (id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES ('t1', 'w1', ?, '', false, false, '', false, '', '', 0)`, table)
	mustExec(t, "INSERT INTO `"+table+"` (id, name, deleted_at) VALUES ('task-1', 'Launch', 0)")

	w := &workspaceRepository{Db: testDB}
	for _, tc := range []struct {
		value any
		want  string
	}{
		{"2026-10-28", "2026-10-28"},
		{"2028-02-29", "2028-02-29"},
		{nil, ""},
	} {
		row, err := w.UpdateTaskTx("w1", "t1", "task-1", "due_date", "", tc.value, true, false, false)
		if err != nil {
			t.Fatalf("%v: %v", tc.value, err)
		}

		var stored string
		if err := testDB.QueryRow("SELECT COALESCE(CAST(due_date AS CHAR), '') FROM `" + table + "` WHERE id = 'task-1'").Scan(&stored); err != nil {
			t.Fatal(err)
		}

		handed := ""
		if day := rowValue(t, (*row)["due_date"]); day != nil {
			handed = fmt.Sprint(day)
		}

		if stored != tc.want || handed != tc.want {
			t.Errorf("%v stored as %q and handed back as %q, want %q", tc.value, stored, handed, tc.want)
		}
	}
}

func TestAssignedSectionsFindAndPageEveryTask(t *testing.T) {
	requireDB(t)
	const table = "zz_days_assigned"
	createDynamicTable(t, table, createTaskTable)

	today := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	groups := []struct {
		name       string
		count      int
		first, gap int
	}{
		{"overdue", 30, -40, 1},
		{"today", 5, 0, 0},
		{"thisWeek", 120, 1, 0},
		{"upcoming", 10, 8, 3},
		{"noDueDate", 7, 0, 0},
	}
	for _, g := range groups {
		for i := range g.count {
			var due any
			switch g.name {
			case "noDueDate":
			case "thisWeek":
				due = today.AddDate(0, 0, 1+i%7).Format(time.DateOnly)
			default:
				due = today.AddDate(0, 0, g.first+i*g.gap).Format(time.DateOnly)
			}

			mustExec(t, "INSERT INTO `"+table+"` (id, name, assignee, due_date, deleted_at) VALUES (?, ?, 'me', ?, 0)", fmt.Sprintf("%s-%03d", g.name, i), g.name, due)
		}
	}

	// The conditions assignedSections builds for a user whose today is 28 October.
	sections := []model.SQLFilter{
		{SQL: "main.due_date < ?", Args: []any{"2026-10-28"}},
		{SQL: "(main.due_date >= ? AND main.due_date < ?)", Args: []any{"2026-10-28", "2026-10-29"}},
		{SQL: "(main.due_date >= ? AND main.due_date < ?)", Args: []any{"2026-10-29", "2026-11-05"}},
		{SQL: "main.due_date >= ?", Args: []any{"2026-11-05"}},
		{SQL: "main.due_date IS NULL"},
	}
	mine := model.SQLFilter{SQL: "main.assignee = ?", Args: []any{"me"}}
	sources := []model.TaskSource{{TableID: "t1", TableName: table, Filter: mine}}

	w := &workspaceRepository{Db: testDB}
	counts, err := w.CountTasksAcrossTables(context.Background(), sources, sections)
	if err != nil {
		t.Fatal(err)
	}

	for i, g := range groups {
		if counts[i] != g.count {
			t.Errorf("%s counts %d, want %d", g.name, counts[i], g.count)
		}

		seen := map[string]bool{}
		filter := mine.And(sections[i])
		for range g.count + 1 {
			rows, err := w.GetTasksAcrossTables(context.Background(), []model.TaskSource{{TableID: "t1", TableName: table, Filter: filter}}, 50)
			if err != nil {
				t.Fatal(err)
			}

			for _, row := range rows {
				id := row["id"].(string)
				if seen[id] || row["name"] != g.name {
					t.Errorf("%s page gave %s again or from another group", g.name, id)
				}

				seen[id] = true
			}

			if len(rows) < 50 {
				break
			}

			last := rows[len(rows)-1]
			filter = mine.And(sections[i]).And(model.RowsAfter("main.due_date", rowValue(t, last["due_date"]), last["id"].(string)))
		}

		if len(seen) != g.count {
			t.Errorf("%s pages gave %d tasks, want %d", g.name, len(seen), g.count)
		}
	}
}

func TestLeapAndClockChangeDaysStayWholeDays(t *testing.T) {
	w := setupFilterTables(t)
	mustExec(t, "UPDATE `"+filterTaskTable+"` SET start_date = CASE id WHEN 't1' THEN '2028-02-28' WHEN 't2' THEN '2028-02-29' WHEN 't3' THEN '2026-03-29' ELSE '2026-10-25' END")
	riga, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		t.Skip("no time zone data")
	}

	for _, tc := range []struct {
		name   string
		filter model.Filter
		want   []string
	}{
		{"the leap day", model.Filter{Field: "start_date", Operator: "is", Value: "Exact date", Date: "2028-02-29"}, []string{"t2"}},
		{"across the leap day", model.Filter{Field: "start_date", Operator: "is", Value: "Date range", Date: "2028-02-28", Date2: "2028-03-01"}, []string{"t1", "t2"}},
		{"after the leap day", model.Filter{Field: "start_date", Operator: "is", Value: "After date", Date: "2028-02-28"}, []string{"t2"}},
		{"the day clocks go forward", model.Filter{Field: "start_date", Operator: "is", Value: "Exact date", Date: "2026-03-29"}, []string{"t3"}},
		{"the day clocks go back", model.Filter{Field: "start_date", Operator: "is", Value: "Exact date", Date: "2026-10-25"}, []string{"t4"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := w.GetTableRowsFiltered(context.Background(), filterTaskTable,
				w.BuildFilterSQLInline(flat(tc.filter), filterHeaders, riga, filterTableNames))
			if err != nil {
				t.Fatal(err)
			}

			if got := rowIDs(rows); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStreamedRowsCarryEachDayAsStored(t *testing.T) {
	requireDB(t)
	const table = "zz_days_stream"
	createDynamicTable(t, table, createTaskTable)
	mustExec(t, "INSERT INTO `"+table+"` (id, name, start_date, due_date, deleted_at) VALUES ('a', 'Launch', '2026-10-27', '2026-10-28', 0), ('b', 'Leap', NULL, '2028-02-29', 0)")

	w := &workspaceRepository{Db: testDB}
	got := map[string][2]any{}
	err := w.StreamTableRows(context.Background(), table, model.SQLFilter{}, "ORDER BY main.id", func(row map[string]interface{}) error {
		got[row["id"].(string)] = [2]any{rowValue(t, row["start_date"]), rowValue(t, row["due_date"])}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got["a"] != [2]any{"2026-10-27", "2026-10-28"} || got["b"] != [2]any{nil, "2028-02-29"} {
		t.Errorf("rows carry %v, want each day as stored and no date as nil", got)
	}
}

func TestTaskFieldValueReadsADayOrNothing(t *testing.T) {
	requireDB(t)
	const table = "zz_days_field_value"
	createDynamicTable(t, table, createTaskTable)
	mustExec(t, "INSERT INTO `"+table+"` (id, name, due_date, deleted_at) VALUES ('task-1', 'Launch', '2026-10-28', 0)")

	w := &workspaceRepository{Db: testDB}
	due, err := w.GetTaskFieldValue(table, "task-1", "due_date")
	if err != nil || len(due) < 10 || due[:10] != "2026-10-28" {
		t.Errorf("due date = %q, %v; want 2026-10-28", due, err)
	}

	start, err := w.GetTaskFieldValue(table, "task-1", "start_date")
	if err != nil || start != "" {
		t.Errorf("start date = %q, %v; want nothing", start, err)
	}
}
