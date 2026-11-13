// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// GetGraphDatasetChoices lists, for each Collimato workspace the user is in,
// the datasets and fields they may build types on: only those they can see,
// and only dimensions, the same as saving a type checks. A workspace whose
// data model can't be read is left out, and the others are still listed.
func (a *App) GetGraphDatasetChoices(ctx context.Context, user model.User) ([]model.GraphDatasetChoice, *model.AppError) {
	appErr := a.guardGraphTypeChange(user)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaces, err := a.Store.Collimato.GetWorkspacesForUser(user.ID)
	if err != nil {
		return nil, graphTypesFailed(err)
	}

	choices := make([]model.GraphDatasetChoice, 0)

	for _, workspace := range workspaces {
		// The user's workspace list still holds workspaces that were deleted.
		if workspace.DeletedAt != 0 {
			continue
		}

		workspaceChoices, err := a.graphWorkspaceChoices(ctx, user, workspace)
		if err != nil {
			tlog.Warnw("Left a workspace out of the graph's dataset choices", "workspace_id", workspace.ID, "error", err)

			continue
		}

		choices = append(choices, workspaceChoices...)
	}

	return choices, nil
}

func (a *App) graphWorkspaceChoices(ctx context.Context, user model.User, workspace model.CollimatoWorkspace) ([]model.GraphDatasetChoice, error) {
	meta, appErr := a.graphVisibleMeta(ctx, user, workspace.ID)
	if appErr != nil {
		return nil, fmt.Errorf("data model: %s", appErr.ID)
	}

	datasets, err := a.Store.Graph.GetAllDatasetsForWorkspace(ctx, workspace.ID)
	if err != nil {
		return nil, err
	}

	fields, err := a.Store.Graph.GetAllDatasetFieldsForWorkspace(ctx, workspace.ID)
	if err != nil {
		return nil, err
	}

	choices := make([]model.GraphDatasetChoice, 0, len(datasets))

	for _, dataset := range datasets {
		if !meta.datasets[dataset.Name] {
			continue
		}

		choice := model.GraphDatasetChoice{
			ID:            dataset.ID,
			Name:          dataset.Name,
			WorkspaceID:   workspace.ID,
			WorkspaceName: workspace.Name,
			Fields:        make([]model.GraphFieldChoice, 0),
		}

		for _, field := range fields {
			if field.DatasetID != dataset.ID || !meta.dimensions[field.Name] {
				continue
			}

			fieldChoice := model.GraphFieldChoice{
				ID:    field.ID,
				Name:  field.Name,
				Title: meta.titles[field.Name],
			}

			choice.Fields = append(choice.Fields, fieldChoice)
		}

		choices = append(choices, choice)
	}

	return choices, nil
}
