// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
	yaml "go.yaml.in/yaml/v3"
)

type generateStore struct {
	store.CollimatoStore
	conn         *model.Connection
	existing     map[string]bool
	saved        map[string]model.CollimatoWorkspaceFile
	roles        []model.CollimatoRole
	updatedRoles []model.CollimatoRole
}

func (g *generateStore) GetEffectiveRolesForUser(_ context.Context, _, _ string) ([]string, error) {
	return []string{"workspace_admin"}, nil
}

func (g *generateStore) GetRolesByName(_ []string, _ string) ([]model.CollimatoRole, error) {
	return []model.CollimatoRole{{
		Name:        "workspace_admin",
		Permissions: []string{model.CollimatoPermissions.PermissionCreateDataModels.Id},
	}}, nil
}

func (g *generateStore) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	return &model.CollimatoWorkspace{ID: id, Status: model.CollimatoWorkspaceStatusDraft}, nil
}

func (g *generateStore) GetConnectionByID(_ string) (*model.Connection, error) {
	return g.conn, nil
}

func (g *generateStore) GetWorkspaceFile(_, name, _ string) (*model.CollimatoWorkspaceFile, error) {
	if g.existing[name] {
		return &model.CollimatoWorkspaceFile{Name: name}, nil
	}
	return nil, nil
}

func (g *generateStore) UpdateWorkspaceFile(f model.CollimatoWorkspaceFile) (*model.CollimatoWorkspaceFile, error) {
	if g.saved == nil {
		g.saved = map[string]model.CollimatoWorkspaceFile{}
	}
	g.saved[f.Name] = f
	return &f, nil
}

func (g *generateStore) GetWorkspaceRoles(_ string) ([]model.CollimatoRole, error) {
	return g.roles, nil
}

func (g *generateStore) UpdateWorkspaceRole(_ string, role model.CollimatoRole) error {
	g.updatedRoles = append(g.updatedRoles, role)
	return nil
}

func generateApp(g *generateStore) *App {
	if g.conn == nil {
		g.conn = &model.Connection{ID: "conn1", WorkspaceID: "ws1", Database: "Analytics", Type: "mysql"}
	}
	return &App{Store: store.Store{Collimato: g}}
}

func payload(columns ...map[string]interface{}) map[string]interface{} {
	cols := make([]interface{}, 0, len(columns))
	for _, c := range columns {
		cols = append(cols, c)
	}
	return map[string]interface{}{
		"conn1": map[string]interface{}{"orders": cols},
	}
}

func col(name, sqlType string) map[string]interface{} {
	return map[string]interface{}{"name": name, "type": sqlType}
}

func TestGenerateCubeFilesWritesBuilderModels(t *testing.T) {
	g := &generateStore{}
	a := generateApp(g)

	result, appErr := a.GenerateCubeFiles(admin(), payload(
		col("id", "int"),
		col("title", "varchar(255)"),
	), "ws1")
	if appErr != nil {
		t.Fatalf("GenerateCubeFiles: %v", appErr)
	}

	file, ok := g.saved["orders.yml"]
	if !ok {
		t.Fatalf("no orders.yml written, got %v", g.saved)
	}

	if file.BuilderModel == nil {
		t.Fatal("BuilderModel is nil, the builder cannot reopen this cube")
	}

	var reopened model.Dataset
	if err := json.Unmarshal([]byte(*file.BuilderModel), &reopened); err != nil {
		t.Fatalf("BuilderModel is not a Dataset: %v", err)
	}
	if reopened.Name != "orders" || reopened.ConnectionID != "conn1" {
		t.Errorf("round trip lost data: %+v", reopened)
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(file.Content), &parsed); err != nil {
		t.Fatalf("content is not valid YAML: %v", err)
	}

	if len(result.Created) != 1 || result.Created[0] != "orders" {
		t.Errorf("created: got %v", result.Created)
	}
}

func TestGenerateCubeFilesSkipsExistingNames(t *testing.T) {
	g := &generateStore{existing: map[string]bool{"orders.yml": true}}
	a := generateApp(g)

	result, appErr := a.GenerateCubeFiles(admin(), payload(col("id", "int")), "ws1")
	if appErr != nil {
		t.Fatalf("GenerateCubeFiles: %v", appErr)
	}

	if len(result.Created) != 0 {
		t.Errorf("should not have overwritten: %v", result.Created)
	}
	if len(result.Skipped) != 1 || result.Skipped[0] != "orders" {
		t.Errorf("skipped: got %v", result.Skipped)
	}
	if _, wrote := g.saved["orders.yml"]; wrote {
		t.Error("an existing model was overwritten")
	}
}

func TestGenerateCubeFilesReportsSkippedColumns(t *testing.T) {
	g := &generateStore{}
	a := generateApp(g)

	result, appErr := a.GenerateCubeFiles(admin(), payload(
		col("id", "int"),
		col("payload", "json"),
	), "ws1")
	if appErr != nil {
		t.Fatalf("GenerateCubeFiles: %v", appErr)
	}

	if got := result.SkippedColumns["orders"]; len(got) != 1 || got[0] != "payload" {
		t.Errorf("skipped columns: got %v", result.SkippedColumns)
	}
	if len(result.Created) != 1 {
		t.Errorf("the table should still be created: %v", result.Created)
	}
}

func TestGenerateCubeFilesSkipsTableWithNoUsableColumns(t *testing.T) {
	g := &generateStore{}
	a := generateApp(g)

	result, appErr := a.GenerateCubeFiles(admin(), payload(col("data", "longblob")), "ws1")
	if appErr != nil {
		t.Fatalf("GenerateCubeFiles: %v", appErr)
	}

	if len(result.Created) != 0 || len(result.Skipped) != 1 {
		t.Errorf("created %v, skipped %v", result.Created, result.Skipped)
	}
}

func TestGenerateCubeFilesRejectsAnotherWorkspacesConnection(t *testing.T) {
	g := &generateStore{
		conn: &model.Connection{ID: "conn1", WorkspaceID: "ws2", Database: "Analytics"},
	}
	a := generateApp(g)

	_, appErr := a.GenerateCubeFiles(admin(), payload(col("id", "int")), "ws1")

	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403 for a connection in another workspace, got %v", appErr)
	}
	if len(g.saved) != 0 {
		t.Error("wrote files despite the forbidden connection")
	}
}

func TestGenerateCubeFilesGrantsPermissionsOnlyForWhatExists(t *testing.T) {
	g := &generateStore{
		existing: map[string]bool{"skipped.yml": true},
		roles:    []model.CollimatoRole{{Name: "viewer", AutoUpdate: true}},
	}
	a := generateApp(g)

	_, appErr := a.GenerateCubeFiles(admin(), map[string]interface{}{
		"conn1": map[string]interface{}{
			"orders":  []interface{}{col("id", "int")},
			"skipped": []interface{}{col("id", "int")},
		},
	}, "ws1")
	if appErr != nil {
		t.Fatalf("GenerateCubeFiles: %v", appErr)
	}

	if len(g.updatedRoles) != 1 {
		t.Fatalf("want one role update, got %d", len(g.updatedRoles))
	}

	perms := g.updatedRoles[0].TablePermissions
	if len(perms) != 1 || perms[0] != "orders" {
		t.Errorf("permissions should cover only created cubes, got %v", perms)
	}
}
