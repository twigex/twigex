// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Keys appear in addresses, so they are kept to a plain, stable form.
var graphTypeKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func graphTypesFailed(err error) *model.AppError {
	tlog.Errorw("Failed to read or save graph types", "error", err)

	return model.NewAppError("graph.types_failed", http.StatusInternalServerError)
}

func (a *App) canManageGraph(user model.User) bool {
	return a.SessionHasPermission(user, model.KnowledgeGraphPermissions.PermissionManageKnowledgeGraph)
}

func (a *App) guardGraphTypeChange(user model.User) *model.AppError {
	if !a.graphEnabled() {
		return model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	if !a.canManageGraph(user) {
		return model.NewAppError("graph.forbidden", http.StatusForbidden)
	}

	return nil
}

// GetGraphTypes gives every signed-in user the type names for the type map.
// Where a type's data comes from is only shown to those who manage types,
// because it names datasets and fields the user may not be allowed to see.
func (a *App) GetGraphTypes(ctx context.Context, user model.User) (*model.GraphTypes, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	entityTypes, err := a.Store.Graph.GetAllEntityTypes(ctx)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	relationTypes, err := a.Store.Graph.GetAllRelationTypes(ctx)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	if a.canManageGraph(user) {
		return &model.GraphTypes{EntityTypes: entityTypes, RelationTypes: relationTypes}, nil
	}

	namesOnly := &model.GraphTypes{
		EntityTypes:   make([]model.GraphEntityType, 0, len(entityTypes)),
		RelationTypes: make([]model.GraphRelationType, 0, len(relationTypes)),
	}

	for _, entityType := range entityTypes {
		entityType.SourceID = ""
		entityType.KeyProperty = ""
		entityType.TitleProperty = ""

		namesOnly.EntityTypes = append(namesOnly.EntityTypes, entityType)
	}

	for _, relationType := range relationTypes {
		relationType.KindConfig = nil

		namesOnly.RelationTypes = append(namesOnly.RelationTypes, relationType)
	}

	return namesOnly, nil
}

func (a *App) CreateGraphEntityType(ctx context.Context, user model.User, input model.GraphEntityTypeInput) (*model.GraphEntityType, *model.AppError) {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	appErr = checkGraphTypeKey(input.Key)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeNames(input.Name)
	if appErr != nil {
		return nil, appErr
	}

	appErr = a.checkGraphTypeFields(ctx, user, input.SourceID, input.KeyProperty, input.TitleProperty)
	if appErr != nil {
		return nil, appErr
	}

	now := time.Now().Unix()

	entityType := model.GraphEntityType{
		ID:            model.NewID(),
		Key:           input.Key,
		Name:          strings.TrimSpace(input.Name),
		Source:        model.GraphSourceDataset,
		SourceID:      input.SourceID,
		KeyProperty:   input.KeyProperty,
		TitleProperty: input.TitleProperty,
		Description:   strings.TrimSpace(input.Description),
		CreatedBy:     user.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := a.Store.Graph.CreateEntityType(ctx, entityType)
	if errors.Is(err, model.ErrGraphTypeKeyTaken) {
		return nil, model.NewAppError("graph.type_key_taken", http.StatusConflict)
	}

	if err != nil {
		return nil, graphTypesFailed(err)
	}

	tlog.Infow("A user created a graph entity type", "user_id", user.ID, "type_key", entityType.Key)

	return &entityType, nil
}

// UpdateGraphEntityType keeps the key, which addresses and relations rely on. The
// dataset may only change while no relation type uses the type, because those
// relations name fields of the old dataset.
func (a *App) UpdateGraphEntityType(ctx context.Context, user model.User, id string, input model.GraphEntityTypeInput) (*model.GraphEntityType, *model.AppError) {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	entityType, appErr := a.editableGraphEntityType(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeNames(input.Name)
	if appErr != nil {
		return nil, appErr
	}

	if input.SourceID != entityType.SourceID {
		appErr = a.checkGraphEntityTypeUnused(ctx, id)
		if appErr != nil {
			return nil, appErr
		}
	}

	appErr = a.checkGraphTypeFields(ctx, user, input.SourceID, input.KeyProperty, input.TitleProperty)
	if appErr != nil {
		return nil, appErr
	}

	entityType.Name = strings.TrimSpace(input.Name)
	entityType.Description = strings.TrimSpace(input.Description)
	entityType.SourceID = input.SourceID
	entityType.KeyProperty = input.KeyProperty
	entityType.TitleProperty = input.TitleProperty
	entityType.UpdatedAt = time.Now().Unix()

	err := a.Store.Graph.UpdateEntityType(ctx, *entityType)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	tlog.Infow("A user updated a graph entity type", "user_id", user.ID, "type_key", entityType.Key)

	return entityType, nil
}

func (a *App) DeleteGraphEntityType(ctx context.Context, user model.User, id string) *model.AppError {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	entityType, appErr := a.editableGraphEntityType(ctx, id)
	if appErr != nil {
		return appErr
	}

	appErr = a.checkGraphEntityTypeUnused(ctx, id)
	if appErr != nil {
		return appErr
	}

	err := a.Store.Graph.DeleteEntityType(ctx, id)
	if err != nil {
		return graphTypesFailed(err)
	}

	tlog.Infow("A user deleted a graph entity type", "user_id", user.ID, "type_key", entityType.Key)

	return nil
}

func (a *App) CreateGraphRelationType(ctx context.Context, user model.User, input model.GraphRelationTypeInput) (*model.GraphRelationType, *model.AppError) {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	appErr = checkGraphTypeKey(input.Key)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeNames(input.Name, input.ReverseName)
	if appErr != nil {
		return nil, appErr
	}

	appErr = a.checkGraphRelationProperties(ctx, user, input.FromTypeID, input.ToTypeID, input.Property, input.FromProperty)
	if appErr != nil {
		return nil, appErr
	}

	now := time.Now().Unix()

	relationType := model.GraphRelationType{
		ID:          model.NewID(),
		Key:         input.Key,
		Name:        strings.TrimSpace(input.Name),
		ReverseName: strings.TrimSpace(input.ReverseName),
		FromTypeID:  input.FromTypeID,
		ToTypeID:    input.ToTypeID,
		Kind:        model.GraphRelationKindProperty,
		KindConfig: &model.GraphRelationKindConfig{
			Property:     input.Property,
			FromProperty: input.FromProperty,
		},
		Description: strings.TrimSpace(input.Description),
		CreatedBy:   user.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := a.Store.Graph.CreateRelationType(ctx, relationType)
	if errors.Is(err, model.ErrGraphTypeKeyTaken) {
		return nil, model.NewAppError("graph.type_key_taken", http.StatusConflict)
	}

	if err != nil {
		return nil, graphTypesFailed(err)
	}

	tlog.Infow("A user created a graph relation type", "user_id", user.ID, "relation_type_key", relationType.Key)

	return &relationType, nil
}

// UpdateGraphRelationType keeps the key and both types; a relation between other
// types is a different relation type.
func (a *App) UpdateGraphRelationType(ctx context.Context, user model.User, id string, input model.GraphRelationTypeInput) (*model.GraphRelationType, *model.AppError) {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	relationType, appErr := a.editableGraphRelationType(ctx, id)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeNames(input.Name, input.ReverseName)
	if appErr != nil {
		return nil, appErr
	}

	appErr = a.checkGraphRelationProperties(ctx, user, relationType.FromTypeID, relationType.ToTypeID, input.Property, input.FromProperty)
	if appErr != nil {
		return nil, appErr
	}

	relationType.Name = strings.TrimSpace(input.Name)
	relationType.ReverseName = strings.TrimSpace(input.ReverseName)
	relationType.Description = strings.TrimSpace(input.Description)
	relationType.KindConfig = &model.GraphRelationKindConfig{
		Property:     input.Property,
		FromProperty: input.FromProperty,
	}
	relationType.UpdatedAt = time.Now().Unix()

	err := a.Store.Graph.UpdateRelationType(ctx, *relationType)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	tlog.Infow("A user updated a graph relation type", "user_id", user.ID, "relation_type_key", relationType.Key)

	return relationType, nil
}

func (a *App) DeleteGraphRelationType(ctx context.Context, user model.User, id string) *model.AppError {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	relationType, appErr := a.editableGraphRelationType(ctx, id)
	if appErr != nil {
		return appErr
	}

	err := a.Store.Graph.DeleteRelationType(ctx, id)
	if err != nil {
		return graphTypesFailed(err)
	}

	tlog.Infow("A user deleted a graph relation type", "user_id", user.ID, "relation_type_key", relationType.Key)

	return nil
}

func checkGraphTypeKey(key string) *model.AppError {
	if !graphTypeKeyPattern.MatchString(key) {
		return model.NewAppError("graph.type_key_invalid", http.StatusBadRequest)
	}

	return nil
}

func checkGraphTypeNames(names ...string) *model.AppError {
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			return model.NewAppError("graph.type_name_required", http.StatusBadRequest)
		}
	}

	return nil
}

func (a *App) checkGraphEntityTypeUnused(ctx context.Context, id string) *model.AppError {
	relationTypes, err := a.Store.Graph.GetAllRelationTypes(ctx)
	if err != nil {
		return graphTypesFailed(err)
	}

	for _, relationType := range relationTypes {
		if relationType.FromTypeID == id || relationType.ToTypeID == id {
			return model.NewAppError("graph.type_in_use", http.StatusConflict)
		}
	}

	return nil
}

func (a *App) editableGraphEntityType(ctx context.Context, id string) (*model.GraphEntityType, *model.AppError) {
	entityType, err := a.Store.Graph.GetEntityTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.NewAppError("graph.type_not_found", http.StatusNotFound)
	}

	if err != nil {
		return nil, graphTypesFailed(err)
	}

	if entityType.BuiltIn {
		return nil, model.NewAppError("graph.type_built_in", http.StatusBadRequest)
	}

	return entityType, nil
}

func (a *App) editableGraphRelationType(ctx context.Context, id string) (*model.GraphRelationType, *model.AppError) {
	relationType, err := a.Store.Graph.GetRelationTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.NewAppError("graph.type_not_found", http.StatusNotFound)
	}

	if err != nil {
		return nil, graphTypesFailed(err)
	}

	if relationType.BuiltIn {
		return nil, model.NewAppError("graph.type_built_in", http.StatusBadRequest)
	}

	return relationType, nil
}

