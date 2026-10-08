// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeTaskWriteStore struct {
	fakeReadAccessStore
	columns map[string]bool
	members map[string]bool
	options map[string]bool
	writes  int
	written interface{}
	created map[string]any
	field   *model.CreateFieldParams
	tables  []createdTable
}

type createdTable struct{ id, name, physical string }

func (f *fakeTaskWriteStore) CreateTable(tableID, _, name, safeName string, _ bool, _, _ string, _, _ bool, _, _, _ string,
	_ []model.KanbanStatusOption, _, _, _ string) (*model.WorkspaceTable, error) {
	f.tables = append(f.tables, createdTable{tableID, name, safeName})
	return &model.WorkspaceTable{ID: tableID, Name: safeName}, nil
}

func (f *fakeTaskWriteStore) CreateTableField(p model.CreateFieldParams) (*model.WorkspaceHeaders, error) {
	f.field = &p
	return &model.WorkspaceHeaders{ID: p.MainFieldID, Name: p.FieldName, DisplayName: p.FieldNameDisplay}, nil
}

func (f *fakeTaskWriteStore) UpdateTaskOrderGridField(_, _ string, _ model.WorkspaceHeaders) error {
	return nil
}

func (f *fakeTaskWriteStore) IsOptionTableOf(_ context.Context, _, optionTableID string) (bool, error) {
	return f.options[optionTableID], nil
}

func (f *fakeTaskWriteStore) GetTableName(string) (string, error) { return "pre_tasks", nil }

func (f *fakeTaskWriteStore) TableHasColumn(_, column string) (bool, error) {
	return f.columns[column], nil
}

func (f *fakeTaskWriteStore) GetColumnNames(context.Context, string) ([]string, error) {
	var names []string
	for name := range f.columns {
		names = append(names, name)
	}
	return names, nil
}

