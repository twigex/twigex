// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/twigex/twigex/model"
)

var IsDev bool = false

const (
	StorageConfigPath = "config/storage.json"
)

type EnvLockedSettings struct {
	SMTP        bool
	Office      bool
	Video       bool
	GIF         bool
	LinkPreview bool
	LDAP        bool
	License     bool
}

type ConfigStore struct {
	Config    *model.ServerConfig
	EnvLocked EnvLockedSettings
}

// NewConfig creates a new ConfigStore entirely from environment variables.
// For local development, use a .env file loaded via godotenv.Load() in main.go.
func NewConfig() (*ConfigStore, error) {
	config := model.ServerConfig{}
	config.SetDefaults()

	l := &ConfigStore{Config: &config}
	l.loadEnvVariables()

	if err := l.validate(); err != nil {
		return nil, err
	}

	return l, nil
}

func (c *ConfigStore) validate() error {
	if *c.Config.SqlSettings.DataSource == "" {
		return fmt.Errorf(
			"database not configured — set TWIGEX_DB_DATASOURCE " +
				"or TWIGEX_DB_HOST + TWIGEX_DB_NAME + TWIGEX_DB_USER + TWIGEX_DB_PASSWORD",
		)
	}

	if *c.Config.ServerSettings.AtRestEncryptKey == "" {
		return fmt.Errorf(
			"TWIGEX_SERVER_ENCRYPT_KEY is required — " +
				"provide a 32-character key for encrypting sensitive values in DB, " +
				"generate with: openssl rand -hex 16",
		)
	}

	return nil
}

func (c *ConfigStore) loadEnvVariables() {
	c.loadServerSettings()
	c.loadDatabaseSettings()
	c.loadRedisSettings()
	c.loadEmailSettings()
	c.loadOfficeSettings()
	c.loadChannelSettings()
	c.loadLDAPSettings()
	c.loadCollimatoSettings()
	c.computeEnvLocked()
}

func (c *ConfigStore) loadServerSettings() {
	setStr(&c.Config.ServerSettings.SiteURL, "TWIGEX_SERVER_SITE_URL")
	setStr(&c.Config.ServerSettings.Port, "TWIGEX_SERVER_PORT")
	setStr(&c.Config.ServerSettings.AtRestEncryptKey, "TWIGEX_SERVER_ENCRYPT_KEY")
	setStr(&c.Config.ServerSettings.TlsCertFile, "TWIGEX_SERVER_TLS_CERT_FILE")
	setStr(&c.Config.ServerSettings.TlsKeyFile, "TWIGEX_SERVER_TLS_KEY_FILE")
	setStr(&c.Config.ServerSettings.DefaultLocale, "TWIGEX_SERVER_DEFAULT_LOCALE")
	setStr(&c.Config.ServerSettings.SessionLengthInDays, "TWIGEX_SERVER_SESSION_LENGTH_DAYS")
	setInt(&c.Config.FileSettings.MaxPublicUploadSize, "TWIGEX_MAX_PUBLIC_UPLOAD_SIZE")
	setStrSlice(&c.Config.ServerSettings.TrustedProxyHeaders, "TWIGEX_TRUSTED_PROXY_HEADERS")
	setStrSlice(&c.Config.ServerSettings.TrustedProxies, "TWIGEX_TRUSTED_PROXIES")
}

func (c *ConfigStore) loadDatabaseSettings() {
	host := os.Getenv("TWIGEX_DB_HOST")
	port := os.Getenv("TWIGEX_DB_PORT")
	user := os.Getenv("TWIGEX_DB_USER")
	password := os.Getenv("TWIGEX_DB_PASSWORD")
	name := os.Getenv("TWIGEX_DB_NAME")
	readTimeout := getEnvOrDefault("TWIGEX_DB_READ_TIMEOUT", "30s")
	writeTimeout := getEnvOrDefault("TWIGEX_DB_WRITE_TIMEOUT", "30s")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&readTimeout=%s&writeTimeout=%s",
		user, password, host, port, name, readTimeout, writeTimeout,
	)
	c.Config.SqlSettings.DataSource = &dsn
	setStr(&c.Config.SqlSettings.DriverName, "TWIGEX_DB_DRIVER")
	setInt(&c.Config.SqlSettings.MaxOpenConns, "TWIGEX_DB_MAX_OPEN_CONNS")
	setInt(&c.Config.SqlSettings.MaxIdleConns, "TWIGEX_DB_MAX_IDLE_CONNS")
	setInt(&c.Config.SqlSettings.ConnMaxLifetimeMinutes, "TWIGEX_DB_CONN_MAX_LIFETIME_MINUTES")
	setInt(&c.Config.SqlSettings.ConnMaxIdleTimeMinutes, "TWIGEX_DB_CONN_MAX_IDLE_TIME_MINUTES")
	setInt(&c.Config.SqlSettings.QueryTimeout, "TWIGEX_DB_QUERY_TIMEOUT")
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultVal
}

func (c *ConfigStore) loadRedisSettings() {
	setStr(&c.Config.RedisSettings.Address, "TWIGEX_REDIS_ADDRESS")
	setStr(&c.Config.RedisSettings.Password, "TWIGEX_REDIS_PASSWORD")
}

