// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "golang.org/x/text/language"

const (
	DefaultSiteURL    = "http://localhost"
	DefaultOfficeHost = "http://localhost:9980"

	OfficeTypeCollabora  = "collabora"
	OfficeTypeEuroOffice = "eurooffice"

	LDAPSecurityNone     = "none"
	LDAPSecurityTLS      = "tls"
	LDAPSecurityStartTLS = "starttls"
)

type ServerConfig struct {
	ServerSettings ServerSettings
	SqlSettings    SqlSettings
	RedisSettings  RedisSettings
	FileSettings   FileSettings

	OfficeSettings    OfficeSettings
	ChannelSettings   ChannelSettings
	EmailSettings     EmailSettings
	AuthSettings      AuthSettings
	LDAPSettings      LDAPSettings
	PasswordSettings  PasswordSettings
	CollimatoSettings CollimatoSettings
}

type ServerSettings struct {
	SiteURL                   *string
	Port                      *string
	SessionLengthInDays       *string
	TlsCertFile               *string
	TlsKeyFile                *string
	DefaultLocale             *string
	AtRestEncryptKey          *string
	AllowUserLanguageOverride *bool
	// Headers that carry the client IP, in priority order (e.g. X-Forwarded-For,
	// X-Real-IP). Only honored when the request's TCP peer is in TrustedProxies.
	TrustedProxyHeaders []string
	// IPs/CIDRs of the reverse proxies in front of the app. Forwarding headers
	// are only trusted from these peers, and their entries are skipped when
	// resolving the client from X-Forwarded-For. Empty means trust no headers and
	// use the TCP peer, so a direct client cannot forge its IP.
	TrustedProxies []string
}

func (s *ServerSettings) SetDefaults() {
	if s.SiteURL == nil {
		s.SiteURL = NewString(DefaultSiteURL)
	}

	if s.Port == nil {
		s.Port = NewString("3000")
	}

	if s.SessionLengthInDays == nil {
		s.SessionLengthInDays = NewString("30")
	}

	if s.TlsCertFile == nil {
		s.TlsCertFile = NewString("")
	}

	if s.TlsKeyFile == nil {
		s.TlsKeyFile = NewString("")
	}

	if s.DefaultLocale == nil {
		s.DefaultLocale = NewString(language.English.String())
	}

	if s.AtRestEncryptKey == nil {
		s.AtRestEncryptKey = NewString("")
	}

	if s.AllowUserLanguageOverride == nil {
		s.AllowUserLanguageOverride = NewBool(true)
	}

	if s.TrustedProxyHeaders == nil {
		s.TrustedProxyHeaders = []string{}
	}

	if s.TrustedProxies == nil {
		s.TrustedProxies = []string{}
	}
}

type LanguageSettings struct {
	DefaultLocale     string `json:"defaultLocale"`
	AllowUserOverride bool   `json:"allowUserOverride"`
}

type AuthSettings struct {
	LocalAuthEnabled *bool
	OidcEnabled      *bool
	OidcFirst        *bool
}

func (s *AuthSettings) SetDefaults() {
	if s.LocalAuthEnabled == nil {
		s.LocalAuthEnabled = NewBool(true)
	}

	if s.OidcEnabled == nil {
		s.OidcEnabled = NewBool(false)
	}

	if s.OidcFirst == nil {
		s.OidcFirst = NewBool(false)
	}
}

type SqlSettings struct {
	DataSource             *string
	DriverName             *string
	MaxOpenConns           *int
	MaxIdleConns           *int
	ConnMaxLifetimeMinutes *int
	ConnMaxIdleTimeMinutes *int
	QueryTimeout           *int
}

func (s *SqlSettings) SetDefaults() {
	if s.DataSource == nil {
		s.DataSource = NewString("")
	}

	if s.DriverName == nil {
		s.DriverName = NewString("mysql")
	}

	if s.MaxOpenConns == nil {
		s.MaxOpenConns = NewInt(25)
	}

	if s.MaxIdleConns == nil {
		s.MaxIdleConns = NewInt(10)
	}

	if s.ConnMaxLifetimeMinutes == nil {
		s.ConnMaxLifetimeMinutes = NewInt(60)
	}

	if s.ConnMaxIdleTimeMinutes == nil {
		s.ConnMaxIdleTimeMinutes = NewInt(30)
	}

	if s.QueryTimeout == nil {
		s.QueryTimeout = NewInt(25)
	}
}

type RedisSettings struct {
	Address  *string
	Password *string
}

func (s *RedisSettings) SetDefaults() {
	if s.Address == nil {
		s.Address = NewString("localhost:6379")
	}

	if s.Password == nil {
		s.Password = NewString("")
	}
}