func (f *fakeTaskWriteStore) GetMemberUserIDs(_ context.Context, _ string, ids []string) ([]string, error) {
	var out []string
	for _, id := range ids {
		if f.members[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

// CreateMemberToTask counts the write and then fails, so a test sees whether
// the checks let an assignment through without faking what follows it.
func (f *fakeTaskWriteStore) CreateMemberToTask(_, _, _, _, _ string) error {
	f.writes++
	return errors.New("stop after the write")
}

func (f *fakeTaskWriteStore) GetTaskAssigneeID(string, string) (string, error) { return "", nil }

func (f *fakeTaskWriteStore) GetPersonFieldNamesForTable(string) ([]string, error) {
	return []string{"reviewer"}, nil
}

func (f *fakeTaskWriteStore) GetTaskFieldValue(_, taskID, _ string) (string, error) {
	if taskID == "missing" {
		return "", sql.ErrNoRows
	}
	return "a task", nil
}

func (f *fakeTaskWriteStore) CreateTask(_, _, _, _, _, _ string, _ int64, fields map[string]any) (*map[string]interface{}, error) {
	f.writes++
	f.created = fields
	return nil, errors.New("stop after the write")
}

func (f *fakeTaskWriteStore) GetTableWorkspaceID(_ context.Context, tableID string) (string, error) {
	if tableID == "t-elsewhere" {
		return "other-ws", nil
	}
	return "ws", nil
}

func (f *fakeTaskWriteStore) GetLinkedTableID(_, field string) (string, error) {
	switch field {
	case "priority":
		return "t-priority", nil
	case "status":
		return "t-status", nil
	}
	return "", nil
}

func (f *fakeTaskWriteStore) IsTableSingleSelect(tableID string) (bool, error) {
	return tableID == "t-priority", nil
}

func (f *fakeTaskWriteStore) GetSingleSelectOptions(_ context.Context, _ string, ids []string) (map[string]*model.TaskOrderField, error) {
	out := map[string]*model.TaskOrderField{}
	for _, id := range ids {
		if f.options[id] {
			out[id] = &model.TaskOrderField{ID: id}
		}
	}
	return out, nil
}

func (f *fakeTaskWriteStore) UpdateTaskTx(_, _, _, _, _ string, value interface{}, _, _, _ bool) (*map[string]interface{}, error) {
	f.writes++
	f.written = value
	return &map[string]interface{}{}, nil
}

func (f *fakeTaskWriteStore) GetFieldType(_ context.Context, _, field string) (string, error) {
	switch field {
	case "website":
		return "url", nil
	case "done":
		return "bool", nil
	}
	return "", nil
}

func (f *fakeTaskWriteStore) GetTableMeta(string) (string, bool, error) {
	return "", false, errors.New("not needed")
}

func taskWriteApp() (*App, *fakeTaskWriteStore) {
	ws := &fakeTaskWriteStore{
		columns: map[string]bool{"id": true, "name": true, "assignee": true, "created_by": true, "deleted_at": true, "priority": true, "updated_at": true,
			"status": true, "start_date": true, "due_date": true, "description": true, "reviewer": true, "website": true, "done": true},
		members: map[string]bool{"member": true},
		options: map[string]bool{"high": true, "todo": true},
	}
	ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	return a, ws
}

func TestTaskUpdatesWriteOnlyTheTablesOwnFields(t *testing.T) {
	admin := workspaceAdmin

	for _, tc := range []struct {
		name, field, value string
		want               int
	}{
		{"a field of the table", "name", "Renamed", http.StatusOK},
		{"a column the table does not have", "name` = (SELECT 1), `name", "x", http.StatusBadRequest},
		{"the id", "id", "other", http.StatusBadRequest},
		{"the id spelled in capitals", "ID", "other", http.StatusBadRequest},
		{"the name spelled in capitals", "NAME", "Renamed", http.StatusBadRequest},
		{"who created it", "created_by", "someone", http.StatusBadRequest},
		{"its deletion", "deleted_at", "1", http.StatusBadRequest},
		{"an option of the field", "priority", "high", http.StatusOK},
		{"clearing the option", "priority", "", http.StatusOK},
		{"an option the field does not have", "priority", "made-up", http.StatusBadRequest},
		{"the assignee, to a member", "assignee", "member", http.StatusOK},
		{"the assignee, to someone who is not a member", "assignee", "stranger", http.StatusBadRequest},
		{"clearing the assignee", "assignee", "", http.StatusOK},
		{"a person field, to a member", "reviewer", "member", http.StatusOK},
		{"a person field, to someone who is not a member", "reviewer", "stranger", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", tc.field, tc.value, admin, "")

			switch {
			case tc.want == http.StatusOK && appErr != nil:
				t.Fatalf("error = %v, want the update", appErr)
			case tc.want != http.StatusOK && (appErr == nil || appErr.Status != tc.want):
				t.Fatalf("error = %v, want status %d", appErr, tc.want)
			}
			if wrote := ws.writes > 0; wrote != (tc.want == http.StatusOK) {
				t.Errorf("wrote %d times", ws.writes)
			}
		})
	}
}

func TestAValueIsStoredAsTypedExceptACheckbox(t *testing.T) {
	admin := workspaceAdmin

	for _, tc := range []struct {
		field, value string
		want         interface{}
	}{
		{"name", "007", "007"},
		{"description", "true", "true"},
		{"done", "true", true},
		{"done", "false", false},
		{"done", "1", "1"},
		{"description", "", nil},
	} {
		a, ws := taskWriteApp()
		if _, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", tc.field, tc.value, admin, ""); appErr != nil {
			t.Fatalf("%s = %q: %v", tc.field, tc.value, appErr)
		}

		if ws.written != tc.want {
			t.Errorf("%s = %q stored %#v, want %#v", tc.field, tc.value, ws.written, tc.want)
		}
	}
}

func TestALinkFieldRefusesLinksThatRunCode(t *testing.T) {
	admin := workspaceAdmin

	for link, ok := range map[string]bool{
		"https://example.com":        true,
		"http://example.com/a?b=c":   true,
		"mailto:someone@example.com": true,
		"example.com/page":           true,
		"javascript:alert(1)":        false,
		" JavaScript:alert(1)":       false,
		"java\tscript:alert(1)":      false,
		"data:text/html,<b>x</b>":    false,
		"vbscript:msgbox(1)":         false,
	} {
		a, ws := taskWriteApp()
		_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", "website", link, admin, "")

		if ok && appErr != nil {
			t.Errorf("%q: %v, want it stored", link, appErr)
		}
		if !ok && (appErr == nil || appErr.Status != http.StatusBadRequest || ws.writes != 0) {
			t.Errorf("%q: error %v after %d writes, want it refused", link, appErr, ws.writes)
		}
	}
}

func TestAssigningNeedsAnAssigneeFieldAndAMember(t *testing.T) {
	admin := workspaceAdmin

	for _, tc := range []struct {
		name, field, user string
		ok                bool
	}{
		{"a member", "assignee", "member", true},
		{"nobody", "assignee", "", true},
		{"someone outside the workspace", "assignee", "stranger", false},
		{"a column the table does not have", "assignee` = 'x', `name", "member", false},
		{"who created it", "created_by", "member", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			_, appErr := a.AddMemberToTask(context.Background(), "ws", "t-a", "task-1", tc.user, tc.field, admin)

			if tc.ok && ws.writes != 1 {
				t.Fatalf("error = %v after %d writes, want the assignment written", appErr, ws.writes)
			}
			if !tc.ok && (appErr == nil || appErr.Status != http.StatusBadRequest || ws.writes != 0) {
				t.Fatalf("error = %v after %d writes, want a bad request and no write", appErr, ws.writes)
			}
		})
	}
}

