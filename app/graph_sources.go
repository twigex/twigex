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
// as Projects tables, is another graphSource; the callers stay the same.
type graphSource interface {
	relations(ctx context.Context, user model.User, typeKey, id string) (*model.GraphEntityRelations, *model.AppError)
}

// graphBuiltInSource serves Collimato's own types: datasets, their fields,
// charts and dashboards.
type graphBuiltInSource struct {
	app *App
	ids *graphBuiltInIDs
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

// graphSourceFor finds the source of a type from the type's own source, not
// from its key. found is false for an unknown type, and for a type whose
// source this version of Twigex can't read.
func (a *App) graphSourceFor(ctx context.Context, typeKey string) (graphSource, bool, error) {
	entityTypes, err := a.Store.Graph.GetAllEntityTypes(ctx)
	if err != nil {
		return nil, false, err
	}

	index := slices.IndexFunc(entityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.Key == typeKey
	})

	if index == -1 {
		return nil, false, nil
	}

	switch entityTypes[index].Source {
	case model.GraphSourceBuiltIn:
		ids, err := a.loadGraphBuiltInIDs(ctx)
		if err != nil {
			return nil, false, err
		}

		return graphBuiltInSource{app: a, ids: ids}, true, nil
	}

	return nil, false, nil
}
