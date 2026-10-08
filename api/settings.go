// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initSettings() {
	a.BaseRoutes.Settings.HandleFunc("/language", a.getLanguageSettings).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/language", a.updateLanguageSettings).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/email-server", a.getEmailServer).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/email-server", a.updateEmailServer).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/office", a.getOfficeSettings).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/office", a.updateOfficeSettings).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/email-server/connection", a.testEmailServerConnection).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/password-policy", a.getSecuritySettings).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/password-policy", a.updateSecuritySettings).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/metadata", a.createFileMetadata).Methods("POST")        // create metadata in settings menu
	a.BaseRoutes.Settings.HandleFunc("/metadata", a.getFileMetadataList).Methods("GET")        // get all available metadata
	a.BaseRoutes.Settings.HandleFunc("/metadata/{id}", a.deleteFileMetadata).Methods("DELETE") // deletes metadata
	a.BaseRoutes.Settings.HandleFunc("/metadata/{id}", a.updateFileMetadata).Methods("PUT")    // edits metadata
	a.BaseRoutes.Settings.HandleFunc("/chat", a.getChatSettings).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/chat/configuration", a.getChatConfig).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/chat/configuration", a.updateChatConfig).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/license", a.getLicense).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/license", a.removeLicense).Methods("DELETE")
	a.BaseRoutes.Settings.HandleFunc("/auth", a.getAuthSettings).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/auth", a.updateAuthSettings).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/auth/ldap/test", a.testLDAPConnection).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/auth/ldap/sync", a.runLDAPSync).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/auth/oidc/providers", a.getOIDCProviders).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/auth/oidc/providers", a.createOIDCProvider).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/auth/oidc/providers/prepare", a.prepareOIDCProvider).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/auth/oidc/providers/{id}", a.updateOIDCProvider).Methods("PUT")
	a.BaseRoutes.Settings.HandleFunc("/auth/oidc/providers/{id}", a.deleteOIDCProvider).Methods("DELETE")

	a.BaseRoutes.Settings.HandleFunc("/storage", a.getStorages).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/storage", a.createStorage).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/storage/{id}", a.updateStorage).Methods("PUT")
	a.BaseRoutes.Settings.HandleFunc("/storage/{id}", a.deleteStorage).Methods("DELETE")
	a.BaseRoutes.Settings.HandleFunc("/storage/{id}/primary", a.setPrimaryStorage).Methods("PUT")

	a.BaseRoutes.Settings.HandleFunc("/roles", a.getSystemRoles).Methods("GET")
	a.BaseRoutes.Settings.HandleFunc("/roles", a.createSystemRole).Methods("POST")
	a.BaseRoutes.Settings.HandleFunc("/roles/{id}", a.updateSystemRole).Methods("PUT")
	a.BaseRoutes.Settings.HandleFunc("/roles/{id}", a.deleteSystemRole).Methods("DELETE")
	a.BaseRoutes.Settings.HandleFunc("/permissions", a.getSystemPermissions).Methods("GET")

	a.BaseRoutes.Settings.Use(a.RequireSession)
	a.BaseRoutes.Settings.Use(a.RequireCSRF)

	Settings := a.BaseRoutes.APIRoot.PathPrefix("/settings").Subrouter()
	Settings.HandleFunc("/password-policy", a.getPasswordPolicy).Methods("GET")
	Settings.HandleFunc("/oauth", a.getOAuthSettings).Methods("GET")
}

