// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeTableOwnerStore struct {
	store.WorkspaceStore
	owners map[string]string
}

func (f fakeTableOwnerStore) GetTableWorkspaceID(_ context.Context, tableID string) (string, error) {
	return f.owners[tableID], nil
}

func TestRequireTableInWorkspace(t *testing.T) {
	if err := i18n.Init("../i18n"); err != nil {
		t.Fatalf("i18n: %v", err)
	}
	a := &API{app: &app.App{
		Store: store.Store{Workspace: fakeTableOwnerStore{owners: map[string]string{"t-a": "ws-a", "t-b": "ws-b"}}},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
	}}

	router := mux.NewRouter()
	reached := 0
	ok := func(w http.ResponseWriter, _ *http.Request) { reached++; w.WriteHeader(http.StatusOK) }
	router.HandleFunc("/{id}/table/{tid}/kanban-data", ok)
	router.HandleFunc("/{id}/members", ok)
	router.Use(a.RequireTableInWorkspace)

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/ws-a/table/t-a/kanban-data", http.StatusOK},
		{"/ws-a/table/t-b/kanban-data", http.StatusForbidden},
		{"/ws-a/table/t-missing/kanban-data", http.StatusForbidden},
		{"/ws-a/members", http.StatusOK},
	} {
		before := reached
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != tc.want {
			t.Errorf("%s: status %d, want %d", tc.path, rec.Code, tc.want)
		}
		if handled := reached > before; handled != (tc.want == http.StatusOK) {
			t.Errorf("%s: handler reached = %v", tc.path, handled)
		}
	}
}
