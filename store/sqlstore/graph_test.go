// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

func resetGraphTables(t *testing.T) {
	cleanTables(t,
		"graph_entity_types",
		"graph_relation_types",
		"graph_relations",
		"graph_datasets",
		"graph_dataset_fields",
	)
}

func seedGraphRelation(t *testing.T, relationTypeID, fromID, toID, origin string) {
	t.Helper()
	mustExec(t,
		`INSERT INTO graph_relations (id, relation_type_id, from_type_id, from_id, to_type_id, to_id, origin, created_at)
		 VALUES (?, ?, 'chart-type', ?, 'field-type', ?, ?, 0)`,
		model.NewID(), relationTypeID, fromID, toID, origin,
	)
}

func relationTargets(t *testing.T, relations []model.GraphRelation) map[string]string {
	t.Helper()
	targets := make(map[string]string, len(relations))
	for _, relation := range relations {
		targets[relation.ToID] = relation.Origin
	}

	return targets
}

func TestCreateOrUpdateEntityTypeKeepsTheIDWhenSavedAgain(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	entityType := model.GraphEntityType{Key: "chart", Name: "Chart", Source: model.GraphSourceBuiltIn, BuiltIn: true}
	if err := repo.CreateOrUpdateEntityType(ctx, entityType); err != nil {
		t.Fatalf("first save: %v", err)
	}

	first, err := repo.GetAllEntityTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllEntityTypes: %v", err)
	}

	if len(first) != 1 {
		t.Fatalf("want 1 entity type after the first save, got %d", len(first))
	}

	entityType.Name = "Diagram"
	if err := repo.CreateOrUpdateEntityType(ctx, entityType); err != nil {
		t.Fatalf("second save: %v", err)
	}

	all, err := repo.GetAllEntityTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllEntityTypes: %v", err)
	}

	if len(all) != 1 {
		t.Fatalf("want 1 entity type, got %d", len(all))
	}

	if all[0].ID != first[0].ID {
		t.Errorf("ID changed from %s to %s", first[0].ID, all[0].ID)
	}

	if all[0].Name != "Diagram" {
		t.Errorf("want name Diagram, got %q", all[0].Name)
	}

	if !all[0].BuiltIn {
		t.Errorf("want built in")
	}
}

func TestCreateOrUpdateRelationTypeKeepsTheIDWhenSavedAgain(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	relationType := model.GraphRelationType{
		Key:         "chart_uses_dataset_field",
		Name:        "uses",
		ReverseName: "used by",
		FromTypeID:  "chart-type",
		ToTypeID:    "field-type",
		Kind:        model.GraphRelationKindStored,
		BuiltIn:     true,
	}
	if err := repo.CreateOrUpdateRelationType(ctx, relationType); err != nil {
		t.Fatalf("first save: %v", err)
	}

	first, err := repo.GetAllRelationTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllRelationTypes: %v", err)
	}

	if len(first) != 1 {
		t.Fatalf("want 1 relation type after the first save, got %d", len(first))
	}

	relationType.ReverseName = "read by"
	if err := repo.CreateOrUpdateRelationType(ctx, relationType); err != nil {
		t.Fatalf("second save: %v", err)
	}

	all, err := repo.GetAllRelationTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllRelationTypes: %v", err)
	}

	if len(all) != 1 {
		t.Fatalf("want 1 relation type, got %d", len(all))
	}

	if all[0].ID != first[0].ID {
		t.Errorf("ID changed from %s to %s", first[0].ID, all[0].ID)
	}

	if all[0].ReverseName != "read by" {
		t.Errorf("want reverse name 'read by', got %q", all[0].ReverseName)
	}
}

