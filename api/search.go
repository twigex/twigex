// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

func (a *API) initSearch() {
	a.BaseRoutes.Search.Use(a.RequireSession)
	a.BaseRoutes.Search.Use(a.RequireCSRF)
}
