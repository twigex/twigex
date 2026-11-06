// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) EnsureGraphBuiltInTypes() {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	for _, entityType := range model.GraphBuiltInEntityTypes() {
		err := a.Store.Graph.CreateOrUpdateEntityType(ctx, entityType)
		if err != nil {
			tlog.Errorw("Failed to save built-in graph entity type", "type", entityType.Key, "error", err)
			return
		}
	}

	entityTypes, err := a.Store.Graph.GetAllEntityTypes(ctx)
	if err != nil {
		tlog.Errorw("Failed to retrieve graph entity types", "error", err)
		return
	}

	typeIDs := make(map[string]string, len(entityTypes))
	for _, entityType := range entityTypes {
		typeIDs[entityType.Key] = entityType.ID
	}

	for _, builtIn := range model.GraphBuiltInRelationTypes() {
		fromTypeID, fromFound := typeIDs[builtIn.FromType]
		toTypeID, toFound := typeIDs[builtIn.ToType]
		if !fromFound || !toFound {
			tlog.Errorw("Built-in graph relation type points to a missing entity type", "relation_type", builtIn.Key)
			continue
		}

		relationType := model.GraphRelationType{
			Key:         builtIn.Key,
			Name:        builtIn.Name,
			ReverseName: builtIn.ReverseName,
			FromTypeID:  fromTypeID,
			ToTypeID:    toTypeID,
			Kind:        builtIn.Kind,
			BuiltIn:     true,
		}

		err := a.Store.Graph.CreateOrUpdateRelationType(ctx, relationType)
		if err != nil {
			tlog.Errorw("Failed to save built-in graph relation type", "relation_type", builtIn.Key, "error", err)
			return
		}
	}
}
