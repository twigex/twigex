// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

const passwordUnchangedSentinel = "********"

func (a *App) GetAuthSettings(user model.User) (*model.AdminAuthSettings, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	cfg := a.ConfigStore.Config
	ldap := cfg.LDAPSettings

	allowedGroups := ""
	if ldap.AllowedGroups != nil {
		allowedGroups = strings.Join(*ldap.AllowedGroups, "|")
	}

	target := ldap.DialTarget()

	ldapEnabled := *ldap.Enabled && a.Server.License.HasLDAP()
	oidcEnabled := *cfg.AuthSettings.OidcEnabled && a.Server.License.HasOAuthProviders()

	syncStatus, err := a.LDAPSyncStatus()
	if err != nil {
		tlog.Errorw("Could not read the last LDAP sync", "error", err)
	}

	return &model.AdminAuthSettings{
		LocalAuthEnabled: *cfg.AuthSettings.LocalAuthEnabled,
		LdapEnabled:      ldapEnabled,
		OidcEnabled:      oidcEnabled,
		OidcFirst:        *cfg.AuthSettings.OidcFirst,
		LDAP: model.AdminLDAPSettings{
			Host:                     target.Host,
			Port:                     target.Port,
			BaseDN:                   *ldap.BaseDN,
			BindDN:                   *ldap.BindDN,
			BindPassword:             passwordUnchangedSentinel,
			UserFilter:               *ldap.UserFilter,
			UsernameAttr:             *ldap.UsernameAttribute,
			IDAttr:                   *ldap.IDAttribute,
			FirstNameAttr:            *ldap.FirstNameAttribute,
			LastNameAttr:             *ldap.LastNameAttribute,
			EmailAttr:                *ldap.EmailAttribute,
			AllowedGroups:            allowedGroups,
			InsecureSkipVerify:       *ldap.InsecureSkipVerify,
			ConnectionSecurity:       *ldap.ConnectionSecurity,
			CACertificate:            *ldap.CACertificate,
			TimeoutSeconds:           *ldap.TimeoutSeconds,
			SyncEnabled:              *ldap.SyncEnabled,
			SyncIntervalMinutes:      *ldap.SyncIntervalMinutes,
			SyncMaxDeactivatePercent: *ldap.SyncMaxDeactivatePercent,
		},
		LdapEnvLocked:  a.ConfigStore.EnvLocked.LDAP,
		LDAPSyncStatus: syncStatus,
	}, nil
}

