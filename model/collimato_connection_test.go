// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "testing"

func str(s string) *string { return &s }

func current() Connection {
	return Connection{
		Type:        "postgres",
		DisplayName: "Warehouse",
		Host:        "db.internal",
		Port:        "5432",
		Username:    "reporting",
		Database:    "shop",
		SSLMode:     SSLModeRequire,
	}
}

func TestRenamingDoesNotNeedTheServer(t *testing.T) {
	patch := ConnectionPatch{DisplayName: str("Sales warehouse")}

	if patch.NeedsVerification(current()) {
		t.Error("a display name change should not require verification")
	}
}

func TestMakingAConnectionDefaultDoesNotNeedTheServer(t *testing.T) {
	yes := true
	patch := ConnectionPatch{Default: &yes}

	if patch.NeedsVerification(current()) {
		t.Error("choosing the default should not require verification")
	}
}

func TestAnEmptyPatchNeedsNothing(t *testing.T) {
	if (ConnectionPatch{}).NeedsVerification(current()) {
		t.Error("an empty patch changes nothing")
	}
}

func TestResendingTheSameValuesNeedsNothing(t *testing.T) {
	c := current()
	patch := ConnectionPatch{
		Type:        str(c.Type),
		DisplayName: str("Sales warehouse"),
		Host:        str(c.Host),
		Port:        str(c.Port),
		Username:    str(c.Username),
		Password:    str(""),
		Database:    str(c.Database),
		SSLMode:     str(c.SSLMode),
	}

	if patch.NeedsVerification(c) {
		t.Error("a form that sends every field unchanged is still only a rename")
	}
}

func TestEveryFieldTheDatabaseIsReachedByNeedsVerifying(t *testing.T) {
	config := ConnectionConfig{"ca_cert": "x"}

	cases := map[string]ConnectionPatch{
		"type":     {Type: str("mysql")},
		"host":     {Host: str("elsewhere")},
		"port":     {Port: str("5433")},
		"username": {Username: str("someone")},
		"password": {Password: str("p")},
		"database": {Database: str("other")},
		"ssl mode": {SSLMode: str(SSLModeVerifyFull)},
		"config":   {Config: &config},
	}

	for name, patch := range cases {
		if !patch.NeedsVerification(current()) {
			t.Errorf("a changed %s must be verified", name)
		}
	}
}

func TestABlankPasswordMeansUnchanged(t *testing.T) {
	patch := ConnectionPatch{Password: str("")}

	if patch.NeedsVerification(current()) {
		t.Error("an empty password is the form saying it was left alone")
	}
}

func TestABlankPasswordAlongsideARealChangeStillVerifies(t *testing.T) {
	patch := ConnectionPatch{Password: str(""), Host: str("elsewhere")}

	if !patch.NeedsVerification(current()) {
		t.Error("the host changed, so it must still be verified")
	}
}

func TestAddingACertificateNeedsVerifying(t *testing.T) {
	c := current()
	c.Config = ConnectionConfig{}
	config := ConnectionConfig{"ca_cert": "x"}

	if !(ConnectionPatch{Config: &config}).NeedsVerification(c) {
		t.Error("a new certificate changes who is trusted")
	}
}

func TestResendingTheSameCertificateNeedsNothing(t *testing.T) {
	c := current()
	c.Config = ConnectionConfig{"ca_cert": "x"}
	same := ConnectionConfig{"ca_cert": "x"}

	if (ConnectionPatch{Config: &same}).NeedsVerification(c) {
		t.Error("the same certificate is not a change")
	}
}

func TestRemovingACertificateNeedsVerifying(t *testing.T) {
	c := current()
	c.Config = ConnectionConfig{"ca_cert": "x"}
	empty := ConnectionConfig{}

	if !(ConnectionPatch{Config: &empty}).NeedsVerification(c) {
		t.Error("dropping the certificate changes who is trusted")
	}
}

func TestOnlyVerifyingModesUseACertificate(t *testing.T) {
	uses := map[string]bool{
		SSLModeDisable:    false,
		SSLModeRequire:    false,
		SSLModeVerifyCA:   true,
		SSLModeVerifyFull: true,
	}

	for mode, want := range uses {
		c := current()
		c.SSLMode = mode
		if got := c.UsesCACertificate(); got != want {
			t.Errorf("Connection %q: got %v, want %v", mode, got, want)
		}

		n := NewConnection{SSLMode: mode}
		if got := n.UsesCACertificate(); got != want {
			t.Errorf("NewConnection %q: got %v, want %v", mode, got, want)
		}
	}
}
