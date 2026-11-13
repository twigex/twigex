// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// graphTypeVisibility decides which types one user may see. A user-made type
// shows only to those who can see its dataset and the fields it reads in
// Collimato, so its name never reaches someone outside the workspace. This is
// the one place that decides it; loadGraphOntology applies it to every
// request.
type graphTypeVisibility struct {
	app  *App
	user model.User

	// Filled one workspace at a time, the first time a type needs it.
	metas    map[string]*graphVisibleMeta
	datasets map[string]model.GraphDataset
	fields   map[string]model.GraphDatasetField
}

func (a *App) newGraphTypeVisibility(user model.User) *graphTypeVisibility {
	return &graphTypeVisibility{
		app:      a,
		user:     user,
		metas:    make(map[string]*graphVisibleMeta),
		datasets: make(map[string]model.GraphDataset),
		fields:   make(map[string]model.GraphDatasetField),
	}
}

func (v *graphTypeVisibility) canSeeEntityType(ctx context.Context, entityType model.GraphEntityType) (bool, error) {
	switch entityType.Source {
	case model.GraphSourceBuiltIn:
		return true, nil
	case model.GraphSourceDataset:
		return v.canSeeFields(ctx, entityType.SourceID, entityType.KeyProperty, entityType.TitleProperty)
	}

	// A source this version can't read has no data to check against, so only
	// managers see the type.
	return v.app.canManageGraph(v.user), nil
}

// canSeeRelationType needs both of its types to be visible, and for a
// property relation, the fields it matches on.
func (v *graphTypeVisibility) canSeeRelationType(ctx context.Context, relationType model.GraphRelationType, visibleTypes map[string]model.GraphEntityType) (bool, error) {
	from, fromVisible := visibleTypes[relationType.FromTypeID]
	to, toVisible := visibleTypes[relationType.ToTypeID]

	if !fromVisible || !toVisible {
		return false, nil
	}

	if relationType.KindConfig == nil {
		return true, nil
	}

	canSee, err := v.canSeeFields(ctx, to.SourceID, relationType.KindConfig.Property)
	if err != nil || !canSee {
		return false, err
	}

	if relationType.KindConfig.FromProperty == "" {
		return true, nil
	}

	return v.canSeeFields(ctx, from.SourceID, relationType.KindConfig.FromProperty)
}

// canSeeFields reports whether the user can see a dataset and some of its
// fields. A dataset that was removed names nothing any more, so its types are
// left to managers, who need to see them to clean them up. A removed field
// likewise doesn't hide what uses it.
func (v *graphTypeVisibility) canSeeFields(ctx context.Context, datasetID string, fieldIDs ...string) (bool, error) {
	dataset, err := v.dataset(ctx, datasetID)
	if errors.Is(err, sql.ErrNoRows) {
		return v.app.canManageGraph(v.user), nil
	}

	if err != nil {
		return false, err
	}

	meta := v.metas[dataset.WorkspaceID]

	if !meta.datasets[dataset.Name] {
		return false, nil
	}

	for _, fieldID := range fieldIDs {
		field, found := v.fields[fieldID]
		if !found {
			continue
		}

		if field.DatasetID != dataset.ID || !meta.fields[field.Name] {
			return false, nil
		}
	}

	return true, nil
}

func (v *graphTypeVisibility) dataset(ctx context.Context, id string) (*model.GraphDataset, error) {
	dataset, found := v.datasets[id]
	if found {
		return &dataset, nil
	}

	stored, err := v.app.Store.Graph.GetDatasetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = v.loadWorkspace(ctx, stored.WorkspaceID)
	if err != nil {
		return nil, err
	}

	v.datasets[stored.ID] = *stored

	return stored, nil
}

// loadWorkspace reads what the user can see in one workspace, and its
// datasets and fields, so the other types there need no further reads. A
// workspace whose data model can't be read shows the user nothing, and the
// rest of the graph still works.
func (v *graphTypeVisibility) loadWorkspace(ctx context.Context, workspaceID string) error {
	_, loaded := v.metas[workspaceID]
	if loaded {
		return nil
	}

	meta, appErr := v.app.graphVisibleMeta(ctx, v.user, workspaceID)
	if appErr != nil {
		tlog.Warnw("Hid a workspace's graph types because its data model could not be read", "workspace_id", workspaceID, "error", appErr.ID)

		meta = &graphVisibleMeta{}
	}

	v.metas[workspaceID] = meta

	datasets, err := v.app.Store.Graph.GetAllDatasetsForWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		v.datasets[dataset.ID] = dataset
	}

	fields, err := v.app.Store.Graph.GetAllDatasetFieldsForWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	for _, field := range fields {
		v.fields[field.ID] = field
	}

	return nil
}
