// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"crypto/subtle"
	"net/http"

	"github.com/twigex/twigex/model"
)

// initCubeInternal registers the internal endpoints the shared Cube.js server
// calls back into to resolve a workspace's data model and database connections
// at query time. These routes are NOT behind the user session/CSRF middleware
// (the caller is Cube.js, authenticated with a shared callback secret) and must
// only be reachable on a private network, since the connection endpoint returns
// database credentials.
func (a *API) initCubeInternal() {
	a.BaseRoutes.CubeInternal.HandleFunc("/schema", a.cubeSchema).Methods("GET")
	a.BaseRoutes.CubeInternal.HandleFunc("/schema-version", a.cubeSchemaVersion).Methods("GET")
	a.BaseRoutes.CubeInternal.HandleFunc("/connection", a.cubeConnection).Methods("GET")

	a.BaseRoutes.CubeInternal.Use(a.RequireCubeCallbackSecret)
}

// RequireCubeCallbackSecret authenticates the shared Cube.js server via a
// constant secret sent in the X-Cube-Callback-Secret header. It fails closed:
// if no secret is configured, every request is rejected.
func (a *API) RequireCubeCallbackSecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := *a.app.ConfigStore.Config.CollimatoSettings.CubeCallbackSecret
		provided := r.Header.Get("X-Cube-Callback-Secret")
		if secret == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(provided)) != 1 {
			respondAppError(w, r, model.NewAppError("session.not_authorized", http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *API) cubeSchema(w http.ResponseWriter, r *http.Request) {
	files, appErr := a.app.GetCubeSchemaFiles(r.URL.Query().Get("workspace"))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, files)
}

func (a *API) cubeSchemaVersion(w http.ResponseWriter, r *http.Request) {
	version, appErr := a.app.GetCubeSchemaVersion(r.URL.Query().Get("workspace"))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64{"version": version})
}

func (a *API) cubeConnection(w http.ResponseWriter, r *http.Request) {
	conn, appErr := a.app.GetCubeConnection(r.URL.Query().Get("workspace"), r.URL.Query().Get("dataSource"))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, conn)
}
