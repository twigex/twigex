// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestGetWorkspaceTablesRefusesToSendEveryRow(t *testing.T) {
	a := readAccessApp(&fakeReadAccessStore{})
	user := model.User{ID: "u1", Role: model.SystemAdminRoleId}

	for _, mode := range []string{"all", "filtred", ""} {
		if _, appErr := a.GetWorkspaceTables("ws-a", user, mode, 1, 0); appErr == nil || appErr.Status != http.StatusBadRequest {
			t.Errorf("%q without a page size = %v, want a bad request", mode, appErr)
		}
	}
}
