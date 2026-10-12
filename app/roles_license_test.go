// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeSystemRoleStore struct {
	store.RoleStore
	roles map[string]*model.Role
}

func (f *fakeSystemRoleStore) GetAll() ([]model.Role, error) {
	roles := []model.Role{}

	for _, role := range f.roles {
		roles = append(roles, *role)
	}

	return roles, nil
}

func (f *fakeSystemRoleStore) GetByID(id string) (*model.Role, error) {
	return f.roles[id], nil
}

func (f *fakeSystemRoleStore) GetByName(string) (*model.Role, error) {
	return nil, nil
}

func (f *fakeSystemRoleStore) CreateOrUpdate(role model.Role) (*model.Role, error) {
	return &role, nil
}

func (f *fakeSystemRoleStore) Delete(id string) error {
	delete(f.roles, id)

	return nil
}

func systemRolesApp(licensed bool) *App {
	a := &App{
		Store: store.Store{
			Roles: &fakeSystemRoleStore{roles: map[string]*model.Role{
				"builtin": {ID: "builtin", Name: model.SystemUserRoleId, BuiltIn: true},
				"custom":  {ID: "custom", Name: "editors"},
			}},
		},
	}

	if licensed {
		a.Server.License = &model.License{
			StartsAt:  time.Now().Add(-time.Hour).Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Features:  &model.Features{SystemRoles: boolPtr(true)},
		}
	}

	return a
}

func TestManagingSystemRolesNeedsALicense(t *testing.T) {
	a := systemRolesApp(false)
	permissions := []string{model.AdminPermissions.PermissionCreateUsers.Id}

	_, appErr := a.CreateSystemRole(systemAdmin, model.Role{Name: "auditors"})
	if appErr == nil || appErr.Status != http.StatusPaymentRequired {
		t.Errorf("creating a role: want 402, got %v", appErr)
	}

	for _, id := range []string{"custom", "builtin"} {
		_, appErr = a.UpdateSystemRole(systemAdmin, id, model.RolePatch{Permissions: &permissions})
		if appErr == nil || appErr.Status != http.StatusPaymentRequired {
			t.Errorf("editing the %s role: want 402, got %v", id, appErr)
		}
	}
}

func TestUnlicensedInstancesCanStillListAndDeleteRoles(t *testing.T) {
	a := systemRolesApp(false)

	roles, appErr := a.GetSystemRoles(systemAdmin)
	if appErr != nil || len(roles) != 2 {
		t.Errorf("listing roles for assignment: want 2, got %d (%v)", len(roles), appErr)
	}

	if appErr := a.DeleteSystemRole(systemAdmin, "custom"); appErr != nil {
		t.Errorf("deleting a custom role: %v", appErr)
	}
}

func TestLicensedInstancesManageSystemRoles(t *testing.T) {
	a := systemRolesApp(true)
	permissions := []string{model.AdminPermissions.PermissionCreateUsers.Id}

	if _, appErr := a.CreateSystemRole(systemAdmin, model.Role{Name: "auditors"}); appErr != nil {
		t.Errorf("creating a role: %v", appErr)
	}

	for _, id := range []string{"custom", "builtin"} {
		if _, appErr := a.UpdateSystemRole(systemAdmin, id, model.RolePatch{Permissions: &permissions}); appErr != nil {
			t.Errorf("editing the %s role: %v", id, appErr)
		}
	}
}