func (a *App) UpdateAuthSettings(user model.User, in model.AdminAuthSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	// Mmust have at least one method enabled.
	// Check against the DB for OIDC since enabling OIDC without providers is equivalent to no OIDC.
	oidcEffective := in.OidcEnabled
	if oidcEffective {
		count, err := a.Store.OIDC.Count()
		if err != nil {
			tlog.Errorw("Failed to count OIDC providers",
				"error", err,
			)
			return model.NewAppError("settings.retrieval_failed", http.StatusInternalServerError)
		}

		if count == 0 {
			oidcEffective = false
		}
	}

	if !in.LocalAuthEnabled && !in.LdapEnabled && !oidcEffective {
		return model.NewAppError("auth.no_methods_enabled", http.StatusBadRequest)
	}

	if in.LdapEnabled && !a.Server.License.HasLDAP() {
		return model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	if !model.ValidLDAPSecurity(in.LDAP.ConnectionSecurity) {
		return model.NewAppError("ldap.invalid_connection_security", http.StatusBadRequest)
	}

	if !model.ValidLDAPTimeout(in.LDAP.TimeoutSeconds) {
		return model.NewAppError("ldap.invalid_timeout", http.StatusBadRequest)
	}

	if !model.ValidLDAPSyncInterval(in.LDAP.SyncIntervalMinutes) {
		return model.NewAppError("ldap.invalid_sync_interval", http.StatusBadRequest)
	}

	if !model.ValidLDAPDeactivatePercent(in.LDAP.SyncMaxDeactivatePercent) {
		return model.NewAppError("ldap.invalid_sync_deactivate_percent", http.StatusBadRequest)
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	batch := map[string]string{
		"auth.local_enabled": strconv.FormatBool(in.LocalAuthEnabled),
		"auth.oidc_enabled":  strconv.FormatBool(in.OidcEnabled),
		"auth.oidc_first":    strconv.FormatBool(in.OidcFirst),
	}

	if !a.ConfigStore.EnvLocked.LDAP {
		batch["ldap.enabled"] = strconv.FormatBool(in.LdapEnabled)
		batch["ldap.host"] = in.LDAP.Host
		batch["ldap.port"] = strconv.Itoa(in.LDAP.Port)
		batch["ldap.base_dn"] = in.LDAP.BaseDN
		batch["ldap.bind_dn"] = in.LDAP.BindDN
		batch["ldap.user_filter"] = in.LDAP.UserFilter
		batch["ldap.username_attr"] = in.LDAP.UsernameAttr
		batch["ldap.id_attr"] = in.LDAP.IDAttr
		batch["ldap.first_name_attr"] = in.LDAP.FirstNameAttr
		batch["ldap.last_name_attr"] = in.LDAP.LastNameAttr
		batch["ldap.email_attr"] = in.LDAP.EmailAttr
		batch["ldap.allowed_groups"] = in.LDAP.AllowedGroups
		batch["ldap.insecure_skip_verify"] = strconv.FormatBool(in.LDAP.InsecureSkipVerify)
		batch["ldap.connection_security"] = in.LDAP.ConnectionSecurity
		batch["ldap.ca_cert"] = in.LDAP.CACertificate
		batch["ldap.timeout_seconds"] = strconv.Itoa(in.LDAP.TimeoutSeconds)
		batch["ldap.sync_enabled"] = strconv.FormatBool(in.LDAP.SyncEnabled)
		batch["ldap.sync_interval_minutes"] = strconv.Itoa(in.LDAP.SyncIntervalMinutes)
		batch["ldap.sync_max_deactivate_percent"] = strconv.Itoa(in.LDAP.SyncMaxDeactivatePercent)

		// LDAP bind password "********" means unchanged
		switch in.LDAP.BindPassword {
		case passwordUnchangedSentinel:
			existing, err := a.Store.SystemSettings.Get("ldap.bind_password")
			if err != nil {
				tlog.Errorw("Failed to retrieve existing LDAP bind password",
					"error", err,
				)
				return model.NewAppError("settings.retrieval_failed", http.StatusInternalServerError)
			}

			batch["ldap.bind_password"] = existing
		case "":
			batch["ldap.bind_password"] = ""
		default:
			encrypted, err := crypto.Encrypt(encryptKey, in.LDAP.BindPassword)
			if err != nil {
				tlog.Errorw("Failed to encrypt LDAP bind password",
					"error", err,
				)
				return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
			}

			batch["ldap.bind_password"] = encrypted
		}
	}

	if err := a.Store.SystemSettings.UpdateBatch(batch); err != nil {
		tlog.Errorw("Failed to update auth settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	// Update in-memory config
	cfg := a.ConfigStore.Config
	*cfg.AuthSettings.LocalAuthEnabled = in.LocalAuthEnabled
	*cfg.AuthSettings.OidcEnabled = in.OidcEnabled
	*cfg.AuthSettings.OidcFirst = in.OidcFirst

	if !a.ConfigStore.EnvLocked.LDAP {
		*cfg.LDAPSettings.Enabled = in.LdapEnabled
		*cfg.LDAPSettings.Host = in.LDAP.Host
		*cfg.LDAPSettings.Port = in.LDAP.Port
		*cfg.LDAPSettings.BaseDN = in.LDAP.BaseDN
		*cfg.LDAPSettings.BindDN = in.LDAP.BindDN
		*cfg.LDAPSettings.UserFilter = in.LDAP.UserFilter
		*cfg.LDAPSettings.UsernameAttribute = in.LDAP.UsernameAttr
		*cfg.LDAPSettings.IDAttribute = in.LDAP.IDAttr
		*cfg.LDAPSettings.FirstNameAttribute = in.LDAP.FirstNameAttr
		*cfg.LDAPSettings.LastNameAttribute = in.LDAP.LastNameAttr
		*cfg.LDAPSettings.EmailAttribute = in.LDAP.EmailAttr
		*cfg.LDAPSettings.InsecureSkipVerify = in.LDAP.InsecureSkipVerify
		*cfg.LDAPSettings.ConnectionSecurity = in.LDAP.ConnectionSecurity
		*cfg.LDAPSettings.CACertificate = in.LDAP.CACertificate
		*cfg.LDAPSettings.TimeoutSeconds = in.LDAP.TimeoutSeconds
		*cfg.LDAPSettings.SyncEnabled = in.LDAP.SyncEnabled
		*cfg.LDAPSettings.SyncIntervalMinutes = in.LDAP.SyncIntervalMinutes
		*cfg.LDAPSettings.SyncMaxDeactivatePercent = in.LDAP.SyncMaxDeactivatePercent

		groups := splitAllowedGroups(in.LDAP.AllowedGroups)
		cfg.LDAPSettings.AllowedGroups = &groups

		if in.LDAP.BindPassword != passwordUnchangedSentinel {
			*cfg.LDAPSettings.BindPassword = in.LDAP.BindPassword
		}
	}

	return nil
}

func splitAllowedGroups(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}

	parts := strings.Split(s, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}

	return out
}

func (a *App) TestLDAPConnection(user model.User, in model.AdminLDAPSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.LDAPAuth == nil {
		return model.NewAppError("auth.ldap_unavailable", http.StatusNotImplemented)
	}

	if in.BindPassword == passwordUnchangedSentinel {
		in.BindPassword = *a.ConfigStore.Config.LDAPSettings.BindPassword
	}

	return a.LDAPAuth.TestConnection(in)
}

// oidcRedirectURL computes the callback URL for a provider from SiteURL.
func (a *App) OidcRedirectURL(providerID string) string {
	siteURL := strings.TrimSuffix(*a.ConfigStore.Config.ServerSettings.SiteURL, "/")
	return fmt.Sprintf("%s/api/auth/oidc/callback/%s", siteURL, providerID)
}

func (a *App) GetOIDCProviders(user model.User) ([]model.OIDCProvider, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.Server.License.HasOAuthProviders() {
		return nil, model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	providers, err := a.Store.OIDC.GetAll()
	if err != nil {
		tlog.Errorw("Failed to retrieve OIDC providers",
			"error", err,
		)
		return nil, model.NewAppError("oidc.retrieval_failed", http.StatusInternalServerError)
	}

	for i := range providers {
		providers[i].ClientSecret = ""
		providers[i].RedirectURL = a.OidcRedirectURL(providers[i].ID)
	}

	return providers, nil
}

type PreparedOIDCProvider struct {
	ID          string `json:"id"`
	RedirectURL string `json:"redirect_url"`
}

func (a *App) PrepareOIDCProvider(user model.User) (*PreparedOIDCProvider, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.Server.License.HasOAuthProviders() {
		return nil, model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	id := model.NewID()
	return &PreparedOIDCProvider{
		ID:          id,
		RedirectURL: a.OidcRedirectURL(id),
	}, nil
}

func (a *App) CreateOIDCProvider(user model.User, p model.OIDCProvider) (*model.OIDCProvider, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.Server.License.HasOAuthProviders() {
		return nil, model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	// RedirectURL is computed, not accepted from the client.
	if strings.TrimSpace(p.Name) == "" ||
		strings.TrimSpace(p.DiscoveryURL) == "" ||
		strings.TrimSpace(p.ClientID) == "" ||
		strings.TrimSpace(p.ClientSecret) == "" {
		return nil, model.NewAppError("oidc.invalid_input", http.StatusBadRequest)
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey
	encrypted, err := crypto.Encrypt(encryptKey, p.ClientSecret)
	if err != nil {
		tlog.Errorw("Failed to encrypt OIDC client secret",
			"error", err,
		)
		return nil, model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
	}

	p.ClientSecret = encrypted

	created, err := a.Store.OIDC.Create(&p)
	if err != nil {
		tlog.Errorw("Failed to create OIDC provider",
			"error", err,
		)
		return nil, model.NewAppError("oidc.create_failed", http.StatusInternalServerError)
	}

	created.ClientSecret = ""
	created.RedirectURL = a.OidcRedirectURL(created.ID)
	return created, nil
}

func (a *App) UpdateOIDCProvider(user model.User, id string, p model.OIDCProvider) (*model.OIDCProvider, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.Server.License.HasOAuthProviders() {
		return nil, model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	existing, err := a.Store.OIDC.GetByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve OIDC provider",
			"provider_id", id,
			"error", err,
		)
		return nil, model.NewAppError("oidc.retrieval_failed", http.StatusInternalServerError)
	}

	if existing == nil {
		return nil, model.NewAppError("oidc.not_found", http.StatusNotFound)
	}

	// Blank fields fall back to existing
	p.ID = existing.ID
	p.CreatedAt = existing.CreatedAt
	if p.Name == "" {
		p.Name = existing.Name
	}

	if p.DiscoveryURL == "" {
		p.DiscoveryURL = existing.DiscoveryURL
	}

	if p.ClientID == "" {
		p.ClientID = existing.ClientID
	}

	if p.Scopes == "" {
		p.Scopes = existing.Scopes
	}

	if _, err := a.Store.OIDC.Update(&p); err != nil {
		tlog.Errorw("Failed to update OIDC provider",
			"provider_id", id,
			"error", err,
		)
		return nil, model.NewAppError("oidc.update_failed", http.StatusInternalServerError)
	}

	// Secret is updated separately, only if non-blank
	if strings.TrimSpace(p.ClientSecret) != "" {
		encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey
		encrypted, err := crypto.Encrypt(encryptKey, p.ClientSecret)
		if err != nil {
			tlog.Errorw("Failed to encrypt OIDC client secret",
				"provider_id", id,
				"error", err,
			)
			return nil, model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		if err := a.Store.OIDC.UpdateSecret(id, encrypted); err != nil {
			tlog.Errorw("Failed to update OIDC client secret",
				"provider_id", id,
				"error", err,
			)
			return nil, model.NewAppError("oidc.update_failed", http.StatusInternalServerError)
		}
	}

	p.ClientSecret = ""
	p.RedirectURL = a.OidcRedirectURL(p.ID)
	return &p, nil
}

func (a *App) DeleteOIDCProvider(user model.User, id string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if !a.Server.License.HasOAuthProviders() {
		return model.NewAppError("license.feature_unavailable", http.StatusPaymentRequired)
	}

	existing, err := a.Store.OIDC.GetByID(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve OIDC provider",
			"provider_id", id,
			"error", err,
		)
		return model.NewAppError("oidc.retrieval_failed", http.StatusInternalServerError)
	}

	if existing == nil {
		return model.NewAppError("oidc.not_found", http.StatusNotFound)
	}

	if err := a.Store.OIDC.Delete(id); err != nil {
		tlog.Errorw("Failed to delete OIDC provider",
			"provider_id", id,
			"error", err,
		)
		return model.NewAppError("oidc.delete_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetPublicAuthSettings() (*model.PublicAuthSettings, *model.AppError) {
	cfg := a.ConfigStore.Config

	ldapLicensed := a.Server.License.HasLDAP()
	oidcLicensed := a.Server.License.HasOAuthProviders()

	out := &model.PublicAuthSettings{
		LocalAuthEnabled:     *cfg.AuthSettings.LocalAuthEnabled,
		LdapEnabled:          *cfg.LDAPSettings.Enabled && ldapLicensed,
		OidcFirst:            *cfg.AuthSettings.OidcFirst,
		PasswordResetEnabled: a.emailEnabled(),
		OIDCProviders:        []model.PublicOIDCProvider{},
	}

	if *cfg.AuthSettings.OidcEnabled && oidcLicensed {
		providers, err := a.Store.OIDC.GetAllEnabled()
		if err != nil {
			tlog.Errorw("Failed to retrieve enabled OIDC providers",
				"error", err,
			)
			return nil, model.NewAppError("oidc.retrieval_failed", http.StatusInternalServerError)
		}

		for _, p := range providers {
			out.OIDCProviders = append(out.OIDCProviders, model.PublicOIDCProvider{
				ID:          p.ID,
				Name:        p.Name,
				ButtonText:  p.ButtonText,
				ButtonColor: p.ButtonColor,
				LoginURL:    fmt.Sprintf("/api/auth/oidc/login/%s", p.ID),
			})
		}
	}

	return out, nil
}