func TestAddDatasetsKeepsExistingRows(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	added, err := repo.AddDatasets(ctx, []model.GraphDataset{
		{WorkspaceID: "ws", Name: "orders", Kind: model.GraphDatasetKindCube},
	})
	if err != nil {
		t.Fatalf("first AddDatasets: %v", err)
	}

	if added != 1 {
		t.Errorf("first add: want 1 added, got %d", added)
	}

	first, err := repo.GetAllDatasetsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetsForWorkspace: %v", err)
	}

	added, err = repo.AddDatasets(ctx, []model.GraphDataset{
		{WorkspaceID: "ws", Name: "orders", Kind: model.GraphDatasetKindCube},
		{WorkspaceID: "ws", Name: "sales_view", Kind: model.GraphDatasetKindView},
	})
	if err != nil {
		t.Fatalf("second AddDatasets: %v", err)
	}

	if added != 1 {
		t.Errorf("second add: want 1 added, got %d", added)
	}

	all, err := repo.GetAllDatasetsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetsForWorkspace: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("want 2 datasets, got %d", len(all))
	}

	if all[0].Name != "orders" || all[0].ID != first[0].ID {
		t.Errorf("orders changed: was %+v, now %+v", first[0], all[0])
	}

	if all[1].Kind != model.GraphDatasetKindView {
		t.Errorf("want sales_view as a view, got %q", all[1].Kind)
	}
}

func TestDeleteDatasetsRemovesTheirFields(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	_, err := repo.AddDatasets(ctx, []model.GraphDataset{
		{WorkspaceID: "ws", Name: "customers", Kind: model.GraphDatasetKindCube},
		{WorkspaceID: "ws", Name: "orders", Kind: model.GraphDatasetKindCube},
	})
	if err != nil {
		t.Fatalf("AddDatasets: %v", err)
	}

	datasets, err := repo.GetAllDatasetsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetsForWorkspace: %v", err)
	}

	kept := datasets[0]
	doomed := datasets[1]

	_, err = repo.AddDatasetFields(ctx, []model.GraphDatasetField{
		{WorkspaceID: "ws", DatasetID: kept.ID, Name: "customers.name"},
		{WorkspaceID: "ws", DatasetID: doomed.ID, Name: "orders.amount"},
	})
	if err != nil {
		t.Fatalf("AddDatasetFields: %v", err)
	}

	if err := repo.DeleteDatasets(ctx, []string{doomed.ID}); err != nil {
		t.Fatalf("DeleteDatasets: %v", err)
	}

	left, err := repo.GetAllDatasetsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetsForWorkspace: %v", err)
	}

	if len(left) != 1 || left[0].ID != kept.ID {
		t.Errorf("want only customers left, got %+v", left)
	}

	fields, err := repo.GetAllDatasetFieldsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetFieldsForWorkspace: %v", err)
	}

	if len(fields) != 1 || fields[0].Name != "customers.name" {
		t.Errorf("want only customers.name left, got %+v", fields)
	}
}

