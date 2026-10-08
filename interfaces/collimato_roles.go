// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package interfaces

import (
	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
)

// CollimatoRoles covers defining roles and enforcing what they grant. Reading
// roles and enforcing the built-in workspace_admin / workspace_user stays in the
// public tree, because an unlicensed instance still needs both.
type CollimatoRoles interface {
	CreateRole(user model.User, workspaceID string, role model.CollimatoRole) (*model.CollimatoRole, *model.AppError)
	UpdateRole(user model.User, workspaceID, roleID string, patch model.CollimatoRolePatch) (*model.CollimatoRole, *model.AppError)
	DeleteRole(user model.User, workspaceID, roleID string) *model.AppError
	AssignUserRoles(workspaceID, memberID string, roles []string) *model.AppError
	AssignGroupRoles(workspaceID, groupID string, roles []string) *model.AppError

	HasPermissionToFields(roles []model.CollimatoRole, query *model.DataQuery) (bool, error)
	FilterMeta(roles []model.CollimatoRole, cubeCollection collimato.CubeCollection) *collimato.CubeCollection
}