func (a *API) getLanguageSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	settings, appErr := a.app.GetLanguageSettings(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (a *API) updateLanguageSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.LanguageSettings
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.UpdateLanguageSettings(*user, req); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getAuthSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	settings, appErr := a.app.GetAuthSettings(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (a *API) updateAuthSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.AdminAuthSettings
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.UpdateAuthSettings(*user, req); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) testLDAPConnection(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.AdminLDAPSettings
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.TestLDAPConnection(*user, req); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Connected successfully"})
}

func (a *API) runLDAPSync(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	job, appErr := a.app.RunLDAPSync(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}

func (a *API) getOIDCProviders(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	providers, appErr := a.app.GetOIDCProviders(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, providers)
}

func (a *API) createOIDCProvider(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.OIDCProvider
	if !decodeBody(w, r, &req) {
		return
	}

	created, appErr := a.app.CreateOIDCProvider(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, created)
}

func (a *API) prepareOIDCProvider(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	result, appErr := a.app.PrepareOIDCProvider(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) updateOIDCProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.OIDCProvider
	if !decodeBody(w, r, &req) {
		return
	}

	updated, appErr := a.app.UpdateOIDCProvider(*user, id, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (a *API) deleteOIDCProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteOIDCProvider(*user, id); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) createFileMetadata(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	m := struct {
		Name   string
		Type   string
		Fields any
	}{}
	if !decodeBody(w, r, &m) {
		return
	}

	meta, appErr := a.app.CreateMetadata(*user, m.Name, m.Type, m.Fields)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

func (a *API) getFileMetadataList(w http.ResponseWriter, r *http.Request) {
	meta, appErr := a.app.GetAllMetadata()
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meta)
}

func (a *API) deleteFileMetadata(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteMetadata(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateFileMetadata(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	m := struct {
		Name   string
		Type   string
		Fields any
	}{}
	if !decodeBody(w, r, &m) {
		return
	}

	appErr = a.app.EditFileMetadata(*user, params["id"], m.Name, m.Type, m.Fields)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getChatSettings(w http.ResponseWriter, r *http.Request) {
	normalizedHost := a.app.NormalizeLiveKitHost(*a.app.ConfigStore.Config.ChannelSettings.Host, model.ProtocolWSS)
	s := struct {
		EnableVideo bool   `json:"enable_video"`
		Host        string `json:"host"`
		GIF         bool   `json:"gif"`
	}{
		EnableVideo: *a.app.ConfigStore.Config.ChannelSettings.Enabled,
		Host:        normalizedHost,
		GIF:         *a.app.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled,
	}

	respondJSON(w, http.StatusOK, s)
}

func (a *API) getChatConfig(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	config, appErr := a.app.GetChatConfig(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, config)
}

func (a *API) updateChatConfig(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	settings := model.ChatSettings{}
	if !decodeBody(w, r, &settings) {
		return
	}

	appErr = a.app.UpdateChatSettings(*user, settings)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getEmailServer(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	server, appErr := a.app.GetEmailServer(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, server)
}

func (a *API) updateEmailServer(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.EmailSettings{}
	if !decodeBody(w, r, &req) {
		return
	}

	appErr = a.app.UpdateEmailServer(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getOfficeSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	settings, appErr := a.app.GetOfficeSettings(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (a *API) updateOfficeSettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.OfficeSettings{}
	if !decodeBody(w, r, &req) {
		return
	}

	if appErr := a.app.UpdateOfficeSettings(*user, req); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) testEmailServerConnection(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.EmailSettings{}
	if !decodeBody(w, r, &req) {
		return
	}

	status, appErr := a.app.TestEmailServerConnection(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, status)
}

func (a *API) getSecuritySettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	settings, appErr := a.app.GetSecuritySettings(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (a *API) updateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := struct {
		PasswordSettings model.PasswordSettings
		SessionLength    string
	}{}
	if !decodeBody(w, r, &req) {
		return
	}

	appErr = a.app.UpdateSecuritySettings(*user, req.PasswordSettings, req.SessionLength)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getLicense(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	info, appErr := a.app.GetLicense(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, info)
}

func (a *API) getPasswordPolicy(w http.ResponseWriter, r *http.Request) {
	policy, appErr := a.app.GetPasswordPolicy()
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	respondJSON(w, http.StatusOK, policy)
}

func (a *API) getStorages(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	storages, appErr := a.app.GetStorages(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, storages)
}

func (a *API) createStorage(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.Storage
	if !decodeBody(w, r, &req) {
		return
	}

	created, appErr := a.app.CreateStorage(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, created)
}

func (a *API) updateStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var req model.Storage
	if !decodeBody(w, r, &req) {
		return
	}

	updated, appErr := a.app.UpdateStorage(r.Context(), *user, id, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (a *API) deleteStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteStorage(r.Context(), *user, id); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) setPrimaryStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.SetPrimaryStorage(*user, id); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getOAuthSettings(w http.ResponseWriter, r *http.Request) {
	settings, appErr := a.app.GetPublicAuthSettings()
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

func (a *API) getSystemRoles(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	roles, appErr := a.app.GetSystemRoles(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

func (a *API) getSystemPermissions(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	sections, appErr := a.app.GetSystemPermissions(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, sections)
}

func (a *API) createSystemRole(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var role model.Role
	if !decodeBody(w, r, &role) {
		return
	}

	created, appErr := a.app.CreateSystemRole(*user, role)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, created)
}

func (a *API) updateSystemRole(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	var patch model.RolePatch
	if !decodeBody(w, r, &patch) {
		return
	}

	updated, appErr := a.app.UpdateSystemRole(*user, id, patch)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (a *API) deleteSystemRole(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteSystemRole(*user, id); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
