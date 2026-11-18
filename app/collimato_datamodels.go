// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Old models written by hand were saved with these extensions. They stay
// listed, readable and deletable, but are not used and cannot be edited.
var handWrittenExtensions = []string{".yml", ".js"}

// GetDataModels lists a workspace's saved data models: the builder's, and old
// ones written by hand with their text.
func (a *App) GetDataModels(user model.User, workspaceID string) ([]*model.DataModelFile, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	dataModels := make([]*model.DataModelFile, 0)

	for _, fileType := range []string{model.CollimatoFileTypeDataset, model.CollimatoFileTypeView} {
		savedFiles, err := a.Store.Collimato.GetWorkspaceFiles(workspace.ID, fileType)
		if err != nil {
			tlog.Errorw("Failed to retrieve workspace files",
				"workspace_id", workspaceID,
				"file_type", fileType,
				"error", err,
			)
			return nil, model.NewAppError("datamodel.file_read_failed", http.StatusInternalServerError)
		}

		for _, savedFile := range savedFiles {
			dataModel := &model.DataModelFile{
				Name:         savedFile.Name,
				Type:         fileType,
				BuilderModel: savedFile.BuilderModel,
			}

			if savedFile.BuilderModel == nil {
				dataModel.Content = savedFile.Content
			}

			dataModels = append(dataModels, dataModel)
		}
	}

	return dataModels, nil
}

func hasDuplicateFieldNames(m model.Dataset) bool {
	seen := make(map[string]bool)

	for _, d := range m.Dimensions {
		if seen[d.Name] {
			return true
		}
		seen[d.Name] = true
	}

	for _, ms := range m.Measures {
		if seen[ms.Name] {
			return true
		}
		seen[ms.Name] = true
	}

	return false
}

// isValidIdentifier reports whether s is a safe dataset or field name
// (^[A-Za-z_][A-Za-z0-9_]*$). Used to reject malformed builder input.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		switch {
		case r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}

	return true
}

var datasetDimensionTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"time":    true,
	"boolean": true,
}

// The builder sends countDistinct for a distinct count, as the first saved
// models do; both spellings are accepted.
var datasetMeasureTypes = map[string]bool{
	"count":                 true,
	"count_distinct":        true,
	"countDistinct":         true,
	"count_distinct_approx": true,
	"sum":                   true,
	"avg":                   true,
	"min":                   true,
	"max":                   true,
	"number":                true,
}

// checkDataModelWrite needs the right to create data models, or to edit them
// when the save replaces a model of the same name and type.
func (a *App) checkDataModelWrite(user model.User, workspaceID, name, fileType string, overwrite bool) *model.AppError {
	permission := model.CollimatoPermissions.PermissionCreateDataModels

	if overwrite {
		savedName, err := a.savedBuilderFileName(workspaceID, name, fileType)
		if err != nil {
			tlog.Errorw("Failed to check data model",
				"workspace_id", workspaceID,
				"name", name,
				"error", err,
			)
			return model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
		}

		existing, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, savedName, fileType)
		if err != nil {
			tlog.Errorw("Failed to check data model",
				"workspace_id", workspaceID,
				"name", name,
				"error", err,
			)
			return model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
		}

		if existing != nil {
			permission = model.CollimatoPermissions.PermissionEditDataModels
		}
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, permission) {
		return model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	return nil
}

// savedBuilderFileName is the name a builder model of this name is saved
// under: its plain name, or its old name ending in .yml when it kept that
// because a file of the plain name existed when Cube.js was removed.
func (a *App) savedBuilderFileName(workspaceID, name, fileType string) (string, error) {
	oldName := name + ".yml"

	existing, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, oldName, fileType)
	if err != nil {
		return "", err
	}

	if existing != nil && existing.BuilderModel != nil {
		return oldName, nil
	}

	return name, nil
}

// nameIsTaken reports whether the workspace already has a dataset or view of
// this name, made in the builder or written by hand: field keys start with it,
// so datasets and views share one set of names. overwrite is set when the
// builder saves the model of that exact name and type again.
func (a *App) nameIsTaken(workspaceID, name, fileType string, overwrite bool) (bool, error) {
	if overwrite {
		savedName, err := a.savedBuilderFileName(workspaceID, name, fileType)
		if err != nil {
			return false, err
		}

		// A model that kept its old name is saved again in its own row; the
		// file of the plain name next to it is not replaced.
		if savedName != name {
			return false, nil
		}
	}

	savedNames := append([]string{name}, handWrittenNames(name)...)

	for _, savedType := range []string{model.CollimatoFileTypeDataset, model.CollimatoFileTypeView} {
		for _, savedName := range savedNames {
			existing, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, savedName, savedType)
			if err != nil {
				return false, err
			}

			if existing == nil {
				continue
			}

			if overwrite && savedType == fileType && savedName == name {
				continue
			}

			return true, nil
		}
	}

	return false, nil
}

