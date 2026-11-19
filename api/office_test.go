// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/model"
)

func newOfficeTestConfig(enabled bool, secret string) *model.ServerConfig {
	serverConfig := model.ServerConfig{}
	serverConfig.SetDefaults()
	serverConfig.OfficeSettings.Enable = model.NewBool(enabled)
	serverConfig.OfficeSettings.Secret = model.NewString(secret)

	return &serverConfig
}

func newOfficeRoutesAPI(enabled bool, secret string) *mux.Router {
	a := &API{
		app:        &app.App{ConfigStore: config.ConfigStore{Config: newOfficeTestConfig(enabled, secret)}},
		BaseRoutes: &Routes{},
	}

	router := mux.NewRouter()
	a.BaseRoutes.APIRoot = router.PathPrefix("/api").Subrouter()
	a.BaseRoutes.Office = a.BaseRoutes.APIRoot.PathPrefix("/office").Subrouter()
	a.initOffice()

	return router
}

func TestWOPIRoutesRefuseWhenOfficeUnavailable(t *testing.T) {
	if err := i18n.Init("../i18n"); err != nil {
		t.Fatalf("i18n: %v", err)
	}

	requests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/office/wopi/files/files/doc-1?access_token=x"},
		{http.MethodGet, "/api/office/wopi/files/files/doc-1/contents?access_token=x"},
		{http.MethodPost, "/api/office/wopi/files/files/doc-1/contents?access_token=x"},
		{http.MethodGet, "/api/office/eurooffice/file/files/doc-1?access_token=x"},
		{http.MethodPost, "/api/office/eurooffice/callback/files/doc-1?access_token=x"},
	}

	configs := []struct {
		name    string
		enabled bool
		secret  string
		want    int
	}{
		{"disabled", false, "office-secret", http.StatusNotFound},
		{"empty secret", true, "", http.StatusNotFound},
		{"whitespace secret", true, "  ", http.StatusNotFound},
		{"enabled with secret", true, "office-secret", http.StatusUnauthorized},
	}

	for _, c := range configs {
		router := newOfficeRoutesAPI(c.enabled, c.secret)

		for _, req := range requests {
			t.Run(c.name+" "+req.method+" "+req.path, func(t *testing.T) {
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(req.method, req.path, nil))

				if rec.Code != c.want {
					t.Errorf("status %d, want %d", rec.Code, c.want)
				}
			})
		}
	}
}

func TestOfficeSupportIsFalseWhenOfficeUnavailable(t *testing.T) {
	if err := i18n.Init("../i18n"); err != nil {
		t.Fatalf("i18n: %v", err)
	}

	configs := []struct {
		name    string
		enabled bool
		secret  string
		want    string
	}{
		{"disabled", false, "office-secret", "false"},
		{"empty secret", true, "", "false"},
		{"enabled with secret", true, "office-secret", "true"},
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			serverConfig := newOfficeTestConfig(c.enabled, c.secret)
			serverConfig.OfficeSettings.Type = model.NewString(model.OfficeTypeEuroOffice)
			api := &API{app: &app.App{ConfigStore: config.ConfigStore{Config: serverConfig}}}

			body := strings.NewReader(`{"officeType":"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}`)
			rec := httptest.NewRecorder()
			api.officeSupport(rec, httptest.NewRequest(http.MethodPost, "/api/office/file/support", body))

			got := strings.TrimSpace(rec.Body.String())
			if got != c.want {
				t.Errorf("support = %s, want %s", got, c.want)
			}
		})
	}
}
