// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"

	"github.com/twigex/twigex/model"
)

const connectionsTable = "collimato_workspace_connections"

func connectionStore(t *testing.T) *collimatoRepository {
	t.Helper()

	store, err := NewCollimatoRepository(requireDB(t))
	if err != nil {
		t.Fatalf("NewCollimatoRepository: %v", err)
	}
	cleanTables(t, connectionsTable)
	t.Cleanup(func() { cleanTables(t, connectionsTable) })

	return store
}

func newConnection() model.NewConnection {
	return model.NewConnection{
		WorkspaceID: "ws-1",
		DisplayName: "Warehouse",
		Type:        "postgres",
		Host:        "db.internal",
		Port:        "5432",
		Database:    "shop",
		Username:    "reporting",
		Password:    "ciphertext",
		SSLMode:     model.SSLModeVerifyFull,
		Config:      model.ConnectionConfig{"ca_cert": "-----BEGIN CERTIFICATE-----\nabc\n-----END CERTIFICATE-----"},
		Default:     true,
		CreatedBy:   "user-1",
	}
}

// Every field must survive a write and a read. This is the test the migration
// needed: the statements used to be positional, so inserting three columns in
// the middle silently shifted ssl_mode into is_default with nothing failing.
func TestAConnectionRoundTrips(t *testing.T) {
	store := connectionStore(t)
	want := newConnection()

	created, err := store.CreateConnection(want)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected an id")
	}

	got, err := store.GetConnectionByID(created.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID: %v", err)
	}

	checks := []struct {
		field string
		want  string
		got   string
	}{
		{"workspace_id", want.WorkspaceID, got.WorkspaceID},
		{"display_name", want.DisplayName, got.DisplayName},
		{"type", want.Type, got.Type},
		{"host", want.Host, got.Host},
		{"port", want.Port, got.Port},
		{"database", want.Database, got.Database},
		{"username", want.Username, got.Username},
		{"password", want.Password, got.Password},
		{"ssl_mode", want.SSLMode, got.SSLMode},
		{"created_by", want.CreatedBy, got.CreatedBy},
		{"ca_cert", want.Config["ca_cert"], got.Config["ca_cert"]},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: want %q, got %q", c.field, c.want, c.got)
		}
	}

	if !got.Default {
		t.Error("is_default: want true, got false")
	}
	if got.CreatedAt == 0 {
		t.Error("created_at was not set")
	}
	if got.DeletedAt != 0 {
		t.Errorf("deleted_at: want 0, got %d", got.DeletedAt)
	}
}

// ssl_mode and is_default are adjacent in the column list and one is a string
// while the other is a bool, so a misalignment shows up here first.
func TestSSLModeIsNotConfusedWithTheDefaultFlag(t *testing.T) {
	store := connectionStore(t)

	conn := newConnection()
	conn.SSLMode = model.SSLModeRequire
	conn.Default = false

	created, err := store.CreateConnection(conn)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	got, err := store.GetConnectionByID(created.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID: %v", err)
	}

	if got.SSLMode != model.SSLModeRequire {
		t.Errorf("ssl_mode: want %q, got %q", model.SSLModeRequire, got.SSLMode)
	}
	if got.Default {
		t.Error("is_default: want false, got true")
	}
}

func TestEverySSLModeSurvivesStorage(t *testing.T) {
	store := connectionStore(t)

	for _, mode := range []string{
		model.SSLModeDisable,
		model.SSLModeRequire,
		model.SSLModeVerifyCA,
		model.SSLModeVerifyFull,
	} {
		conn := newConnection()
		conn.SSLMode = mode

		created, err := store.CreateConnection(conn)
		if err != nil {
			t.Fatalf("%s: CreateConnection: %v", mode, err)
		}

		got, err := store.GetConnectionByID(created.ID)
		if err != nil {
			t.Fatalf("%s: GetConnectionByID: %v", mode, err)
		}
		if got.SSLMode != mode {
			t.Errorf("want %q, got %q", mode, got.SSLMode)
		}
	}
}

func TestAConnectionWithNoCertificateReadsBackEmpty(t *testing.T) {
	store := connectionStore(t)

	conn := newConnection()
	conn.SSLMode = model.SSLModeRequire
	conn.Config = nil

	created, err := store.CreateConnection(conn)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	got, err := store.GetConnectionByID(created.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID: %v", err)
	}

	if len(got.Config) != 0 {
		t.Errorf("want no config, got %v", got.Config)
	}
	if got.Config["ca_cert"] != "" {
		t.Errorf("want empty ca_cert, got %q", got.Config["ca_cert"])
	}
}