type FileSettings struct {
	Directory           *string
	MaxPublicUploadSize *int
}

func (s *FileSettings) SetDefaults() {
	if s.Directory == nil {
		s.Directory = NewString("data/")
	}

	if s.MaxPublicUploadSize == nil {
		s.MaxPublicUploadSize = NewInt(2048)
	}
}

type OfficeSettings struct {
	Enable    *bool
	Type      *string
	Host      *string
	Secret    *string
	EnvLocked bool `json:"env_locked"`
}

func (s *OfficeSettings) SetDefaults() {
	if s.Enable == nil {
		s.Enable = NewBool(false)
	}

	if s.Type == nil {
		s.Type = NewString(OfficeTypeCollabora)
	}

	if s.Host == nil {
		s.Host = NewString(DefaultOfficeHost)
	}

	if s.Secret == nil {
		s.Secret = NewString("")
	}
}

type OfficeOpenResult struct {
	Type   string            `json:"type"`
	URL    string            `json:"url,omitempty"`
	Host   string            `json:"host,omitempty"`
	Config *EuroOfficeConfig `json:"config,omitempty"`
}

type EuroOfficeConfig struct {
	Document     EuroOfficeDocument     `json:"document"`
	DocumentType string                 `json:"documentType"`
	EditorConfig EuroOfficeEditorConfig `json:"editorConfig"`
	Token        string                 `json:"token"`
}

type EuroOfficeDocument struct {
	FileType    string                `json:"fileType"`
	Key         string                `json:"key"`
	Title       string                `json:"title"`
	URL         string                `json:"url"`
	Permissions EuroOfficePermissions `json:"permissions"`
}

type EuroOfficePermissions struct {
	Edit     bool `json:"edit"`
	Download bool `json:"download"`
	Print    bool `json:"print"`
	Chat     bool `json:"chat"`
}

type EuroOfficeEditorConfig struct {
	CallbackURL string         `json:"callbackUrl"`
	Mode        string         `json:"mode"`
	User        EuroOfficeUser `json:"user"`
}

type EuroOfficeUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ChannelSettings struct {
	Enabled      *bool
	Host         *string
	Key          *string
	Secret       *string
	KlipyGIF     KlipyGIFSettings
	LinkPreviews *bool
}

func (s *ChannelSettings) SetDefaults() {
	if s.Enabled == nil {
		s.Enabled = NewBool(false)
	}

	if s.Host == nil {
		s.Host = NewString("")
	}

	if s.Key == nil {
		s.Key = NewString("")
	}

	if s.Secret == nil {
		s.Secret = NewString("")
	}

	if s.LinkPreviews == nil {
		s.LinkPreviews = NewBool(true)
	}

	s.KlipyGIF.SetDefaults()
}

type KlipyGIFSettings struct {
	Enabled *bool
	APIKey  *string
}

func (s *KlipyGIFSettings) SetDefaults() {
	if s.Enabled == nil {
		s.Enabled = NewBool(false)
	}

	if s.APIKey == nil {
		s.APIKey = NewString("")
	}
}

type EmailSettings struct {
	// Nil until an admin chooses, which then means enabled when an SMTP server is set.
	EnableEmail        *bool
	SMTPUsername       *string
	SMTPPassword       *string
	SMTPServer         *string
	SMTPPort           *string
	ConnectionSecurity *string
	EnableSMTPAuth     *bool
	FromAddress        *string
	FromName           *string
	ReplyToAddress     *string
	EnvLocked          bool `json:"env_locked"`
}

func (s *EmailSettings) SetDefaults() {
	if s.SMTPUsername == nil {
		s.SMTPUsername = NewString("")
	}

	if s.SMTPPassword == nil {
		s.SMTPPassword = NewString("")
	}

	if s.SMTPServer == nil {
		s.SMTPServer = NewString("")
	}

	if s.SMTPPort == nil {
		s.SMTPPort = NewString("")
	}

	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("TLS")
	}

	if s.EnableSMTPAuth == nil {
		s.EnableSMTPAuth = NewBool(true)
	}

	if s.FromAddress == nil {
		s.FromAddress = NewString("")
	}

	if s.FromName == nil {
		s.FromName = NewString("Twigex")
	}

	if s.ReplyToAddress == nil {
		s.ReplyToAddress = NewString("")
	}
}

