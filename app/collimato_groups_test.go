// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeCollimatoStore struct {
	store.CollimatoStore
	effectiveRoles []string
	knownRoles     []model.CollimatoRole
}

func (f *fakeCollimatoStore) GetEffectiveRolesForUser(_ context.Context, _, _ string) ([]string, error) {
	return f.effectiveRoles, nil
}

func (f *fakeCollimatoStore) GetRolesByName(names []string, _ string) ([]model.CollimatoRole, error) {
	out := []model.CollimatoRole{}
	for _, r := range f.knownRoles {
		for _, n := range names {
			if r.Name == n {
				out = append(out, r)
				break
			}
		}
	}
	return out, nil
}

type fakeGroupStore struct {
	store.GroupStore
	groups []model.Group
}

func (f *fakeGroupStore) GetByIDs(_ context.Context, _ []string) ([]model.Group, error) {
	return f.groups, nil
}

func boolPtr(b bool) *bool { return &b }

func collimatoApp(c *fakeCollimatoStore, g *fakeGroupStore) *App {
	return &App{
		Store: store.Store{Collimato: c, Groups: g},
		Server: Server{
			License: &model.License{
				StartsAt:  time.Now().Add(-time.Hour).Unix(),
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Features:  &model.Features{Groups: boolPtr(true)},
			},
		},
	}
}

func memberWithPermission(permID string) *fakeCollimatoStore {
	return &fakeCollimatoStore{
		effectiveRoles: []string{"workspace_admin"},
		knownRoles: []model.CollimatoRole{{
			Name:        "workspace_admin",
			Permissions: []string{permID},
		}},
	}
}

func admin() model.User {
	return model.User{ID: "admin", Role: model.SystemAdminRoleId}
}

func regularUser() model.User {
	return model.User{ID: "u1", Role: model.SystemUserRoleId}
}

func TestAddGroupsToWorkspaceNoLicense(t *testing.T) {
	a := &App{Store: store.Store{Collimato: &fakeCollimatoStore{}}}

	_, appErr := a.AddGroupsToWorkspace(admin(), "ws1", []string{"g1"}, []string{"viewer"})
	if appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Fatalf("expected 402 without a license, got %v", appErr)
	}
}

func TestAddGroupsToWorkspaceForbiddenForNonMember(t *testing.T) {
	c := &fakeCollimatoStore{effectiveRoles: nil}
	a := collimatoApp(c, &fakeGroupStore{})

	_, appErr := a.AddGroupsToWorkspace(regularUser(), "ws1", []string{"g1"}, []string{"viewer"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", appErr)
	}
}

func TestAddGroupsToWorkspaceEmptyGroups(t *testing.T) {
	a := collimatoApp(memberWithPermission(model.CollimatoPermissions.PermissionAddUsers.Id), &fakeGroupStore{})

	_, appErr := a.AddGroupsToWorkspace(regularUser(), "ws1", nil, []string{"viewer"})
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty groups, got %v", appErr)
	}
}

func TestAddGroupsToWorkspaceEmptyRoles(t *testing.T) {
	a := collimatoApp(memberWithPermission(model.CollimatoPermissions.PermissionAddUsers.Id), &fakeGroupStore{})

	_, appErr := a.AddGroupsToWorkspace(regularUser(), "ws1", []string{"g1"}, nil)
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty roles, got %v", appErr)
	}
}

func TestAddGroupsToWorkspaceMissingGroup(t *testing.T) {
	g := &fakeGroupStore{groups: nil}
	a := collimatoApp(memberWithPermission(model.CollimatoPermissions.PermissionAddUsers.Id), g)

	_, appErr := a.AddGroupsToWorkspace(regularUser(), "ws1", []string{"ghost"}, []string{"viewer"})
	if appErr == nil || appErr.Status != http.StatusNotFound {
		t.Fatalf("expected 404 for missing group, got %v", appErr)
	}
}

func TestAddGroupsToWorkspaceUnknownRole(t *testing.T) {
	g := &fakeGroupStore{groups: []model.Group{{ID: "g1"}}}
	a := collimatoApp(memberWithPermission(model.CollimatoPermissions.PermissionAddUsers.Id), g)

	_, appErr := a.AddGroupsToWorkspace(regularUser(), "ws1", []string{"g1"}, []string{"nope"})
	if appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown role, got %v", appErr)
	}
}

func TestRemoveGroupFromWorkspaceForbiddenForNonMember(t *testing.T) {
	c := &fakeCollimatoStore{effectiveRoles: nil}
	a := collimatoApp(c, &fakeGroupStore{})

	appErr := a.RemoveGroupFromWorkspace(regularUser(), "ws1", "g1")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", appErr)
	}
}
