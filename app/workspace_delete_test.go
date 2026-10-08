// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestDeleteProjectWorkspaceIgnoresTheOldDeletePermission(t *testing.T) {
	ws := &fakeAccessWorkspaceStore{
		memberRoleName: "manager",
		roles:          []model.ProjectWorkspaceRole{{Name: "manager", Permissions: []string{"delete_workspace"}}},
	}
	a := accessApp(ws)

	_, appErr := a.DeleteProjectWorkspace(context.Background(), "w1", model.User{ID: "u1"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 for a member without the admin role, got %v", appErr)
	}
}
