// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestAssignedSectionsSplitTheDayInTheUsersTimeZone(t *testing.T) {
	riga, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		t.Skip("no time zone data")
	}

	now := time.Date(2026, 9, 28, 23, 30, 0, 0, riga)
	todayStart := time.Date(2026, 9, 28, 0, 0, 0, 0, riga).Unix()
	tomorrow := time.Date(2026, 9, 29, 0, 0, 0, 0, riga).Unix()
	weekEnd := time.Date(2026, 10, 6, 0, 0, 0, 0, riga).Unix()

	var keys []string
	bounds := map[string][]any{}
	for _, s := range assignedSections(now) {
		keys = append(keys, s.key)
		bounds[s.key] = s.filter.Args
	}
	if want := []string{"overdue", "today", "thisWeek", "upcoming", "noDueDate"}; !slices.Equal(keys, want) {
		t.Errorf("sections = %v, want %v", keys, want)
	}
	for key, want := range map[string][]any{
		"overdue":   {todayStart},
		"today":     {todayStart, tomorrow},
		"thisWeek":  {tomorrow, weekEnd},
		"upcoming":  {weekEnd},
		"noDueDate": nil,
	} {
		if !slices.Equal(bounds[key], want) {
			t.Errorf("%s bounds = %v, want %v", key, bounds[key], want)
		}
	}
}

type fakeAssignedStore struct {
	fakeReadAccessStore
	counts    []int
	counted   int
	pageCalls int
	lastPage  model.SQLFilter
	rows      []map[string]interface{}
}

func (f *fakeAssignedStore) GetTableMetasForUser(string) ([]model.WorkspaceTable, error) {
	return []model.WorkspaceTable{{
		ID: "t-a", WorkspaceID: "ws-a", Name: "pre_tasks", Prefix: "pre_", WorkspaceName: "Workspace A",
		DisplayName: sql.NullString{String: "Tasks", Valid: true},
	}}, nil
}

func (f *fakeAssignedStore) GetStatusTypeMap(string) (map[string]string, error) {
	return map[string]string{}, nil
}

// CountTasksAcrossTables counts the sections from the first, so a single
// section asked for must be overdue.
func (f *fakeAssignedStore) CountTasksAcrossTables(_ context.Context, _ []model.TaskSource, groups []model.SQLFilter) ([]int, error) {
	f.counted = len(groups)
	return f.counts[:len(groups)], nil
}

func (f *fakeAssignedStore) GetTasksAcrossTables(_ context.Context, sources []model.TaskSource, limit int) ([]map[string]interface{}, error) {
	f.pageCalls++
	f.lastPage = sources[0].Filter
	return f.rows[:min(limit, len(f.rows))], nil
}

func (f *fakeAssignedStore) GetLinkedTableIDByName(string, string) string { return "" }

func (f *fakeAssignedStore) GetMainViewIDs(map[string]string) (map[string]string, error) {
	return map[string]string{"t-a": "grid-a"}, nil
}

func TestGetAssignedToMeReturnsCountsAndOnePagePerSection(t *testing.T) {
	ws := &fakeAssignedStore{
		counts: []int{3, 0, 0, 0, 1},
		rows:   []map[string]interface{}{{"id": "task-1", "table_id": "t-a", "due_date": int64(7)}},
	}
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws

	page, appErr := a.GetAssignedToMe(context.Background(), "u1", "UTC", "", "", 50)
	if appErr != nil {
		t.Fatal(appErr)
	}
	var totals []int
	for _, s := range page.Sections {
		totals = append(totals, s.Total)
		if s.Next != "" {
			t.Errorf("%s has a next page, want none past its one task", s.Key)
		}
	}
	if !slices.Equal(totals, []int{3, 0, 0, 0, 1}) || ws.pageCalls != 2 {
		t.Errorf("totals = %v after %d page queries, want every section's count and a query only for the two with tasks", totals, ws.pageCalls)
	}
	if got := page.Tables["t-a"]; got != (model.AssignedTable{WorkspaceID: "ws-a", WorkspaceName: "Workspace A", TableName: "Tasks", ViewID: "grid-a"}) {
		t.Errorf("table = %+v", got)
	}
}

func TestGetAssignedToMeLoadsMoreAfterTheLastTaskShown(t *testing.T) {
	ws := &fakeAssignedStore{
		counts: []int{3},
		rows: []map[string]interface{}{
			{"id": "task-1", "table_id": "t-a", "due_date": int64(5)},
			{"id": "task-2", "table_id": "t-a", "due_date": "9"},
			{"id": "task-3", "table_id": "t-a", "due_date": int64(9)},
		},
	}
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws

	page, appErr := a.GetAssignedToMe(context.Background(), "u1", "UTC", "overdue", "", 2)
	if appErr != nil {
		t.Fatal(appErr)
	}
	first := page.Sections[0]
	if len(page.Sections) != 1 || len(first.Tasks) != 2 || first.Next == "" || ws.counted != 1 {
		t.Fatalf("section = %+v after counting %d sections, want overdue alone with two tasks and more to come", page.Sections, ws.counted)
	}

	ws.rows = ws.rows[2:]
	page, appErr = a.GetAssignedToMe(context.Background(), "u1", "UTC", "overdue", first.Next, 2)
	if appErr != nil {
		t.Fatal(appErr)
	}
	due := int64(9)
	after := model.RowsAfter("main.due_date", &due, "task-2")
	if !strings.HasSuffix(ws.lastPage.SQL, after.SQL) || !slices.Equal(ws.lastPage.Args[len(ws.lastPage.Args)-3:], after.Args) {
		t.Errorf("next page read with %+v, want it to start after task-2", ws.lastPage)
	}
	if next := page.Sections[0]; len(next.Tasks) != 1 || next.Next != "" {
		t.Errorf("next page = %+v, want the last task and no page after it", next)
	}

	if _, appErr := a.GetAssignedToMe(context.Background(), "u1", "UTC", "overdue", "not a cursor", 2); appErr == nil || appErr.Status != 400 {
		t.Errorf("error = %v, want a bad request for a cursor no page returned", appErr)
	}
}
