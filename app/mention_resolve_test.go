// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeMentionChannelStore struct {
	store.ChannelStore
	byUsername map[string]model.User
	byID       map[string]model.User
	nameCalls  [][]string
	idCalls    [][]string
	nameErr    error
	idErr      error
}

func (f *fakeMentionChannelStore) GetMembersByUsernames(_ context.Context, _ string, usernames []string) ([]model.User, error) {
	f.nameCalls = append(f.nameCalls, usernames)
	if f.nameErr != nil {
		return nil, f.nameErr
	}
	users := make([]model.User, 0, len(usernames))
	for _, n := range usernames {
		if u, ok := f.byUsername[n]; ok {
			users = append(users, u)
		}
	}
	return users, nil
}

func (f *fakeMentionChannelStore) GetMembersByIDs(_ context.Context, _ string, ids []string) ([]model.User, error) {
	f.idCalls = append(f.idCalls, ids)
	if f.idErr != nil {
		return nil, f.idErr
	}
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		if u, ok := f.byID[id]; ok {
			users = append(users, u)
		}
	}
	return users, nil
}

type fakeMentionUserStore struct {
	store.UserStore
	byID map[string]model.User
	err  error
}

func (f *fakeMentionUserStore) GetByIDs(ids []string) ([]model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		if u, ok := f.byID[id]; ok {
			users = append(users, u)
		}
	}
	return users, nil
}

func mentionApp(members ...model.User) (*App, *fakeMentionChannelStore) {
	s := &fakeMentionChannelStore{
		byUsername: make(map[string]model.User, len(members)),
		byID:       make(map[string]model.User, len(members)),
	}
	u := &fakeMentionUserStore{byID: make(map[string]model.User, len(members))}
	for _, m := range members {
		s.byUsername[m.Username] = m
		s.byID[m.ID] = m
		u.byID[m.ID] = m
	}
	return &App{
		Store: store.Store{Channels: s, User: u},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}, s
}

func sorted(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

func TestResolveHandlesNameForm(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, _ := mentionApp(jane)

	got := a.resolvePostMentions(model.Post{Message: "hey @jane"})
	if want := []string{testMentionID}; !reflect.DeepEqual(got.UserIDs, want) {
		t.Errorf("got %v, want %v", got.UserIDs, want)
	}
	if got.Broadcast {
		t.Error("broadcast set without @all/@here")
	}
}

func TestResolveHandlesIDForm(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, s := mentionApp(jane)

	got := a.resolvePostMentions(model.Post{Message: "hey <@" + testMentionID + ">"})
	if want := []string{testMentionID}; !reflect.DeepEqual(got.UserIDs, want) {
		t.Errorf("got %v, want %v", got.UserIDs, want)
	}
	if len(s.nameCalls) != 0 {
		t.Errorf("username lookup ran for an id-only message: %v", s.nameCalls)
	}
}

func TestResolveHandlesDeduplicatesAcrossForms(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, _ := mentionApp(jane)

	got := a.resolvePostMentions(model.Post{Message: "@jane and <@" + testMentionID + ">"})
	if want := []string{testMentionID}; !reflect.DeepEqual(got.UserIDs, want) {
		t.Errorf("got %v, want %v", got.UserIDs, want)
	}
}

func TestResolveHandlesMixedFormsResolveBoth(t *testing.T) {
	const bobID = "11111111111111111111111111111111"
	jane := model.User{ID: testMentionID, Username: "jane"}
	bob := model.User{ID: bobID, Username: "bob"}
	a, _ := mentionApp(jane, bob)

	got := a.resolvePostMentions(model.Post{Message: "@jane and <@" + bobID + ">"})
	want := sorted([]string{bobID, testMentionID})
	if !reflect.DeepEqual(sorted(got.UserIDs), want) {
		t.Errorf("got %v, want %v", sorted(got.UserIDs), want)
	}
}

func TestResolveHandlesBroadcastKeywords(t *testing.T) {
	a, s := mentionApp()

	got := a.resolvePostMentions(model.Post{Message: "morning @here"})
	if !got.Broadcast {
		t.Error("@here did not set broadcast")
	}
	if len(s.idCalls) != 0 {
		t.Errorf("id lookup ran for a keyword-only message: %v", s.idCalls)
	}
}

func TestResolveHandlesKeepsIDsWhenUsernameLookupFails(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, s := mentionApp(jane)
	s.nameErr = errors.New("db down")

	got := a.resolvePostMentions(model.Post{Message: "@bob and <@" + testMentionID + ">"})
	if want := []string{testMentionID}; !reflect.DeepEqual(got.UserIDs, want) {
		t.Errorf("got %v, want %v", got.UserIDs, want)
	}
}

// A crafted post must not be able to notify an arbitrary user.
func TestResolveHandlesIgnoresNonMemberID(t *testing.T) {
	a, _ := mentionApp()

	got := a.resolvePostMentions(model.Post{Message: "hey <@" + testMentionID + ">"})
	if len(got.UserIDs) != 0 {
		t.Errorf("got %v, want none", got.UserIDs)
	}
}