func handWrittenNames(name string) []string {
	names := make([]string, 0, len(handWrittenExtensions))

	for _, extension := range handWrittenExtensions {
		names = append(names, name+extension)
	}

	return names
}

// SaveDataset saves a dataset made in the builder.
func (a *App) SaveDataset(user model.User, workspaceID string, m model.Dataset, overwrite bool) (*model.DataModelFile, *model.AppError) {
	if appErr := a.checkDataModelWrite(user, workspaceID, m.Name, model.CollimatoFileTypeDataset, overwrite); appErr != nil {
		return nil, appErr
	}

	if !isValidIdentifier(m.Name) || m.SqlTable == "" || len(m.Dimensions) == 0 {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	for _, d := range m.Dimensions {
		if !isValidIdentifier(d.Name) || d.SQL == "" || !datasetDimensionTypes[d.Type] {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	for _, ms := range m.Measures {
		if !isValidIdentifier(ms.Name) || !datasetMeasureTypes[ms.Type] {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	if hasDuplicateFieldNames(m) {
		return nil, model.NewAppError("datamodel.duplicate_name", http.StatusBadRequest)
	}

	for _, j := range m.Joins {
		validRel := j.Relationship == "many_to_one" || j.Relationship == "one_to_many" || j.Relationship == "one_to_one"
		hasCondition := j.SQL != "" || (isValidIdentifier(j.ThisColumn) && isValidIdentifier(j.OtherColumn))
		if !isValidIdentifier(j.Name) || !validRel || !hasCondition {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}
	}

	conn, err := a.Store.Collimato.GetConnectionByID(m.ConnectionID)
	if err != nil {
		tlog.Errorw("Failed to retrieve connection",
			"workspace_id", workspaceID,
			"connection_id", m.ConnectionID,
			"error", err,
		)
		return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
	}

	if conn.WorkspaceID != workspaceID {
		return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
	}

	return a.saveBuilderModel(user, workspaceID, m.Name, model.CollimatoFileTypeDataset, m, overwrite)
}

// SaveView saves a view made in the builder.
func (a *App) SaveView(user model.User, workspaceID string, m model.View, overwrite bool) (*model.DataModelFile, *model.AppError) {
	if appErr := a.checkDataModelWrite(user, workspaceID, m.Name, model.CollimatoFileTypeView, overwrite); appErr != nil {
		return nil, appErr
	}

	if !isValidIdentifier(m.Name) || len(m.Parts) == 0 {
		return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
	}

	for _, part := range m.Parts {
		if part.JoinPath == "" || len(part.Includes) == 0 {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}

		for _, step := range strings.Split(part.JoinPath, ".") {
			if !isValidIdentifier(step) {
				return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
			}
		}

		for _, include := range part.Includes {
			if include != "*" && !isValidIdentifier(include) {
				return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
			}
		}
	}

	return a.saveBuilderModel(user, workspaceID, m.Name, model.CollimatoFileTypeView, m, overwrite)
}

// saveBuilderModel saves a checked dataset or view under its own name and
// brings roles and the knowledge graph up to date.
func (a *App) saveBuilderModel(user model.User, workspaceID, name, fileType string, builderModel any, overwrite bool) (*model.DataModelFile, *model.AppError) {
	taken, err := a.nameIsTaken(workspaceID, name, fileType, overwrite)
	if err != nil {
		tlog.Errorw("Failed to check data model name",
			"workspace_id", workspaceID,
			"name", name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if taken {
		return nil, model.NewAppError("datamodel.name_taken", http.StatusConflict)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	modelJSON, err := json.Marshal(builderModel)
	if err != nil {
		tlog.Errorw("Failed to marshal builder model",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	modelText := string(modelJSON)

	savedName := name
	if overwrite {
		savedName, err = a.savedBuilderFileName(workspaceID, name, fileType)
		if err != nil {
			tlog.Errorw("Failed to check data model",
				"workspace_id", workspaceID,
				"name", name,
				"error", err,
			)
			return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
		}
	}

	if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
		WorkspaceID:  workspaceID,
		Name:         savedName,
		FileType:     fileType,
		BuilderModel: &modelText,
	}); err != nil {
		tlog.Errorw("Failed to save data model",
			"workspace_id", workspaceID,
			"name", name,
			"error", err,
		)
		return nil, model.NewAppError("datamodel.file_write_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	a.graphDataModelChanged(workspaceID)

	saved := &model.DataModelFile{
		Name:         savedName,
		Type:         fileType,
		BuilderModel: &modelText,
	}

	return saved, nil
}

// datasetDependents returns the names of builder models that use a dataset:
// datasets joining to it and views whose path goes through it.
func (a *App) datasetDependents(workspaceID, datasetName string) ([]string, error) {
	var dependents []string

	datasets, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, model.CollimatoFileTypeDataset)
	if err != nil {
		return nil, err
	}

	for _, f := range datasets {
		if f.BuilderModel == nil || dataModelName(f.Name) == datasetName {
			continue
		}

		var saved model.Dataset
		if json.Unmarshal([]byte(*f.BuilderModel), &saved) != nil {
			continue
		}

		for _, j := range saved.Joins {
			if j.Name == datasetName {
				dependents = append(dependents, f.Name)
				break
			}
		}
	}

	views, err := a.Store.Collimato.GetWorkspaceFiles(workspaceID, model.CollimatoFileTypeView)
	if err != nil {
		return nil, err
	}

	for _, f := range views {
		if f.BuilderModel == nil {
			continue
		}

		var saved model.View
		if json.Unmarshal([]byte(*f.BuilderModel), &saved) != nil {
			continue
		}

		for _, part := range saved.Parts {
			if slices.Contains(strings.Split(part.JoinPath, "."), datasetName) {
				dependents = append(dependents, f.Name)
				break
			}
		}
	}

	return dependents, nil
}

// dataModelName is a saved model's name without the extension an old one
// written by hand has.
func dataModelName(savedName string) string {
	return strings.TrimSuffix(savedName, path.Ext(savedName))
}

func (a *App) DeleteDataModel(user model.User, name, fileType, workspaceID string) *model.AppError {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteDataModels) {
		return model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	saved, err := a.Store.Collimato.GetWorkspaceFile(workspaceID, name, fileType)
	if err != nil {
		tlog.Errorw("Failed to read data model",
			"workspace_id", workspaceID,
			"name", name,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	// A dataset other datasets join to, or views go through, would leave them
	// broken, so it is kept until they are changed or removed. An old file
	// written by hand is not used, so deleting it breaks nothing.
	isBuilderDataset := fileType == model.CollimatoFileTypeDataset && saved != nil && saved.BuilderModel != nil
	if isBuilderDataset {
		dependents, err := a.datasetDependents(workspaceID, dataModelName(name))
		if err != nil {
			tlog.Errorw("Failed to find the models using a dataset",
				"workspace_id", workspaceID,
				"dataset", name,
				"error", err,
			)
			return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		if len(dependents) > 0 {
			return model.NewAppError("datamodel.delete_has_dependents", http.StatusConflict)
		}
	}

	if err = a.Store.Collimato.DeleteWorkspaceFile(workspaceID, name, fileType); err != nil {
		tlog.Errorw("Failed to delete data model",
			"workspace_id", workspaceID,
			"name", name,
			"error", err,
		)
		return model.NewAppError("datamodel.file_delete_failed", http.StatusInternalServerError)
	}

	if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
		tlog.Errorw("Failed to update workspace role permissions",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	a.graphDataModelChanged(workspaceID)

	return nil
}

// GenerateDatasets makes a dataset for each chosen table, with a field for
// each column of a type the builder knows.
func (a *App) GenerateDatasets(user model.User, m map[string]interface{}, workspaceID string) (*model.GenerateModelsResult, *model.AppError) {
	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateDataModels) {
		return nil, model.NewAppError("datamodel.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	result := &model.GenerateModelsResult{
		Created:        make([]string, 0),
		Skipped:        make([]string, 0),
		SkippedColumns: make(map[string][]string),
	}

	for connID, v := range m {
		conn, err := a.Store.Collimato.GetConnectionByID(connID)
		if err != nil {
			tlog.Errorw("Failed to retrieve connection",
				"workspace_id", workspaceID,
				"connection_id", connID,
				"error", err,
			)
			return nil, model.NewAppError("connection.retrieval_failed", http.StatusInternalServerError)
		}

		if conn.WorkspaceID != workspaceID {
			return nil, model.NewAppError("connection.forbidden", http.StatusForbidden)
		}

		tables, ok := v.(map[string]interface{})
		if !ok {
			return nil, model.NewAppError("datamodel.invalid_model", http.StatusBadRequest)
		}

		for table, cols := range tables {
			columns, ok := columnDetails(cols)
			if !ok {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			datasetName := collimato.DatasetNameForTable(table)
			if !isValidIdentifier(datasetName) {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			taken, err := a.nameIsTaken(workspaceID, datasetName, model.CollimatoFileTypeDataset, false)
			if err != nil {
				tlog.Errorw("Failed to check data model name",
					"workspace_id", workspaceID,
					"name", datasetName,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			if taken {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			dataset, skippedColumns := datasetFromTable(conn, datasetName, table, columns)
			if len(dataset.Dimensions) == 0 {
				result.Skipped = append(result.Skipped, table)
				continue
			}

			if len(skippedColumns) > 0 {
				result.SkippedColumns[table] = skippedColumns
			}

			modelJSON, err := json.Marshal(dataset)
			if err != nil {
				tlog.Errorw("Failed to marshal builder model",
					"workspace_id", workspaceID,
					"name", dataset.Name,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			modelText := string(modelJSON)

			if _, err = a.Store.Collimato.UpdateWorkspaceFile(model.CollimatoWorkspaceFile{
				WorkspaceID:  workspaceID,
				Name:         dataset.Name,
				FileType:     model.CollimatoFileTypeDataset,
				BuilderModel: &modelText,
			}); err != nil {
				tlog.Errorw("Failed to save a generated dataset",
					"workspace_id", workspaceID,
					"name", dataset.Name,
					"error", err,
				)
				return nil, model.NewAppError("datamodel.generate_failed", http.StatusInternalServerError)
			}

			result.Created = append(result.Created, dataset.Name)
		}
	}

	roles, err := a.Store.Collimato.GetWorkspaceRoles(workspace.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	for _, role := range roles {
		if !role.AutoUpdate {
			continue
		}

		role.TablePermissions = append(role.TablePermissions, result.Created...)

		if err = a.Store.Collimato.UpdateWorkspaceRole(workspaceID, role); err != nil {
			tlog.Errorw("Failed to update workspace role",
				"workspace_id", workspaceID,
				"role_id", role.ID,
				"error", err,
			)
			return nil, model.NewAppError("role.update_failed", http.StatusInternalServerError)
		}
	}

	if workspace.Status == model.CollimatoWorkspaceStatusFinished {
		if err = a.UpdateWorkspaceRolePermissions(*workspace); err != nil {
			tlog.Errorw("Failed to update workspace role permissions",
				"workspace_id", workspaceID,
				"user_id", user.ID,
				"error", err,
			)
			return nil, model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
		}

		a.graphDataModelChanged(workspaceID)
	}

	return result, nil
}

func columnDetails(cols interface{}) ([]model.ColumnDetail, bool) {
	raw, ok := cols.([]interface{})
	if !ok {
		return nil, false
	}

	columns := make([]model.ColumnDetail, 0, len(raw))
	for _, c := range raw {
		entry, ok := c.(map[string]interface{})
		if !ok {
			return nil, false
		}

		name, nameOK := entry["name"].(string)
		columnType, typeOK := entry["type"].(string)
		if !nameOK || !typeOK {
			return nil, false
		}

		columns = append(columns, model.ColumnDetail{Name: name, Type: columnType})
	}

	return columns, true
}

func datasetFromTable(conn *model.Connection, datasetName string, table string, columns []model.ColumnDetail) (model.Dataset, []string) {
	dataset := model.Dataset{
		Name:         datasetName,
		SqlTable:     table,
		ConnectionID: conn.ID,
		Dimensions:   make([]model.DatasetDimension, 0, len(columns)),
		Measures:     []model.DatasetMeasure{{Name: "count", Type: "count"}},
	}

	skipped := make([]string, 0)

	for _, c := range columns {
		t := collimato.SuggestBuilderType(conn.Type, c.Type)
		if !datasetDimensionTypes[t] || !isValidIdentifier(c.Name) {
			skipped = append(skipped, c.Name)
			continue
		}

		dataset.Dimensions = append(dataset.Dimensions, model.DatasetDimension{
			Name: c.Name,
			SQL:  c.Name,
			Type: t,
		})
	}

	return dataset, skipped
}

// GetDatasetNames lists the names of a workspace's usable datasets and views.
func (a *App) GetDatasetNames(workspace model.CollimatoWorkspace) ([]string, error) {
	fieldList, err := a.QueryEngine.FieldList(context.Background(), workspace.ID)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fieldList.Datasets))
	for _, dataset := range fieldList.Datasets {
		names = append(names, dataset.Name)
	}

	return names, nil
}
