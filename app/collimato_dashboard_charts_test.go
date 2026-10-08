// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeDashboardChartsStore struct {
	fakeMetaCollimatoStore
	saved bool
}

func (f *fakeDashboardChartsStore) GetDashboardByID(id string) (*model.Dashboard, error) {
	return &model.Dashboard{ID: id, WorkspaceID: "ws"}, nil
}

func (f *fakeDashboardChartsStore) GetCharts(string) ([]model.Chart, error) {
	return []model.Chart{{ID: "own"}}, nil
}

func (f *fakeDashboardChartsStore) AddDashboardCharts(string, string, []model.DashboardChartPatch) error {
	f.saved = true
	return nil
}

func TestUpdateDashboardChartsRefusesChartsFromAnotherWorkspace(t *testing.T) {
	collimato := &fakeDashboardChartsStore{
		fakeMetaCollimatoStore: fakeMetaCollimatoStore{
			permissions: []string{model.CollimatoPermissions.PermissionEditDashboards.Id},
		},
	}

	a := metaApp()
	a.Store.Collimato = collimato

	patch := []model.DashboardChartPatch{{ID: "own"}, {ID: "foreign"}}

	appErr := a.UpdateDashboardCharts(model.User{ID: "u1"}, "ws", "d1", patch)
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("want 403 for a chart from another workspace, got %v", appErr)
	}

	if collimato.saved {
		t.Fatal("expected nothing saved")
	}

	if appErr := a.UpdateDashboardCharts(model.User{ID: "u1"}, "ws", "d1", patch[:1]); appErr != nil {
		t.Fatalf("own chart: %v", appErr)
	}

	if !collimato.saved {
		t.Error("expected the charts saved")
	}
}
