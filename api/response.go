// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"encoding/json"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func respondAppError(w http.ResponseWriter, r *http.Request, appErr *model.AppError) {
	locale := getLocale(r)

	b, err := json.Marshal(map[string]string{
		"error": appErr.Translate(locale),
	})
	if err != nil {
		tlog.Errorw("Failed to marshal app error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	if _, err := w.Write(b); err != nil {
		tlog.Errorw("Failed to write response", "error", err)
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(b); err != nil {
		tlog.Errorw("Failed to write response", "error", err)
	}
}
