// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeGraphStore struct {
	store.GraphStore
	entityTypes   []model.GraphEntityType
	relationTypes []model.GraphRelationType
	datasets      []model.GraphDataset
	fields        []model.GraphDatasetField
	unrelatedIDs  []string
	chartFields   map[string][]string
}

func (f *fakeGraphStore) CreateOrUpdateEntityType(_ context.Context, entityType model.GraphEntityType) error {
	entityType.ID = "id-" + entityType.Key
	f.entityTypes = append(f.entityTypes, entityType)

	return nil
}

func (f *fakeGraphStore) GetAllEntityTypes(_ context.Context) ([]model.GraphEntityType, error) {
	return f.entityTypes, nil
}

func (f *fakeGraphStore) CreateOrUpdateRelationType(_ context.Context, relationType model.GraphRelationType) error {
	relationType.ID = "id-" + relationType.Key
	f.relationTypes = append(f.relationTypes, relationType)

	return nil
}

func (f *fakeGraphStore) GetAllRelationTypes(_ context.Context) ([]model.GraphRelationType, error) {
	return f.relationTypes, nil
}

func (f *fakeGraphStore) AddDatasets(_ context.Context, datasets []model.GraphDataset) (int64, error) {
	var added int64
	for _, dataset := range datasets {
		exists := slices.ContainsFunc(f.datasets, func(existing model.GraphDataset) bool {
			return existing.WorkspaceID == dataset.WorkspaceID && existing.Name == dataset.Name
		})
		if exists {
			continue
		}

		dataset.ID = "ds-" + dataset.Name
		f.datasets = append(f.datasets, dataset)
		added++
	}

	return added, nil
}

func (f *fakeGraphStore) GetAllDatasetsForWorkspace(_ context.Context, workspaceID string) ([]model.GraphDataset, error) {
	datasets := make([]model.GraphDataset, 0)
	for _, dataset := range f.datasets {
		if dataset.WorkspaceID == workspaceID {
			datasets = append(datasets, dataset)
		}
	}

	return datasets, nil
}

func (f *fakeGraphStore) DeleteDatasets(_ context.Context, ids []string) error {
	f.datasets = slices.DeleteFunc(f.datasets, func(dataset model.GraphDataset) bool {
		return slices.Contains(ids, dataset.ID)
	})

	f.fields = slices.DeleteFunc(f.fields, func(field model.GraphDatasetField) bool {
		return slices.Contains(ids, field.DatasetID)
	})

	return nil
}

func (f *fakeGraphStore) AddDatasetFields(_ context.Context, fields []model.GraphDatasetField) (int64, error) {
	for _, field := range fields {
		field.ID = "f-" + field.Name
		f.fields = append(f.fields, field)
	}

	return int64(len(fields)), nil
}

