// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"
)

func TestTableExists(t *testing.T) {
	db := requireDB(t)

	repo, err := NewMigrationRepository(db)
	if err != nil {
		t.Fatalf("NewMigrationRepository: %v", err)
	}

	tests := []struct {
		table string
		want  bool
	}{
		{"system_settings", true},
		{"app_migrations", true},
		{"schema_migrations", true},
		{"no_such_table", false},
	}

	for _, tt := range tests {
		got, err := repo.TableExists(context.Background(), tt.table)
		if err != nil {
			t.Fatalf("TableExists(%q): %v", tt.table, err)
		}
		if got != tt.want {
			t.Errorf("TableExists(%q) = %t, want %t", tt.table, got, tt.want)
		}
	}
}

func TestSchemaVersion(t *testing.T) {
	db := requireDB(t)

	repo, err := NewMigrationRepository(db)
	if err != nil {
		t.Fatalf("NewMigrationRepository: %v", err)
	}

	version, dirty, err := repo.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if dirty {
		t.Errorf("dirty = true on a freshly migrated schema")
	}
	if version <= 0 {
		t.Errorf("version = %d, want the applied migration on a freshly migrated schema", version)
	}
}
