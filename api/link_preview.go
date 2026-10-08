// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) initLinkPreview() {
	a.BaseRoutes.LinkPreview.HandleFunc("/image/{hash}", a.serveLinkPreviewImage).Methods("GET")
	a.BaseRoutes.LinkPreview.Use(a.RequireSession)
}

func (a *API) serveLinkPreviewImage(w http.ResponseWriter, r *http.Request) {
	hash, err := strconv.ParseInt(mux.Vars(r)["hash"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if appErr := a.app.ServeLinkPreviewImage(hash, w, r); appErr != nil {
		respondAppError(w, r, appErr)
	}
}
