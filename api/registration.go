// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) resetPassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	p := struct {
		Password        string
		ConfirmPassword string
	}{}
	if !decodeBody(w, r, &p) {
		return
	}

	if p.Password != p.ConfirmPassword {
		respondAppError(w, r, model.NewAppError("user.password_mismatch", http.StatusBadRequest))
		return
	}

	appErr := a.app.ResetPassword(params["token"], p.Password)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) sendResetPasswordEmail(w http.ResponseWriter, r *http.Request) {
	d := struct {
		Email string
	}{}
	if !decodeBody(w, r, &d) {
		return
	}

	appErr := a.app.SendResetPasswordEmail(d.Email)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