func TestCreatingATaskChecksItsParentAndSection(t *testing.T) {
	admin := workspaceAdmin

	for _, tc := range []struct {
		name, parent, option, section string
		ok                            bool
	}{
		{"a plain task", "", "", "", true},
		{"a subtask", "task-1", "", "", true},
		{"a subtask of a task not in the table", "missing", "", "", false},
		{"into a column", "", "high", "priority", true},
		{"into Unassigned", "", "0", "priority", true},
		{"into an option the field does not have", "", "made-up", "priority", false},
		{"into a column that is not a field", "", "high", "priority` = 'x', `name", false},
		{"into a column the server keeps", "", "high", "created_by", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			_, appErr := a.CreateWorkspaceTask(context.Background(), "ws", "t-a", "New task", tc.option, tc.section, model.NewTaskFields{}, admin, tc.parent)

			if tc.ok && ws.writes != 1 {
				t.Fatalf("error = %v after %d writes, want the task created", appErr, ws.writes)
			}
			if !tc.ok && (appErr == nil || appErr.Status != http.StatusBadRequest || ws.writes != 0) {
				t.Fatalf("error = %v after %d writes, want a bad request and no write", appErr, ws.writes)
			}
		})
	}
}

func TestCreatingATaskWithFieldsChecksEachOne(t *testing.T) {
	admin := workspaceAdmin
	all := model.NewTaskFields{Status: "todo", Assignee: "member", StartDate: 100, DueDate: 200, Description: "Notes"}

	for _, tc := range []struct {
		name            string
		fields          model.NewTaskFields
		option, section string
		want            int
	}{
		{"every field", all, "", "", http.StatusOK},
		{"with a Kanban column of another field", all, "high", "priority", http.StatusOK},
		{"a status the table does not have", model.NewTaskFields{Status: "made-up"}, "", "", http.StatusBadRequest},
		{"a status unlike the column it was added in", model.NewTaskFields{Status: "todo"}, "high", "status", http.StatusBadRequest},
		{"someone who is not a member", model.NewTaskFields{Assignee: "stranger"}, "", "", http.StatusBadRequest},
		{"a start after the due date", model.NewTaskFields{StartDate: 300, DueDate: 200}, "", "", http.StatusBadRequest},
		{"a negative date", model.NewTaskFields{DueDate: -1}, "", "", http.StatusBadRequest},
		{"a description too long for the column", model.NewTaskFields{Description: string(make([]byte, maxTaskDescription+1))}, "", "", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			_, appErr := a.CreateWorkspaceTask(context.Background(), "ws", "t-a", "New task", tc.option, tc.section, tc.fields, admin, "")

			if tc.want != http.StatusOK {
				if appErr == nil || appErr.Status != tc.want || ws.writes != 0 {
					t.Fatalf("error = %v after %d writes, want status %d and no write", appErr, ws.writes, tc.want)
				}
				return
			}
			if ws.writes != 1 {
				t.Fatalf("error = %v after %d writes, want the task created", appErr, ws.writes)
			}
			want := map[string]any{"status": "todo", "assignee": "member", "start_date": int64(100), "due_date": int64(200), "description": "Notes"}
			if tc.section != "" {
				want[tc.section] = tc.option
			}
			for column, value := range want {
				if ws.created[column] != value {
					t.Errorf("%s created as %v, want %v", column, ws.created[column], value)
				}
			}
			if len(ws.created) != len(want) {
				t.Errorf("created with %v, want exactly %v", ws.created, want)
			}
		})
	}

	a, ws := taskWriteApp()
	if _, appErr := a.CreateWorkspaceTask(context.Background(), "ws", "t-a", "   ", "", "", all, admin, ""); appErr == nil || ws.writes != 0 {
		t.Errorf("a blank name gave %v after %d writes, want it refused", appErr, ws.writes)
	}
}

