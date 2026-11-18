// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeFieldListCollimatoStore struct {
	store.CollimatoStore
	permissions []string
}

func (f *fakeFieldListCollimatoStore) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	return &model.CollimatoWorkspace{ID: id}, nil
}

func (f *fakeFieldListCollimatoStore) GetEffectiveRolesForUser(context.Context, string, string) ([]string, error) {
	return []string{"editor"}, nil
}

func (f *fakeFieldListCollimatoStore) GetRolesByName([]string, string) ([]model.CollimatoRole, error) {
	return []model.CollimatoRole{{Name: "editor", Permissions: f.permissions}}, nil
}

type fakeFieldListEngine struct {
	QueryEngine
}

func (fakeFieldListEngine) FieldList(context.Context, string) (*collimato.FieldList, error) {
	datasets := []collimato.DatasetFields{{Name: "orders"}, {Name: "salaries"}}
	modelErrors := []collimato.ModelError{{Name: "legacy.yml", Reason: collimato.UnsupportedWrittenByHand}}

	return &collimato.FieldList{Datasets: datasets, ModelErrors: modelErrors}, nil
}

type fakeFieldListFilter struct {
	interfaces.CollimatoRoles
}

func (fakeFieldListFilter) FilterFieldList(_ []model.CollimatoRole, c collimato.FieldList) *collimato.FieldList {
	c.Datasets = c.Datasets[:1]
	return &c
}

func fieldListApp(permissions ...string) *App {
	return &App{
		Store:          store.Store{Collimato: &fakeFieldListCollimatoStore{permissions: permissions}},
		QueryEngine:    fakeFieldListEngine{},
		CollimatoRoles: fakeFieldListFilter{},
	}
}

func TestRoleFieldListNeedsARolePermission(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	if _, appErr := a.GetWorkspaceRoleFieldList(context.Background(), model.User{ID: "u1"}, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without a role permission, got %v", appErr)
	}
}

func TestRoleFieldListForViewersFollowsTheirOwnRoles(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewRoles.Id)

	fieldList, appErr := a.GetWorkspaceRoleFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoleFieldList: %v", appErr)
	}

	if len(fieldList.Datasets) != 1 {
		t.Errorf("expected the filtered datasets, got %v", fieldList.Datasets)
	}
}

func TestRoleFieldListIsNotFilteredForRoleEditors(t *testing.T) {
	for _, permission := range []string{
		model.CollimatoPermissions.PermissionCreateRoles.Id,
		model.CollimatoPermissions.PermissionEditRoles.Id,
	} {
		a := fieldListApp(permission)

		fieldList, appErr := a.GetWorkspaceRoleFieldList(context.Background(), model.User{ID: "u1"}, "ws")
		if appErr != nil {
			t.Fatalf("%s: %v", permission, appErr)
		}

		if len(fieldList.Datasets) != 2 {
			t.Errorf("%s: expected every dataset, got %v", permission, fieldList.Datasets)
		}
	}
}

func TestFieldListStaysFiltered(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	fieldList, appErr := a.GetWorkspaceFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceFieldList: %v", appErr)
	}

	if len(fieldList.Datasets) != 1 {
		t.Errorf("expected the filtered datasets, got %v", fieldList.Datasets)
	}

	a = fieldListApp(model.CollimatoPermissions.PermissionViewRoles.Id)

	if _, appErr := a.GetWorkspaceFieldList(context.Background(), model.User{ID: "u1"}, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without view charts or dashboards, got %v", appErr)
	}
}

func TestDashboardViewersCanQueryFilteredData(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewDashboards.Id)

	fieldList, appErr := a.GetWorkspaceFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceFieldList: %v", appErr)
	}

	if len(fieldList.Datasets) != 1 {
		t.Errorf("expected the filtered datasets, got %v", fieldList.Datasets)
	}

	if !a.canQueryWorkspaceData(model.User{ID: "u1"}, "ws") {
		t.Error("expected a dashboard viewer to load dashboard data")
	}
}

func TestFieldListHidesModelErrorsWithoutDataModelAccess(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	fieldList, appErr := a.GetWorkspaceFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceFieldList: %v", appErr)
	}

	if fieldList.ModelErrors != nil {
		t.Errorf("expected the model errors hidden, got %v", fieldList.ModelErrors)
	}

	a = fieldListApp(
		model.CollimatoPermissions.PermissionViewCharts.Id,
		model.CollimatoPermissions.PermissionViewDataModels.Id,
	)

	fieldList, appErr = a.GetWorkspaceFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceFieldList: %v", appErr)
	}

	if len(fieldList.ModelErrors) != 1 {
		t.Errorf("expected a data model reader to see the model errors, got %v", fieldList.ModelErrors)
	}
}

func TestSQLPreviewNeedsDataModelAccess(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	_, appErr := a.GetWorkspaceSQL(context.Background(), model.User{ID: "u1"}, "ws", model.DataQuery{})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without view data models, got %v", appErr)
	}
}

func TestRoleFieldListHidesModelErrorsFromRoleEditorsWithoutDataModelAccess(t *testing.T) {
	a := fieldListApp(model.CollimatoPermissions.PermissionCreateRoles.Id)

	fieldList, appErr := a.GetWorkspaceRoleFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoleFieldList: %v", appErr)
	}

	if fieldList.ModelErrors != nil {
		t.Errorf("expected the model errors hidden, got %v", fieldList.ModelErrors)
	}

	a = fieldListApp(
		model.CollimatoPermissions.PermissionCreateRoles.Id,
		model.CollimatoPermissions.PermissionViewDataModels.Id,
	)

	fieldList, appErr = a.GetWorkspaceRoleFieldList(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoleFieldList: %v", appErr)
	}

	if len(fieldList.ModelErrors) != 1 {
		t.Errorf("expected a data model reader to see the model errors, got %v", fieldList.ModelErrors)
	}
}

func TestTheDatabasesOwnMessageOnlyForDataModelReaders(t *testing.T) {
	databaseError := &QueryError{DatabaseMessage: "Error 1054: Unknown column 'orders.salary' in 'field list'"}

	a := fieldListApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	appErr := a.queryAppError(model.User{ID: "u1"}, "ws", databaseError)
	if appErr.ID != "collimato.query_failed" {
		t.Errorf("want collimato.query_failed for a chart viewer, got %+v", appErr)
	}

	a = fieldListApp(
		model.CollimatoPermissions.PermissionViewCharts.Id,
		model.CollimatoPermissions.PermissionViewDataModels.Id,
	)

	appErr = a.queryAppError(model.User{ID: "u1"}, "ws", databaseError)
	if appErr.Message != databaseError.DatabaseMessage {
		t.Errorf("want the database's message for a data model reader, got %+v", appErr)
	}
}
