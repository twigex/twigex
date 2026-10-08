// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/klipy"
	"github.com/twigex/twigex/internal/parse"
	"github.com/twigex/twigex/internal/tmail"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetEmailServer(user model.User) (*model.EmailSettings, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	settings := a.ConfigStore.Config.EmailSettings
	pw := "********"
	settings.SMTPPassword = &pw
	enabled := a.emailEnabled()
	settings.EnableEmail = &enabled
	settings.EnvLocked = a.ConfigStore.EnvLocked.SMTP
	return &settings, nil
}

func (a *App) UpdateEmailServer(user model.User, settings model.EmailSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.ConfigStore.EnvLocked.SMTP {
		return model.NewAppError("settings.env_locked", http.StatusForbidden)
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	batch := map[string]string{
		"email.smtp_host":         *settings.SMTPServer,
		"email.smtp_port":         *settings.SMTPPort,
		"email.smtp_username":     *settings.SMTPUsername,
		"email.smtp_security":     *settings.ConnectionSecurity,
		"email.smtp_auth":         strconv.FormatBool(*settings.EnableSMTPAuth),
		"email.smtp_from_address": *settings.FromAddress,
		"email.smtp_from_name":    *settings.FromName,
		"email.smtp_reply_to":     *settings.ReplyToAddress,
	}

	if settings.EnableEmail != nil {
		batch["email.enabled"] = strconv.FormatBool(*settings.EnableEmail)
	}

	// "********" means password was not changed
	if *settings.SMTPPassword == "********" {
		existing, err := a.Store.SystemSettings.Get("email.smtp_password")
		if err != nil {
			tlog.Errorw("Failed to retrieve existing SMTP password",
				"error", err,
			)
			return model.NewAppError("settings.retrieval_failed", http.StatusInternalServerError)
		}

		batch["email.smtp_password"] = existing
	} else {
		encryptedPassword, err := crypto.Encrypt(encryptKey, *settings.SMTPPassword)
		if err != nil {
			tlog.Errorw("Failed to encrypt SMTP password",
				"error", err,
			)
			return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		batch["email.smtp_password"] = encryptedPassword
	}

	if err := a.Store.SystemSettings.UpdateBatch(batch); err != nil {
		tlog.Errorw("Failed to update email settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	if settings.EnableEmail != nil {
		a.ConfigStore.Config.EmailSettings.EnableEmail = settings.EnableEmail
	}

	a.ConfigStore.Config.EmailSettings.SMTPServer = settings.SMTPServer
	a.ConfigStore.Config.EmailSettings.SMTPPort = settings.SMTPPort
	a.ConfigStore.Config.EmailSettings.SMTPUsername = settings.SMTPUsername
	a.ConfigStore.Config.EmailSettings.ConnectionSecurity = settings.ConnectionSecurity
	a.ConfigStore.Config.EmailSettings.EnableSMTPAuth = settings.EnableSMTPAuth
	a.ConfigStore.Config.EmailSettings.FromAddress = settings.FromAddress
	a.ConfigStore.Config.EmailSettings.FromName = settings.FromName
	a.ConfigStore.Config.EmailSettings.ReplyToAddress = settings.ReplyToAddress

	if *settings.SMTPPassword != "********" {
		a.ConfigStore.Config.EmailSettings.SMTPPassword = settings.SMTPPassword
	}

	return nil
}

func (a *App) GetOfficeSettings(user model.User) (*model.OfficeSettings, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	settings := a.ConfigStore.Config.OfficeSettings
	secret := "********"
	settings.Secret = &secret
	settings.EnvLocked = a.ConfigStore.EnvLocked.Office
	return &settings, nil
}

func (a *App) UpdateOfficeSettings(user model.User, settings model.OfficeSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.ConfigStore.EnvLocked.Office {
		return model.NewAppError("settings.env_locked", http.StatusForbidden)
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	batch := map[string]string{
		"office.enabled": strconv.FormatBool(*settings.Enable),
		"office.type":    *settings.Type,
		"office.host":    *settings.Host,
	}

	if *settings.Secret == "********" {
		existing, err := a.Store.SystemSettings.Get("office.secret")
		if err != nil {
			tlog.Errorw("Failed to retrieve existing office secret",
				"error", err,
			)
			return model.NewAppError("settings.retrieval_failed", http.StatusInternalServerError)
		}

		batch["office.secret"] = existing
	} else {
		encrypted, err := crypto.Encrypt(encryptKey, *settings.Secret)
		if err != nil {
			tlog.Errorw("Failed to encrypt office secret",
				"error", err,
			)
			return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		batch["office.secret"] = encrypted
	}

	if err := a.Store.SystemSettings.UpdateBatch(batch); err != nil {
		tlog.Errorw("Failed to update office settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	a.ConfigStore.Config.OfficeSettings.Enable = settings.Enable
	a.ConfigStore.Config.OfficeSettings.Type = settings.Type
	a.ConfigStore.Config.OfficeSettings.Host = settings.Host

	if *settings.Secret != "********" {
		a.ConfigStore.Config.OfficeSettings.Secret = settings.Secret
	}

	return nil
}

func (a *App) TestEmailServerConnection(user model.User, settings model.EmailSettings) (bool, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return false, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	status, err := tmail.TestConnection(a.smtpConfigFromSettings(settings))
	if err != nil {
		tlog.Errorw("Failed to test email server connection",
			"error", err,
		)
		return false, model.NewAppError("settings.email_test_failed", http.StatusInternalServerError)
	}

	return status, nil
}

func (a *App) GetSecuritySettings(user model.User) (*struct {
	PasswordSettings *model.PasswordSettings
	SessionLength    *string
}, *model.AppError,
) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	passwordSettings := a.GetPasswordSettings()

	return &struct {
		PasswordSettings *model.PasswordSettings
		SessionLength    *string
	}{
		PasswordSettings: passwordSettings,
		SessionLength:    a.ConfigStore.Config.ServerSettings.SessionLengthInDays,
	}, nil
}

func (a *App) GetPasswordSettings() *model.PasswordSettings {
	return &a.ConfigStore.Config.PasswordSettings
}

func (a *App) UpdateSecuritySettings(user model.User, passwordSettings model.PasswordSettings, sessionLength string) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	minLength := strconv.Itoa(*passwordSettings.PasswordMinLength)
	err := a.Store.SystemSettings.UpdateBatch(map[string]string{
		"password.min_length":         minLength,
		"password.require_mixed_case": strconv.FormatBool(*passwordSettings.UpperLowerCharacters),
		"password.require_number":     strconv.FormatBool(*passwordSettings.NumericCharacters),
		"password.require_symbol":     strconv.FormatBool(*passwordSettings.SpecialCharacters),
		"server.session_length_days":  sessionLength,
	})
	if err != nil {
		tlog.Errorw("Failed to update security settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	// Update in-memory config so changes take effect immediately without restart
	a.ConfigStore.Config.PasswordSettings = passwordSettings
	a.ConfigStore.Config.ServerSettings.SessionLengthInDays = &sessionLength

	return nil
}

func (a *App) GetPasswordPolicy() (*model.PasswordSettings, *model.AppError) {
	return a.GetPasswordSettings(), nil
}

func (a *App) GetLicense(user model.User) (*model.LicenseInfo, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	return &model.LicenseInfo{
		License:   a.Server.License,
		EnvLocked: a.ConfigStore.EnvLocked.License,
	}, nil
}

func (a *App) GetChatConfig(user model.User) (*model.ChatSettings, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	return &model.ChatSettings{
		VideoSettings: model.VideoSettings{
			Enabled: *a.ConfigStore.Config.ChannelSettings.Enabled,
			Host:    *a.ConfigStore.Config.ChannelSettings.Host,
			Key:     *a.ConfigStore.Config.ChannelSettings.Key,
		},
		GIFSettings: model.GIFSettings{
			Enabled: *a.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled,
		},
		LinkPreviewSettings: model.LinkPreviewSettings{
			Enabled: *a.ConfigStore.Config.ChannelSettings.LinkPreviews,
		},
		VideoEnvLocked:       a.ConfigStore.EnvLocked.Video,
		GIFEnvLocked:         a.ConfigStore.EnvLocked.GIF,
		LinkPreviewEnvLocked: a.ConfigStore.EnvLocked.LinkPreview,
	}, nil
}

func (a *App) UpdateChatSettings(user model.User, settings model.ChatSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.ConfigStore.EnvLocked.Video || a.ConfigStore.EnvLocked.GIF || a.ConfigStore.EnvLocked.LinkPreview {
		return model.NewAppError("settings.env_locked", http.StatusForbidden)
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey

	batch := map[string]string{
		"channel.enabled":       strconv.FormatBool(settings.VideoSettings.Enabled),
		"channel.host":          settings.VideoSettings.Host,
		"channel.link_previews": strconv.FormatBool(settings.LinkPreviewSettings.Enabled),
		"channel.gif_enabled":   strconv.FormatBool(settings.GIFSettings.Enabled),
	}

	if settings.VideoSettings.Key != "" {
		encryptedKey, err := crypto.Encrypt(encryptKey, settings.VideoSettings.Key)
		if err != nil {
			tlog.Errorw("Failed to encrypt Livekit key", "error", err)
			return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		batch["channel.key"] = encryptedKey
	}

	if settings.VideoSettings.Secret != "" {
		encryptedSecret, err := crypto.Encrypt(encryptKey, settings.VideoSettings.Secret)
		if err != nil {
			tlog.Errorw("Failed to encrypt Livekit secret", "error", err)
			return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		batch["channel.secret"] = encryptedSecret
	}

	if settings.GIFSettings.APIKey != "" {
		encryptedAPIKey, err := crypto.Encrypt(encryptKey, strings.TrimSpace(settings.GIFSettings.APIKey))
		if err != nil {
			tlog.Errorw("Failed to encrypt GIF API key", "error", err)
			return model.NewAppError("settings.encrypt_failed", http.StatusInternalServerError)
		}

		batch["channel.gif_api_key"] = encryptedAPIKey
	}

	if err := a.Store.SystemSettings.UpdateBatch(batch); err != nil {
		tlog.Errorw("Failed to update chat settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	// Update in-memory config immediately
	*a.ConfigStore.Config.ChannelSettings.Enabled = settings.VideoSettings.Enabled
	*a.ConfigStore.Config.ChannelSettings.Host = settings.VideoSettings.Host
	*a.ConfigStore.Config.ChannelSettings.LinkPreviews = settings.LinkPreviewSettings.Enabled
	*a.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled = settings.GIFSettings.Enabled

	if settings.VideoSettings.Key != "" {
		*a.ConfigStore.Config.ChannelSettings.Key = settings.VideoSettings.Key
	}

	if settings.VideoSettings.Secret != "" {
		*a.ConfigStore.Config.ChannelSettings.Secret = settings.VideoSettings.Secret
	}

	if settings.GIFSettings.APIKey != "" {
		*a.ConfigStore.Config.ChannelSettings.KlipyGIF.APIKey = strings.TrimSpace(settings.GIFSettings.APIKey)
	}

	// Update GIF client in memory immediately
	if settings.GIFSettings.Enabled {
		key := strings.TrimSpace(settings.GIFSettings.APIKey)
		if a.GIF == nil {
			a.GIF = klipy.New(key)
		} else {
			a.GIF.APIKey = key
		}
	}

	return nil
}

func (a *App) LoadSystemSettings() error {
	all, err := a.Store.SystemSettings.GetByPrefix("")
	if err != nil {
		return err
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey
	cfg := a.ConfigStore.Config

	// Each setting use DB value if exists, otherwise keep env var default
	if v := all["password.min_length"]; v != "" {
		minLen := parse.Int(v, 8)
		cfg.PasswordSettings.PasswordMinLength = &minLen
	}

	if v := all["password.require_mixed_case"]; v != "" {
		b := parse.Bool(v, true)
		cfg.PasswordSettings.UpperLowerCharacters = &b
	}

	if v := all["password.require_number"]; v != "" {
		b := parse.Bool(v, true)
		cfg.PasswordSettings.NumericCharacters = &b
	}

	if v := all["password.require_symbol"]; v != "" {
		b := parse.Bool(v, false)
		cfg.PasswordSettings.SpecialCharacters = &b
	}

	if v := all["server.session_length_days"]; v != "" {
		cfg.ServerSettings.SessionLengthInDays = &v
	}

	// When env-locked, env values must win, so ignore the stored GUI values.
	if !a.ConfigStore.EnvLocked.SMTP {
		if v := all["email.enabled"]; v != "" {
			b := parse.Bool(v, true)
			cfg.EmailSettings.EnableEmail = &b
		}

		if v := all["email.smtp_host"]; v != "" {
			cfg.EmailSettings.SMTPServer = &v
		}

		if v := all["email.smtp_port"]; v != "" {
			cfg.EmailSettings.SMTPPort = &v
		}

		if v := all["email.smtp_username"]; v != "" {
			cfg.EmailSettings.SMTPUsername = &v
		}

		if v := all["email.smtp_security"]; v != "" {
			cfg.EmailSettings.ConnectionSecurity = &v
		}

		if v := all["email.smtp_auth"]; v != "" {
			b := parse.Bool(v, true)
			cfg.EmailSettings.EnableSMTPAuth = &b
		}

		if v := all["email.smtp_password"]; v != "" {
			password, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt SMTP password: %w", err)
			}

			cfg.EmailSettings.SMTPPassword = &password
		}

		if v := all["email.smtp_from_address"]; v != "" {
			cfg.EmailSettings.FromAddress = &v
		}

		if v := all["email.smtp_from_name"]; v != "" {
			cfg.EmailSettings.FromName = &v
		}

		if v := all["email.smtp_reply_to"]; v != "" {
			cfg.EmailSettings.ReplyToAddress = &v
		}
	}

	if !a.ConfigStore.EnvLocked.Video {
		if v := all["channel.enabled"]; v != "" {
			b := parse.Bool(v, false)
			cfg.ChannelSettings.Enabled = &b
		}

		if v := all["channel.host"]; v != "" {
			cfg.ChannelSettings.Host = &v
		}

		if v := all["channel.key"]; v != "" {
			key, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt channel key: %w", err)
			}

			cfg.ChannelSettings.Key = &key
		}

		if v := all["channel.secret"]; v != "" {
			secret, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt channel secret: %w", err)
			}

			cfg.ChannelSettings.Secret = &secret
		}
	}

	if !a.ConfigStore.EnvLocked.LinkPreview {
		if v := all["channel.link_previews"]; v != "" {
			b := parse.Bool(v, true)
			cfg.ChannelSettings.LinkPreviews = &b
		}
	}

	if !a.ConfigStore.EnvLocked.GIF {
		if v := all["channel.gif_enabled"]; v != "" {
			b := parse.Bool(v, false)
			cfg.ChannelSettings.KlipyGIF.Enabled = &b
		}

		if v := all["channel.gif_api_key"]; v != "" {
			apiKey, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt GIF API key: %w", err)
			}

			cfg.ChannelSettings.KlipyGIF.APIKey = &apiKey
		}
	}

	if !a.ConfigStore.EnvLocked.Office {
		if v := all["office.enabled"]; v != "" {
			b := parse.Bool(v, false)
			cfg.OfficeSettings.Enable = &b
		}

		if v := all["office.type"]; v != "" {
			cfg.OfficeSettings.Type = &v
		}

		if v := all["office.host"]; v != "" {
			cfg.OfficeSettings.Host = &v
		}

		if v := all["office.secret"]; v != "" {
			secret, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt office secret: %w", err)
			}

			cfg.OfficeSettings.Secret = &secret
		}
	}

	if v := all["server.default_locale"]; v != "" {
		cfg.ServerSettings.DefaultLocale = &v
	}

	if v := all["server.allow_user_language_override"]; v != "" {
		b := parse.Bool(v, true)
		cfg.ServerSettings.AllowUserLanguageOverride = &b
	}

	if v := all["auth.local_enabled"]; v != "" {
		b := parse.Bool(v, true)
		cfg.AuthSettings.LocalAuthEnabled = &b
	}

	if v := all["auth.oidc_enabled"]; v != "" {
		b := parse.Bool(v, false)
		cfg.AuthSettings.OidcEnabled = &b
	}

	if v := all["auth.oidc_first"]; v != "" {
		b := parse.Bool(v, false)
		cfg.AuthSettings.OidcFirst = &b
	}

	if !a.ConfigStore.EnvLocked.LDAP {
		if v := all["ldap.enabled"]; v != "" {
			b := parse.Bool(v, false)
			cfg.LDAPSettings.Enabled = &b
		}

		if v := all["ldap.base_dn"]; v != "" {
			cfg.LDAPSettings.BaseDN = &v
		}

		if v := all["ldap.bind_dn"]; v != "" {
			cfg.LDAPSettings.BindDN = &v
		}

		if v := all["ldap.bind_password"]; v != "" {
			password, err := crypto.Decrypt(encryptKey, v)
			if err != nil {
				return fmt.Errorf("failed to decrypt LDAP bind password: %w", err)
			}

			cfg.LDAPSettings.BindPassword = &password
		}

		if v := all["ldap.user_filter"]; v != "" {
			cfg.LDAPSettings.UserFilter = &v
		}

		if v := all["ldap.username_attr"]; v != "" {
			cfg.LDAPSettings.UsernameAttribute = &v
		}

		if v := all["ldap.id_attr"]; v != "" {
			cfg.LDAPSettings.IDAttribute = &v
		}

		if v := all["ldap.first_name_attr"]; v != "" {
			cfg.LDAPSettings.FirstNameAttribute = &v
		}

		if v := all["ldap.last_name_attr"]; v != "" {
			cfg.LDAPSettings.LastNameAttribute = &v
		}

		if v := all["ldap.email_attr"]; v != "" {
			cfg.LDAPSettings.EmailAttribute = &v
		}

		if v := all["ldap.allowed_groups"]; v != "" {
			groups := splitAllowedGroups(v)
			cfg.LDAPSettings.AllowedGroups = &groups
		}

		if v := all["ldap.insecure_skip_verify"]; v != "" {
			b := parse.Bool(v, false)
			cfg.LDAPSettings.InsecureSkipVerify = &b
		}

		if v := all["ldap.host"]; v != "" {
			cfg.LDAPSettings.Host = &v
		}

		if v := all["ldap.port"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				cfg.LDAPSettings.Port = &n
			}
		}

		if v := all["ldap.connection_security"]; v != "" {
			cfg.LDAPSettings.ConnectionSecurity = &v
		}

		if v := all["ldap.ca_cert"]; v != "" {
			cfg.LDAPSettings.CACertificate = &v
		}

		if v := all["ldap.sync_enabled"]; v != "" {
			b := parse.Bool(v, false)
			cfg.LDAPSettings.SyncEnabled = &b
		}

		if v := all["ldap.sync_interval_minutes"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil && model.ValidLDAPSyncInterval(n) {
				cfg.LDAPSettings.SyncIntervalMinutes = &n
			}
		}

		if v := all["ldap.sync_max_deactivate_percent"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil && model.ValidLDAPDeactivatePercent(n) {
				cfg.LDAPSettings.SyncMaxDeactivatePercent = &n
			}
		}

		if v := all["ldap.timeout_seconds"]; v != "" {
			if n, err := strconv.Atoi(v); err == nil && model.ValidLDAPTimeout(n) {
				cfg.LDAPSettings.TimeoutSeconds = &n
			}
		}
	}

	return nil
}

func (a *App) GetLanguageSettings(user model.User) (*model.LanguageSettings, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	return &model.LanguageSettings{
		DefaultLocale:     *a.ConfigStore.Config.ServerSettings.DefaultLocale,
		AllowUserOverride: *a.ConfigStore.Config.ServerSettings.AllowUserLanguageOverride,
	}, nil
}

func (a *App) UpdateLanguageSettings(user model.User, settings model.LanguageSettings) *model.AppError {
	if user.Role != model.SystemAdminRoleId {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if err := a.Store.SystemSettings.UpdateBatch(map[string]string{
		"server.default_locale":               settings.DefaultLocale,
		"server.allow_user_language_override": strconv.FormatBool(settings.AllowUserOverride),
	}); err != nil {
		tlog.Errorw("Failed to update language settings",
			"error", err,
		)
		return model.NewAppError("settings.update_failed", http.StatusInternalServerError)
	}

	a.ConfigStore.Config.ServerSettings.DefaultLocale = &settings.DefaultLocale
	a.ConfigStore.Config.ServerSettings.AllowUserLanguageOverride = &settings.AllowUserOverride

	return nil
}

func generateMissingNotifications(userID string, preferences []model.Preference) []model.Preference {
	missing := make([]model.Preference, 0)

	for k, v := range model.DefaultNotifications {
		exists := false
		for _, preference := range preferences {
			if k == preference.Name {
				exists = true
			}
		}

		if !exists {
			missing = append(missing, model.Preference{
				UserID:   userID,
				Category: model.PreferencesCategoryNotifications,
				Name:     k,
				Value:    v,
			})
		}
	}

	return missing
}
