// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

func insertSavedFilter(t *testing.T, id, workspaceID, tableID, viewID, userID string, private, active bool, createdAt int64, settings string) {
	t.Helper()
	mustExec(t, `INSERT INTO workspace_filters (id, workspace_id, table_id, filter_settings, created_at, updated_at, name, is_private, created_by)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)`, "f-"+id, workspaceID, tableID, settings, createdAt, "filter "+id, private, userID)
	mustExec(t, `INSERT INTO workspace_saved_filters (id, workspace_id, table_id, view_id, filter_id, user_id, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)`, id, workspaceID, tableID, viewID, "f-"+id, userID, active, createdAt)
}

func TestGetTableFiltersReadsStoredCreationTime(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_filters", "workspace_saved_filters")
	insertSavedFilter(t, "s1", "w1", "t1", "v1", "u1", false, true, 1727000000, `{"flatFilters":[]}`)

	filters, err := (&workspaceRepository{Db: testDB}).GetTableFilters("w1", "t1", "u1")
	if err != nil {
		t.Fatalf("GetTableFilters: %v", err)
	}
	if len(filters) != 1 || filters[0].CreatedAt != 1727000000 {
		t.Fatalf("filters = %+v, want one created at 1727000000", filters)
	}
}

func nameFilter(value string) string {
	return `{"flatFilters":[{"field":"name","operator":"is","value":"` + value + `"}]}`
}

func setupViewMetadata(t *testing.T) {
	t.Helper()
	requireDB(t)
	cleanTables(t, "workspace_view", "workspace_sort", "workspace_filters", "workspace_saved_filters")

	for _, v := range []struct {
		id, workspace, table, viewType string
		main                           bool
	}{
		{"v1", "w1", "t1", "kanban", false},
		{"v2", "w1", "t1", "grid", false},
		{"v3", "w1", "t1", "kanban", false},
		{"g1", "w1", "t1", "grid", true},
		{"v4", "w1", "t2", "list", false},
		{"v9", "w1", "t2", "form", false},
		{"g2", "w1", "t2", "grid", true},
		{"v8", "w2", "t1", "kanban", false},
	} {
		mustExec(t, `INSERT INTO workspace_view (id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, main_view, created_at, updated_at, deleted_at, is_public)
			VALUES (?, ?, ?, ?, '', ?, '[]', 'u1', ?, 0, 0, 0, 1)`, v.id, v.workspace, v.table, v.viewType, "view "+v.id, v.main)
	}

	insertSavedFilter(t, "s1", "w1", "t1", "v1", "u1", false, true, 100, nameFilter("older"))
	insertSavedFilter(t, "s2", "w1", "t1", "v1", "u1", false, true, 200, nameFilter("newest active"))
	insertSavedFilter(t, "s3", "w1", "t1", "v1", "u1", false, false, 300, nameFilter("inactive"))
	insertSavedFilter(t, "s4", "w1", "t1", "v1", "u2", false, true, 400, nameFilter("other user"))
	insertSavedFilter(t, "s5", "w1", "t1", "g1", "u1", false, true, 50, nameFilter("grid"))
	insertSavedFilter(t, "s6", "w1", "t1", "v2", "u2", true, true, 500, nameFilter("private"))
	insertSavedFilter(t, "s7", "w1", "t2", "v4", "u1", false, true, 10, nameFilter("second table"))

	mustExec(t, `INSERT INTO workspace_sort (id, workspace_id, table_id, view_id, sort_settings, created_at, updated_at) VALUES
		('o1', 'w1', 't1', 'v2', '[{"field":"name","direction":"asc"}]', 0, 0),
		('o2', 'w1', 't1', 'g1', '[{"field":"due_date","direction":"desc"}]', 0, 0)`)
}

func describeViews(views []model.WorkspaceView) []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		s := v.ID
		if v.Filter != nil && len(v.Filter.FlatFilters) > 0 {
			s += " filter=" + v.Filter.FlatFilters[0].Value
		}
		if len(v.Sort) > 0 {
			s += fmt.Sprintf(" sort=%v", v.Sort[0]["field"])
		}
		out = append(out, s)
	}
	return out
}

func savedFilterIDs(filters []model.SavedFilterWithPayload) []string {
	ids := make([]string, 0, len(filters))
	for _, f := range filters {
		ids = append(ids, f.ID)
	}
	return ids
}