func TestANewFieldCannotReachOtherTables(t *testing.T) {
	admin := workspaceAdmin

	for _, tc := range []struct {
		name, linkedTable, selectedType, display, secondDisplay string
		both                                                    bool
		want                                                    int
	}{
		{"a link to another workspace's table", "t-elsewhere", "link", "Related", "", false, http.StatusForbidden},
		{"an unknown field type", "", "anything", "Related", "", false, http.StatusBadRequest},
		{"no name", "t-b", "link", "   ", "", false, http.StatusBadRequest},
		{"a two-way link with no name on the other side", "t-b", "link", "Related", "", true, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			_, appErr := a.CreateWorkspaceTableField(context.Background(), "ws", "t-a", "link", tc.linkedTable, admin,
				tc.selectedType, tc.both, tc.display, tc.secondDisplay, "", nil)
			if appErr == nil || appErr.Status != tc.want {
				t.Fatalf("error = %v, want status %d", appErr, tc.want)
			}
			if ws.field != nil {
				t.Error("a refused field reached the store")
			}
		})
	}
}

func TestOnlyALinkFieldKeepsTheTableItPointsAt(t *testing.T) {
	admin := workspaceAdmin

	a, ws := taskWriteApp()
	if _, appErr := a.CreateWorkspaceTableField(context.Background(), "ws", "t-a", "single select", "t-b", admin,
		"single select", false, "Priority", "", "", nil); appErr != nil {
		t.Fatal(appErr)
	}

	if ws.field == nil || ws.field.LinkedTableID != "" {
		t.Errorf("a single select was created pointing at %v, want no table", ws.field)
	}
}

func TestANewTableIsNamedFromItsIDNotItsName(t *testing.T) {
	admin := workspaceAdmin
	generated := regexp.MustCompile(`^t[0-9a-f]{32}$`)
	name := "Client `Projects`; DROP TABLE users; --"

	a, ws := taskWriteApp()
	if _, appErr := a.CreateWorkspaceTable(context.Background(), "ws", name, false, "", admin, false, false, "", "", true); appErr != nil {
		t.Fatal(appErr)
	}

	if len(ws.tables) != 2 {
		t.Fatalf("created %d tables, want the table and its subtasks", len(ws.tables))
	}
	for _, table := range ws.tables {
		if !generated.MatchString(table.physical) || table.physical != physicalName("t", table.id) {
			t.Errorf("table %q is named %q, not from its id", table.name, table.physical)
		}
	}
	if ws.tables[0].name != name || ws.tables[1].name != name+"_subtasks" {
		t.Errorf("names = %q and %q, want what was typed", ws.tables[0].name, ws.tables[1].name)
	}
}

func TestANewFieldsColumnsAreNamedFromItsIDNotItsName(t *testing.T) {
	admin := workspaceAdmin
	generated := regexp.MustCompile(`^c[0-9a-f]{32}$`)

	for _, name := range []string{
		"Budget (€)",
		"x` TEXT, DROP COLUMN `created_by",
		"'; DROP TABLE users; --",
		"Pārdošana 2026",
	} {
		t.Run(name, func(t *testing.T) {
			a, ws := taskWriteApp()
			if _, appErr := a.CreateWorkspaceTableField(context.Background(), "ws", "t-a", "link", "t-b", admin,
				"link", true, "  "+name+"  ", name, "", nil); appErr != nil {
				t.Fatal(appErr)
			}

			p := ws.field
			if p.FieldNameDisplay != name || p.FieldNameInSecondTableDisplay != name {
				t.Errorf("labels = %q and %q, want %q", p.FieldNameDisplay, p.FieldNameInSecondTableDisplay, name)
			}
			if !generated.MatchString(p.FieldName) || !generated.MatchString(p.FieldNameInSecondTable) {
				t.Errorf("columns %q and %q are not generated names", p.FieldName, p.FieldNameInSecondTable)
			}
			if p.FieldName != physicalName("c", p.MainFieldID) || p.FieldNameInSecondTable != physicalName("c", p.SecondFieldID) {
				t.Errorf("columns %q and %q are not named from the fields' ids %q and %q",
					p.FieldName, p.FieldNameInSecondTable, p.MainFieldID, p.SecondFieldID)
			}
		})
	}
}

