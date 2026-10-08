// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeTaskReportWorkspaceStore struct {
	store.WorkspaceStore
	saved []model.SavedFilterWithPayload
}

func (f *fakeTaskReportWorkspaceStore) GetTableMetas([]string) ([]model.WorkspaceTable, error) {
	return nil, nil
}

func (f *fakeTaskReportWorkspaceStore) GetTableFilters(workspaceID, tableID, _ string) ([]model.SavedFilterWithPayload, error) {
	if workspaceID != "all" || tableID != "detailed-task-report" {
		return nil, nil
	}
	return f.saved, nil
}

func savedReportFilter(id, filterID, settings string) model.SavedFilterWithPayload {
	return model.SavedFilterWithPayload{SavedFilterInfo: model.SavedFilterInfo{ID: id, FilterID: filterID, FilterSettings: settings}}
}

type fakeReportAccessStore struct {
	store.WorkspaceStore
	metaCalls int
}

func (f *fakeReportAccessStore) GetTableMetas([]string) ([]model.WorkspaceTable, error) {
	f.metaCalls++
	return nil, nil
}

func (f *fakeReportAccessStore) GetTableFilters(_, _, _ string) ([]model.SavedFilterWithPayload, error) {
	return nil, nil
}

func TestTaskReportIsForSystemAdminsOnly(t *testing.T) {
	ws := &fakeReportAccessStore{}
	a := &App{
		Store: store.Store{Workspace: ws},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
	}

	for _, ids := range [][]string{nil, {"someone-elses-workspace"}} {
		_, appErr := a.GetAllWorkspaceTasks(context.Background(), ids, model.User{ID: "u1", Role: model.SystemUserRoleId}, 1, 100, true, model.FilterPayload{})
		if appErr == nil || appErr.Status != http.StatusForbidden {
			t.Errorf("regular user with workspace_ids %v: got %v, want 403", ids, appErr)
		}
	}
	if ws.metaCalls != 0 {
		t.Errorf("tables were read %d times for a forbidden request", ws.metaCalls)
	}

	if _, appErr := a.GetAllWorkspaceTasks(context.Background(), nil, model.User{ID: "admin", Role: model.SystemAdminRoleId}, 1, 100, true, model.FilterPayload{}); appErr != nil {
		t.Errorf("system admin: %v", appErr)
	}
}

func TestTaskReportReturnsSavedFiltersWithoutMatchingTasks(t *testing.T) {
	ws := &fakeTaskReportWorkspaceStore{saved: []model.SavedFilterWithPayload{
		savedReportFilter("s1", "f1", `{"flatFilters":[{"field":"name","operator":"is","value":"a"}]}`),
		savedReportFilter("s2", "f1", `{"flatFilters":[]}`),
		savedReportFilter("s3", "f2", `not json`),
	}}
	a := &App{
		Store: store.Store{Workspace: ws},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
	}

	page, appErr := a.GetAllWorkspaceTasks(context.Background(), []string{"w1"}, model.User{ID: "u1", Role: model.SystemAdminRoleId}, 1, 100, true, model.FilterPayload{})
	if appErr != nil {
		t.Fatalf("GetAllWorkspaceTasks: %v", appErr)
	}
	if len(page.DataBase) != 0 {
		t.Fatalf("rows = %d, want none", len(page.DataBase))
	}
	if len(page.Filters) != 2 || page.Filters[0].ID != "s1" || page.Filters[1].ID != "s3" {
		t.Fatalf("filters = %+v, want s1 and s3, one per saved filter", page.Filters)
	}
	if page.Filters[0].Filters == nil || len(page.Filters[0].Filters.FlatFilters) != 1 {
		t.Errorf("s1 payload = %+v, want its one flat filter", page.Filters[0].Filters)
	}
	if page.Filters[1].Filters != nil {
		t.Errorf("s3 payload = %+v, want none for unreadable settings", page.Filters[1].Filters)
	}
}

type fakeReportCountStore struct {
	fakeTaskReportWorkspaceStore
	counts int
}

func (f *fakeReportCountStore) GetTableMetas([]string) ([]model.WorkspaceTable, error) {
	return []model.WorkspaceTable{{ID: "t1", Name: "tasks_t1", WorkspaceID: "w1"}}, nil
}

func (f *fakeReportCountStore) GetUserTimezone(string) (*model.UserTimezone, error) {
	return nil, nil
}

func (f *fakeReportCountStore) GetTableColumnTypes(string) ([]*sql.ColumnType, error) {
	return nil, nil
}

func (f *fakeReportCountStore) CountTasksAcrossTables(context.Context, []model.TaskSource, []model.SQLFilter) ([]int, error) {
	f.counts++
	return []int{3, 5}, nil
}

func (f *fakeReportCountStore) GetTaskPageKeys(context.Context, []model.TaskSource, []model.SortParam, int, int) ([]model.TaskKey, error) {
	return nil, nil
}

func (f *fakeReportCountStore) GetMainViewIDs(map[string]string) (map[string]string, error) {
	return nil, nil
}

func TestTaskReportCountsOnlyWhenAsked(t *testing.T) {
	ws := &fakeReportCountStore{}
	a := &App{
		Store: store.Store{Workspace: ws},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
	}
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}

	page, appErr := a.GetAllWorkspaceTasks(context.Background(), []string{"w1"}, admin, 1, 100, true, model.FilterPayload{})
	if appErr != nil {
		t.Fatalf("GetAllWorkspaceTasks: %v", appErr)
	}
	if page.Total == nil || *page.Total != 8 || page.RootTotal == nil || *page.RootTotal != 3 {
		t.Errorf("counted page totals = %v, %v; want 8 and 3", page.Total, page.RootTotal)
	}

	page, appErr = a.GetAllWorkspaceTasks(context.Background(), []string{"w1"}, admin, 2, 100, false, model.FilterPayload{})
	if appErr != nil {
		t.Fatalf("GetAllWorkspaceTasks: %v", appErr)
	}
	if page.Total != nil || page.RootTotal != nil {
		t.Errorf("uncounted page totals = %v, %v; want none", page.Total, page.RootTotal)
	}
	if ws.counts != 1 {
		t.Errorf("counted %d times, want once", ws.counts)
	}
}

func TestTaskReportFiltersComeWithoutTasks(t *testing.T) {
	ws := &fakeTaskReportWorkspaceStore{saved: []model.SavedFilterWithPayload{savedReportFilter("s1", "f1", `{"flatFilters":[]}`)}}
	a := &App{Store: store.Store{Workspace: ws}}

	if _, appErr := a.GetTaskReportFilters(model.User{ID: "u1", Role: model.SystemUserRoleId}); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("regular user: %v, want forbidden", appErr)
	}
	page, appErr := a.GetTaskReportFilters(model.User{ID: "admin", Role: model.SystemAdminRoleId})
	if appErr != nil {
		t.Fatal(appErr)
	}
	if len(page.Filters) != 1 || len(page.DataBase) != 0 || page.Total != nil {
		t.Errorf("page = %+v, want the saved filter, no tasks and no count", page)
	}
}
