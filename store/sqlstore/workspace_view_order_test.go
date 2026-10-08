// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestChangeViewOrderLetsConcurrentChangesAllLand(t *testing.T) {
	requireDB(t)
	cleanTables(t, "workspace_view")
	mustExec(t, `INSERT INTO workspace_view (id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, main_view, created_at, updated_at, deleted_at, is_public)
		VALUES ('v1', 'w1', 't1', 'kanban', '', 'Board', '', 'u1', 0, 0, 0, 0, 1)`)
	w := &workspaceRepository{Db: testDB}
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := w.ChangeViewOrder(ctx, "v1", "w1", "t1", func(order string) (string, error) {
				return order + fmt.Sprintf("[%d]", i), nil
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	var order string
	if err := testDB.QueryRow("SELECT item_order FROM workspace_view WHERE id = 'v1'").Scan(&order); err != nil {
		t.Fatal(err)
	}

	for i := range 20 {
		if !strings.Contains(order, fmt.Sprintf("[%d]", i)) {
			t.Errorf("change %d was overwritten: %s", i, order)
		}
	}

	if _, err := w.ChangeViewOrder(ctx, "v1", "other", "t1", func(order string) (string, error) { return "x", nil }); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("a view of another workspace: %v, want no rows", err)
	}
	if _, err := w.ChangeViewOrder(ctx, "v1", "w1", "t1", func(string) (string, error) { return "", errors.New("refused") }); err == nil {
		t.Error("a failed change was reported as done")
	}
	var after string
	_ = testDB.QueryRow("SELECT item_order FROM workspace_view WHERE id = 'v1'").Scan(&after)
	if after != order {
		t.Error("a failed change still wrote the order")
	}
}

func TestKanbanRestIDsFollowTheBoardsOrderUpToTheCard(t *testing.T) {
	requireDB(t)
	createDynamicTable(t, "zz_test_kanban_rest", createTaskTable)
	for _, r := range []struct {
		id, status string
		created    int
		deleted    int
	}{
		{"old", "s1", 1, 0}, {"same-b", "s1", 2, 0}, {"same-a", "s1", 2, 0}, {"gone", "s1", 3, 1},
		{"new", "s1", 4, 0}, {"elsewhere", "s2", 1, 0}, {"newest", "s1", 5, 0},
	} {
		mustExec(t, "INSERT INTO zz_test_kanban_rest (id, name, status, created_at, deleted_at) VALUES (?, ?, ?, ?, ?)", r.id, r.id, r.status, r.created, r.deleted)
	}
	w := &workspaceRepository{Db: testDB}
	inS1 := model.SQLFilter{SQL: "main.status = ?", Args: []any{"s1"}}

	for _, tc := range []struct {
		name, upTo string
		limit      int
		want       []string
	}{
		{"up to a card", "new", 100, []string{"old", "same-a", "same-b", "new"}},
		{"up to the first", "old", 100, []string{"old"}},
		{"a card not in the column", "elsewhere", 100, []string{"old", "same-a", "same-b", "new", "newest"}},
		{"at most the limit", "newest", 2, []string{"old", "same-a"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := w.GetKanbanRestIDs(context.Background(), "zz_test_kanban_rest", inS1, tc.upTo, tc.limit)
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(ids, tc.want) {
				t.Errorf("ids = %v, want %v", ids, tc.want)
			}
		})
	}
}