func TestRenamingAnOptionStaysInTheWorkspace(t *testing.T) {
	admin := workspaceAdmin
	a, _ := taskWriteApp()

	_, appErr := a.UpdateWorkspaceSingleSelectName(context.Background(), "ws", "t-a", "option-1", "name", "Renamed", admin, "t-elsewhere")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("renaming in another workspace's table gave %v, want forbidden", appErr)
	}

	_, appErr = a.UpdateWorkspaceSingleSelectName(context.Background(), "ws", "t-a", "task-1", "name", "Renamed", admin, "t-b")
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("renaming a row of a task table gave %v, want a bad request", appErr)
	}
}

func TestAnOptionIsOnlyDeletedFromTheTablesOwnOptions(t *testing.T) {
	admin := workspaceAdmin
	a, _ := taskWriteApp()

	_, appErr := a.DeleteWorkspaceTableSingleField(context.Background(), "ws", "t-a", "row", "someone-elses-table", admin)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("error = %v, want forbidden before anything is deleted", appErr)
	}
}

func TestASortOnlyOrdersByTheTablesColumns(t *testing.T) {
	known := func(name string) bool { return name == "due_date" || name == "c8b21e4f0a9d" }

	for _, tc := range []struct {
		name string
		sort []model.SortParam
		want string
	}{
		{"a column of the table", []model.SortParam{{Field: "due_date", Direction: "desc"}}, "ORDER BY main.`due_date` DESC, main.id DESC"},
		{"the id follows the last field", []model.SortParam{{Field: "due_date", Direction: "desc"}, {Field: "c8b21e4f0a9d"}}, "ORDER BY main.`due_date` DESC, main.`c8b21e4f0a9d` ASC, main.id ASC"},
		{"a generated column", []model.SortParam{{Field: "c8b21e4f0a9d"}}, "ORDER BY main.`c8b21e4f0a9d` ASC, main.id ASC"},
		{"a column the table does not have", []model.SortParam{{Field: "created_by"}}, "ORDER BY main.id ASC"},
		{"a name that is not a name", []model.SortParam{{Field: "due_date` DESC, (SELECT 1) --"}}, "ORDER BY main.id ASC"},
		{"a mix", []model.SortParam{{Field: "nope"}, {Field: "due_date"}}, "ORDER BY main.`due_date` ASC, main.id ASC"},
	} {
		if got := buildSortSQL(tc.sort, known); got != tc.want {
			t.Errorf("%s: buildSortSQL = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestAGridWithoutASortShowsTheNewestFirst(t *testing.T) {
	withCreated := func(name string) bool { return name == "created_at" || name == "due_date" }
	without := func(name string) bool { return name == "due_date" }

	if got, want := gridSortSQL(nil, withCreated), "ORDER BY main.`created_at` DESC, main.id DESC"; got != want {
		t.Errorf("no sort = %q, want %q", got, want)
	}
	if got, want := gridSortSQL([]model.SortParam{{Field: "nope"}}, withCreated), "ORDER BY main.`created_at` DESC, main.id DESC"; got != want {
		t.Errorf("only an unknown field = %q, want %q", got, want)
	}
	if got, want := gridSortSQL([]model.SortParam{{Field: "due_date"}}, withCreated), "ORDER BY main.`due_date` ASC, main.id ASC"; got != want {
		t.Errorf("a chosen sort = %q, want %q", got, want)
	}
	if got, want := gridSortSQL(nil, without), "ORDER BY main.id ASC"; got != want {
		t.Errorf("a table without created_at = %q, want %q", got, want)
	}
}
