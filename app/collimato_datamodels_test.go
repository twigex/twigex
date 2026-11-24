// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
	"github.com/twigex/twigex/store"
)

func TestDatasetFromTable(t *testing.T) {
	conn := &model.Connection{ID: "conn1", Database: "Analytics", Type: "mysql"}
	columns := []model.ColumnDetail{
		{Name: "id", Type: "int"},
		{Name: "title", Type: "varchar(255)"},
		{Name: "created_at", Type: "datetime"},
		{Name: "active", Type: "tinyint(1)"},
		{Name: "rank", Type: "tinyint(10)"},
		{Name: "payload", Type: "json"},
		{Name: "order-id", Type: "int"},
	}

	dataset, skipped := datasetFromTable(conn, "orders", "orders", columns)

	if dataset.Name != "orders" || dataset.SqlTable != "orders" {
		t.Errorf("name/table: got %q/%q", dataset.Name, dataset.SqlTable)
	}
	if dataset.ConnectionID != "conn1" {
		t.Errorf("connection: got %q, want conn1", dataset.ConnectionID)
	}

	want := []model.DatasetDimension{
		{Name: "id", SQL: "id", Type: "number"},
		{Name: "title", SQL: "title", Type: "string"},
		{Name: "created_at", SQL: "created_at", Type: "time"},
		{Name: "active", SQL: "active", Type: "boolean"},
		{Name: "rank", SQL: "rank", Type: "number"},
	}
	if len(dataset.Dimensions) != len(want) {
		t.Fatalf("dimensions: got %d, want %d: %+v", len(dataset.Dimensions), len(want), dataset.Dimensions)
	}
	for i, d := range want {
		if dataset.Dimensions[i] != d {
			t.Errorf("dimension %d: got %+v, want %+v", i, dataset.Dimensions[i], d)
		}
	}

	if len(skipped) != 2 || skipped[0] != "payload" || skipped[1] != "order-id" {
		t.Errorf("skipped: got %v, want [payload order-id]", skipped)
	}

	if len(dataset.Measures) != 1 || dataset.Measures[0].Type != "count" {
		t.Errorf("measures: got %+v", dataset.Measures)
	}
}

func TestDatasetFromTableSkipsEveryColumn(t *testing.T) {
	conn := &model.Connection{ID: "conn1", Database: "shop", Type: "mysql"}

	dataset, skipped := datasetFromTable(conn, "blobs", "blobs", []model.ColumnDetail{
		{Name: "data", Type: "longblob"},
	})

	if len(dataset.Dimensions) != 0 {
		t.Errorf("want no dimensions, got %+v", dataset.Dimensions)
	}
	if len(skipped) != 1 {
		t.Errorf("want 1 skipped, got %v", skipped)
	}
}

func TestHasDuplicateFieldNames(t *testing.T) {
	unique := model.Dataset{
		Dimensions: []model.DatasetDimension{{Name: "id"}, {Name: "full_name"}},
		Measures:   []model.DatasetMeasure{{Name: "count"}, {Name: "total"}},
	}
	if hasDuplicateFieldNames(unique) {
		t.Errorf("unique names reported as duplicates")
	}

	twoDimensions := model.Dataset{
		Dimensions: []model.DatasetDimension{{Name: "full_name"}, {Name: "full_name"}},
	}
	if !hasDuplicateFieldNames(twoDimensions) {
		t.Errorf("two dimensions with one name were not caught")
	}

	dimensionAndMeasure := model.Dataset{
		Dimensions: []model.DatasetDimension{{Name: "total"}},
		Measures:   []model.DatasetMeasure{{Name: "count"}, {Name: "total"}},
	}
	if !hasDuplicateFieldNames(dimensionAndMeasure) {
		t.Errorf("a dimension and a measure with one name were not caught")
	}
}

func TestDatasetFromTableInAnotherPostgresSchema(t *testing.T) {
	conn := &model.Connection{ID: "conn1", Type: "postgres"}

	dataset, _ := datasetFromTable(conn, "archive_regions", "archive.regions", []model.ColumnDetail{
		{Name: "name", Type: "text"},
	})

	if dataset.Name != "archive_regions" || dataset.SqlTable != "archive.regions" {
		t.Errorf("want dataset archive_regions on table archive.regions, got %q on %q", dataset.Name, dataset.SqlTable)
	}
}

