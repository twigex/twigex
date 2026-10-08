// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/model"
	"golang.org/x/text/language"
)

func (a *API) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		sidCookie, err := r.Cookie(model.SessionCookieToken)
		if err != nil {
			respondAppError(w, r, model.NewAppError("session.invalid", http.StatusUnauthorized))
			return
		}

		sd, appErr := a.app.GetSessionDetails(sidCookie.Value)
		if appErr != nil {
			respondAppError(w, r, appErr)
			return
		}

		csrf := r.Header.Get("X-CSRF-TOKEN")
		if csrf == "" || csrf != sd.CSRF {
			respondAppError(w, r, model.NewAppError("session.not_authorized", http.StatusForbidden))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *API) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(model.SessionCookieToken)
		if err != nil {
			respondAppError(w, r, model.NewAppError("session.invalid", http.StatusUnauthorized))
			return
		}

		sd, appErr := a.app.ValidateSession(cookie.Value)
		if appErr != nil {
			if appErr.Status == http.StatusUnauthorized {
				a.app.InvalidateClientSession(w)
			}

			respondAppError(w, r, appErr)
			return
		}

		a.app.UpdateSession(sd, w)
		next.ServeHTTP(w, r)
	})
}

// RecordOriginClient passes the tab id the client sends on to the app, which
// stamps it on the project changes the request causes.
func (a *API) RecordOriginClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get("X-Client-ID"); id != "" {
			r = r.WithContext(app.WithOriginClient(r.Context(), id))
		}

		next.ServeHTTP(w, r)
	})
}

// RequireTableInWorkspace refuses a route that names a table from another
// workspace than the one in its path, so every such route is covered without
// each handler repeating the check.
func (a *API) RequireTableInWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if tableID := vars["tid"]; tableID != "" {
			if appErr := a.app.CheckTableInWorkspace(r.Context(), vars["id"], tableID); appErr != nil {
				respondAppError(w, r, appErr)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (a *API) LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := "en"

		tags, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		if len(tags) > 0 {
			locale = tags[0].String()
		}

		ctx := context.WithValue(r.Context(), localeKey{}, locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) VersionHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-App-Deploy", model.BuildTimestamp)
		w.Header().Set("X-App-Version", model.Version)
		w.Header().Set("X-App-Min-Version", model.MinClientVersion)
		next.ServeHTTP(w, r)
	})
}

func (a *API) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(self), microphone=(self), display-capture=(self), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
