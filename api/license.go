// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/twigex/twigex/model"
)

func (a *API) initLicense() {
	a.BaseRoutes.License.HandleFunc("/add", a.addLicense).Methods("POST")

	a.BaseRoutes.License.Use(a.RequireSession)
	a.BaseRoutes.License.Use(a.RequireCSRF)

	License := a.BaseRoutes.APIRoot.PathPrefix("/license").Subrouter()
	License.HandleFunc("", a.getActiveLicense).Methods("GET")
}

func (a *API) getActiveLicense(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var m *map[string]any

	if user == nil {
		m = a.app.GetLicenseMap(false)
	} else {
		m = a.app.GetLicenseMap(true)
	}

	respondJSON(w, http.StatusOK, m)
}

func (a *API) addLicense(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondAppError(w, r, model.NewAppError("license.file_too_large", http.StatusRequestEntityTooLarge))
			return
		}

		respondAppError(w, r, model.NewAppError("license.upload_failed", http.StatusInternalServerError))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondAppError(w, r, model.NewAppError("license.upload_failed", http.StatusInternalServerError))
		return
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondAppError(w, r, model.NewAppError("license.upload_failed", http.StatusInternalServerError))
		return
	}

	license, appErr := a.app.AddLicense(*user, data)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	license.SetDefaults()
	licenseMap := a.app.GetLicenseMap(true)

	respondJSON(w, http.StatusOK, struct {
		*model.License `json:"license"`
		Map            *map[string]any `json:"map"`
	}{
		License: license,
		Map:     licenseMap,
	})
}

func (a *API) removeLicense(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.RemoveActiveLicense(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	licenseMap := a.app.GetLicenseMap(false)
	respondJSON(w, http.StatusOK, licenseMap)
}
