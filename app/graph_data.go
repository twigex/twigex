// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Above this many neighbours a group shows "1000+" instead of a number.
const graphCountCap = 1000

var errGraphTypeBroken = errors.New("the type's dataset or field no longer exists")

// graphDataType is a user-defined entity type with its dataset's workspace
// and its field names looked up, ready to read entities with.
type graphDataType struct {
	entityType  model.GraphEntityType
	workspaceID string
	keyField    string
	titleField  string
}

// graphPropertyRelation is a property relation type with its field names looked up.
// The "to" type holds property, which carries the "from" entity's key, or
// the value of fromProperty when one is set.
type graphPropertyRelation struct {
	relationType model.GraphRelationType
	from         *graphDataType
	to           *graphDataType
	property     string
	fromProperty string
}

func (relation graphPropertyRelation) fromMatchField() string {
	if relation.fromProperty != "" {
		return relation.fromProperty
	}

	return relation.from.keyField
}

func graphText(value any) string {
	if value == nil {
		return ""
	}

	return fmt.Sprint(value)
}

func graphEquals(field, value string) model.QueryFilter {
	return model.QueryFilter{Member: field, Operator: "equals", Values: []string{value}}
}

// graphDataItem is a relation found in a field of the data.
func graphDataItem(entity model.GraphEntityRef, source string) model.GraphRelationItem {
	return model.GraphRelationItem{Entity: entity, Origin: model.GraphRelationOriginStructured, Source: source}
}

func (t *graphDataType) ref(row map[string]any) model.GraphEntityRef {
	return model.GraphEntityRef{
		Type:  t.entityType.Key,
		ID:    graphText(row[t.keyField]),
		Title: graphText(row[t.titleField]),
	}
}

func (a *App) graphFieldName(ctx context.Context, fieldID string) (string, error) {
	field, err := a.Store.Graph.GetDatasetFieldByID(ctx, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errGraphTypeBroken
	}

	if err != nil {
		return "", err
	}

	return field.Name, nil
}

func (a *App) loadGraphDataType(ctx context.Context, entityType model.GraphEntityType) (*graphDataType, error) {
	dataset, err := a.Store.Graph.GetDatasetByID(ctx, entityType.SourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errGraphTypeBroken
	}

	if err != nil {
		return nil, err
	}

	keyField, err := a.graphFieldName(ctx, entityType.KeyProperty)
	if err != nil {
		return nil, err
	}

	titleField, err := a.graphFieldName(ctx, entityType.TitleProperty)
	if err != nil {
		return nil, err
	}

	dataType := &graphDataType{
		entityType:  entityType,
		workspaceID: dataset.WorkspaceID,
		keyField:    keyField,
		titleField:  titleField,
	}

	return dataType, nil
}

// findGraphDataType finds a user-defined type by its key. A type whose
// dataset or field was removed is reported as not found and logged, rather
// than shown with no entities, which would look like real data.
func (a *App) findGraphDataType(ctx context.Context, ontology *graphOntology, key string) (*graphDataType, *model.AppError) {
	for _, entityType := range ontology.entityTypes {
		if entityType.Key != key || entityType.Source != model.GraphSourceDataset {
			continue
		}

		dataType, err := a.loadGraphDataType(ctx, entityType)
		if errors.Is(err, errGraphTypeBroken) {
			tlog.Warnw("A graph type points to a dataset or field that no longer exists", "type_key", key)

			return nil, graphEntityNotFound()
		}

		if err != nil {
			return nil, graphRelationsFailed("", err)
		}

		return dataType, nil
	}

	return nil, graphEntityNotFound()
}