func (f *fakeGraphStore) GetAllDatasetFieldsForWorkspace(_ context.Context, workspaceID string) ([]model.GraphDatasetField, error) {
	fields := make([]model.GraphDatasetField, 0)
	for _, field := range f.fields {
		if field.WorkspaceID == workspaceID {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (f *fakeGraphStore) GetDatasetFieldsByNames(_ context.Context, workspaceID string, names []string) ([]model.GraphDatasetField, error) {
	fields := make([]model.GraphDatasetField, 0)
	for _, field := range f.fields {
		if field.WorkspaceID == workspaceID && slices.Contains(names, field.Name) {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (f *fakeGraphStore) DeleteDatasetFields(_ context.Context, ids []string) error {
	f.fields = slices.DeleteFunc(f.fields, func(field model.GraphDatasetField) bool {
		return slices.Contains(ids, field.ID)
	})

	return nil
}

func (f *fakeGraphStore) DeleteEntityRelations(_ context.Context, _ string, entityIDs []string) error {
	f.unrelatedIDs = append(f.unrelatedIDs, entityIDs...)

	return nil
}

func (f *fakeGraphStore) UpdateStructuredRelations(_ context.Context, _, _, fromID, _ string, toIDs []string) (int64, int64, error) {
	if f.chartFields == nil {
		f.chartFields = make(map[string][]string)
	}

	f.chartFields[fromID] = toIDs

	return int64(len(toIDs)), 0, nil
}

type fakeGraphEngine struct {
	QueryEngine
	meta map[string]*collimato.CubeCollection
}

func (f fakeGraphEngine) Meta(_ context.Context, workspaceID string) (*collimato.CubeCollection, error) {
	collection, ok := f.meta[workspaceID]
	if !ok {
		return nil, errors.New("cube is down")
	}

	return collection, nil
}

type fakeGraphCollimatoStore struct {
	store.CollimatoStore
	workspaces []model.CollimatoWorkspace
	charts     []model.Chart
}

func (f *fakeGraphCollimatoStore) GetWorkspaces() ([]model.CollimatoWorkspace, error) {
	return f.workspaces, nil
}

func (f *fakeGraphCollimatoStore) GetCharts(workspaceID string) ([]model.Chart, error) {
	charts := make([]model.Chart, 0)
	for _, chart := range f.charts {
		if chart.WorkspaceID == workspaceID {
			charts = append(charts, chart)
		}
	}

	return charts, nil
}

type fakeGraphSettingsStore struct {
	store.SystemSettingsStore
	values map[string]string
}

func (f *fakeGraphSettingsStore) Get(name string) (string, error) {
	return f.values[name], nil
}

func (f *fakeGraphSettingsStore) UpdateBatch(settings map[string]string) error {
	for name, value := range settings {
		f.values[name] = value
	}

	return nil
}

type fakeGraphRoleStore struct {
	store.RoleStore
	roles []model.Role
}

func (f *fakeGraphRoleStore) GetByNames(_ []string) ([]model.Role, error) {
	return f.roles, nil
}

type fakeGraphJobsStore struct {
	store.JobsStore
	active  int
	created []model.Job
}

func (f *fakeGraphJobsStore) CountActiveByType(_ string) (int, error) {
	return f.active, nil
}

func (f *fakeGraphJobsStore) Create(job model.Job) (*model.Job, error) {
	f.created = append(f.created, job)

	return &job, nil
}

type graphTestApp struct {
	app       *App
	graph     *fakeGraphStore
	collimato *fakeGraphCollimatoStore
	settings  *fakeGraphSettingsStore
	roles     *fakeGraphRoleStore
	jobs      *fakeGraphJobsStore
	meta      map[string]*collimato.CubeCollection
}

func newGraphTestApp(enabled bool) *graphTestApp {
	g := &graphTestApp{
		graph:     &fakeGraphStore{},
		collimato: &fakeGraphCollimatoStore{},
		settings:  &fakeGraphSettingsStore{values: make(map[string]string)},
		roles:     &fakeGraphRoleStore{},
		jobs:      &fakeGraphJobsStore{},
		meta:      make(map[string]*collimato.CubeCollection),
	}

	g.app = &App{
		Store: store.Store{
			Graph:          g.graph,
			Collimato:      g.collimato,
			SystemSettings: g.settings,
			Roles:          g.roles,
			Jobs:           g.jobs,
		},
		QueryEngine: fakeGraphEngine{meta: g.meta},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings:  model.SqlSettings{QueryTimeout: model.NewInt(25)},
				FeatureFlags: model.FeatureFlags{KnowledgeGraph: model.NewBool(enabled)},
			},
		},
	}

	g.app.EnsureGraphBuiltInTypes()

	return g
}

func ordersCube(fields ...string) collimato.Cube {
	cube := collimato.Cube{Name: "orders", Type: "cube"}
	for _, field := range fields {
		cube.Dimensions = append(cube.Dimensions, collimato.Dimension{Name: field})
	}

	return cube
}

func chartUsing(id, workspaceID string, measures ...string) model.Chart {
	return model.Chart{
		ID:          id,
		WorkspaceID: workspaceID,
		Data: map[string]interface{}{
			"query": map[string]interface{}{"measures": measures},
		},
	}
}

func TestEnsureGraphBuiltInTypesRelationsPointToTheirEntityTypes(t *testing.T) {
	g := newGraphTestApp(true)

	if len(g.graph.entityTypes) != 4 {
		t.Fatalf("want 4 entity types, got %d", len(g.graph.entityTypes))
	}

	if len(g.graph.relationTypes) != 4 {
		t.Fatalf("want 4 relation types, got %d", len(g.graph.relationTypes))
	}

	for _, relationType := range g.graph.relationTypes {
		if !relationType.BuiltIn {
			t.Errorf("%s is not marked built in", relationType.Key)
		}

		if relationType.Key != model.GraphRelationChartUsesDatasetField {
			continue
		}

		if relationType.FromTypeID != "id-chart" {
			t.Errorf("want from type id-chart, got %q", relationType.FromTypeID)
		}

		if relationType.ToTypeID != "id-dataset_field" {
			t.Errorf("want to type id-dataset_field, got %q", relationType.ToTypeID)
		}

		if relationType.Kind != model.GraphRelationKindStored {
			t.Errorf("want kind stored, got %q", relationType.Kind)
		}
	}
}

func TestSyncGraphDatasetsMatchesByName(t *testing.T) {
	g := newGraphTestApp(true)

	g.graph.datasets = []model.GraphDataset{
		{ID: "ds-orders", WorkspaceID: "ws", Name: "orders", Kind: model.GraphDatasetKindCube},
		{ID: "ds-old", WorkspaceID: "ws", Name: "old", Kind: model.GraphDatasetKindCube},
	}
	g.graph.fields = []model.GraphDatasetField{
		{ID: "kept", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.amount"},
		{ID: "gone", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.discount"},
		{ID: "old-field", WorkspaceID: "ws", DatasetID: "ds-old", Name: "old.id"},
	}

	salesView := collimato.Cube{
		Name:     "sales_view",
		Type:     "view",
		Measures: []collimato.Measure{{Name: "sales_view.total"}},
	}

	g.meta["ws"] = &collimato.CubeCollection{
		Cubes: []collimato.Cube{ordersCube("orders.amount", "orders.status"), salesView},
	}

	ids, err := g.app.loadGraphBuiltInIDs(context.Background())
	if err != nil {
		t.Fatalf("loadGraphBuiltInIDs: %v", err)
	}

	result, err := g.app.syncGraphDatasets(context.Background(), "ws", ids)
	if err != nil {
		t.Fatalf("syncGraphDatasets: %v", err)
	}

	want := model.GraphRebuildResult{DatasetsAdded: 1, DatasetsRemoved: 1, FieldsAdded: 2, FieldsRemoved: 2}
	if result != want {
		t.Errorf("want %+v, got %+v", want, result)
	}

	fieldIDs := make(map[string]string)
	for _, field := range g.graph.fields {
		fieldIDs[field.Name] = field.ID
	}

	if fieldIDs["orders.amount"] != "kept" {
		t.Errorf("orders.amount lost its ID: %v", fieldIDs)
	}

	if fieldIDs["sales_view.total"] == "" || fieldIDs["orders.status"] == "" {
		t.Errorf("new fields missing: %v", fieldIDs)
	}

	if _, found := fieldIDs["orders.discount"]; found {
		t.Errorf("orders.discount still exists")
	}

	slices.Sort(g.graph.unrelatedIDs)
	if !slices.Equal(g.graph.unrelatedIDs, []string{"gone", "old-field"}) {
		t.Errorf("want the relations of the removed fields deleted, got %v", g.graph.unrelatedIDs)
	}

	for _, dataset := range g.graph.datasets {
		if dataset.Name == "sales_view" && dataset.Kind != model.GraphDatasetKindView {
			t.Errorf("want sales_view stored as a view, got %q", dataset.Kind)
		}
	}
}

func TestSyncGraphDatasetsDeletesNothingWhenCubeReportsAnError(t *testing.T) {
	g := newGraphTestApp(true)

	g.graph.datasets = []model.GraphDataset{{ID: "ds-orders", WorkspaceID: "ws", Name: "orders"}}
	g.graph.fields = []model.GraphDatasetField{{ID: "f1", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.amount"}}

	g.meta["ws"] = &collimato.CubeCollection{Error: "Compile errors: orders.yml"}

	ids, err := g.app.loadGraphBuiltInIDs(context.Background())
	if err != nil {
		t.Fatalf("loadGraphBuiltInIDs: %v", err)
	}

	_, err = g.app.syncGraphDatasets(context.Background(), "ws", ids)
	if err == nil {
		t.Fatalf("want an error when Cube reports one")
	}

	if len(g.graph.datasets) != 1 || len(g.graph.fields) != 1 {
		t.Errorf("datasets or fields were deleted: %+v %+v", g.graph.datasets, g.graph.fields)
	}
}

func TestRebuildGraphWorkspaceConnectsChartsToTheirFields(t *testing.T) {
	g := newGraphTestApp(true)

	g.meta["ws"] = &collimato.CubeCollection{Cubes: []collimato.Cube{ordersCube("orders.amount", "orders.status")}}
	g.collimato.charts = []model.Chart{chartUsing("c1", "ws", "orders.amount", "orders.missing")}

	ids, err := g.app.loadGraphBuiltInIDs(context.Background())
	if err != nil {
		t.Fatalf("loadGraphBuiltInIDs: %v", err)
	}

	result, err := g.app.rebuildGraphWorkspace(context.Background(), "ws", ids)
	if err != nil {
		t.Fatalf("rebuildGraphWorkspace: %v", err)
	}

	if !slices.Equal(g.graph.chartFields["c1"], []string{"f-orders.amount"}) {
		t.Errorf("want c1 related to orders.amount only, got %v", g.graph.chartFields["c1"])
	}

	if result.RelationsAdded != 1 {
		t.Errorf("want 1 relation added, got %d", result.RelationsAdded)
	}
}

func TestRebuildGraphWorkspaceConnectsAMapChartToItsLayerFields(t *testing.T) {
	g := newGraphTestApp(true)

	g.meta["ws"] = &collimato.CubeCollection{Cubes: []collimato.Cube{
		ordersCube("orders.count", "orders.lat", "orders.long", "orders.amount"),
	}}

	mapChart := model.Chart{
		ID:          "map1",
		WorkspaceID: "ws",
		ChartType:   model.ChartTypeMap,
		Data:        map[string]interface{}{"query": nil},
		Configuration: map[string]interface{}{
			"layers": []interface{}{
				map[string]interface{}{
					"model": "orders",
					"config": map[string]interface{}{
						"coordinates":  map[string]interface{}{"lat": "orders.lat", "long": "orders.long"},
						"radius_field": nil,
					},
				},
			},
		},
	}
	g.collimato.charts = []model.Chart{mapChart}

	ids, err := g.app.loadGraphBuiltInIDs(context.Background())
	if err != nil {
		t.Fatalf("loadGraphBuiltInIDs: %v", err)
	}

	_, err = g.app.rebuildGraphWorkspace(context.Background(), "ws", ids)
	if err != nil {
		t.Fatalf("rebuildGraphWorkspace: %v", err)
	}

	got := g.graph.chartFields["map1"]
	slices.Sort(got)

	want := []string{"f-orders.count", "f-orders.lat", "f-orders.long"}

	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestRebuildGraphMarksTheGraphBuiltEvenWhenAWorkspaceFails(t *testing.T) {
	g := newGraphTestApp(true)

	g.collimato.workspaces = []model.CollimatoWorkspace{{ID: "ws"}, {ID: "broken"}}
	g.meta["ws"] = &collimato.CubeCollection{Cubes: []collimato.Cube{ordersCube("orders.amount")}}

	result, err := g.app.RebuildGraph(context.Background())
	if err != nil {
		t.Fatalf("RebuildGraph: %v", err)
	}

	if result.Workspaces != 2 || result.FailedWorkspaces != 1 {
		t.Errorf("want 2 workspaces with 1 failed, got %+v", result)
	}

	if g.settings.values[model.GraphBuiltSetting] != "true" {
		t.Errorf("want the graph marked built")
	}
}

func TestGraphChartSavedDoesNothingWhileTheFlagIsOff(t *testing.T) {
	g := newGraphTestApp(false)

	g.graph.fields = []model.GraphDatasetField{{ID: "f1", WorkspaceID: "ws", Name: "orders.amount"}}

	g.app.graphChartSaved(chartUsing("c1", "ws", "orders.amount"))

	if len(g.graph.chartFields) != 0 {
		t.Errorf("want no relations written, got %v", g.graph.chartFields)
	}
}

func TestRequestGraphRebuild(t *testing.T) {
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}
	member := model.User{ID: "member", Role: "member"}

	t.Run("flag off", func(t *testing.T) {
		g := newGraphTestApp(false)

		_, appErr := g.app.RequestGraphRebuild(admin)
		if appErr == nil || appErr.Status != http.StatusNotFound {
			t.Errorf("want 404, got %v", appErr)
		}
	})

	t.Run("without the permission", func(t *testing.T) {
		g := newGraphTestApp(true)

		_, appErr := g.app.RequestGraphRebuild(member)
		if appErr == nil || appErr.Status != http.StatusForbidden {
			t.Errorf("want 403, got %v", appErr)
		}
	})

	t.Run("with the permission from a role", func(t *testing.T) {
		g := newGraphTestApp(true)
		g.roles.roles = []model.Role{{Name: "member", Permissions: []string{"manage_knowledge_graph"}}}

		_, appErr := g.app.RequestGraphRebuild(member)
		if appErr != nil {
			t.Fatalf("RequestGraphRebuild: %v", appErr)
		}

		if len(g.jobs.created) != 1 || g.jobs.created[0].Type != model.JobTypeGraphRebuild {
			t.Errorf("want one rebuild job, got %+v", g.jobs.created)
		}
	})

	t.Run("already running", func(t *testing.T) {
		g := newGraphTestApp(true)
		g.jobs.active = 1

		_, appErr := g.app.RequestGraphRebuild(admin)
		if appErr == nil || appErr.Status != http.StatusConflict {
			t.Errorf("want 409, got %v", appErr)
		}
	})
}

func TestQueueGraphRebuildIfNeededSkipsABuiltGraph(t *testing.T) {
	g := newGraphTestApp(true)
	g.settings.values[model.GraphBuiltSetting] = "true"

	g.app.QueueGraphRebuildIfNeeded()

	if len(g.jobs.created) != 0 {
		t.Errorf("want no job for a built graph, got %d", len(g.jobs.created))
	}

	g.app.ClearGraphBuilt()
	g.app.QueueGraphRebuildIfNeeded()

	if len(g.jobs.created) != 1 {
		t.Errorf("want one job after the status was cleared, got %d", len(g.jobs.created))
	}
}

func TestGetSystemPermissionsHidesTheGraphWhileTheFlagIsOff(t *testing.T) {
	admin := model.User{ID: "admin", Role: model.SystemAdminRoleId}

	for _, enabled := range []bool{true, false} {
		g := newGraphTestApp(enabled)

		sections, appErr := g.app.GetSystemPermissions(admin)
		if appErr != nil {
			t.Fatalf("GetSystemPermissions: %v", appErr)
		}

		shown := slices.ContainsFunc(sections, func(section model.PermissionSection) bool {
			return section.ID == model.KnowledgeGraphPermissionSectionID
		})

		if shown != enabled {
			t.Errorf("flag %v: want the section shown %v, got %v", enabled, enabled, shown)
		}
	}
}
