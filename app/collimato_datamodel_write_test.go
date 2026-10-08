// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeDataModelStore struct {
	fakeMetaCollimatoStore
	files map[string]bool
}

func (f *fakeDataModelStore) GetWorkspaceFile(_, name, _ string) (*model.CollimatoWorkspaceFile, error) {
	if !f.files[name] {
		return nil, nil
	}

	return &model.CollimatoWorkspaceFile{}, nil
}

func TestDataModelWriteNeedsEditToReplaceAndCreateToAdd(t *testing.T) {
	edit := model.CollimatoPermissions.PermissionEditDataModels.Id
	create := model.CollimatoPermissions.PermissionCreateDataModels.Id

	cases := []struct {
		name       string
		permission string
		model      string
		overwrite  bool
		allowed    bool
	}{
		{"edit replaces an existing model", edit, "orders", true, true},
		{"edit cannot add a model", edit, "fresh", false, false},
		{"edit cannot add a model through overwrite", edit, "fresh", true, false},
		{"create adds a model", create, "fresh", false, true},
		{"create cannot replace a model", create, "orders", true, false},
	}

	for _, c := range cases {
		a := metaApp()
		a.Store.Collimato = &fakeDataModelStore{
			fakeMetaCollimatoStore: fakeMetaCollimatoStore{permissions: []string{c.permission}},
			files:                  map[string]bool{"orders.yml": true},
		}

		appErr := a.checkDataModelWrite(model.User{ID: "u1"}, "ws", c.model, model.CollimatoFileTypeCube, c.overwrite)

		if c.allowed && appErr != nil {
			t.Errorf("%s: %v", c.name, appErr)
		}

		if !c.allowed && (appErr == nil || appErr.Status != http.StatusForbidden) {
			t.Errorf("%s: want 403, got %v", c.name, appErr)
		}
	}
}
