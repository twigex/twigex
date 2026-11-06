// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
)

func (f *fakeGraphStore) GetDatasetByID(_ context.Context, id string) (*model.GraphDataset, error) {
	for _, dataset := range f.datasets {
		if dataset.ID == id {
			return &dataset, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphStore) GetDatasetFieldByID(_ context.Context, id string) (*model.GraphDatasetField, error) {
	for _, field := range f.fields {
		if field.ID == id {
			return &field, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphStore) GetRelationsFrom(_ context.Context, _, entityID string) ([]model.GraphRelation, error) {
	relations := make([]model.GraphRelation, 0)
	for _, fieldID := range f.chartFields[entityID] {
		relations = append(relations, f.chartFieldRelation(entityID, fieldID))
	}

	return relations, nil
}

func (f *fakeGraphStore) GetRelationsTo(_ context.Context, _, entityID string) ([]model.GraphRelation, error) {
	relations := make([]model.GraphRelation, 0)
	for chartID, fieldIDs := range f.chartFields {
		if slices.Contains(fieldIDs, entityID) {
			relations = append(relations, f.chartFieldRelation(chartID, entityID))
		}
	}

	return relations, nil
}

func (f *fakeGraphStore) chartFieldRelation(chartID, fieldID string) model.GraphRelation {
	origin := model.GraphRelationOriginStructured
	if f.manualFields[chartID+">"+fieldID] {
		origin = model.GraphRelationOriginManual
	}

	return model.GraphRelation{
		RelationTypeID: "id-" + model.GraphRelationChartUsesDatasetField,
		FromID:         chartID,
		ToID:           fieldID,
		Origin:         origin,
	}
}

func (f *fakeGraphCollimatoStore) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	return &model.CollimatoWorkspace{ID: id}, nil
}

func (f *fakeGraphCollimatoStore) GetEffectiveRolesForUser(_ context.Context, _, userID string) ([]string, error) {
	if userID == "outsider" {
		return nil, nil
	}

	return []string{"member"}, nil
}

func (f *fakeGraphCollimatoStore) GetRolesByName(_ []string, _ string) ([]model.CollimatoRole, error) {
	return []model.CollimatoRole{{Name: "member", Permissions: f.permissions}}, nil
}

func (f *fakeGraphCollimatoStore) GetChartByID(id string) (*model.Chart, error) {
	for _, chart := range f.charts {
		if chart.ID == id {
			return &chart, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphCollimatoStore) GetDashboards(workspaceID string) ([]model.Dashboard, error) {
	dashboards := make([]model.Dashboard, 0)
	for _, dashboard := range f.dashboards {
		if dashboard.WorkspaceID == workspaceID {
			dashboards = append(dashboards, dashboard)
		}
	}

	return dashboards, nil
}

func (f *fakeGraphCollimatoStore) GetDashboardByID(id string) (*model.Dashboard, error) {
	for _, dashboard := range f.dashboards {
		if dashboard.ID == id {
			return &dashboard, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphCollimatoStore) GetDashboardIDsForChart(chartID string) ([]string, error) {
	ids := make([]string, 0)
	for _, dashboard := range f.dashboards {
		for _, chart := range dashboard.Charts {
			if chart.ID == chartID {
				ids = append(ids, dashboard.ID)
			}
		}
	}

	return ids, nil
}

func (f *fakeGraphCollimatoStore) GetDashboardIDsForFilterColumn(_, column string) ([]string, error) {
	ids := make([]string, 0)
	for _, dashboard := range f.dashboards {
		for _, filter := range dashboard.Filters {
			if filter.Column == column {
				ids = append(ids, dashboard.ID)
			}
		}
	}

	return ids, nil
}

// fakeColumnRules hides fields the way the enterprise column rules do: they
// are left out of the meta, and a query that reads one is refused.
type fakeColumnRules struct {
	interfaces.CollimatoRoles
	hidden []string
}

func (f fakeColumnRules) FilterMeta(_ []model.CollimatoRole, collection collimato.CubeCollection) *collimato.CubeCollection {
	cubes := make([]collimato.Cube, 0, len(collection.Cubes))

	for _, cube := range collection.Cubes {
		dimensions := make([]collimato.Dimension, 0, len(cube.Dimensions))
		for _, dimension := range cube.Dimensions {
			if !slices.Contains(f.hidden, dimension.Name) {
				dimensions = append(dimensions, dimension)
			}
		}

		cube.Dimensions = dimensions
		cubes = append(cubes, cube)
	}

	collection.Cubes = cubes

	return &collection
}

func (f fakeColumnRules) HasPermissionToFields(_ []model.CollimatoRole, query *model.DataQuery) (bool, error) {
	for _, field := range model.GraphQueryFields(*query) {
		if slices.Contains(f.hidden, field) {
			return false, nil
		}
	}

	return true, nil
}

var graphViewerPermissions = []string{
	model.CollimatoPermissions.PermissionViewCharts.Id,
	model.CollimatoPermissions.PermissionViewDashboards.Id,
	model.CollimatoPermissions.PermissionViewDashboardFilters.Id,
}

// newGraphRelationsTestApp builds an "orders" dataset where orders.salary is
// hidden by a column rule. Chart c1 reads orders.amount; chart c2 also reads
// orders.salary, so the user may not see it. Dashboard d1 holds both charts
// and filters on orders.amount.
func newGraphRelationsTestApp(permissions ...string) *graphTestApp {
	g := newGraphTestApp(true)

	g.app.CollimatoRoles = fakeColumnRules{hidden: []string{"orders.salary"}}
	g.collimato.permissions = permissions

	g.meta["ws"] = &collimato.CubeCollection{Cubes: []collimato.Cube{ordersCube("orders.amount", "orders.salary")}}

	g.graph.datasets = []model.GraphDataset{{ID: "ds-orders", WorkspaceID: "ws", Name: "orders"}}
	g.graph.fields = []model.GraphDatasetField{
		{ID: "f-orders.amount", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.amount"},
		{ID: "f-orders.salary", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.salary"},
	}

	c1 := chartUsing("c1", "ws", "orders.amount")
	c1.Name = "Revenue"
	c2 := chartUsing("c2", "ws", "orders.amount", "orders.salary")
	c2.Name = "Pay"
	g.collimato.charts = []model.Chart{c1, c2}

	g.graph.chartFields = map[string][]string{
		"c1": {"f-orders.amount"},
		"c2": {"f-orders.amount", "f-orders.salary"},
	}

	g.collimato.dashboards = []model.Dashboard{{
		ID:          "d1",
		Title:       "Sales",
		WorkspaceID: "ws",
		Charts:      []model.Chart{c1, c2},
		Filters:     []model.DashboardFilter{{ID: "flt1", Column: "orders.amount"}},
	}}

	return g
}

func graphGroup(relations *model.GraphEntityRelations, relationType string) *model.GraphRelationGroup {
	for _, group := range relations.Groups {
		if group.RelationType == relationType {
			return &group
		}
	}

	return nil
}

func graphItemIDs(group *model.GraphRelationGroup) []string {
	if group == nil {
		return nil
	}

	ids := make([]string, 0, len(group.Items))
	for _, item := range group.Items {
		ids = append(ids, item.Entity.ID)
	}

	return ids
}

func TestGraphRelationsForAFieldListOnlyWhatTheUserMaySee(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeDatasetField, "f-orders.amount")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	charts := graphGroup(relations, model.GraphRelationChartUsesDatasetField)
	if !slices.Equal(graphItemIDs(charts), []string{"c1"}) || charts.Count != 1 {
		t.Errorf("want only c1 counted and listed, got %+v", charts)
	}

	dashboards := graphGroup(relations, model.GraphRelationDashboardUsesDatasetField)
	if !slices.Equal(graphItemIDs(dashboards), []string{"d1"}) {
		t.Errorf("want the filtering dashboard d1, got %+v", dashboards)
	}

	datasets := graphGroup(relations, model.GraphRelationDatasetFieldBelongsToDataset)
	if !slices.Equal(graphItemIDs(datasets), []string{"ds-orders"}) {
		t.Errorf("want the orders dataset, got %+v", datasets)
	}
}

func TestGraphRelationsAnswerNotFoundForAHiddenEntity(t *testing.T) {
	cases := []struct {
		name       string
		user       string
		entityType string
		entityID   string
	}{
		{"a field hidden by a column rule", "anna", model.GraphTypeDatasetField, "f-orders.salary"},
		{"a chart that reads a hidden field", "anna", model.GraphTypeChart, "c2"},
		{"a chart outside the user's workspaces", "outsider", model.GraphTypeChart, "c1"},
		{"a dataset outside the user's workspaces", "outsider", model.GraphTypeDataset, "ds-orders"},
		{"a dashboard outside the user's workspaces", "outsider", model.GraphTypeDashboard, "d1"},
		{"an entity that does not exist", "anna", model.GraphTypeChart, "missing"},
		{"an unknown type", "anna", "invoice", "c1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGraphRelationsTestApp(graphViewerPermissions...)

			_, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: tc.user}, tc.entityType, tc.entityID)
			if appErr == nil || appErr.ID != "graph.entity_not_found" {
				t.Errorf("want graph.entity_not_found, got %v", appErr)
			}
		})
	}
}

func TestGraphRelationsForAChartLeaveOutHiddenFields(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	g.graph.chartFields["c1"] = []string{"f-orders.amount", "f-orders.salary"}

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeChart, "c1")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	fields := graphGroup(relations, model.GraphRelationChartUsesDatasetField)
	if !slices.Equal(graphItemIDs(fields), []string{"f-orders.amount"}) || fields.Count != 1 {
		t.Errorf("want only orders.amount, got %+v", fields)
	}

	dashboards := graphGroup(relations, model.GraphRelationDashboardContainsChart)
	if !slices.Equal(graphItemIDs(dashboards), []string{"d1"}) {
		t.Errorf("want the dashboard holding c1, got %+v", dashboards)
	}
}

func TestGraphRelationsSayWhereEachRelationComesFrom(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	g.graph.manualFields = map[string]bool{"c1>f-orders.amount": true}

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeChart, "c1")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	fields := graphGroup(relations, model.GraphRelationChartUsesDatasetField)
	if fields == nil || fields.Items[0].Origin != model.GraphRelationOriginManual {
		t.Errorf("want the field relation a person made marked manual, got %+v", fields)
	}

	dashboards := graphGroup(relations, model.GraphRelationDashboardContainsChart)
	if dashboards == nil || dashboards.Items[0].Origin != model.GraphRelationOriginStructured {
		t.Errorf("want the dashboard read from its table marked structured, got %+v", dashboards)
	}
}

func TestGraphRelationsAnswerNotFoundForATypeFromAnUnknownSource(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	g.graph.entityTypes = append(g.graph.entityTypes, model.GraphEntityType{ID: "id-task", Key: "task", Source: "projects"})

	for _, typeKey := range []string{"task", "no_such_type"} {
		_, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, typeKey, "t1")
		if appErr == nil || appErr.Status != http.StatusNotFound {
			t.Errorf("%s: want not found, got %v", typeKey, appErr)
		}
	}
}

func TestGraphRelationsForADashboardLeaveOutHiddenCharts(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeDashboard, "d1")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	charts := graphGroup(relations, model.GraphRelationDashboardContainsChart)
	if !slices.Equal(graphItemIDs(charts), []string{"c1"}) || charts.Count != 1 {
		t.Errorf("want only c1, got %+v", charts)
	}

	fields := graphGroup(relations, model.GraphRelationDashboardUsesDatasetField)
	if !slices.Equal(graphItemIDs(fields), []string{"f-orders.amount"}) {
		t.Errorf("want the filter field, got %+v", fields)
	}
}

func TestGraphRelationsLeaveOutDashboardFiltersWithoutTheirPermission(t *testing.T) {
	g := newGraphRelationsTestApp(
		model.CollimatoPermissions.PermissionViewCharts.Id,
		model.CollimatoPermissions.PermissionViewDashboards.Id,
	)

	fieldRelations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeDatasetField, "f-orders.amount")
	if appErr != nil {
		t.Fatalf("field relations: %v", appErr)
	}

	if graphGroup(fieldRelations, model.GraphRelationDashboardUsesDatasetField) != nil {
		t.Errorf("want no filter group for the field")
	}

	dashboardRelations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeDashboard, "d1")
	if appErr != nil {
		t.Fatalf("dashboard relations: %v", appErr)
	}

	if graphGroup(dashboardRelations, model.GraphRelationDashboardUsesDatasetField) != nil {
		t.Errorf("want no filter group for the dashboard")
	}
}

func TestGraphRelationsSendOnlyTheCountForALargeGroup(t *testing.T) {
	g := newGraphRelationsTestApp(graphViewerPermissions...)

	for i := 0; i < graphGroupItemLimit+1; i++ {
		id := fmt.Sprintf("bulk-%02d", i)
		g.collimato.charts = append(g.collimato.charts, chartUsing(id, "ws", "orders.amount"))
		g.graph.chartFields[id] = []string{"f-orders.amount"}
	}

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeDatasetField, "f-orders.amount")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	charts := graphGroup(relations, model.GraphRelationChartUsesDatasetField)
	if charts.Count != graphGroupItemLimit+2 || len(charts.Items) != 0 {
		t.Errorf("want %d counted and no items, got count %d with %d items", graphGroupItemLimit+2, charts.Count, len(charts.Items))
	}
}

func TestGraphRelationsAreHiddenWhileTheFlagIsOff(t *testing.T) {
	g := newGraphTestApp(false)

	_, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "anna"}, model.GraphTypeChart, "c1")
	if appErr == nil || appErr.Status != http.StatusNotFound || appErr.ID != "graph.disabled" {
		t.Errorf("want graph.disabled, got %v", appErr)
	}
}