// checkGraphTypeFields checks that every field belongs to the dataset and is a
// dimension the user can see in Collimato. That one check covers both rules:
// nobody builds a type on data they may not see, and a primary key Cube hides
// (it does so by default) can't be used as a key.
func (a *App) checkGraphTypeFields(ctx context.Context, user model.User, datasetID string, fieldIDs ...string) *model.AppError {
	dataset, err := a.Store.Graph.GetDatasetByID(ctx, datasetID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.NewAppError("graph.type_dataset_not_found", http.StatusBadRequest)
	}

	if err != nil {
		return graphTypesFailed(err)
	}

	meta, appErr := a.graphVisibleMeta(ctx, user, dataset.WorkspaceID)
	if appErr != nil {
		return appErr
	}

	if !meta.datasets[dataset.Name] {
		return model.NewAppError("graph.type_dataset_not_found", http.StatusBadRequest)
	}

	for _, fieldID := range fieldIDs {
		field, err := a.Store.Graph.GetDatasetFieldByID(ctx, fieldID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.NewAppError("graph.type_field_not_found", http.StatusBadRequest)
		}

		if err != nil {
			return graphTypesFailed(err)
		}

		if field.DatasetID != dataset.ID {
			return model.NewAppError("graph.type_field_not_found", http.StatusBadRequest)
		}

		if meta.fields[field.Name] && !meta.dimensions[field.Name] {
			return model.NewAppError("graph.type_field_is_measure", http.StatusBadRequest)
		}

		if !meta.dimensions[field.Name] {
			return model.NewAppError("graph.type_field_not_visible", http.StatusBadRequest)
		}
	}

	return nil
}