// graphPropertyRelationsFor lists the property relation types that touch a type. A
// relation type whose field was removed is skipped and logged, never shown as an
// empty group.
func (a *App) graphPropertyRelationsFor(ctx context.Context, ontology *graphOntology, dataType *graphDataType) ([]graphPropertyRelation, *model.AppError) {
	typesByID := ontology.entityTypesByID()
	relations := make([]graphPropertyRelation, 0)

	for _, relationType := range ontology.relationTypes {
		if relationType.Kind != model.GraphRelationKindProperty || relationType.KindConfig == nil {
			continue
		}

		touches := relationType.FromTypeID == dataType.entityType.ID || relationType.ToTypeID == dataType.entityType.ID
		if !touches {
			continue
		}

		relation, err := a.loadGraphPropertyRelation(ctx, relationType, typesByID)
		if errors.Is(err, errGraphTypeBroken) {
			tlog.Warnw("Skipped a graph relation type whose type or field no longer exists", "relation_type_key", relationType.Key)

			continue
		}

		if err != nil {
			return nil, graphRelationsFailed("", err)
		}

		relations = append(relations, *relation)
	}

	return relations, nil
}

func (a *App) loadGraphPropertyRelation(ctx context.Context, relationType model.GraphRelationType, typesByID map[string]model.GraphEntityType) (*graphPropertyRelation, error) {
	fromType, fromFound := typesByID[relationType.FromTypeID]
	toType, toFound := typesByID[relationType.ToTypeID]

	if !fromFound || !toFound {
		return nil, errGraphTypeBroken
	}

	from, err := a.loadGraphDataType(ctx, fromType)
	if err != nil {
		return nil, err
	}

	to, err := a.loadGraphDataType(ctx, toType)
	if err != nil {
		return nil, err
	}

	property, err := a.graphFieldName(ctx, relationType.KindConfig.Property)
	if err != nil {
		return nil, err
	}

	relation := &graphPropertyRelation{
		relationType: relationType,
		from:         from,
		to:           to,
		property:     property,
	}

	if relationType.KindConfig.FromProperty != "" {
		fromProperty, err := a.graphFieldName(ctx, relationType.KindConfig.FromProperty)
		if err != nil {
			return nil, err
		}

		relation.fromProperty = fromProperty
	}

	return relation, nil
}

func (a *App) graphRows(ctx context.Context, user model.User, dataType *graphDataType, fields []string, filters []model.QueryFilter, after string, limit int) ([]map[string]any, *model.AppError) {
	request := model.CollimatoRowsRequest{
		Fields:  fields,
		Filters: filters,
		OrderBy: dataType.keyField,
		After:   after,
		Limit:   limit,
	}

	return a.CollimatoGetRows(ctx, user, dataType.workspaceID, request)
}

// graphOwnRow reads the entity itself, its key and name only. No row means
// the entity is missing or a rule hides it; both are answered the same way.
func (a *App) graphOwnRow(ctx context.Context, user model.User, dataType *graphDataType, id string) (map[string]any, *model.AppError) {
	fields := []string{dataType.keyField, dataType.titleField}
	filters := []model.QueryFilter{graphEquals(dataType.keyField, id)}

	rows, appErr := a.graphRows(ctx, user, dataType, fields, filters, "", 1)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return nil, graphEntityNotFound()
		}

		return nil, appErr
	}

	if len(rows) == 0 {
		return nil, graphEntityNotFound()
	}

	return rows[0], nil
}

// graphDataGroup reads a group in at most two short queries and no COUNT. The
// first page tells a small group apart; for a big one, a read of keys only,
// up to the cap, gives the number. A COUNT under a row rule has to open every
// row of the group (3.2 s for a million orders); this stops at the cap.
//
// A refused read, such as a property hidden by a column rule, means the user
// may not see the relation, so no group is returned.
func (a *App) graphDataGroup(ctx context.Context, user model.User, target *graphDataType, matchField, value, source string) (*model.GraphRelationGroup, *model.AppError) {
	filters := []model.QueryFilter{graphEquals(matchField, value)}
	fields := []string{target.keyField, target.titleField}

	rows, appErr := a.graphRows(ctx, user, target, fields, filters, "", graphGroupItemLimit+1)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return nil, nil
		}

		return nil, appErr
	}

	group := &model.GraphRelationGroup{
		EntityType: target.entityType.Key,
		Items:      make([]model.GraphRelationItem, 0),
	}

	if len(rows) <= graphGroupItemLimit {
		for _, row := range rows {
			group.Items = append(group.Items, graphDataItem(target.ref(row), source))
		}

		group.Count = len(group.Items)

		return group, nil
	}

	keys, appErr := a.graphRows(ctx, user, target, []string{target.keyField}, filters, "", graphCountCap+1)
	if appErr != nil {
		return nil, appErr
	}

	group.Count = len(keys)

	if group.Count > graphCountCap {
		group.Count = graphCountCap
		group.CountCapped = true
	}

	return group, nil
}

