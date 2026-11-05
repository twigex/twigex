// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"slices"

	"github.com/twigex/twigex/model"
)

// graphSource reads the entities of the types one source serves, as one user
// sees them. Each source checks access the way its own app does, and answers
// an entity the user may not see as not found. Another kind of source, such
// as Projects tables, is another graphSource, with a case in graphSourceIn and
// in canSeeEntityType; the callers stay the same.
type graphSource interface {
	relations(ctx context.Context, user model.User, typeKey, id string) (*model.GraphEntityRelations, *model.AppError)

	// relationPage reads one group too large to send with its entity.
	relationPage(ctx context.Context, user model.User, typeKey, id, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError)
}

// graphSearchableSource is a source whose entities can be listed and searched
// by name, a page at a time.
type graphSearchableSource interface {
	search(ctx context.Context, user model.User, typeKey, search, after string) (*model.GraphEntityPage, *model.AppError)
}

// graphBuiltInSource serves Collimato's own types: datasets, their fields,
// charts and dashboards.
type graphBuiltInSource struct {
	app *App
	ids *graphBuiltInIDs
}

func (s graphBuiltInSource) relationPage(ctx context.Context, user model.User, typeKey, id, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError) {
	relations, appErr := s.relations(ctx, user, typeKey, id)
	if appErr != nil {
		return nil, appErr
	}

	return graphGroupPage(relations, relationTypeKey, direction, after)
}

func (s graphBuiltInSource) relations(ctx context.Context, user model.User, typeKey, id string) (*model.GraphEntityRelations, *model.AppError) {
	switch typeKey {
	case model.GraphTypeChart:
		return s.app.chartGraphRelations(ctx, user, s.ids, id)
	case model.GraphTypeDatasetField:
		return s.app.datasetFieldGraphRelations(ctx, user, s.ids, id)
	case model.GraphTypeDataset:
		return s.app.datasetGraphRelations(ctx, user, id)
	case model.GraphTypeDashboard:
		return s.app.dashboardGraphRelations(ctx, user, id)
	}

	return nil, graphEntityNotFound()
}

// graphDatasetSource serves the user-defined types read from a dataset.
type graphDatasetSource struct {
	app      *App
	ontology *graphOntology
}

func (s graphDatasetSource) relations(ctx context.Context, user model.User, typeKey, id string) (*model.GraphEntityRelations, *model.AppError) {
	return s.app.dataEntityGraphRelations(ctx, user, s.ontology, typeKey, id)
}

func (s graphDatasetSource) relationPage(ctx context.Context, user model.User, typeKey, id, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError) {
	return s.app.dataRelationPage(ctx, user, s.ontology, typeKey, id, relationTypeKey, direction, after)
}

func (s graphDatasetSource) search(ctx context.Context, user model.User, typeKey, search, after string) (*model.GraphEntityPage, *model.AppError) {
	return s.app.dataEntitySearch(ctx, user, s.ontology, typeKey, search, after)
}

// graphOntology is the graph's entity and relation types, read together and
// handed down to the steps of a request.
type graphOntology struct {
	entityTypes   []model.GraphEntityType
	relationTypes []model.GraphRelationType
}

// loadAllGraphTypes reads every type, for work no user asked for, such as a
// rebuild. A user's request reads them with loadGraphOntology instead.
func (a *App) loadAllGraphTypes(ctx context.Context) (*graphOntology, error) {
	entityTypes, err := a.Store.Graph.GetAllEntityTypes(ctx)
	if err != nil {
		return nil, err
	}

	relationTypes, err := a.Store.Graph.GetAllRelationTypes(ctx)
	if err != nil {
		return nil, err
	}

	return &graphOntology{entityTypes: entityTypes, relationTypes: relationTypes}, nil
}

// loadGraphOntology reads the types one user may see. Every user request
// starts here, so a type left out is gone for everything after it: the type
// list, its entities, search and the relations that lead to it.
func (a *App) loadGraphOntology(ctx context.Context, user model.User) (*graphOntology, error) {
	all, err := a.loadAllGraphTypes(ctx)
	if err != nil {
		return nil, err
	}

	visibility := a.newGraphTypeVisibility(user)

	visible := &graphOntology{
		entityTypes:   make([]model.GraphEntityType, 0, len(all.entityTypes)),
		relationTypes: make([]model.GraphRelationType, 0, len(all.relationTypes)),
	}

	visibleByID := make(map[string]model.GraphEntityType, len(all.entityTypes))

	for _, entityType := range all.entityTypes {
		canSee, err := visibility.canSeeEntityType(ctx, entityType)
		if err != nil {
			return nil, err
		}

		if !canSee {
			continue
		}

		visible.entityTypes = append(visible.entityTypes, entityType)
		visibleByID[entityType.ID] = entityType
	}

	for _, relationType := range all.relationTypes {
		canSee, err := visibility.canSeeRelationType(ctx, relationType, visibleByID)
		if err != nil {
			return nil, err
		}

		if canSee {
			visible.relationTypes = append(visible.relationTypes, relationType)
		}
	}

	return visible, nil
}

func (o *graphOntology) entityTypesByID() map[string]model.GraphEntityType {
	byID := make(map[string]model.GraphEntityType, len(o.entityTypes))

	for _, entityType := range o.entityTypes {
		byID[entityType.ID] = entityType
	}

	return byID
}

// graphSourceFor reads the types the user may see and finds the source of one
// of them.
func (a *App) graphSourceFor(ctx context.Context, user model.User, typeKey string) (graphSource, bool, error) {
	ontology, err := a.loadGraphOntology(ctx, user)
	if err != nil {
		return nil, false, err
	}

	return a.graphSourceIn(ontology, typeKey)
}

// graphSourceIn finds the source of a type from the type's own source, not
// from its key. found is false for an unknown type, and for a type whose
// source this version of Twigex can't read.
func (a *App) graphSourceIn(ontology *graphOntology, typeKey string) (graphSource, bool, error) {
	index := slices.IndexFunc(ontology.entityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.Key == typeKey
	})

	if index == -1 {
		return nil, false, nil
	}

	switch ontology.entityTypes[index].Source {
	case model.GraphSourceBuiltIn:
		ids, err := graphBuiltInIDsOf(ontology)
		if err != nil {
			return nil, false, err
		}

		return graphBuiltInSource{app: a, ids: ids}, true, nil
	case model.GraphSourceDataset:
		return graphDatasetSource{app: a, ontology: ontology}, true, nil
	}

	return nil, false, nil
}
