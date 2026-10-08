// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// fakeAddMemberWorkspaceStore behaves like the unique (workspace_id, user_id)
// key: a user already in the project keeps their stored row.
type fakeAddMemberWorkspaceStore struct {
	fakeAccessWorkspaceStore
	existing map[string]model.WorkspaceMember
}

func (f *fakeAddMemberWorkspaceStore) AddMember(members []model.WorkspaceMember) ([]model.WorkspaceMember, error) {
	out := make([]model.WorkspaceMember, 0, len(members))
	seen := map[string]bool{}
	for _, m := range members {
		if seen[m.UserID] {
			continue
		}
		seen[m.UserID] = true
		if stored, ok := f.existing[m.UserID]; ok {
			out = append(out, stored)
			continue
		}
		f.existing[m.UserID] = m
		out = append(out, m)
	}
	return out, nil
}

func (f *fakeAddMemberWorkspaceStore) GetNameByID(string) (*string, error) {
	name := "Roadmap"
	return &name, nil
}

type fakeInviteActivityStore struct {
	store.ActivityStore
	invited []string
}

func (f *fakeInviteActivityStore) Create(_, _, _, _, _, parameters string) error {
	var data struct {
		AddedUser string `json:"AddedUser"`
	}
	if err := json.Unmarshal([]byte(parameters), &data); err != nil {
		return err
	}
	f.invited = append(f.invited, data.AddedUser)
	return nil
}

type fakeNoUserStore struct {
	store.UserStore
}

func (fakeNoUserStore) Get(string) (*model.User, error) { return nil, nil }

func TestAddMemberToWorkspaceInvitesOnlyNewMembers(t *testing.T) {
	ws := &fakeAddMemberWorkspaceStore{fakeAccessWorkspaceStore: workspaceAdminAccess(), existing: map[string]model.WorkspaceMember{
		"anna": {ID: "existing-row", UserID: "anna", WorkspaceID: "w1", Role: "admin"},
	}}
	activity := &fakeInviteActivityStore{}
	a := &App{Store: store.Store{Workspace: ws, Activity: activity, User: fakeNoUserStore{}}}

	if _, appErr := a.AddMemberToWorkspace("w1", []string{"anna", "bob", "bob"}, workspaceAdmin); appErr != nil {
		t.Fatalf("AddMemberToWorkspace: %v", appErr)
	}

	sort.Strings(activity.invited)
	if got := strings.Join(activity.invited, ","); got != "bob" {
		t.Fatalf("invited = %q, want only the new member bob, once", got)
	}
}