// graphPropertyValue reads one property of the entity on its own. Read
// together with the entity, a property hidden by a column rule would make
// Collimato refuse the whole read and hide the entity; read apart, it only
// hides the relation that needs it, which then answers with an empty value.
func (a *App) graphPropertyValue(ctx context.Context, user model.User, dataType *graphDataType, id, property string) (string, *model.AppError) {
	fields := []string{dataType.keyField, property}
	filters := []model.QueryFilter{graphEquals(dataType.keyField, id)}

	rows, appErr := a.graphRows(ctx, user, dataType, fields, filters, "", 1)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return "", nil
		}

		return "", appErr
	}

	if len(rows) == 0 {
		return "", nil
	}

	return graphText(rows[0][property]), nil
}

// graphRelationTarget says, for one end of a relation, which entities are on the other
// end: the type to read, the field to match on, and the value to match. An
// empty value means there is nothing to show at that end.
func (a *App) graphRelationTarget(ctx context.Context, user model.User, dataType *graphDataType, relation graphPropertyRelation, direction, id string) (*graphDataType, string, string, *model.AppError) {
	if direction == model.GraphDirectionOut && relation.fromProperty == "" {
		return relation.to, relation.property, id, nil
	}

	if direction == model.GraphDirectionOut {
		value, appErr := a.graphPropertyValue(ctx, user, dataType, id, relation.fromProperty)

		return relation.to, relation.property, value, appErr
	}

	value, appErr := a.graphPropertyValue(ctx, user, dataType, id, relation.property)

	return relation.from, relation.fromMatchField(), value, appErr
}

// graphRelationDirections lists the ends of a relation this type is on; both, for a
// relation between two entities of the same type.
func graphRelationDirections(relation graphPropertyRelation, dataType *graphDataType) []string {
	directions := make([]string, 0, 2)

	if relation.from.entityType.ID == dataType.entityType.ID {
		directions = append(directions, model.GraphDirectionOut)
	}

	if relation.to.entityType.ID == dataType.entityType.ID {
		directions = append(directions, model.GraphDirectionIn)
	}

	return directions
}

func (a *App) dataEntityGraphRelations(ctx context.Context, user model.User, ontology *graphOntology, typeKey, id string) (*model.GraphEntityRelations, *model.AppError) {
	dataType, appErr := a.findGraphDataType(ctx, ontology, typeKey)
	if appErr != nil {
		return nil, appErr
	}

	relations, appErr := a.graphPropertyRelationsFor(ctx, ontology, dataType)
	if appErr != nil {
		return nil, appErr
	}

	row, appErr := a.graphOwnRow(ctx, user, dataType, id)
	if appErr != nil {
		return nil, appErr
	}

	result := &model.GraphEntityRelations{
		Entity: dataType.ref(row),
		Groups: make([]model.GraphRelationGroup, 0),
	}

	for _, relation := range relations {
		for _, direction := range graphRelationDirections(relation, dataType) {
			target, matchField, value, appErr := a.graphRelationTarget(ctx, user, dataType, relation, direction, id)
			if appErr != nil {
				return nil, appErr
			}

			if value == "" {
				continue
			}

			group, appErr := a.graphDataGroup(ctx, user, target, matchField, value, relation.property)
			if appErr != nil {
				return nil, appErr
			}

			if group == nil || group.Count == 0 {
				continue
			}

			group.RelationType = relation.relationType.Key
			group.Direction = direction

			result.Groups = append(result.Groups, *group)
		}
	}

	return result, nil
}

