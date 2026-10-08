// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for system-role grants on groups: persistence through Create/Update and
// the effective-role resolution used to augment a user's permissions at load.
// Runs against real MySQL, see testhelper_test.go. Reuses seedGroupMember and
// the table list from channel_groups_helpers_test.go (same package).

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func resetGroupRoleTables(t *testing.T) {
	cleanTables(t, "group_members", "user_groups")
}

func TestCreateGroup_PersistsRoles(t *testing.T) {
	db := requireDB(t)
	resetGroupRoleTables(t)

	repo := &groupRepository{Db: db}
	ctx := context.Background()

	created, err := repo.Create(ctx, "owner1", "Analysts", "desc", []string{"analyst", "auditor"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("Get returned nil")
	}
	want := []string{"analyst", "auditor"}
	if !reflect.DeepEqual(got.Roles, want) {
		t.Fatalf("roles = %v, want %v", got.Roles, want)
	}
}

func TestCreateGroup_NoRoles_EmptySlice(t *testing.T) {
	db := requireDB(t)
	resetGroupRoleTables(t)

	repo := &groupRepository{Db: db}
	ctx := context.Background()

	created, err := repo.Create(ctx, "owner1", "Plain", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Roles) != 0 {
		t.Fatalf("roles = %v, want empty", got.Roles)
	}
}

func TestUpdateGroup_ReplacesRoles(t *testing.T) {
	db := requireDB(t)
	resetGroupRoleTables(t)

	repo := &groupRepository{Db: db}
	ctx := context.Background()

	created, err := repo.Create(ctx, "owner1", "G", "", []string{"analyst"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, created.ID, "G", "", []string{"auditor", "editor"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := []string{"auditor", "editor"}
	if !reflect.DeepEqual(got.Roles, want) {
		t.Fatalf("roles = %v, want %v", got.Roles, want)
	}

	// Clearing roles leaves no grants.
	if err := repo.Update(ctx, created.ID, "G", "", nil); err != nil {
		t.Fatalf("Update clear: %v", err)
	}
	got, err = repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Roles) != 0 {
		t.Fatalf("roles after clear = %v, want empty", got.Roles)
	}
}

func TestGetGroupRolesForUser_UnionsDedupesAndExcludes(t *testing.T) {
	db := requireDB(t)
	resetGroupRoleTables(t)

	repo := &groupRepository{Db: db}
	ctx := context.Background()

	user := "userX"

	// Two groups the user belongs to, with an overlapping role.
	g1, err := repo.Create(ctx, "owner1", "G1", "", []string{"analyst", "auditor"})
	if err != nil {
		t.Fatalf("Create g1: %v", err)
	}
	g2, err := repo.Create(ctx, "owner1", "G2", "", []string{"auditor", "editor"})
	if err != nil {
		t.Fatalf("Create g2: %v", err)
	}
	// A group with roles the user is NOT a member of: must be excluded.
	gOther, err := repo.Create(ctx, "owner1", "GOther", "", []string{"admin_helper"})
	if err != nil {
		t.Fatalf("Create gOther: %v", err)
	}
	// A group the user IS a member of but that grants no roles.
	gEmpty, err := repo.Create(ctx, "owner1", "GEmpty", "", nil)
	if err != nil {
		t.Fatalf("Create gEmpty: %v", err)
	}

	seedGroupMember(t, db, g1.ID, user)
	seedGroupMember(t, db, g2.ID, user)
	seedGroupMember(t, db, gEmpty.ID, user)
	_ = gOther // user deliberately not a member

	got, err := repo.GetRolesForUser(ctx, user)
	if err != nil {
		t.Fatalf("GetRolesForUser: %v", err)
	}
	sort.Strings(got)
	want := []string{"analyst", "auditor", "editor"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
}

func TestGetGroupRolesForUser_ExcludesDeletedGroup(t *testing.T) {
	db := requireDB(t)
	resetGroupRoleTables(t)

	repo := &groupRepository{Db: db}
	ctx := context.Background()

	user := "userY"
	g, err := repo.Create(ctx, "owner1", "G", "", []string{"analyst"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	seedGroupMember(t, db, g.ID, user)

	if err := repo.SoftDelete(ctx, g.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	got, err := repo.GetRolesForUser(ctx, user)
	if err != nil {
		t.Fatalf("GetRolesForUser: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("roles = %v, want empty after group delete", got)
	}
}
