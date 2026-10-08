// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestTaskExportCellsReadLikeTheReport(t *testing.T) {
	riga, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		t.Skip("no time zone data")
	}

	e := &TaskExport{loc: riga, dateLayout: "02.01.2006", tableNames: map[string]string{"t-clients": "Clients"}}
	lookups := exportLookups{
		users: map[string]string{"u1": "Anna Berzina"},
		files: map[string]map[string][]string{"docs": {"task-1": {"b.pdf", "a.png"}}},
	}
	// 1790546400 is 2026-09-28 01:00 in Riga, 2026-09-27 22:00 UTC.
	for _, tc := range []struct {
		header model.WorkspaceHeaders
		value  interface{}
		want   string
	}{
		{model.WorkspaceHeaders{Name: "due_date"}, "1790546400", "28.09.2026"},
		{model.WorkspaceHeaders{Name: "updated_at"}, "1790546400000", "28.09.2026"},
		{model.WorkspaceHeaders{Name: "due_date"}, "0", ""},
		{model.WorkspaceHeaders{Name: "deadline", HeaderUsage: "date", HeaderType: "DATE"}, "2026-10-05", "05.10.2026"},
		{model.WorkspaceHeaders{Name: "deadline", HeaderUsage: "date", HeaderType: "DATE"}, time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), "05.10.2026"},
		{model.WorkspaceHeaders{Name: "assignee"}, "u1", "Anna Berzina"},
		{model.WorkspaceHeaders{Name: "owner", HeaderUsage: "assignee"}, "u1", "Anna Berzina"},
		{model.WorkspaceHeaders{Name: "status", SingleSelect: true}, &model.TaskOrderField{Name: "Done"}, "Done"},
		{model.WorkspaceHeaders{Name: "status", SingleSelect: true}, (*model.TaskOrderField)(nil), ""},
		{model.WorkspaceHeaders{Name: "project", LinkedID: "l1"}, []model.LinkedItem{{Name: "Alpha"}, {Name: "Beta"}}, "Alpha, Beta"},
		{model.WorkspaceHeaders{Name: "client", HeaderUsage: "master link"}, "t-clients", "Clients"},
		{model.WorkspaceHeaders{Name: "docs", HeaderUsage: "file"}, "", "b.pdf, a.png"},
		{model.WorkspaceHeaders{Name: "done", HeaderType: "TINYINT"}, "1", "true"},
		{model.WorkspaceHeaders{Name: "price", HeaderUsage: "decimal"}, "12.50", "12.50"},
		{model.WorkspaceHeaders{Name: "notes"}, nil, ""},
		{model.WorkspaceHeaders{Name: "notes"}, "a, \"quoted\" note", "a, \"quoted\" note"},
	} {
		row := map[string]interface{}{"id": "task-1", tc.header.Name: tc.value}
		if got := e.cell(tc.header, row, lookups); got != tc.want {
			t.Errorf("%s %v = %q, want %q", tc.header.Name, tc.value, got, tc.want)
		}
	}

	e.dateLayout = "01/02/2006"
	if got := e.cell(model.WorkspaceHeaders{Name: "due_date"}, map[string]interface{}{"due_date": "1790546400"}, lookups); got != "09/28/2026" {
		t.Errorf("12h date = %q, want 09/28/2026", got)
	}
}

func TestTaskExportCellsNeverRunAsFormulas(t *testing.T) {
	for in, want := range map[string]string{
		`=HYPERLINK("http://x")`: `'=HYPERLINK("http://x")`,
		"+1+1":                   "'+1+1",
		"@SUM(A1)":               "'@SUM(A1)",
		"-2+3":                   "'-2+3",
		"-12.5":                  "-12.5",
		"Plain task":             "Plain task",
		"":                       "",
	} {
		if got := inertCell(in); got != want {
			t.Errorf("inertCell(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTaskExportFileNameIsTheWorkspaceWhenThereIsOne(t *testing.T) {
	e := &TaskExport{loc: time.UTC}
	one := model.WorkspaceTable{WorkspaceID: "w1", WorkspaceName: "Marketing"}
	other := model.WorkspaceTable{WorkspaceID: "w2", WorkspaceName: "Sales"}

	e.tables = []taskExportTable{{table: one}, {table: one}}
	if got := e.FileName(); !strings.HasPrefix(got, "Marketing ") || !strings.HasSuffix(got, ".csv") {
		t.Errorf("one workspace = %q", got)
	}
	e.tables = []taskExportTable{{table: one}, {table: other}}
	if got := e.FileName(); !strings.HasPrefix(got, "Task report ") {
		t.Errorf("two workspaces = %q", got)
	}
}