// GetGraphRelationPage lists one group of an entity's relations, 50 at a
// time, for a group too large to send with the entity. Pages follow the
// group's own order, and after is the id of the last entity already shown.
func (a *App) GetGraphRelationPage(ctx context.Context, user model.User, typeKey, id, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	source, found, err := a.graphSourceFor(ctx, typeKey)
	if err != nil {
		return nil, graphRelationsFailed(id, err)
	}

	if !found {
		return nil, graphEntityNotFound()
	}

	return source.relationPage(ctx, user, typeKey, id, relationTypeKey, direction, after)
}

func (a *App) dataRelationPage(ctx context.Context, user model.User, ontology *graphOntology, typeKey, id, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError) {
	dataType, appErr := a.findGraphDataType(ctx, ontology, typeKey)
	if appErr != nil {
		return nil, appErr
	}

	relations, appErr := a.graphPropertyRelationsFor(ctx, ontology, dataType)
	if appErr != nil {
		return nil, appErr
	}

	index := slices.IndexFunc(relations, func(relation graphPropertyRelation) bool {
		return relation.relationType.Key == relationTypeKey
	})

	if index == -1 || !slices.Contains(graphRelationDirections(relations[index], dataType), direction) {
		return nil, graphEntityNotFound()
	}

	_, appErr = a.graphOwnRow(ctx, user, dataType, id)
	if appErr != nil {
		return nil, appErr
	}

	target, matchField, value, appErr := a.graphRelationTarget(ctx, user, dataType, relations[index], direction, id)
	if appErr != nil {
		return nil, appErr
	}

	if value == "" {
		return &model.GraphRelationPage{Items: make([]model.GraphRelationItem, 0)}, nil
	}

	filters := []model.QueryFilter{graphEquals(matchField, value)}
	fields := []string{target.keyField, target.titleField}

	rows, appErr := a.graphRows(ctx, user, target, fields, filters, after, graphGroupItemLimit+1)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return nil, graphEntityNotFound()
		}

		return nil, appErr
	}

	page := &model.GraphRelationPage{Items: make([]model.GraphRelationItem, 0, graphGroupItemLimit)}

	for _, row := range rows {
		if len(page.Items) == graphGroupItemLimit {
			page.Next = page.Items[graphGroupItemLimit-1].Entity.ID

			break
		}

		page.Items = append(page.Items, graphDataItem(target.ref(row), relations[index].property))
	}

	return page, nil
}

// SearchGraphEntities finds entities of a user-defined type by name, for
// picking where to start. A user who may not read the dataset gets no
// results rather than an error.
func (a *App) SearchGraphEntities(ctx context.Context, user model.User, typeKey, search string) ([]model.GraphEntityRef, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	ontology, err := a.loadGraphOntology(ctx)
	if err != nil {
		return nil, graphRelationsFailed("", err)
	}

	dataType, appErr := a.findGraphDataType(ctx, ontology, typeKey)
	if appErr != nil {
		return nil, appErr
	}

	filters := make([]model.QueryFilter, 0)
	if search != "" {
		contains := model.QueryFilter{Member: dataType.titleField, Operator: "contains", Values: []string{search}}

		filters = append(filters, contains)
	}

	fields := []string{dataType.keyField, dataType.titleField}

	rows, appErr := a.graphRows(ctx, user, dataType, fields, filters, "", graphGroupItemLimit)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return make([]model.GraphEntityRef, 0), nil
		}

		return nil, appErr
	}

	entities := make([]model.GraphEntityRef, 0, len(rows))
	for _, row := range rows {
		entities = append(entities, dataType.ref(row))
	}

	return entities, nil
}
