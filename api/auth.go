// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initAuth() {
	a.BaseRoutes.Auth.HandleFunc("/logout", a.logOut).Methods("POST")
	a.BaseRoutes.Auth.HandleFunc("/session", a.sessionLogin).Methods("GET")
	a.BaseRoutes.Auth.Use(a.RequireSession)
	a.BaseRoutes.Auth.Use(a.RequireCSRF)

	reset := a.BaseRoutes.APIRoot.PathPrefix("/password").Subrouter()
	reset.HandleFunc("/policy", a.passwordPolicy).Methods("GET")
	reset.HandleFunc("/reset", a.sendResetPasswordEmail).Methods("POST")
	reset.HandleFunc("/reset/{token}", a.resetPassword).Methods("POST")

	Auth := a.BaseRoutes.APIRoot.PathPrefix("/auth").Subrouter()
	Auth.HandleFunc("/login", a.signIn).Methods("POST")
	Auth.HandleFunc("/oidc/login/{id}", a.handleOIDCLogin).Methods("GET")
	Auth.HandleFunc("/oidc/callback/{id}", a.handleOIDCCallback).Methods("GET")
}

func (a *API) signIn(w http.ResponseWriter, r *http.Request) {
	var c model.Credentials
	if !decodeBody(w, r, &c) {
		return
	}

	user, appErr := a.app.ValidateCredentials(c, a.app.ClientIP(r))
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.LogIn(*user, false, w, r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (a *API) logOut(w http.ResponseWriter, r *http.Request) {
	sid, err := r.Cookie(model.SessionCookieToken)
	if err != nil {
		respondAppError(w, r, model.NewAppError("auth.logout_failed", http.StatusUnauthorized))
		return
	}

	appErr := a.app.LogOut(sid.Value, w)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *API) sessionLogin(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (a *API) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if !a.app.Server.License.HasOAuthProviders() {
		respondAppError(w, r, model.NewAppError("auth.oauth_unavailable", http.StatusForbidden))
		return
	}

	state := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, state); err != nil {
		respondAppError(w, r, model.NewAppError("auth.oauth_failed", http.StatusForbidden))
		return
	}

	nonce := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		respondAppError(w, r, model.NewAppError("auth.oauth_failed", http.StatusForbidden))
		return
	}

	statebase64 := base64.RawURLEncoding.EncodeToString(state)
	noncebase64 := base64.RawURLEncoding.EncodeToString(nonce)

	authURL, appErr := a.app.OIDCAuthCodeURL(r.Context(), id, statebase64, noncebase64)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	setOIDCCookie(w, "state", statebase64)
	setOIDCCookie(w, "nonce", noncebase64)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (a *API) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// Redirect to login view if errors so users are not left on blank page
	redirectPath := "/login?error=oauth_failed"

	state, err := r.Cookie("state")
	if err != nil {
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	clearOIDCCookie(w, "state")

	if r.URL.Query().Get("state") != state.Value {
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	clearOIDCCookie(w, "nonce")

	user, appErr := a.app.ExchangeOauth2Token(r.Context(), id, r.URL.Query().Get("code"), nonce.Value)
	if appErr != nil {
		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	if appErr := a.app.LogIn(*user, true, w, r); appErr != nil {
		if appErr.ID == "auth.user_deactivated" {
			http.Redirect(w, r, "/login?error=account_disabled", http.StatusFound)
			return
		}

		http.Redirect(w, r, redirectPath, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func setOIDCCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
	})
}

func clearOIDCCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) passwordPolicy(w http.ResponseWriter, r *http.Request) {
	pw, appErr := a.app.GetPasswordPolicy()
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, pw)
}