func TestAMultiLineCertificateIsNotMangled(t *testing.T) {
	store := connectionStore(t)

	pem := "-----BEGIN CERTIFICATE-----\n" +
		"MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw\n" +
		"DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow\n" +
		"-----END CERTIFICATE-----\n"

	conn := newConnection()
	conn.Config = model.ConnectionConfig{"ca_cert": pem}

	created, err := store.CreateConnection(conn)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	got, err := store.GetConnectionByID(created.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID: %v", err)
	}

	if got.Config["ca_cert"] != pem {
		t.Errorf("certificate changed in storage:\nwant %q\ngot  %q", pem, got.Config["ca_cert"])
	}
}

func TestUpdatingAConnectionKeepsTheOtherColumns(t *testing.T) {
	store := connectionStore(t)

	created, err := store.CreateConnection(newConnection())
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	created.SSLMode = model.SSLModeVerifyCA
	created.Config = model.ConnectionConfig{"ca_cert": "another cert"}
	created.DisplayName = "Renamed"

	if err := store.UpdateConnection(*created); err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}

	got, err := store.GetConnectionByID(created.ID)
	if err != nil {
		t.Fatalf("GetConnectionByID: %v", err)
	}

	if got.SSLMode != model.SSLModeVerifyCA {
		t.Errorf("ssl_mode: want %q, got %q", model.SSLModeVerifyCA, got.SSLMode)
	}
	if got.Config["ca_cert"] != "another cert" {
		t.Errorf("ca_cert: got %q", got.Config["ca_cert"])
	}
	if got.DisplayName != "Renamed" {
		t.Errorf("display_name: got %q", got.DisplayName)
	}
	if got.Host != "db.internal" || got.Database != "shop" || got.Username != "reporting" {
		t.Errorf("an untouched column changed: host=%q database=%q username=%q",
			got.Host, got.Database, got.Username)
	}
	if got.UpdatedAt == 0 {
		t.Error("updated_at was not set")
	}
}

func TestClearingTheCertificateWritesNull(t *testing.T) {
	store := connectionStore(t)

	created, err := store.CreateConnection(newConnection())
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	created.SSLMode = model.SSLModeRequire
	created.Config = nil

	if err := store.UpdateConnection(*created); err != nil {
		t.Fatalf("UpdateConnection: %v", err)
	}

	var raw *string
	row := requireDB(t).QueryRow("SELECT config FROM "+connectionsTable+" WHERE id = ?", created.ID)
	if err := row.Scan(&raw); err != nil {
		t.Fatalf("scan config: %v", err)
	}
	if raw != nil {
		t.Errorf("want NULL config, got %q", *raw)
	}
}

func TestGetConnectionsListsOnlyTheGivenWorkspace(t *testing.T) {
	store := connectionStore(t)

	mine := newConnection()
	if _, err := store.CreateConnection(mine); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	theirs := newConnection()
	theirs.WorkspaceID = "ws-2"
	theirs.DisplayName = "Someone else"
	if _, err := store.CreateConnection(theirs); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	got, err := store.GetConnections("ws-1")
	if err != nil {
		t.Fatalf("GetConnections: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("want 1 connection, got %d", len(got))
	}
	if got[0].DisplayName != "Warehouse" {
		t.Errorf("got the wrong workspace's connection: %q", got[0].DisplayName)
	}
	if got[0].SSLMode != model.SSLModeVerifyFull {
		t.Errorf("ssl_mode did not survive the list query: %q", got[0].SSLMode)
	}
	if got[0].Config["ca_cert"] == "" {
		t.Error("config did not survive the list query")
	}
}

func TestADeletedConnectionIsNotListedOrFetched(t *testing.T) {
	store := connectionStore(t)

	created, err := store.CreateConnection(newConnection())
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	if err := store.DeleteConnection(created.ID); err != nil {
		t.Fatalf("DeleteConnection: %v", err)
	}

	got, err := store.GetConnections("ws-1")
	if err != nil {
		t.Fatalf("GetConnections: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want no connections, got %d", len(got))
	}

	if _, err := store.GetConnectionByID(created.ID); err == nil {
		t.Error("expected an error fetching a deleted connection")
	}
}
