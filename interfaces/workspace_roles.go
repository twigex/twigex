// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package interfaces

import "github.com/twigex/twigex/model"

type AccessInput struct {
	Roles      []model.ProjectWorkspaceRole
	TablePerms []model.TablePermission
}

type WorkspaceAccess interface {
	VisibleTables(tables []model.WorkspaceTable) []model.WorkspaceTable
	AssignedOnlyAll() bool
	AssignedOnly(tableID string) bool
	CanPerformRowAction(tableID, workspacePerm, tableAction string) bool
	TableHasPermission(tableID, action string) bool
}

type WorkspaceRoles interface {
	Resolve(in AccessInput) WorkspaceAccess
}
