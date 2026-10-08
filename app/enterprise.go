// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import "github.com/twigex/twigex/interfaces"

var ldapInterface func(*App) interfaces.LDAP
var oidcInterface func(*App) interfaces.OIDC
var collimatoRolesInterface func(*App) interfaces.CollimatoRoles
var workspaceRolesInterface func(*App) interfaces.WorkspaceRoles

func RegisterLDAPInterface(f func(*App) interfaces.LDAP) {
	ldapInterface = f
}

func RegisterOIDCInterface(f func(*App) interfaces.OIDC) {
	oidcInterface = f
}

func RegisterCollimatoRolesInterface(f func(*App) interfaces.CollimatoRoles) {
	collimatoRolesInterface = f
}

func RegisterWorkspaceRolesInterface(f func(*App) interfaces.WorkspaceRoles) {
	workspaceRolesInterface = f
}
