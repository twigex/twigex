// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeRoleListStore struct {
	fakeMetaCollimatoStore
}

func (f *fakeRoleListStore) GetWorkspaceRoles(string) ([]model.CollimatoRole, error) {
	return []model.CollimatoRole{{
		ID:               "r1",
		Name:             "finance",
		Description:      "Finance team",
		Permissions:      []string{"view_charts"},
		TablePermissions: []string{"salaries"},
	}}, nil
}

func roleListApp(permissions ...string) *App {
	a := metaApp()
	a.Store.Collimato = &fakeRoleListStore{fakeMetaCollimatoStore{permissions: permissions}}

	return a
}

func TestAssignersSeeRoleNamesWithoutWhatTheyGrant(t *testing.T) {
	roles, appErr := roleListApp(model.CollimatoPermissions.PermissionAssignRoles.Id).GetWorkspaceRoles(model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoles: %v", appErr)
	}

	if len(roles) != 1 || roles[0].Name != "finance" || roles[0].Description != "Finance team" {
		t.Fatalf("expected the role name and description, got %+v", roles)
	}

	if roles[0].Permissions != nil || roles[0].TablePermissions != nil {
		t.Errorf("expected what the role grants left out, got %+v", roles[0])
	}
}

func TestRoleViewersSeeTheWholeRole(t *testing.T) {
	roles, appErr := roleListApp(model.CollimatoPermissions.PermissionViewRoles.Id).GetWorkspaceRoles(model.User{ID: "u1"}, "ws")
	if appErr != nil {
		t.Fatalf("GetWorkspaceRoles: %v", appErr)
	}

	if len(roles[0].Permissions) != 1 || len(roles[0].TablePermissions) != 1 {
		t.Errorf("expected the full role, got %+v", roles[0])
	}

	_, appErr = roleListApp(model.CollimatoPermissions.PermissionViewCharts.Id).GetWorkspaceRoles(model.User{ID: "u1"}, "ws")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 without view or assign roles, got %v", appErr)
	}
}