func (c *ConfigStore) loadEmailSettings() {
	setBool(&c.Config.EmailSettings.EnableEmail, "TWIGEX_SMTP_ENABLED")
	setStr(&c.Config.EmailSettings.SMTPServer, "TWIGEX_SMTP_HOST")
	setStr(&c.Config.EmailSettings.SMTPPort, "TWIGEX_SMTP_PORT")
	setStr(&c.Config.EmailSettings.SMTPUsername, "TWIGEX_SMTP_USERNAME")
	setStr(&c.Config.EmailSettings.SMTPPassword, "TWIGEX_SMTP_PASSWORD")
	setStr(&c.Config.EmailSettings.ConnectionSecurity, "TWIGEX_SMTP_SECURITY")
	setBool(&c.Config.EmailSettings.EnableSMTPAuth, "TWIGEX_SMTP_AUTH_ENABLED")
	setStr(&c.Config.EmailSettings.FromAddress, "TWIGEX_SMTP_FROM_ADDRESS")
	setStr(&c.Config.EmailSettings.FromName, "TWIGEX_SMTP_FROM_NAME")
	setStr(&c.Config.EmailSettings.ReplyToAddress, "TWIGEX_SMTP_REPLY_TO")
}

func (c *ConfigStore) loadOfficeSettings() {
	setBool(&c.Config.OfficeSettings.Enable, "TWIGEX_OFFICE_ENABLE")
	setStr(&c.Config.OfficeSettings.Type, "TWIGEX_OFFICE_TYPE")
	setStr(&c.Config.OfficeSettings.Host, "TWIGEX_OFFICE_HOST")
	setStr(&c.Config.OfficeSettings.Secret, "TWIGEX_OFFICE_SECRET")
}

func (c *ConfigStore) loadChannelSettings() {
	setBool(&c.Config.ChannelSettings.Enabled, "TWIGEX_CHANNEL_ENABLED")
	setStr(&c.Config.ChannelSettings.Host, "TWIGEX_CHANNEL_HOST")
	setStr(&c.Config.ChannelSettings.Key, "TWIGEX_CHANNEL_KEY")
	setStr(&c.Config.ChannelSettings.Secret, "TWIGEX_CHANNEL_SECRET")
	setBool(&c.Config.ChannelSettings.LinkPreviews, "TWIGEX_CHANNEL_LINK_PREVIEWS")

	// Klipy GIF (nested in ChannelSettings)
	setBool(&c.Config.ChannelSettings.KlipyGIF.Enabled, "TWIGEX_CHANNEL_GIF_ENABLED")
	setStr(&c.Config.ChannelSettings.KlipyGIF.APIKey, "TWIGEX_CHANNEL_GIF_API_KEY")
}

func (c *ConfigStore) loadCollimatoSettings() {
	setStr(&c.Config.CollimatoSettings.CubeAPIURL, "TWIGEX_COLLIMATO_CUBE_API_URL")
	setStr(&c.Config.CollimatoSettings.CubeAPISecret, "TWIGEX_COLLIMATO_CUBE_API_SECRET")
	setStr(&c.Config.CollimatoSettings.CubeCallbackSecret, "TWIGEX_COLLIMATO_CUBE_CALLBACK_SECRET")
}

func (c *ConfigStore) loadLDAPSettings() {
	setBool(&c.Config.LDAPSettings.Enabled, "TWIGEX_LDAP_ENABLED")
	setStr(&c.Config.LDAPSettings.Host, "TWIGEX_LDAP_HOST")
	setInt(&c.Config.LDAPSettings.Port, "TWIGEX_LDAP_PORT")
	setStr(&c.Config.LDAPSettings.BaseDN, "TWIGEX_LDAP_BASE_DN")
	setStr(&c.Config.LDAPSettings.BindDN, "TWIGEX_LDAP_BIND_DN")
	setStr(&c.Config.LDAPSettings.BindPassword, "TWIGEX_LDAP_BIND_PASSWORD")
	setStr(&c.Config.LDAPSettings.ConnectionSecurity, "TWIGEX_LDAP_CONNECTION_SECURITY")
	setStr(&c.Config.LDAPSettings.CACertificate, "TWIGEX_LDAP_CA_CERT")
	setBool(&c.Config.LDAPSettings.SyncEnabled, "TWIGEX_LDAP_SYNC_ENABLED")
	setInt(&c.Config.LDAPSettings.SyncIntervalMinutes, "TWIGEX_LDAP_SYNC_INTERVAL_MINUTES")
	setInt(&c.Config.LDAPSettings.SyncMaxDeactivatePercent, "TWIGEX_LDAP_SYNC_MAX_DEACTIVATE_PERCENT")
	setInt(&c.Config.LDAPSettings.TimeoutSeconds, "TWIGEX_LDAP_TIMEOUT_SECONDS")
	setBool(&c.Config.LDAPSettings.InsecureSkipVerify, "TWIGEX_LDAP_INSECURE_SKIP_VERIFY")
	setStr(&c.Config.LDAPSettings.UserFilter, "TWIGEX_LDAP_USER_FILTER")
	setStr(&c.Config.LDAPSettings.UsernameAttribute, "TWIGEX_LDAP_USERNAME_ATTR")
	setStr(&c.Config.LDAPSettings.IDAttribute, "TWIGEX_LDAP_ID_ATTR")
	setStr(&c.Config.LDAPSettings.FirstNameAttribute, "TWIGEX_LDAP_FIRST_NAME_ATTR")
	setStr(&c.Config.LDAPSettings.LastNameAttribute, "TWIGEX_LDAP_LAST_NAME_ATTR")
	setStr(&c.Config.LDAPSettings.EmailAttribute, "TWIGEX_LDAP_EMAIL_ATTR")

	// AllowedGroups is a pipe-separated list.
	if v := os.Getenv("TWIGEX_LDAP_ALLOWED_GROUPS"); v != "" {
		groups := strings.Split(v, "|")
		c.Config.LDAPSettings.AllowedGroups = &groups
	}
}