// checkGraphRelationProperties checks a property relation: the property is a field of
// the "to" type's dataset, and the optional from-property a field of the
// "from" type's dataset. Both types must read their entities from datasets.
func (a *App) checkGraphRelationProperties(ctx context.Context, user model.User, fromTypeID, toTypeID, property, fromProperty string) *model.AppError {
	fromType, appErr := a.graphDatasetEntityType(ctx, fromTypeID)
	if appErr != nil {
		return appErr
	}

	toType, appErr := a.graphDatasetEntityType(ctx, toTypeID)
	if appErr != nil {
		return appErr
	}

	appErr = a.checkGraphTypeFields(ctx, user, toType.SourceID, property)
	if appErr != nil {
		return appErr
	}

	if fromProperty == "" {
		return nil
	}

	return a.checkGraphTypeFields(ctx, user, fromType.SourceID, fromProperty)
}

func (a *App) graphDatasetEntityType(ctx context.Context, id string) (*model.GraphEntityType, *model.AppError) {
	entityType, err := a.Store.Graph.GetEntityTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.NewAppError("graph.type_not_found", http.StatusBadRequest)
	}

	if err != nil {
		return nil, graphTypesFailed(err)
	}

	if entityType.Source != model.GraphSourceDataset {
		return nil, model.NewAppError("graph.relation_type_needs_dataset_types", http.StatusBadRequest)
	}

	return entityType, nil
}
