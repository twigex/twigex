// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeTaskLinkStore struct {
	fakeReadAccessStore
	calls       int
	field       string
	add, remove []string
	err         error
}

func (f *fakeTaskLinkStore) ChangeTaskLinks(_ context.Context, _, _, _, field string, add, remove []string) (*map[string]interface{}, error) {
	f.calls++
	f.field, f.add, f.remove = field, add, remove
	if f.err != nil {
		return nil, f.err
	}

	return &map[string]interface{}{}, nil
}

func (f *fakeTaskLinkStore) GetTableMeta(string) (string, bool, error) {
	return "", false, errors.New("not needed")
}

func TestChangeTaskLinksChecksTheRequestBeforeTheStore(t *testing.T) {
	admin := workspaceAdmin
	tooMany := make([]string, maxLinkChanges+1)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("t%d", i)
	}

	for _, tc := range []struct {
		name        string
		task, field string
		add, remove []string
	}{
		{"an id added and removed", "a1", "related", []string{"b1"}, []string{"b1"}},
		{"an id that is not an id", "a1", "related", []string{"b1'; DROP TABLE x"}, nil},
		{"an empty id", "a1", "related", nil, []string{""}},
		{"no field", "a1", "", []string{"b1"}, nil},
		{"an undefined task", "undefined task", "related", []string{"b1"}, nil},
		{"too many changes", "a1", "related", tooMany, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := &fakeTaskLinkStore{}
			ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
			a := readAccessApp(&ws.fakeReadAccessStore)
			a.Store.Workspace = ws

			_, appErr := a.ChangeTaskLinks(context.Background(), "ws", "t-a", tc.task, tc.field, tc.add, tc.remove, admin)
			if appErr == nil || appErr.Status != http.StatusBadRequest || ws.calls != 0 {
				t.Errorf("error = %v after %d store calls, want a bad request before the store", appErr, ws.calls)
			}
		})
	}

	ws := &fakeTaskLinkStore{}
	ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	if _, appErr := a.ChangeTaskLinks(context.Background(), "ws", "t-a", "a1", "related", []string{"b1", "b-2"}, []string{"b3"}, admin); appErr != nil {
		t.Fatal(appErr)
	}
	if ws.field != "related" || !slices.Equal(ws.add, []string{"b1", "b-2"}) || !slices.Equal(ws.remove, []string{"b3"}) {
		t.Errorf("store got field %q add %v remove %v", ws.field, ws.add, ws.remove)
	}
}

func TestChangeTaskLinksAnswersForWhatTheStoreRefused(t *testing.T) {
	admin := workspaceAdmin
	for _, tc := range []struct {
		err  error
		want int
	}{
		{sql.ErrNoRows, http.StatusNotFound},
		{model.ErrLinkTargetMissing, http.StatusBadRequest},
		{errors.New("connection lost"), http.StatusInternalServerError},
	} {
		ws := &fakeTaskLinkStore{err: tc.err}
		ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
		a := readAccessApp(&ws.fakeReadAccessStore)
		a.Store.Workspace = ws

		_, appErr := a.ChangeTaskLinks(context.Background(), "ws", "t-a", "a1", "related", []string{"b1"}, nil, admin)
		if appErr == nil || appErr.Status != tc.want {
			t.Errorf("store error %v gave %v, want status %d", tc.err, appErr, tc.want)
		}
	}
}
