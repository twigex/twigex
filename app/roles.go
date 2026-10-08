// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"regexp"
	"slices"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

var validRoleName = regexp.MustCompile(`^[a-z0-9_-]+$`)

func (a *App) GetSystemRoles(user model.User) ([]model.Role, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionCreateUsers) &&
		!a.SessionHasPermission(user, model.AdminPermissions.PermissionEditUsers) &&
		!a.SessionHasPermission(user, model.AdminPermissions.PermissionManageRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	roles, err := a.Store.Roles.GetAll()
	if err != nil {
		tlog.Errorw("Failed to retrieve system roles", "error", err)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	return roles, nil
}

func (a *App) GetSystemPermissions(user model.User) ([]model.PermissionSection, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	return model.GetSystemPermissionSections(), nil
}

func (a *App) CreateSystemRole(user model.User, role model.Role) (*model.Role, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	if role.Name == "" {
		return nil, model.NewAppError("role.name_required", http.StatusBadRequest)
	}

	if !validRoleName.MatchString(role.Name) {
		return nil, model.NewAppError("role.name_invalid", http.StatusBadRequest)
	}

	existing, err := a.Store.Roles.GetByName(role.Name)
	if err != nil {
		tlog.Errorw("Failed to check role name uniqueness", "name", role.Name, "error", err)
		return nil, model.NewAppError("role.check_failed", http.StatusInternalServerError)
	}

	if existing != nil {
		return nil, model.NewAppError("role.name_taken", http.StatusBadRequest)
	}

	role.BuiltIn = false
	created, err := a.Store.Roles.CreateOrUpdate(role)
	if err != nil {
		tlog.Errorw("Failed to create system role", "name", role.Name, "error", err)
		return nil, model.NewAppError("role.create_failed", http.StatusInternalServerError)
	}

	return created, nil
}

func (a *App) UpdateSystemRole(user model.User, id string, patch model.RolePatch) (*model.Role, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	role, err := a.Store.Roles.GetByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve system role", "id", id, "error", err)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	if role == nil {
		return nil, model.NewAppError("role.not_found", http.StatusNotFound)
	}

	if appErr := guardSystemRoleEdit(user, role); appErr != nil {
		return nil, appErr
	}

	if role.BuiltIn {
		// Built-in roles: only permissions can change, name/display/description are locked
		if patch.Permissions != nil {
			role.Permissions = *patch.Permissions
		}
	} else {
		role.Patch(patch)
	}

	updated, err := a.Store.Roles.CreateOrUpdate(*role)
	if err != nil {
		tlog.Errorw("Failed to update system role", "id", id, "error", err)
		return nil, model.NewAppError("role.update_failed", http.StatusInternalServerError)
	}

	return updated, nil
}

func (a *App) DeleteSystemRole(user model.User, id string) *model.AppError {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionManageRoles) {
		return model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	role, err := a.Store.Roles.GetByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve system role", "id", id, "error", err)
		return model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	if role == nil {
		return model.NewAppError("role.not_found", http.StatusNotFound)
	}

	if role.BuiltIn {
		return model.NewAppError("role.cannot_delete_builtin", http.StatusForbidden)
	}

	if err := a.Store.Roles.Delete(id); err != nil {
		tlog.Errorw("Failed to delete system role", "id", id, "error", err)
		return model.NewAppError("role.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

// EnsureSystemRoles creates any built-in system role that is missing and marks
// the existing ones as built in. It runs on every start, since a fresh database
// holds no roles and permissions depend on them.
func (a *App) EnsureSystemRoles() {
	existing, err := a.Store.Roles.GetAll()
	if err != nil {
		tlog.Errorw("Failed to retrieve existing roles", "error", err)
		return
	}

	for _, role := range model.MakeDefaultRoles() {
		i := slices.IndexFunc(existing, func(r model.Role) bool { return r.Name == role.Name })

		saved := *role
		if i >= 0 {
			saved = existing[i]
			saved.BuiltIn = true
		}

		if _, err := a.Store.Roles.CreateOrUpdate(saved); err != nil {
			tlog.Errorw("Failed to save built-in role", "role", role.Name, "error", err)
		}
	}
}