func TestGetTableViewMetadata(t *testing.T) {
	setupViewMetadata(t)

	got, err := (&workspaceRepository{Db: testDB}).GetTableViewMetadata("w1", "u1", []string{"t1", "t2", "t-empty"})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := map[string]struct{ views, grid, saved []string }{
		"t1": {
			views: []string{"v2 sort=name", "v1 filter=newest active", "v3"},
			grid:  []string{"g1 filter=grid sort=due_date"},
			saved: []string{"s4", "s3", "s2", "s1", "s5"},
		},
		"t2": {
			views: []string{"v4 filter=second table"},
			grid:  []string{"g2"},
			saved: []string{"s7"},
		},
		"t-empty": {views: []string{}, grid: []string{}, saved: []string{}},
	}
	for tableID, w := range want {
		m := got[tableID]
		if m == nil {
			t.Errorf("%s: no metadata", tableID)
			continue
		}
		if m.Views == nil || m.GridSettings == nil {
			t.Errorf("%s: views or grid settings are nil, want empty lists", tableID)
		}
		if g := describeViews(m.Views); !slices.Equal(g, w.views) {
			t.Errorf("%s views = %q, want %q", tableID, g, w.views)
		}
		if g := describeViews(m.GridSettings); !slices.Equal(g, w.grid) {
			t.Errorf("%s grid settings = %q, want %q", tableID, g, w.grid)
		}
		if g := savedFilterIDs(m.SavedFilters); !slices.Equal(g, w.saved) {
			t.Errorf("%s saved filters = %q, want %q", tableID, g, w.saved)
		}
		for _, f := range m.SavedFilters {
			if f.TableID != tableID || f.UserID != "u1" {
				t.Errorf("%s saved filter %s has table %q user %q", tableID, f.ID, f.TableID, f.UserID)
			}
		}
	}
}

func TestGetViewVisibility(t *testing.T) {
	setupViewMetadata(t)
	mustExec(t, "UPDATE workspace_view SET is_public = 0, created_by = 'owner' WHERE id = 'v2'")
	w := &workspaceRepository{Db: testDB}

	if public, owner, err := w.GetViewVisibility(context.Background(), "v2"); err != nil || public || owner != "owner" {
		t.Errorf("private view = %v, %q, %v; want false, owner", public, owner, err)
	}

	if public, _, err := w.GetViewVisibility(context.Background(), "v1"); err != nil || !public {
		t.Errorf("public view = %v, %v; want true", public, err)
	}

	if _, _, err := w.GetViewVisibility(context.Background(), "missing"); err == nil {
		t.Error("a missing view reported no error")
	}
}

func TestGetView(t *testing.T) {
	setupViewMetadata(t)
	mustExec(t, "UPDATE workspace_view SET is_public = 0 WHERE id = 'v2'")
	w := &workspaceRepository{Db: testDB}

	v, err := w.GetView(context.Background(), "v2")
	if err != nil {
		t.Fatalf("GetView: %v", err)
	}

	if v.ID != "v2" || v.WorkspaceID != "w1" || v.TableID != "t1" || v.ViewType != "grid" || v.Name != "view v2" || v.CreatedBy != "u1" || v.IsPublic || v.MainView {
		t.Errorf("view = %+v", v)
	}
	if main, err := w.GetView(context.Background(), "g1"); err != nil || !main.MainView {
		t.Errorf("main view = %+v, %v; want main_view set", main, err)
	}

	if _, err := w.GetView(context.Background(), "missing"); err == nil {
		t.Error("a missing view reported no error")
	}
}

func TestGetMainViewIDs(t *testing.T) {
	setupViewMetadata(t)
	w := &workspaceRepository{Db: testDB}
	tables := map[string]string{"t1": "w1", "t2": "w1", "t-empty": "w1", "t-moved": "w9"}
	mustExec(t, `INSERT INTO workspace_view (id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, main_view, created_at, updated_at, deleted_at, is_public)
		VALUES ('g9', 'w1', 't-moved', 'grid', '', 'main', '[]', 'u1', 1, 0, 0, 0, 1)`)

	got, err := w.GetMainViewIDs(tables)
	if err != nil {
		t.Fatalf("GetMainViewIDs: %v", err)
	}
	for tableID, workspaceID := range tables {
		want, err := w.GetMainViewByTableID(workspaceID, tableID)
		if err != nil {
			t.Fatalf("GetMainViewByTableID(%s): %v", tableID, err)
		}
		if got[tableID] != want {
			t.Errorf("%s: main view = %q, want %q as GetMainViewByTableID gives", tableID, got[tableID], want)
		}
	}
	if got["t1"] != "g1" || got["t2"] != "g2" {
		t.Errorf("main views = %v, want t1:g1 t2:g2", got)
	}
}

func TestViewMetadataQueryCountDoesNotGrowWithTables(t *testing.T) {
	setupViewMetadata(t)
	db := pinnedConn(t)
	repo := &workspaceRepository{Db: db}

	count := func(tableIDs []string) int {
		before := queriesRun(t, db)
		if _, err := repo.GetTableViewMetadata("w1", "u1", tableIDs); err != nil {
			t.Fatalf("load: %v", err)
		}
		return queriesRun(t, db) - before - 1
	}

	one, all := count([]string{"t2"}), count([]string{"t1", "t2", "t-empty"})
	if one != all {
		t.Errorf("one table took %d queries and three took %d, want the same", one, all)
	}
}