func TestGetDatasetFieldsByNamesOnlyReadsTheWorkspace(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	_, err := repo.AddDatasetFields(ctx, []model.GraphDatasetField{
		{WorkspaceID: "ws-a", DatasetID: "ds-a", Name: "orders.amount"},
		{WorkspaceID: "ws-b", DatasetID: "ds-b", Name: "orders.amount"},
	})
	if err != nil {
		t.Fatalf("AddDatasetFields: %v", err)
	}

	got, err := repo.GetDatasetFieldsByNames(ctx, "ws-a", []string{"orders.amount", "orders.missing"})
	if err != nil {
		t.Fatalf("GetDatasetFieldsByNames: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("want 1 field, got %d", len(got))
	}

	if got[0].WorkspaceID != "ws-a" {
		t.Errorf("want the field from ws-a, got %s", got[0].WorkspaceID)
	}
}

func TestUpdateStructuredRelationsReplacesOnlyStructuredRelations(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	seedGraphRelation(t, "uses", "c1", "f-manual", model.GraphRelationOriginManual)
	seedGraphRelation(t, "uses", "c1", "f-old", model.GraphRelationOriginStructured)

	added, removed, err := repo.UpdateStructuredRelations(ctx, "uses", "chart-type", "c1", "field-type", []string{"f-manual", "f-new"})
	if err != nil {
		t.Fatalf("UpdateStructuredRelations: %v", err)
	}

	if added != 1 || removed != 1 {
		t.Errorf("want 1 added and 1 removed, got %d and %d", added, removed)
	}

	relations, err := repo.GetRelationsFrom(ctx, "chart-type", "c1")
	if err != nil {
		t.Fatalf("GetRelationsFrom: %v", err)
	}

	targets := relationTargets(t, relations)

	if len(targets) != 2 {
		t.Fatalf("want 2 relations, got %v", targets)
	}

	if targets["f-manual"] != model.GraphRelationOriginManual {
		t.Errorf("manual relation changed: %v", targets)
	}

	if targets["f-new"] != model.GraphRelationOriginStructured {
		t.Errorf("new structured relation missing: %v", targets)
	}
}

func TestUpdateStructuredRelationsKeepsAManualRelationToTheSameField(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	seedGraphRelation(t, "uses", "c1", "f1", model.GraphRelationOriginManual)

	added, removed, err := repo.UpdateStructuredRelations(ctx, "uses", "chart-type", "c1", "field-type", []string{"f1", "f1"})
	if err != nil {
		t.Fatalf("UpdateStructuredRelations: %v", err)
	}

	if added != 0 || removed != 0 {
		t.Errorf("want nothing added or removed, got %d and %d", added, removed)
	}

	relations, err := repo.GetRelationsFrom(ctx, "chart-type", "c1")
	if err != nil {
		t.Fatalf("GetRelationsFrom: %v", err)
	}

	if len(relations) != 1 {
		t.Fatalf("want 1 relation, got %d", len(relations))
	}

	if relations[0].Origin != model.GraphRelationOriginManual {
		t.Errorf("want the manual relation kept, got origin %q", relations[0].Origin)
	}
}

func TestUpdateStructuredRelationsKeepsARelationThatIsStillUsed(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	seedGraphRelation(t, "uses", "c1", "f1", model.GraphRelationOriginStructured)

	before, err := repo.GetRelationsFrom(ctx, "chart-type", "c1")
	if err != nil {
		t.Fatalf("GetRelationsFrom: %v", err)
	}

	added, removed, err := repo.UpdateStructuredRelations(ctx, "uses", "chart-type", "c1", "field-type", []string{"f1", "f2"})
	if err != nil {
		t.Fatalf("UpdateStructuredRelations: %v", err)
	}

	if added != 1 || removed != 0 {
		t.Errorf("want 1 added and 0 removed, got %d and %d", added, removed)
	}

	after, err := repo.GetRelationsFrom(ctx, "chart-type", "c1")
	if err != nil {
		t.Fatalf("GetRelationsFrom: %v", err)
	}

	if len(after) != 2 {
		t.Fatalf("want 2 relations, got %d", len(after))
	}

	if after[0].ID != before[0].ID {
		t.Errorf("the relation to f1 was replaced: %s became %s", before[0].ID, after[0].ID)
	}
}

func TestUpdateStructuredRelationsLeavesOtherChartsAlone(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	seedGraphRelation(t, "uses", "c2", "f1", model.GraphRelationOriginStructured)

	_, _, err := repo.UpdateStructuredRelations(ctx, "uses", "chart-type", "c1", "field-type", []string{})
	if err != nil {
		t.Fatalf("UpdateStructuredRelations: %v", err)
	}

	relations, err := repo.GetRelationsTo(ctx, "field-type", "f1")
	if err != nil {
		t.Fatalf("GetRelationsTo: %v", err)
	}

	if len(relations) != 1 || relations[0].FromID != "c2" {
		t.Errorf("want c2's relation kept, got %+v", relations)
	}
}

func TestDeleteEntityRelationsRemovesBothDirections(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	seedGraphRelation(t, "uses", "c1", "f1", model.GraphRelationOriginStructured)
	seedGraphRelation(t, "uses", "c2", "f2", model.GraphRelationOriginManual)
	seedGraphRelation(t, "uses", "c3", "f3", model.GraphRelationOriginStructured)

	if err := repo.DeleteEntityRelations(ctx, "chart-type", []string{"c1"}); err != nil {
		t.Fatalf("DeleteEntityRelations from: %v", err)
	}

	if err := repo.DeleteEntityRelations(ctx, "field-type", []string{"f2"}); err != nil {
		t.Fatalf("DeleteEntityRelations to: %v", err)
	}

	if n := countWhere(t, "SELECT COUNT(*) FROM graph_relations WHERE from_id IN ('c1', 'c2')"); n != 0 {
		t.Errorf("%d relations of deleted entities remain", n)
	}

	if n := countWhere(t, "SELECT COUNT(*) FROM graph_relations WHERE from_id = 'c3'"); n != 1 {
		t.Errorf("other relation: got %d rows, want 1", n)
	}
}

func TestGetDatasetAndFieldByID(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	_, err := repo.AddDatasets(ctx, []model.GraphDataset{{WorkspaceID: "ws", Name: "orders", Kind: model.GraphDatasetKindCube}})
	if err != nil {
		t.Fatalf("AddDatasets: %v", err)
	}

	datasets, err := repo.GetAllDatasetsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetsForWorkspace: %v", err)
	}

	_, err = repo.AddDatasetFields(ctx, []model.GraphDatasetField{{WorkspaceID: "ws", DatasetID: datasets[0].ID, Name: "orders.amount"}})
	if err != nil {
		t.Fatalf("AddDatasetFields: %v", err)
	}

	fields, err := repo.GetAllDatasetFieldsForWorkspace(ctx, "ws")
	if err != nil {
		t.Fatalf("GetAllDatasetFieldsForWorkspace: %v", err)
	}

	dataset, err := repo.GetDatasetByID(ctx, datasets[0].ID)
	if err != nil || dataset.Name != "orders" {
		t.Errorf("want the orders dataset, got %+v (%v)", dataset, err)
	}

	field, err := repo.GetDatasetFieldByID(ctx, fields[0].ID)
	if err != nil || field.Name != "orders.amount" || field.DatasetID != datasets[0].ID {
		t.Errorf("want orders.amount in orders, got %+v (%v)", field, err)
	}

	_, err = repo.GetDatasetByID(ctx, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("want sql.ErrNoRows for a missing dataset, got %v", err)
	}

	_, err = repo.GetDatasetFieldByID(ctx, "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("want sql.ErrNoRows for a missing field, got %v", err)
	}
}

