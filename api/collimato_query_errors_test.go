// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/model"
)

// A chart the engine cannot answer gets its reason, with what it is about, in
// the reader's language.
func TestAnErrorWithValuesIsWrittenInTheReadersLanguage(t *testing.T) {
	if err := i18n.Init("../i18n"); err != nil {
		t.Fatalf("i18n: %v", err)
	}

	request := httptest.NewRequest("POST", "/data", nil)
	request = request.WithContext(context.WithValue(request.Context(), localeKey{}, "pl"))

	appError := model.NewAppError("collimato.query.unknown_field", http.StatusBadRequest)
	appError.Values = map[string]any{"Detail": "orders.profit"}

	recorder := httptest.NewRecorder()
	respondAppError(recorder, request, appError)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", recorder.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	want := "Ten wykres używa pola orders.profit, którego nie ma już w modelu danych."
	if body["error"] != want {
		t.Errorf("want %q, got %q", want, body["error"])
	}
}
