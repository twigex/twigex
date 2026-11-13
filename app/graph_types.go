// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Keys appear in addresses, so they are kept to a plain, stable form.
var graphTypeKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// An icon is named, like "building-office"; the screen draws the icons it
// knows and a plain one for any other name.
var graphTypeIconPattern = regexp.MustCompile(`^[a-z0-9-]{1,32}$`)

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

// GetGraphTypes gives each user the types they may see, for the type map.
// Where a type's data comes from is only shown to those who manage types.
func (a *App) GetGraphTypes(ctx context.Context, user model.User) (*model.GraphTypes, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	ontology, err := a.loadGraphOntology(ctx, user)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	if a.canManageGraph(user) {
		return a.namedGraphTypes(ctx, ontology.entityTypes, ontology.relationTypes), nil
	}

	namesOnly := &model.GraphTypes{
		EntityTypes:   make([]model.GraphEntityType, 0, len(ontology.entityTypes)),
		RelationTypes: make([]model.GraphRelationType, 0, len(ontology.relationTypes)),
	}

	for _, entityType := range ontology.entityTypes {
		entityType.SourceID = ""
		entityType.KeyProperty = ""
		entityType.TitleProperty = ""

		namesOnly.EntityTypes = append(namesOnly.EntityTypes, entityType)
	}

	for _, relationType := range ontology.relationTypes {
		relationType.KindConfig = nil

		namesOnly.RelationTypes = append(namesOnly.RelationTypes, relationType)
	}

	return namesOnly, nil
}