func TestGetDashboardIDsForChartAndFilterColumn(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "collimato_dashboards", "collimato_dashboard_charts", "collimato_dashboard_filters")

	for _, dashboard := range []struct{ id, workspaceID string }{{"d1", "ws"}, {"d2", "ws"}, {"other", "ws-b"}} {
		mustExec(t, `INSERT INTO collimato_dashboards (id, title, owner_id, workspace_id, created_at)
			VALUES (?, 'd', 'alice', ?, 0)`, dashboard.id, dashboard.workspaceID)
	}

	mustExec(t, `INSERT INTO collimato_dashboard_charts (id, dashboard_id, chart_id, created_by, position, created_at)
		VALUES ('dc1', 'd1', 'c1', 'alice', '{}', 0), ('dc2', 'd2', 'c2', 'alice', '{}', 0)`)

	mustExec(t, `INSERT INTO collimato_dashboard_filters (id, dashboard_id, column_name, operator)
		VALUES ('f1', 'd2', 'orders.amount', 'equals'), ('f2', 'other', 'orders.amount', 'equals')`)

	repo := &collimatoRepository{Db: db}

	byChart, err := repo.GetDashboardIDsForChart("c1")
	if err != nil {
		t.Fatalf("GetDashboardIDsForChart: %v", err)
	}

	if !slices.Equal(byChart, []string{"d1"}) {
		t.Errorf("want d1 for c1, got %v", byChart)
	}

	byFilter, err := repo.GetDashboardIDsForFilterColumn("ws", "orders.amount")
	if err != nil {
		t.Fatalf("GetDashboardIDsForFilterColumn: %v", err)
	}

	if !slices.Equal(byFilter, []string{"d2"}) {
		t.Errorf("want only d2 from ws, got %v", byFilter)
	}
}

func TestCreateEntityTypeReportsATakenKey(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	company := model.GraphEntityType{
		ID:        model.NewID(),
		Key:       "company",
		Name:      "Company",
		Source:    model.GraphSourceDataset,
		CreatedAt: 1,
		UpdatedAt: 1,
	}

	err := repo.CreateEntityType(ctx, company)
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}

	company.ID = model.NewID()

	err = repo.CreateEntityType(ctx, company)
	if !errors.Is(err, model.ErrGraphTypeKeyTaken) {
		t.Errorf("want ErrGraphTypeKeyTaken, got %v", err)
	}
}

