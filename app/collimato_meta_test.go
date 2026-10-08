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

type fakeMetaCollimatoStore struct {
	store.CollimatoStore
	permissions []string
}

func (f *fakeMetaCollimatoStore) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	return &model.CollimatoWorkspace{ID: id}, nil
}

func (f *fakeMetaCollimatoStore) GetEffectiveRolesForUser(context.Context, string, string) ([]string, error) {
	return []string{"editor"}, nil
}

func (f *fakeMetaCollimatoStore) GetRolesByName([]string, string) ([]model.CollimatoRole, error) {
	return []model.CollimatoRole{{Name: "editor", Permissions: f.permissions}}, nil
}

type fakeMetaEngine struct {
	QueryEngine
}

func (fakeMetaEngine) Meta(context.Context, string) (*collimato.CubeCollection, error) {
	meta := map[string]any{"owner": "finance"}

	return &collimato.CubeCollection{Cubes: []collimato.Cube{{Name: "orders", Meta: meta}, {Name: "salaries"}}}, nil
}

type fakeMetaFilter struct {
	interfaces.CollimatoRoles
}

func (fakeMetaFilter) FilterMeta(_ []model.CollimatoRole, c collimato.CubeCollection) *collimato.CubeCollection {
	c.Cubes = c.Cubes[:1]
	return &c
}

func metaApp(permissions ...string) *App {
	return &App{
		Store:          store.Store{Collimato: &fakeMetaCollimatoStore{permissions: permissions}},
		QueryEngine:    fakeMetaEngine{},
		CollimatoRoles: fakeMetaFilter{},
	}
}

func TestRoleMetaNeedsARolePermission(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	if _, appErr := a.GetWorkspaceRoleMeta(context.Background(), model.User{ID: "u1"}, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without a role permission, got %v", appErr)
	}
}

func TestRoleMetaForViewersFollowsTheirOwnRoles(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewRoles.Id)

	meta, appErr := a.GetWorkspaceRoleMeta(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoleMeta: %v", appErr)
	}

	if len(meta.Cubes) != 1 || meta.Cubes[0].Meta != nil {
		t.Errorf("expected the filtered cubes without meta, got %v", meta.Cubes)
	}
}

func TestRoleMetaIsNotFilteredForRoleEditors(t *testing.T) {
	for _, permission := range []string{
		model.CollimatoPermissions.PermissionCreateRoles.Id,
		model.CollimatoPermissions.PermissionEditRoles.Id,
	} {
		a := metaApp(permission)

		meta, appErr := a.GetWorkspaceRoleMeta(context.Background(), model.User{ID: "u1"}, "ws")
		if appErr != nil {
			t.Fatalf("%s: %v", permission, appErr)
		}

		if len(meta.Cubes) != 2 {
			t.Errorf("%s: expected every cube, got %v", permission, meta.Cubes)
		}
	}
}

func TestQueryMetaStaysFiltered(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	meta, appErr := a.GetWorkspaceMeta(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceMeta: %v", appErr)
	}

	if len(meta.Cubes) != 1 {
		t.Errorf("expected the filtered cubes, got %v", meta.Cubes)
	}

	a = metaApp(model.CollimatoPermissions.PermissionViewRoles.Id)

	if _, appErr := a.GetWorkspaceMeta(context.Background(), model.User{ID: "u1"}, "ws"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without view charts or dashboards, got %v", appErr)
	}
}

func TestDashboardViewersCanQueryFilteredData(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewDashboards.Id)

	meta, appErr := a.GetWorkspaceMeta(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceMeta: %v", appErr)
	}

	if len(meta.Cubes) != 1 {
		t.Errorf("expected the filtered cubes, got %v", meta.Cubes)
	}

	if !a.canQueryWorkspaceData(model.User{ID: "u1"}, "ws") {
		t.Error("expected a dashboard viewer to load dashboard data")
	}
}

func TestQueryMetaHidesModelMetaWithoutDataModelAccess(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	meta, appErr := a.GetWorkspaceMeta(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceMeta: %v", appErr)
	}

	if meta.Cubes[0].Meta != nil {
		t.Errorf("expected the meta block hidden, got %v", meta.Cubes[0].Meta)
	}

	a = metaApp(
		model.CollimatoPermissions.PermissionViewCharts.Id,
		model.CollimatoPermissions.PermissionViewDataModels.Id,
	)

	meta, appErr = a.GetWorkspaceMeta(context.Background(), model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceMeta: %v", appErr)
	}

	if meta.Cubes[0].Meta == nil {
		t.Error("expected a data model reader to keep the meta block")
	}
}

func TestSQLPreviewNeedsDataModelAccess(t *testing.T) {
	a := metaApp(model.CollimatoPermissions.PermissionViewCharts.Id)

	_, appErr := a.GetWorkspaceSQL(context.Background(), model.User{ID: "u1"}, "ws", model.DataQuery{})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without view data models, got %v", appErr)
	}
}