func TestTableListSuggestsEachColumnsType(t *testing.T) {
	tables := []queryengine.Table{{
		Name: "stores",
		Columns: []queryengine.TableColumn{
			{Name: "is_open", ColumnType: "tinyint(1)"},
			{Name: "opened_at", ColumnType: "datetime"},
			{Name: "notes", ColumnType: "json"},
		},
	}}

	listed := toModelTables("mysql", tables)

	want := []model.ColumnDetail{
		{Name: "is_open", Type: "tinyint(1)", SuggestedType: "boolean"},
		{Name: "opened_at", Type: "datetime", SuggestedType: "time"},
		{Name: "notes", Type: "json", SuggestedType: ""},
	}

	if !reflect.DeepEqual(listed[0].Columns, want) {
		t.Errorf("want columns\n%+v\ngot\n%+v", want, listed[0].Columns)
	}
}

// fakeDataModelFiles keeps saved files by type and name. Names compare
// without case, as the table's collation does.
type fakeDataModelFiles struct {
	store.CollimatoStore
	files   map[string]model.CollimatoWorkspaceFile
	deleted []string
}

func fileKey(name, fileType string) string {
	return fileType + "/" + strings.ToLower(name)
}

func newFakeDataModelFiles(files ...model.CollimatoWorkspaceFile) *fakeDataModelFiles {
	fake := &fakeDataModelFiles{files: map[string]model.CollimatoWorkspaceFile{}}

	for _, file := range files {
		fake.files[fileKey(file.Name, file.FileType)] = file
	}

	return fake
}

func (f *fakeDataModelFiles) GetEffectiveRolesForUser(context.Context, string, string) ([]string, error) {
	return []string{"workspace_admin"}, nil
}

func (f *fakeDataModelFiles) GetRolesByName([]string, string) ([]model.CollimatoRole, error) {
	permissions := []string{
		model.CollimatoPermissions.PermissionCreateDataModels.Id,
		model.CollimatoPermissions.PermissionEditDataModels.Id,
		model.CollimatoPermissions.PermissionDeleteDataModels.Id,
	}

	return []model.CollimatoRole{{Name: "workspace_admin", Permissions: permissions}}, nil
}

func (f *fakeDataModelFiles) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	return &model.CollimatoWorkspace{ID: id}, nil
}

func (f *fakeDataModelFiles) GetConnectionByID(id string) (*model.Connection, error) {
	return &model.Connection{ID: id, WorkspaceID: "ws1", Type: "mysql"}, nil
}

func (f *fakeDataModelFiles) GetWorkspaceRoles(string) ([]model.CollimatoRole, error) {
	return nil, nil
}

func (f *fakeDataModelFiles) GetWorkspaceFile(_, name, fileType string) (*model.CollimatoWorkspaceFile, error) {
	file, found := f.files[fileKey(name, fileType)]
	if !found {
		return nil, nil
	}

	return &file, nil
}

func (f *fakeDataModelFiles) GetWorkspaceFiles(_, fileType string) ([]model.CollimatoWorkspaceFile, error) {
	var files []model.CollimatoWorkspaceFile

	for _, file := range f.files {
		if file.FileType == fileType {
			files = append(files, file)
		}
	}

	return files, nil
}

func (f *fakeDataModelFiles) UpdateWorkspaceFile(file model.CollimatoWorkspaceFile) (*model.CollimatoWorkspaceFile, error) {
	f.files[fileKey(file.Name, file.FileType)] = file

	return &file, nil
}

func (f *fakeDataModelFiles) DeleteWorkspaceFile(_, name, fileType string) error {
	f.deleted = append(f.deleted, name)
	delete(f.files, fileKey(name, fileType))

	return nil
}

type emptyFieldListEngine struct {
	QueryEngine
}

func (emptyFieldListEngine) FieldList(context.Context, string) (*collimato.FieldList, error) {
	return &collimato.FieldList{}, nil
}