func TestUserEntityTypesCanBeChangedButBuiltInsCannot(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	err := repo.CreateOrUpdateEntityType(ctx, model.GraphEntityType{Key: "chart", Name: "Chart", Source: model.GraphSourceBuiltIn, BuiltIn: true})
	if err != nil {
		t.Fatalf("save the built-in type: %v", err)
	}

	company := model.GraphEntityType{
		ID:            model.NewID(),
		Key:           "company",
		Name:          "Company",
		Source:        model.GraphSourceDataset,
		SourceID:      "ds-companies",
		KeyProperty:   "f-key",
		TitleProperty: "f-name",
		CreatedAt:     1,
		UpdatedAt:     1,
	}

	err = repo.CreateEntityType(ctx, company)
	if err != nil {
		t.Fatalf("CreateEntityType: %v", err)
	}

	company.Name = "Client"
	company.TitleProperty = "f-title"
	company.UpdatedAt = 2

	err = repo.UpdateEntityType(ctx, company)
	if err != nil {
		t.Fatalf("UpdateEntityType: %v", err)
	}

	saved, err := repo.GetEntityTypeByID(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetEntityTypeByID: %v", err)
	}

	if saved.Name != "Client" || saved.TitleProperty != "f-title" || saved.Key != "company" || saved.BuiltIn {
		t.Errorf("want the renamed user type, got %+v", saved)
	}

	all, err := repo.GetAllEntityTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllEntityTypes: %v", err)
	}

	builtIn := all[0]
	if builtIn.Key != "chart" {
		builtIn = all[1]
	}

	builtIn.Name = "Changed"

	err = repo.UpdateEntityType(ctx, builtIn)
	if err != nil {
		t.Fatalf("UpdateEntityType on the built-in: %v", err)
	}

	err = repo.DeleteEntityType(ctx, builtIn.ID)
	if err != nil {
		t.Fatalf("DeleteEntityType on the built-in: %v", err)
	}

	kept, err := repo.GetEntityTypeByID(ctx, builtIn.ID)
	if err != nil || kept.Name != "Chart" {
		t.Errorf("want the built-in type unchanged, got %+v (%v)", kept, err)
	}

	err = repo.DeleteEntityType(ctx, company.ID)
	if err != nil {
		t.Fatalf("DeleteEntityType: %v", err)
	}

	_, err = repo.GetEntityTypeByID(ctx, company.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("want the user type gone, got %v", err)
	}
}

func TestRelationTypeKindConfigIsSavedAndReadBack(t *testing.T) {
	db := requireDB(t)
	resetGraphTables(t)

	repo := &graphRepository{Db: db}
	ctx := context.Background()

	err := repo.CreateOrUpdateRelationType(ctx, model.GraphRelationType{
		Key:         "chart_uses_dataset_field",
		Name:        "uses",
		ReverseName: "used by",
		FromTypeID:  "chart-type",
		ToTypeID:    "field-type",
		Kind:        model.GraphRelationKindStored,
		BuiltIn:     true,
	})
	if err != nil {
		t.Fatalf("save the built-in relation type: %v", err)
	}

	places := model.GraphRelationType{
		ID:          model.NewID(),
		Key:         "company_places_order",
		Name:        "places",
		ReverseName: "placed by",
		FromTypeID:  "company-type",
		ToTypeID:    "order-type",
		Kind:        model.GraphRelationKindProperty,
		KindConfig:  &model.GraphRelationKindConfig{Property: "f-company-id"},
		CreatedAt:   1,
		UpdatedAt:   1,
	}

	err = repo.CreateRelationType(ctx, places)
	if err != nil {
		t.Fatalf("CreateRelationType: %v", err)
	}

	places.KindConfig = &model.GraphRelationKindConfig{Property: "f-buyer-id", FromProperty: "f-reg-no"}
	places.UpdatedAt = 2

	err = repo.UpdateRelationType(ctx, places)
	if err != nil {
		t.Fatalf("UpdateRelationType: %v", err)
	}

	saved, err := repo.GetRelationTypeByID(ctx, places.ID)
	if err != nil {
		t.Fatalf("GetRelationTypeByID: %v", err)
	}

	if saved.KindConfig == nil || saved.KindConfig.Property != "f-buyer-id" || saved.KindConfig.FromProperty != "f-reg-no" {
		t.Errorf("want the new relation settings, got %+v", saved.KindConfig)
	}

	all, err := repo.GetAllRelationTypes(ctx)
	if err != nil {
		t.Fatalf("GetAllRelationTypes: %v", err)
	}

	for _, relationType := range all {
		if relationType.BuiltIn && relationType.KindConfig != nil {
			t.Errorf("want no settings on the built-in relation type, got %+v", relationType.KindConfig)
		}
	}
}