type LDAPSettings struct {
	Enabled                  *bool
	BaseDN                   *string
	BindDN                   *string
	Host                     *string
	Port                     *int
	BindPassword             *string
	InsecureSkipVerify       *bool
	ConnectionSecurity       *string
	CACertificate            *string
	TimeoutSeconds           *int
	SyncEnabled              *bool
	SyncIntervalMinutes      *int
	SyncMaxDeactivatePercent *int
	UserFilter               *string
	UsernameAttribute        *string
	IDAttribute              *string
	FirstNameAttribute       *string
	LastNameAttribute        *string
	EmailAttribute           *string
	AllowedGroups            *[]string
}

func (s *LDAPSettings) SetDefaults() {
	if s.Enabled == nil {
		s.Enabled = NewBool(false)
	}

	if s.BaseDN == nil {
		s.BaseDN = NewString("")
	}

	if s.BindDN == nil {
		s.BindDN = NewString("")
	}

	if s.BindPassword == nil {
		s.BindPassword = NewString("")
	}

	if s.InsecureSkipVerify == nil {
		s.InsecureSkipVerify = NewBool(false)
	}

	if s.Host == nil {
		s.Host = NewString("")
	}

	if s.Port == nil {
		s.Port = NewInt(0)
	}

	if s.ConnectionSecurity == nil {
		s.ConnectionSecurity = NewString("")
	}

	if s.CACertificate == nil {
		s.CACertificate = NewString("")
	}

	if s.SyncEnabled == nil {
		s.SyncEnabled = NewBool(false)
	}

	if s.SyncIntervalMinutes == nil {
		s.SyncIntervalMinutes = NewInt(60)
	}

	if s.SyncMaxDeactivatePercent == nil {
		s.SyncMaxDeactivatePercent = NewInt(20)
	}

	if s.TimeoutSeconds == nil {
		s.TimeoutSeconds = NewInt(10)
	}

	if s.UserFilter == nil {
		s.UserFilter = NewString("")
	}

	if s.UsernameAttribute == nil {
		s.UsernameAttribute = NewString("")
	}

	if s.IDAttribute == nil {
		s.IDAttribute = NewString("")
	}

	if s.FirstNameAttribute == nil {
		s.FirstNameAttribute = NewString("")
	}

	if s.LastNameAttribute == nil {
		s.LastNameAttribute = NewString("")
	}

	if s.EmailAttribute == nil {
		s.EmailAttribute = NewString("")
	}

	if s.AllowedGroups == nil {
		s.AllowedGroups = &[]string{}
	}
}

type PasswordSettings struct {
	PasswordMinLength    *int
	UpperLowerCharacters *bool
	NumericCharacters    *bool
	SpecialCharacters    *bool
}

func (s *PasswordSettings) SetDefaults() {
	if s.PasswordMinLength == nil {
		s.PasswordMinLength = NewInt(8)
	}

	if s.UpperLowerCharacters == nil {
		s.UpperLowerCharacters = NewBool(true)
	}

	if s.NumericCharacters == nil {
		s.NumericCharacters = NewBool(true)
	}

	if s.SpecialCharacters == nil {
		s.SpecialCharacters = NewBool(false)
	}
}

// CollimatoSettings configures the shared Cube.js analytics backend used by
// Collimato. A single Cube.js deployment serves every workspace and resolves
// each workspace's schema and database connections at request time, so no
// per-workspace provisioning is required.
type CollimatoSettings struct {
	// CubeAPIURL is the base URL of the shared Cube.js API, e.g.
	// "http://cube:4000". Twigex proxies workspace queries to it.
	CubeAPIURL *string
	// CubeAPISecret is the shared secret used to sign the JWT that Twigex
	// sends to Cube.js. It must match Cube.js CUBEJS_API_SECRET.
	CubeAPISecret *string
	// CubeCallbackSecret authenticates Cube.js when it calls back into
	// Twigex's internal endpoints to fetch schema files and connections.
	CubeCallbackSecret *string
}

func (s *CollimatoSettings) SetDefaults() {
	if s.CubeAPIURL == nil {
		s.CubeAPIURL = NewString("")
	}

	if s.CubeAPISecret == nil {
		s.CubeAPISecret = NewString("")
	}

	if s.CubeCallbackSecret == nil {
		s.CubeCallbackSecret = NewString("")
	}
}

func (c *ServerConfig) SetDefaults() {
	c.ServerSettings.SetDefaults()
	c.SqlSettings.SetDefaults()
	c.RedisSettings.SetDefaults()
	c.FileSettings.SetDefaults()

	c.OfficeSettings.SetDefaults()
	c.ChannelSettings.SetDefaults()
	c.EmailSettings.SetDefaults()
	c.AuthSettings.SetDefaults()
	c.LDAPSettings.SetDefaults()
	c.PasswordSettings.SetDefaults()
	c.CollimatoSettings.SetDefaults()
}
