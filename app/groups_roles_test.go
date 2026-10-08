// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeRoleStore struct {
	store.RoleStore
	known map[string]bool
}

func (f *fakeRoleStore) GetByNames(names []string) ([]model.Role, error) {
	out := []model.Role{}
	for _, n := range names {
		if f.known[n] {
			out = append(out, model.Role{Name: n})
		}
	}
	return out, nil
}

type permFakeRoleStore struct {
	store.RoleStore
	perms map[string][]string // role name -> permission ids
}

func (f *permFakeRoleStore) GetByNames(names []string) ([]model.Role, error) {
	out := []model.Role{}
	for _, n := range names {
		if p, ok := f.perms[n]; ok {
			out = append(out, model.Role{Name: n, Permissions: p})
		}
	}
	return out, nil
}

type rolesFakeGroupStore struct {
	store.GroupStore
	groupRoles []string
}

func (f *rolesFakeGroupStore) GetRolesForUser(_ context.Context, _ string) ([]string, error) {
	return f.groupRoles, nil
}

func roleApp(known ...string) *App {
	m := make(map[string]bool, len(known))
	for _, k := range known {
		m[k] = true
	}
	return &App{Store: store.Store{Roles: &fakeRoleStore{known: m}}}
}

func TestValidateGroupRoles_RejectsSystemAdmin(t *testing.T) {
	a := roleApp("analyst")

	_, appErr := a.validateGroupRoles([]string{"analyst", model.SystemAdminRoleId})
	if appErr == nil || appErr.Status != http.StatusBadRequest || appErr.ID != "group.role_admin_forbidden" {
		t.Fatalf("expected group.role_admin_forbidden 400, got %v", appErr)
	}
}

func TestValidateGroupRoles_RejectsUnknownRole(t *testing.T) {
	a := roleApp("analyst")

	_, appErr := a.validateGroupRoles([]string{"analyst", "does_not_exist"})
	if appErr == nil || appErr.Status != http.StatusBadRequest || appErr.ID != "group.invalid_role" {
		t.Fatalf("expected group.invalid_role 400, got %v", appErr)
	}
}

func TestValidateGroupRoles_NormalizesTrimAndDedupe(t *testing.T) {
	a := roleApp("analyst", "auditor")

	got, appErr := a.validateGroupRoles([]string{"  analyst ", "analyst", "", "auditor"})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	want := []string{"analyst", "auditor"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cleaned roles = %v, want %v", got, want)
	}
}

func TestValidateGroupRoles_EmptyIsAllowed(t *testing.T) {
	a := roleApp("analyst")

	got, appErr := a.validateGroupRoles(nil)
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty roles, got %v", got)
	}
}

func TestApplyGroupRoles_UnionsIntoUserRole(t *testing.T) {
	a := &App{Store: store.Store{Groups: &rolesFakeGroupStore{groupRoles: []string{"analyst", "system_user", "auditor"}}}}

	user := &model.User{ID: "u1", Role: "system_user"}
	a.applyGroupRoles(context.Background(), user)

	// system_user already present (not duplicated); analyst + auditor appended.
	want := []string{"system_user", "analyst", "auditor"}
	got := strings.Fields(user.Role)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("effective role = %q (%v), want %v", user.Role, got, want)
	}
}

func TestApplyGroupRoles_SkipsSystemAdmin(t *testing.T) {
	a := &App{Store: store.Store{Groups: &rolesFakeGroupStore{groupRoles: []string{"analyst"}}}}

	user := &model.User{ID: "admin", Role: model.SystemAdminRoleId}
	a.applyGroupRoles(context.Background(), user)

	if user.Role != model.SystemAdminRoleId {
		t.Fatalf("admin role mutated to %q", user.Role)
	}
}

func TestApplyGroupRoles_NoGroupRolesLeavesUserUntouched(t *testing.T) {
	a := &App{Store: store.Store{Groups: &rolesFakeGroupStore{groupRoles: nil}}}

	user := &model.User{ID: "u1", Role: "system_user"}
	a.applyGroupRoles(context.Background(), user)

	if user.Role != "system_user" {
		t.Fatalf("role changed to %q, want unchanged", user.Role)
	}
}

func TestGetMyPermissions_UnionsAcrossMultipleRoles(t *testing.T) {
	a := &App{Store: store.Store{Roles: &permFakeRoleStore{perms: map[string][]string{
		"system_user": {"view_files", "send_message"},
		"analyst":     {"view_files", "view_charts"}, // view_files overlaps
	}}}}

	// The effective role string GetCurrentUser produces after folding in a
	// group-granted "analyst" role, the exact input that used to return empty.
	user := model.User{ID: "u1", Role: "system_user analyst"}

	perms, appErr := a.GetMyPermissions(user)
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	sort.Strings(perms)
	want := []string{"send_message", "view_charts", "view_files"}
	if !reflect.DeepEqual(perms, want) {
		t.Fatalf("permissions = %v, want %v", perms, want)
	}
}

func TestGetMyPermissions_GroupOnlyRole(t *testing.T) {
	a := &App{Store: store.Store{Roles: &permFakeRoleStore{perms: map[string][]string{
		"analyst": {"view_charts"},
	}}}}

	// User whose only role is group-granted (direct default role removed).
	user := model.User{ID: "u1", Role: "analyst"}

	perms, appErr := a.GetMyPermissions(user)
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if !reflect.DeepEqual(perms, []string{"view_charts"}) {
		t.Fatalf("permissions = %v, want [view_charts]", perms)
	}
}

func TestGetMyPermissions_EmptyRole(t *testing.T) {
	a := &App{Store: store.Store{Roles: &permFakeRoleStore{perms: map[string][]string{}}}}

	perms, appErr := a.GetMyPermissions(model.User{ID: "u1", Role: ""})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if len(perms) != 0 {
		t.Fatalf("permissions = %v, want empty", perms)
	}
}
