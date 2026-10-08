// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/twigex/twigex/tlog"
)

func setStr(target **string, key string) {
	if v := os.Getenv(key); v != "" {
		s := v
		*target = &s
	}
}

func setInt(target **int, key string) {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			tlog.Errorw("invalid value for %s: %s, using default", key, v)
			return
		}

		*target = &i
	}
}

func setBool(target **bool, key string) {
	if v := os.Getenv(key); v != "" {
		b := v == "true"
		*target = &b
	}
}

func setStrSlice(target *[]string, key string) {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		*target = parts
	}
}

func hasEnv(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) != "" {
			return true
		}
	}

	return false
}

func (c *ConfigStore) computeEnvLocked() {
	c.EnvLocked.SMTP = hasEnv(
		"TWIGEX_SMTP_ENABLED", "TWIGEX_SMTP_HOST", "TWIGEX_SMTP_PORT", "TWIGEX_SMTP_USERNAME",
		"TWIGEX_SMTP_PASSWORD", "TWIGEX_SMTP_SECURITY", "TWIGEX_SMTP_AUTH_ENABLED",
	)
	c.EnvLocked.Office = hasEnv(
		"TWIGEX_OFFICE_ENABLE", "TWIGEX_OFFICE_TYPE", "TWIGEX_OFFICE_HOST", "TWIGEX_OFFICE_SECRET",
	)
	c.EnvLocked.Video = hasEnv(
		"TWIGEX_CHANNEL_ENABLED", "TWIGEX_CHANNEL_HOST",
		"TWIGEX_CHANNEL_KEY", "TWIGEX_CHANNEL_SECRET",
	)
	c.EnvLocked.GIF = hasEnv(
		"TWIGEX_CHANNEL_GIF_ENABLED", "TWIGEX_CHANNEL_GIF_API_KEY",
	)
	c.EnvLocked.LinkPreview = hasEnv("TWIGEX_CHANNEL_LINK_PREVIEWS")
	c.EnvLocked.LDAP = hasEnv(
		"TWIGEX_LDAP_ENABLED", "TWIGEX_LDAP_HOST", "TWIGEX_LDAP_PORT", "TWIGEX_LDAP_BASE_DN",
		"TWIGEX_LDAP_BIND_DN", "TWIGEX_LDAP_BIND_PASSWORD", "TWIGEX_LDAP_INSECURE_SKIP_VERIFY",
		"TWIGEX_LDAP_USER_FILTER", "TWIGEX_LDAP_USERNAME_ATTR", "TWIGEX_LDAP_ID_ATTR",
		"TWIGEX_LDAP_FIRST_NAME_ATTR",
		"TWIGEX_LDAP_LAST_NAME_ATTR", "TWIGEX_LDAP_EMAIL_ATTR", "TWIGEX_LDAP_ALLOWED_GROUPS",
		"TWIGEX_LDAP_CONNECTION_SECURITY", "TWIGEX_LDAP_TIMEOUT_SECONDS",
		"TWIGEX_LDAP_CA_CERT", "TWIGEX_LDAP_SYNC_ENABLED",
		"TWIGEX_LDAP_SYNC_INTERVAL_MINUTES", "TWIGEX_LDAP_SYNC_MAX_DEACTIVATE_PERCENT",
	)
	c.EnvLocked.License = hasEnv("TWIGEX_LICENSE")
}