// namedGraphTypes copies the types with the dataset and field names behind
// their IDs added, for the screen. A dataset or field that no longer exists
// just gets no name; the list itself still loads.
func (a *App) namedGraphTypes(ctx context.Context, storedEntityTypes []model.GraphEntityType, storedRelationTypes []model.GraphRelationType) *model.GraphTypes {
	entityTypes := slices.Clone(storedEntityTypes)
	relationTypes := slices.Clone(storedRelationTypes)

	names := make(map[string]string)

	fieldName := func(fieldID string) string {
		if fieldID == "" {
			return ""
		}

		name, known := names[fieldID]
		if known {
			return name
		}

		name, err := a.graphFieldName(ctx, fieldID)
		if err != nil && !errors.Is(err, errGraphTypeBroken) {
			tlog.Warnw("Failed to read a graph type's field name", "field_id", fieldID, "error", err)
		}

		names[fieldID] = name

		return name
	}

	for i := range entityTypes {
		if entityTypes[i].SourceID == "" {
			continue
		}

		dataset, err := a.Store.Graph.GetDatasetByID(ctx, entityTypes[i].SourceID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			tlog.Warnw("Failed to read a graph type's dataset name", "dataset_id", entityTypes[i].SourceID, "error", err)
		}

		if err == nil {
			entityTypes[i].SourceName = dataset.Name
		}

		entityTypes[i].KeyPropertyName = fieldName(entityTypes[i].KeyProperty)
		entityTypes[i].TitlePropertyName = fieldName(entityTypes[i].TitleProperty)
	}

	for i := range relationTypes {
		if relationTypes[i].KindConfig == nil {
			continue
		}

		kindConfig := *relationTypes[i].KindConfig
		kindConfig.PropertyName = fieldName(kindConfig.Property)
		kindConfig.FromPropertyName = fieldName(kindConfig.FromProperty)

		relationTypes[i].KindConfig = &kindConfig
	}

	return &model.GraphTypes{EntityTypes: entityTypes, RelationTypes: relationTypes}
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

	appErr = checkGraphTypeLook(input)
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
		PluralName:    strings.TrimSpace(input.PluralName),
		Icon:          input.Icon,
		Color:         input.Color,
		Status:        graphStatus(input),
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

	entityType, appErr := a.editableGraphEntityType(ctx, user, id)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeNames(input.Name)
	if appErr != nil {
		return nil, appErr
	}

	appErr = checkGraphTypeLook(input)
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
	entityType.PluralName = strings.TrimSpace(input.PluralName)
	entityType.Icon = input.Icon
	entityType.Color = input.Color
	entityType.Status = graphStatus(input)
	entityType.Description = strings.TrimSpace(input.Description)
	entityType.SourceID = input.SourceID
	entityType.KeyProperty = input.KeyProperty
	entityType.TitleProperty = input.TitleProperty
	entityType.UpdatedBy = user.ID
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

	entityType, appErr := a.editableGraphEntityType(ctx, user, id)
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

	cardinality, appErr := graphCardinality(input)
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
		Cardinality: cardinality,
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

	relationType, appErr := a.editableGraphRelationType(ctx, user, id)
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

	cardinality, appErr := graphCardinality(input)
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
	relationType.Cardinality = cardinality
	relationType.UpdatedBy = user.ID
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

	relationType, appErr := a.editableGraphRelationType(ctx, user, id)
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

// checkGraphTypeLook checks the values that only change how a type looks and
// is labelled. Each may be left empty: no icon, a colour from the list, and
// "active" for the status.
func checkGraphTypeLook(input model.GraphEntityTypeInput) *model.AppError {
	if utf8.RuneCountInString(input.PluralName) > graphTypeNameLimit {
		return model.NewAppError("graph.type_name_too_long", http.StatusBadRequest)
	}

	if input.Status != "" && !slices.Contains(model.GraphStatuses, input.Status) {
		return model.NewAppError("graph.type_status_invalid", http.StatusBadRequest)
	}

	if input.Color != "" && !slices.Contains(model.GraphColors, input.Color) {
		return model.NewAppError("graph.type_color_invalid", http.StatusBadRequest)
	}

	if input.Icon != "" && !graphTypeIconPattern.MatchString(input.Icon) {
		return model.NewAppError("graph.type_icon_invalid", http.StatusBadRequest)
	}

	return nil
}

// A type saved without a status is active, the same as the column's default,
// so the answer to the save says what the database holds.
func graphStatus(input model.GraphEntityTypeInput) string {
	if input.Status == "" {
		return model.GraphStatusActive
	}

	return input.Status
}

// A property relation's "to" type holds the "from" entity's key, so each "to"
// entity has one "from" entity, and one "from" entity has many: one to many,
// unless the admin says otherwise.
func graphCardinality(input model.GraphRelationTypeInput) (string, *model.AppError) {
	if input.Cardinality == "" {
		return model.GraphOneToMany, nil
	}

	if !slices.Contains(model.GraphCardinalities, input.Cardinality) {
		return "", model.NewAppError("graph.relation_type_cardinality_invalid", http.StatusBadRequest)
	}

	return input.Cardinality, nil
}

// Names are stored in columns of 255 characters.
const graphTypeNameLimit = 255

func checkGraphTypeNames(names ...string) *model.AppError {
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			return model.NewAppError("graph.type_name_required", http.StatusBadRequest)
		}

		if utf8.RuneCountInString(name) > graphTypeNameLimit {
			return model.NewAppError("graph.type_name_too_long", http.StatusBadRequest)
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

// editableGraphEntityType finds a type among those the user may see, so a
// manager can't change or delete a type built on data hidden from them.
func (a *App) editableGraphEntityType(ctx context.Context, user model.User, id string) (*model.GraphEntityType, *model.AppError) {
	ontology, err := a.loadGraphOntology(ctx, user)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	index := slices.IndexFunc(ontology.entityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.ID == id
	})

	if index == -1 {
		return nil, model.NewAppError("graph.type_not_found", http.StatusNotFound)
	}

	entityType := ontology.entityTypes[index]

	if entityType.BuiltIn {
		return nil, model.NewAppError("graph.type_built_in", http.StatusBadRequest)
	}

	return &entityType, nil
}

func (a *App) editableGraphRelationType(ctx context.Context, user model.User, id string) (*model.GraphRelationType, *model.AppError) {
	ontology, err := a.loadGraphOntology(ctx, user)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	index := slices.IndexFunc(ontology.relationTypes, func(relationType model.GraphRelationType) bool {
		return relationType.ID == id
	})

	if index == -1 {
		return nil, model.NewAppError("graph.type_not_found", http.StatusNotFound)
	}

	relationType := ontology.relationTypes[index]

	if relationType.BuiltIn {
		return nil, model.NewAppError("graph.type_built_in", http.StatusBadRequest)
	}

	return &relationType, nil
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
	ontology, err := a.loadGraphOntology(ctx, user)
	if err != nil {
		return graphTypesFailed(err)
	}

	fromType, appErr := graphDatasetEntityType(ontology, fromTypeID)
	if appErr != nil {
		return appErr
	}

	toType, appErr := graphDatasetEntityType(ontology, toTypeID)
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

func graphDatasetEntityType(ontology *graphOntology, id string) (*model.GraphEntityType, *model.AppError) {
	index := slices.IndexFunc(ontology.entityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.ID == id
	})

	if index == -1 {
		return nil, model.NewAppError("graph.type_not_found", http.StatusBadRequest)
	}

	entityType := ontology.entityTypes[index]

	if entityType.Source != model.GraphSourceDataset {
		return nil, model.NewAppError("graph.relation_type_needs_dataset_types", http.StatusBadRequest)
	}

	return &entityType, nil
}
