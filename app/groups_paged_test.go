// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakePagedGroupStore struct {
	store.GroupStore
	limit  int
	offset int
	query  string
}

func (f *fakePagedGroupStore) GetAllPaged(_ context.Context, query string, _ model.Sort, limit, offset int) ([]model.Group, error) {
	f.query, f.limit, f.offset = query, limit, offset
	return []model.Group{{ID: "g1"}}, nil
}

func (f *fakePagedGroupStore) CountAll(context.Context, string) (int, error) {
	return 42, nil
}

func TestListGroupsPagedIsAdminOnly(t *testing.T) {
	a := &App{Store: store.Store{Groups: &fakePagedGroupStore{}}}

	if _, _, appErr := a.ListGroupsPaged(context.Background(), regularUser(), "", model.Sort{}, 20, 0); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 for a regular user, got %v", appErr)
	}
}

func TestListGroupsPagedClampsAndReturnsTotal(t *testing.T) {
	groups := &fakePagedGroupStore{}
	a := &App{Store: store.Store{Groups: groups}}

	got, total, appErr := a.ListGroupsPaged(context.Background(), admin(), "  sales ", model.Sort{}, 5000, -3)
	if appErr != nil {
		t.Fatalf("ListGroupsPaged: %v", appErr)
	}

	if len(got) != 1 || total != 42 {
		t.Errorf("expected one group and a total of 42, got %d and %d", len(got), total)
	}

	if groups.limit != groupsPageDefaultLimit || groups.offset != 0 || groups.query != "sales" {
		t.Errorf("expected limit %d, offset 0, query \"sales\", got %d, %d, %q",
			groupsPageDefaultLimit, groups.limit, groups.offset, groups.query)
	}
}

type fakeSearchGroupStore struct {
	store.GroupStore
	calledAll  bool
	calledUser string
}

func (f *fakeSearchGroupStore) GetAllPaged(context.Context, string, model.Sort, int, int) ([]model.Group, error) {
	f.calledAll = true
	return nil, nil
}

func (f *fakeSearchGroupStore) SearchForUser(_ context.Context, userID, _ string, _ int) ([]model.Group, error) {
	f.calledUser = userID
	return nil, nil
}

func TestSearchGroupsLimitsRegularUsersToTheirOwn(t *testing.T) {
	groups := &fakeSearchGroupStore{}
	a := &App{Store: store.Store{Groups: groups}}

	if _, appErr := a.SearchGroups(context.Background(), regularUser(), "x", 0); appErr != nil {
		t.Fatalf("SearchGroups: %v", appErr)
	}

	if groups.calledAll || groups.calledUser != regularUser().ID {
		t.Errorf("expected a regular user to search only their own groups, got all=%v user=%q", groups.calledAll, groups.calledUser)
	}

	groups.calledUser = ""

	if _, appErr := a.SearchGroups(context.Background(), admin(), "x", 0); appErr != nil {
		t.Fatalf("SearchGroups: %v", appErr)
	}

	if !groups.calledAll || groups.calledUser != "" {
		t.Errorf("expected an admin to search every group, got all=%v user=%q", groups.calledAll, groups.calledUser)
	}
}
