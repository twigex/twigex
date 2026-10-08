// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"encoding/json"
	"net/http"

	"github.com/twigex/twigex/model"
)

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return false
	}

	return true
}

func sortFromQuery(r *http.Request) model.Sort {
	return model.Sort{
		Key:  r.URL.Query().Get("sort"),
		Desc: r.URL.Query().Get("dir") == "desc",
	}
}

func (a *API) requirePermission(w http.ResponseWriter, r *http.Request, user model.User, permission *model.SystemPermission) bool {
	if !a.app.SessionHasPermission(user, permission) {
		respondAppError(w, r, model.NewAppError("permission.forbidden", http.StatusForbidden))
		return false
	}

	return true
}
