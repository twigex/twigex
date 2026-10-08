// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/json"
	"sort"
	"time"
)

const (
	LicenseTierBusiness   = "business"
	LicenseTierEnterprise = "enterprise"
)

// Self-hosted customers supply their own infrastructure, so only SaaS tenants
// carry usage ceilings.
const (
	DeploymentSelfHosted = "self_hosted"
	DeploymentSaaS       = "saas"
)

// Whether exceeding the seat count refuses the next user or merely records an
// overage to settle at renewal.
const (
	EnforcementBlock  = "block"
	EnforcementTrueUp = "true_up"
)

var tierRank = map[string]int{
	LicenseTierBusiness:   1,
	LicenseTierEnterprise: 2,
}

type License struct {
	Id           string    `json:"id"`
	IssuedAt     int64     `json:"issued_at"`
	StartsAt     int64     `json:"starts_at"`
	ExpiresAt    int64     `json:"expires_at"`
	Customer     *Customer `json:"customer"`
	Features     *Features `json:"features"`
	SkuName      string    `json:"sku_name"`
	SkuShortName string    `json:"sku_short_name"`
	Deployment   string    `json:"deployment"`
	Limits       *Limits   `json:"limits,omitempty"`

	EnforcementMode string `json:"enforcement_mode"`
	ExtraUsers      *int   `json:"extra_users,omitempty"`

	Next *License `json:"-"`
}

// Limits are usage ceilings rather than entitlements: Features answers "may
// you", Limits answers "how much". Users predates the split and stays in
// Features so already-issued licences keep parsing.
type Limits struct {
	StorageGB *int `json:"storage_gb,omitempty"`
}

type LicenseInfo struct {
	License   *License `json:"license"`
	EnvLocked bool     `json:"env_locked"`
}

type Customer struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
}

type Features struct {
	Users                        *int  `json:"users"`
	MultipleStorages             *bool `json:"multiple_storages"`
	CollimatoUnlimitedWorkspaces *bool `json:"collimato_unlimited_workspaces"`
	CollimatoRoles               *bool `json:"collimato_roles"`
	Groups                       *bool `json:"groups"`
	WorkspaceRoles               *bool `json:"workspace_roles"`
	OAuthProviders               *bool `json:"oauth_providers"`
	LDAP                         *bool `json:"ldap"`
}

type ActiveLicense struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Active  bool   `json:"active"`
	Bytes   string `json:"bytes"`
}

func (l *License) IsActive() bool {
	t := time.Now().Unix()

	return l.StartsAt <= t && t <= l.ExpiresAt
}

func (l *License) ActiveLicense() *License {
	for c := l; c != nil; c = c.Next {
		if c.IsActive() {
			return c
		}
	}

	return nil
}

// enabled is nil-receiver safe so guards can read a feature off an absent
// license, and it resolves the term on every call so a license that expires
// while the server is running stops granting features without a restart.
func (l *License) enabled(pick func(*Features) *bool) bool {
	l = l.ActiveLicense()
	if l == nil || l.Features == nil {
		return false
	}

	v := pick(l.Features)

	return v != nil && *v
}

func (l *License) HasMultipleStorages() bool {
	return l.enabled(func(f *Features) *bool { return f.MultipleStorages })
}

func (l *License) HasCollimatoUnlimitedWorkspaces() bool {
	return l.enabled(func(f *Features) *bool { return f.CollimatoUnlimitedWorkspaces })
}

func (l *License) HasCollimatoRoles() bool {
	return l.enabled(func(f *Features) *bool { return f.CollimatoRoles })
}

func (l *License) HasGroups() bool {
	return l.enabled(func(f *Features) *bool { return f.Groups })
}

func (l *License) HasWorkspaceRoles() bool {
	return l.enabled(func(f *Features) *bool { return f.WorkspaceRoles })
}

func (l *License) HasOAuthProviders() bool {
	return l.enabled(func(f *Features) *bool { return f.OAuthProviders })
}

func (l *License) HasLDAP() bool {
	return l.enabled(func(f *Features) *bool { return f.LDAP })
}

// SeatLimit returns the licensed number of active users, or 0 when there is no
// limit. Unlicensed and expired instances are unlimited: the seat count is a
// commercial term of a paid plan, not a restriction the free product carries.
func (l *License) SeatLimit() int {
	l = l.ActiveLicense()
	if l == nil || l.Features == nil || l.Features.Users == nil {
		return 0
	}

	return *l.Features.Users
}

// SeatHardLimit is the point at which user creation is refused. Between
// SeatLimit and this the instance is over its licence but still usable, so the
// admin has time to buy seats or deactivate someone.
func (l *License) SeatHardLimit() int {
	l = l.ActiveLicense()

	limit := l.SeatLimit()
	if limit == 0 {
		return 0
	}

	if l.ExtraUsers != nil && *l.ExtraUsers > 0 {
		limit += *l.ExtraUsers
	}

	return limit
}

// BlocksOnSeatLimit reports whether exceeding the seat count refuses new users.
// Licences issued before the mode existed carry no value and keep the original
// blocking behaviour rather than silently loosening.
func (l *License) BlocksOnSeatLimit() bool {
	l = l.ActiveLicense()
	if l == nil {
		return false
	}

	return l.EnforcementMode != EnforcementTrueUp
}

// TooManyUsers reads the seat terms off the license itself rather than through
// ActiveLicense, so an upload can be judged before its term starts. A renewal
// resolves to no license in force, and would otherwise report no seat limit.
func (l *License) TooManyUsers(active int) bool {
	if l == nil || l.Features == nil || l.Features.Users == nil {
		return false
	}

	if l.EnforcementMode == EnforcementTrueUp {
		return false
	}

	limit := *l.Features.Users
	if limit == 0 {
		return false
	}

	if l.ExtraUsers != nil && *l.ExtraUsers > 0 {
		limit += *l.ExtraUsers
	}

	return active > limit
}