func dataModelApp(files *fakeDataModelFiles) *App {
	return &App{
		Store:       store.Store{Collimato: files},
		QueryEngine: emptyFieldListEngine{},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings:  model.SqlSettings{QueryTimeout: model.NewInt(25)},
			FeatureFlags: model.FeatureFlags{KnowledgeGraph: model.NewBool(false)},
		}},
	}
}

func builderModelText(t *testing.T, builderModel any) *string {
	t.Helper()

	encoded, err := json.Marshal(builderModel)
	if err != nil {
		t.Fatal(err)
	}

	text := string(encoded)

	return &text
}

func storesDataset() model.Dataset {
	return model.Dataset{
		Name:         "stores",
		SqlTable:     "stores",
		ConnectionID: "c1",
		Dimensions:   []model.DatasetDimension{{Name: "id", SQL: "id", Type: "number"}},
	}
}

// A builder model that kept its .yml name, because an old hand-written file
// of the plain name was there when Cube.js was removed, is saved back under
// its own name; the hand-written file is left alone.
func TestSavingADatasetThatKeptItsOldName(t *testing.T) {
	handWritten := model.CollimatoWorkspaceFile{Name: "Stores", FileType: model.CollimatoFileTypeDataset, Content: "cube(`stores`, {})"}
	keptOldName := model.CollimatoWorkspaceFile{Name: "stores.yml", FileType: model.CollimatoFileTypeDataset, BuilderModel: builderModelText(t, storesDataset())}

	files := newFakeDataModelFiles(handWritten, keptOldName)
	a := dataModelApp(files)

	changed := storesDataset()
	changed.Dimensions = append(changed.Dimensions, model.DatasetDimension{Name: "city", SQL: "city", Type: "string"})

	saved, appErr := a.SaveDataset(admin(), "ws1", changed, true)
	if appErr != nil {
		t.Fatalf("SaveDataset: %v", appErr)
	}

	if saved.Name != "stores.yml" {
		t.Errorf("want it saved under stores.yml, got %q", saved.Name)
	}

	if !strings.Contains(*files.files[fileKey("stores.yml", model.CollimatoFileTypeDataset)].BuilderModel, "city") {
		t.Error("the change was not saved in the model's own row")
	}

	if files.files[fileKey("Stores", model.CollimatoFileTypeDataset)].Content != handWritten.Content {
		t.Error("the hand-written file was changed")
	}

	if _, appErr := a.SaveDataset(admin(), "ws1", storesDataset(), false); appErr == nil || appErr.Status != http.StatusConflict {
		t.Errorf("want a new dataset of the same name refused, got %v", appErr)
	}
}

func TestDeletingDataModelsAndTheDatasetsThatUseThem(t *testing.T) {
	orders := model.Dataset{Name: "orders", Joins: []model.DatasetJoin{{Name: "stores"}, {Name: "legacy"}}}

	files := newFakeDataModelFiles(
		model.CollimatoWorkspaceFile{Name: "orders", FileType: model.CollimatoFileTypeDataset, BuilderModel: builderModelText(t, orders)},
		model.CollimatoWorkspaceFile{Name: "stores", FileType: model.CollimatoFileTypeDataset, BuilderModel: builderModelText(t, storesDataset())},
		model.CollimatoWorkspaceFile{Name: "legacy.yml", FileType: model.CollimatoFileTypeDataset, Content: "cubes: []"},
	)
	a := dataModelApp(files)

	appErr := a.DeleteDataModel(admin(), "stores", model.CollimatoFileTypeDataset, "ws1")
	if appErr == nil || appErr.Status != http.StatusConflict {
		t.Errorf("want a dataset another joins to kept, got %v", appErr)
	}

	appErr = a.DeleteDataModel(admin(), "legacy.yml", model.CollimatoFileTypeDataset, "ws1")
	if appErr != nil {
		t.Errorf("want an old hand-written file deleted even though orders names it, got %v", appErr)
	}

	if !slices.Equal(files.deleted, []string{"legacy.yml"}) {
		t.Errorf("want only legacy.yml deleted, got %v", files.deleted)
	}
}
