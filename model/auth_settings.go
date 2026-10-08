// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type AdminAuthSettings struct {
	LocalAuthEnabled bool              `json:"localAuthEnabled"`
	LdapEnabled      bool              `json:"ldapEnabled"`
	OidcEnabled      bool              `json:"oidcEnabled"`
	OidcFirst        bool              `json:"oidcFirst"`
	LDAP             AdminLDAPSettings `json:"ldap"`
	LdapEnvLocked    bool              `json:"ldapEnvLocked"`
	LDAPSyncStatus   *LDAPSyncStatus   `json:"ldapSyncStatus,omitempty"`
}

const LDAPSyncNeverRun = "never"

type LDAPSyncStatus struct {
	Status      string `json:"status"`
	Reason      string `json:"reason,omitempty"`
	Deactivated int    `json:"deactivated"`
	Reactivated int    `json:"reactivated"`
	Updated     int    `json:"updated"`
	Failed      int    `json:"failed"`
	StartedAt   int64  `json:"startedAt"`
	FinishedAt  int64  `json:"finishedAt"`
	Stale       bool   `json:"stale"`
}

type AdminLDAPSettings struct {
	Host                     string `json:"host"`
	Port                     int    `json:"port"`
	BaseDN                   string `json:"baseDN"`
	BindDN                   string `json:"bindDN"`
	BindPassword             string `json:"bindPassword"`
	UserFilter               string `json:"userFilter"`
	UsernameAttr             string `json:"usernameAttr"`
	IDAttr                   string `json:"idAttr"`
	FirstNameAttr            string `json:"firstNameAttr"`
	LastNameAttr             string `json:"lastNameAttr"`
	EmailAttr                string `json:"emailAttr"`
	AllowedGroups            string `json:"allowedGroups"` // pipe-separated: group DNs contain commas
	InsecureSkipVerify       bool   `json:"insecureSkipVerify"`
	ConnectionSecurity       string `json:"connectionSecurity"`
	CACertificate            string `json:"caCertificate"`
	TimeoutSeconds           int    `json:"timeoutSeconds"`
	SyncEnabled              bool   `json:"syncEnabled"`
	SyncIntervalMinutes      int    `json:"syncIntervalMinutes"`
	SyncMaxDeactivatePercent int    `json:"syncMaxDeactivatePercent"`
}

const (
	LDAPTimeoutSecondsMin = 1
	LDAPTimeoutSecondsMax = 120
)

func ValidLDAPTimeout(seconds int) bool {
	return seconds >= LDAPTimeoutSecondsMin && seconds <= LDAPTimeoutSecondsMax
}

const (
	LDAPSyncIntervalMinutesMin = 5
	LDAPSyncIntervalMinutesMax = 1440

	LDAPSyncDeactivatePercentMin = 1
	LDAPSyncDeactivatePercentMax = 100
)

func ValidLDAPSyncInterval(minutes int) bool {
	return minutes >= LDAPSyncIntervalMinutesMin && minutes <= LDAPSyncIntervalMinutesMax
}

func ValidLDAPDeactivatePercent(percent int) bool {
	return percent >= LDAPSyncDeactivatePercentMin && percent <= LDAPSyncDeactivatePercentMax
}

func ValidLDAPSecurity(s string) bool {
	switch s {
	case "", LDAPSecurityNone, LDAPSecurityTLS, LDAPSecurityStartTLS:
		return true
	}

	return false
}

type PublicAuthSettings struct {
	LocalAuthEnabled     bool                 `json:"localAuthEnabled"`
	LdapEnabled          bool                 `json:"ldapEnabled"`
	OidcFirst            bool                 `json:"oidcFirst"`
	PasswordResetEnabled bool                 `json:"passwordResetEnabled"`
	OIDCProviders        []PublicOIDCProvider `json:"oidcProviders"`
}

type OIDCProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DiscoveryURL string `json:"discovery_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RedirectURL  string `json:"redirect_url"`
	Scopes       string `json:"scopes"`
	ButtonText   string `json:"button_text"`
	ButtonColor  string `json:"button_color"`
	CreatedAt    int64  `json:"created_at"`
}

type PublicOIDCProvider struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ButtonText  string `json:"button_text"`
	ButtonColor string `json:"button_color"`
	LoginURL    string `json:"login_url"`
}