func (l *License) IsSaaS() bool {
	return l != nil && l.Deployment == DeploymentSaaS
}

// StorageLimitBytes returns the tenant storage ceiling, or 0 when unlimited.
// Self-hosted licences never carry one, and that is enforced here rather than
// trusting the issuer: a malformed licence must not throttle a customer who
// supplies their own storage.
func (l *License) StorageLimitBytes() int64 {
	l = l.ActiveLicense()
	if l == nil || !l.IsSaaS() || l.Limits == nil || l.Limits.StorageGB == nil {
		return 0
	}

	if *l.Limits.StorageGB <= 0 {
		return 0
	}

	return int64(*l.Limits.StorageGB) * 1024 * 1024 * 1024
}

func SelectLicense(candidates []*License, now int64) *License {
	var best *License

	for _, c := range candidates {
		if c == nil || c.StartsAt > now || c.ExpiresAt < now {
			continue
		}

		if best == nil || c.IssuedAt > best.IssuedAt ||
			(c.IssuedAt == best.IssuedAt && c.ExpiresAt > best.ExpiresAt) {
			best = c
		}
	}

	return best
}

func (l *License) Queued() *License {
	now := time.Now().Unix()

	for c := l; c != nil; c = c.Next {
		if c.StartsAt > now {
			return c
		}
	}

	return nil
}

func OrderLicenses(current *License, all []*License, now int64) *License {
	queued := []*License{}

	for _, c := range all {
		if c != nil && c != current && c.StartsAt > now {
			queued = append(queued, c)
		}
	}

	sort.Slice(queued, func(i, j int) bool {
		return queued[i].StartsAt < queued[j].StartsAt
	})

	head := current
	if head == nil {
		if len(queued) == 0 {
			return nil
		}

		head = queued[0]
		queued = queued[1:]
	}

	tail := head
	for _, q := range queued {
		tail.Next = q
		tail = q
	}
	tail.Next = nil

	return head
}

func (l *License) IsTierAtLeast(tier string) bool {
	wantedRank, ok := tierRank[tier]
	if !ok {
		return false
	}

	currentRank, ok := tierRank[l.SkuShortName]
	if !ok {
		return false
	}

	return currentRank >= wantedRank
}

func (l *License) SetDefaults() {
	if l.Features == nil {
		l.Features = &Features{}
	}

	if l.Features.Users == nil {
		l.Features.Users = NewInt(5)
	}

	if l.Deployment == "" {
		l.Deployment = DeploymentSelfHosted
	}

	if l.EnforcementMode == "" {
		l.EnforcementMode = EnforcementBlock
	}

	if l.Features.MultipleStorages == nil {
		l.Features.MultipleStorages = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.CollimatoUnlimitedWorkspaces == nil {
		l.Features.CollimatoUnlimitedWorkspaces = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.CollimatoRoles == nil {
		l.Features.CollimatoRoles = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.Groups == nil {
		l.Features.Groups = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.WorkspaceRoles == nil {
		l.Features.WorkspaceRoles = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.OAuthProviders == nil {
		l.Features.OAuthProviders = NewBool(l.MinimumLicenseBusiness())
	}

	if l.Features.LDAP == nil {
		l.Features.LDAP = NewBool(l.MinimumLicenseEnterprise())
	}
}

func (l *License) MinimumLicenseBusiness() bool {
	return l.IsTierAtLeast(LicenseTierBusiness)
}

func (l *License) MinimumLicenseEnterprise() bool {
	return l.IsTierAtLeast(LicenseTierEnterprise)
}

// ToMap derives the client-facing keys from the Features json tags so the two
// cannot drift apart.
func (l *License) ToMap() map[string]any {
	m := map[string]any{}

	b, err := json.Marshal(l.Features)
	if err != nil {
		return m
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{}
	}

	return m
}

var PB []byte = []byte(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtlXo9YHuW6xMp6F7nqjy
9325mWhENvqjeqqx8i3P1PI9tn7+9N+PiLylQGugISPqqfxJH+Dp6sWER50qX2Bp
88Moed6LcaZZ/MtryoROjkvYQ/z2yypNSpXlWe6RBSQyqiWCwRiR7sJsQQibXvIw
cyZ/7ODM8dE0c4m24drscqegWvQVhmRR9C1JRKbj65tA9HyMqaqGUQHyH2e29RJT
IYlQLEZxYCFNiyYUQ/RJ9XGknzu2wAud16JesnJH0E90HbrKbvVIvUx1kE8UO6g0
hrl6CxlN3tcjDIRytsSiUIW9Mw2u24Pc4k1ahucnG//y2F2yllOKDfkeAPnOqfYl
/vXkDcEp3Y66JsRpRikCVawPpdRfhDqsOkn32WNaDb0kZH2UouUZURY9XE6FY/75
TJ+/M3U6sis/v6RwbFnrlrSCOrjfhdopJn84p2ffJGkVjk5gg9s3nXPbkVRKmcd1
SwJpSHMgfL8WP48GvymBjZGhI7925p2bVTRaDh8skr2zxUd4MssXpjQK0dh2e0D2
CMz15qiYmVpdWoLbx8lIliW/OwHFBUkhht95Le9zKFbQZgwz6z7R2kHxj7C5uPrP
PAF44lgUnVZs2FQ4QR+IJsW1sNsxZZayEHNJYMQaSMv1bfxiiGHKzkHNG2N6Q/kT
pHyaumraCBGd+A2XUQzqxoUCAwEAAQ==
-----END PUBLIC KEY-----`)
