// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/twigex/twigex/app/jobs"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// EnsureGraphBuiltInTypes saves the built-in types by key. A key that a
// user's own type already holds is left alone, because saving over it would
// turn their type into a built-in one; the clash is logged instead.
func (a *App) EnsureGraphBuiltInTypes() {
	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	existingEntityTypes, err := a.Store.Graph.GetAllEntityTypes(ctx)
	if err != nil {
		tlog.Errorw("Failed to retrieve graph entity types", "error", err)
		return
	}

	userEntityKeys := make(map[string]bool)
	for _, entityType := range existingEntityTypes {
		if !entityType.BuiltIn {
			userEntityKeys[entityType.Key] = true
		}
	}

	for _, entityType := range model.GraphBuiltInEntityTypes() {
		if userEntityKeys[entityType.Key] {
			tlog.Errorw("A user's graph entity type has the key a built-in type needs", "type", entityType.Key)
			continue
		}

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
		if entityType.BuiltIn {
			typeIDs[entityType.Key] = entityType.ID
		}
	}

	existingRelationTypes, err := a.Store.Graph.GetAllRelationTypes(ctx)
	if err != nil {
		tlog.Errorw("Failed to retrieve graph relation types", "error", err)
		return
	}

	userRelationKeys := make(map[string]bool)
	for _, relationType := range existingRelationTypes {
		if !relationType.BuiltIn {
			userRelationKeys[relationType.Key] = true
		}
	}

	for _, builtIn := range model.GraphBuiltInRelationTypes() {
		if userRelationKeys[builtIn.Key] {
			tlog.Errorw("A user's graph relation type has the key a built-in type needs", "relation_type", builtIn.Key)
			continue
		}

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
			Cardinality: builtIn.Cardinality,
			BuiltIn:     true,
		}

		err := a.Store.Graph.CreateOrUpdateRelationType(ctx, relationType)
		if err != nil {
			tlog.Errorw("Failed to save built-in graph relation type", "relation_type", builtIn.Key, "error", err)
			return
		}
	}
}

func (a *App) graphEnabled() bool {
	return *a.ConfigStore.Config.FeatureFlags.KnowledgeGraph
}

type graphBuiltInIDs struct {
	chartType              string
	datasetFieldType       string
	chartUsesFieldRelation string
}

func (a *App) loadGraphBuiltInIDs(ctx context.Context) (*graphBuiltInIDs, error) {
	ontology, err := a.loadGraphOntology(ctx)
	if err != nil {
		return nil, err
	}

	return graphBuiltInIDsOf(ontology)
}

func graphBuiltInIDsOf(ontology *graphOntology) (*graphBuiltInIDs, error) {
	ids := graphBuiltInIDs{}

	for _, entityType := range ontology.entityTypes {
		if !entityType.BuiltIn {
			continue
		}

		switch entityType.Key {
		case model.GraphTypeChart:
			ids.chartType = entityType.ID
		case model.GraphTypeDatasetField:
			ids.datasetFieldType = entityType.ID
		}
	}

	for _, relationType := range ontology.relationTypes {
		if !relationType.BuiltIn {
			continue
		}

		if relationType.Key == model.GraphRelationChartUsesDatasetField {
			ids.chartUsesFieldRelation = relationType.ID
		}
	}

	if ids.chartType == "" || ids.datasetFieldType == "" || ids.chartUsesFieldRelation == "" {
		return nil, errors.New("the built-in graph types are missing")
	}

	return &ids, nil
}

func graphChartFields(chart model.Chart) ([]string, error) {
	if chart.ChartType == model.ChartTypeMap {
		raw, err := json.Marshal(chart.Configuration)
		if err != nil {
			return nil, err
		}

		configuration := model.GraphMapConfiguration{}
		err = json.Unmarshal(raw, &configuration)
		if err != nil {
			return nil, err
		}

		return model.GraphMapFields(configuration), nil
	}

	raw, err := json.Marshal(chart.Data["query"])
	if err != nil {
		return nil, err
	}

	query := model.DataQuery{}
	err = json.Unmarshal(raw, &query)
	if err != nil {
		return nil, err
	}

	return model.GraphQueryFields(query), nil
}

func (a *App) updateChartGraphRelations(ctx context.Context, ids *graphBuiltInIDs, chart model.Chart) (int64, int64, error) {
	fieldNames, err := graphChartFields(chart)
	if err != nil {
		return 0, 0, err
	}

	fields, err := a.Store.Graph.GetDatasetFieldsByNames(ctx, chart.WorkspaceID, fieldNames)
	if err != nil {
		return 0, 0, err
	}

	fieldIDs := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldIDs = append(fieldIDs, field.ID)
	}

	return a.Store.Graph.UpdateStructuredRelations(ctx, ids.chartUsesFieldRelation, ids.chartType, chart.ID, ids.datasetFieldType, fieldIDs)
}

func (a *App) syncGraphDatasets(ctx context.Context, workspaceID string, ids *graphBuiltInIDs) (model.GraphRebuildResult, error) {
	result := model.GraphRebuildResult{}

	collection, err := a.QueryEngine.Meta(ctx, workspaceID)
	if err != nil {
		return result, err
	}

	// A data model that fails to compile comes back as an error with no
	// datasets, and reading that as "every dataset was removed" would delete
	// them all.
	if collection.Error != "" {
		return result, errors.New(collection.Error)
	}

	wantedKinds := make(map[string]string, len(collection.Cubes))
	wantedFields := make(map[string]string)

	for _, cube := range collection.Cubes {
		kind := model.GraphDatasetKindCube
		if cube.Type == model.GraphDatasetKindView {
			kind = model.GraphDatasetKindView
		}

		wantedKinds[cube.Name] = kind

		for _, dimension := range cube.Dimensions {
			wantedFields[dimension.Name] = cube.Name
		}

		for _, measure := range cube.Measures {
			wantedFields[measure.Name] = cube.Name
		}
	}

	existingDatasets, err := a.Store.Graph.GetAllDatasetsForWorkspace(ctx, workspaceID)
	if err != nil {
		return result, err
	}

	existingFields, err := a.Store.Graph.GetAllDatasetFieldsForWorkspace(ctx, workspaceID)
	if err != nil {
		return result, err
	}

	knownDatasets := make(map[string]bool, len(existingDatasets))
	removedDatasetIDs := make([]string, 0)

	for _, dataset := range existingDatasets {
		knownDatasets[dataset.Name] = true

		if _, wanted := wantedKinds[dataset.Name]; !wanted {
			removedDatasetIDs = append(removedDatasetIDs, dataset.ID)
		}
	}

	knownFields := make(map[string]bool, len(existingFields))
	removedFieldIDs := make([]string, 0)

	for _, field := range existingFields {
		knownFields[field.Name] = true

		if _, wanted := wantedFields[field.Name]; !wanted {
			removedFieldIDs = append(removedFieldIDs, field.ID)
		}
	}

	err = a.Store.Graph.DeleteEntityRelations(ctx, ids.datasetFieldType, removedFieldIDs)
	if err != nil {
		return result, err
	}

	err = a.Store.Graph.DeleteDatasetFields(ctx, removedFieldIDs)
	if err != nil {
		return result, err
	}

	err = a.Store.Graph.DeleteDatasets(ctx, removedDatasetIDs)
	if err != nil {
		return result, err
	}

	result.FieldsRemoved = int64(len(removedFieldIDs))
	result.DatasetsRemoved = int64(len(removedDatasetIDs))

	newDatasets := make([]model.GraphDataset, 0)
	for name, kind := range wantedKinds {
		if knownDatasets[name] {
			continue
		}

		newDatasets = append(newDatasets, model.GraphDataset{WorkspaceID: workspaceID, Name: name, Kind: kind})
	}

	result.DatasetsAdded, err = a.Store.Graph.AddDatasets(ctx, newDatasets)
	if err != nil {
		return result, err
	}

	datasets, err := a.Store.Graph.GetAllDatasetsForWorkspace(ctx, workspaceID)
	if err != nil {
		return result, err
	}

	datasetIDs := make(map[string]string, len(datasets))
	for _, dataset := range datasets {
		datasetIDs[dataset.Name] = dataset.ID
	}

	newFields := make([]model.GraphDatasetField, 0)
	for name, datasetName := range wantedFields {
		if knownFields[name] {
			continue
		}

		newFields = append(newFields, model.GraphDatasetField{WorkspaceID: workspaceID, DatasetID: datasetIDs[datasetName], Name: name})
	}

	result.FieldsAdded, err = a.Store.Graph.AddDatasetFields(ctx, newFields)
	if err != nil {
		return result, err
	}

	return result, nil
}

// rebuildGraphWorkspace rebuilds the relations of every chart after the fields are synced,
// because a chart saved while one of its fields was missing has no relation to it.
// Each step gets its own query timeout, as a workspace with many charts takes
// longer in total than a single timeout allows.
func (a *App) rebuildGraphWorkspace(ctx context.Context, workspaceID string, ids *graphBuiltInIDs) (model.GraphRebuildResult, error) {
	syncCtx, cancel := a.dbCtx(ctx)
	result, err := a.syncGraphDatasets(syncCtx, workspaceID, ids)
	cancel()
	if err != nil {
		return result, err
	}

	charts, err := a.Store.Collimato.GetCharts(workspaceID)
	if err != nil {
		return result, err
	}

	for _, chart := range charts {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		chartCtx, cancel := a.dbCtx(ctx)
		added, removed, err := a.updateChartGraphRelations(chartCtx, ids, chart)
		cancel()
		if err != nil {
			return result, err
		}

		result.RelationsAdded += added
		result.RelationsRemoved += removed
	}

	return result, nil
}

func (a *App) RebuildGraph(ctx context.Context) (model.GraphRebuildResult, error) {
	result := model.GraphRebuildResult{}

	idsCtx, cancel := a.dbCtx(ctx)
	ids, err := a.loadGraphBuiltInIDs(idsCtx)
	cancel()
	if err != nil {
		return result, err
	}

	workspaces, err := a.Store.Collimato.GetWorkspaces()
	if err != nil {
		return result, err
	}

	for _, workspace := range workspaces {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		if workspace.DeletedAt != 0 {
			continue
		}

		result.Workspaces++

		workspaceResult, err := a.rebuildGraphWorkspace(ctx, workspace.ID, ids)

		result.Add(workspaceResult)

		if err != nil {
			tlog.Errorw("Failed to rebuild the graph for a workspace", "workspace_id", workspace.ID, "error", err)
			result.FailedWorkspaces++
		}
	}

	err = a.Store.SystemSettings.UpdateBatch(map[string]string{model.GraphBuiltSetting: "true"})
	if err != nil {
		return result, err
	}

	return result, nil
}

// QueueGraphRebuildIfNeeded queues a rebuild at startup when the graph has not
// been built: the first time the flag is on, or after it was off, since saves
// made while it was off wrote no relations.
func (a *App) QueueGraphRebuildIfNeeded() {
	built, err := a.Store.SystemSettings.Get(model.GraphBuiltSetting)
	if err != nil {
		tlog.Errorw("Failed to read the graph status", "error", err)
		return
	}

	if built == "true" {
		return
	}

	job, err := a.enqueueGraphRebuild()
	if err != nil {
		tlog.Errorw("Could not queue a graph rebuild", "error", err)
		return
	}

	if job == nil {
		return
	}

	tlog.Infow("Queued a graph rebuild because the graph has not been built")
}

// enqueueGraphRebuild queues a rebuild unless one is already waiting or
// running, in which case the job is nil.
func (a *App) enqueueGraphRebuild() (*model.Job, error) {
	active, err := a.Store.Jobs.CountActiveByType(model.JobTypeGraphRebuild)
	if err != nil {
		return nil, err
	}

	if active > 0 {
		return nil, nil
	}

	return jobs.Enqueue(a.Store.Jobs, model.JobTypeGraphRebuild)
}

// ClearGraphBuilt runs while the flag is off. Saves write no relations then, so the
// graph goes stale and must be rebuilt when the flag comes back on.
func (a *App) ClearGraphBuilt() {
	err := a.Store.SystemSettings.UpdateBatch(map[string]string{model.GraphBuiltSetting: ""})
	if err != nil {
		tlog.Errorw("Failed to reset the graph status", "error", err)
	}
}

func (a *App) RequestGraphRebuild(user model.User) (*model.Job, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	if !a.SessionHasPermission(user, model.KnowledgeGraphPermissions.PermissionManageKnowledgeGraph) {
		return nil, model.NewAppError("graph.forbidden", http.StatusForbidden)
	}

	job, err := a.enqueueGraphRebuild()
	if err != nil {
		tlog.Errorw("Could not queue a graph rebuild", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("graph.rebuild_enqueue_failed", http.StatusInternalServerError)
	}

	if job == nil {
		return nil, model.NewAppError("graph.rebuild_already_running", http.StatusConflict)
	}

	tlog.Infow("A user requested a graph rebuild", "user_id", user.ID, "job_id", job.ID)

	return job, nil
}

func (a *App) GetGraphRebuildStatus(user model.User) (*model.GraphRebuildStatus, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	if !a.SessionHasPermission(user, model.KnowledgeGraphPermissions.PermissionManageKnowledgeGraph) {
		return nil, model.NewAppError("graph.forbidden", http.StatusForbidden)
	}

	built, err := a.Store.SystemSettings.Get(model.GraphBuiltSetting)
	if err != nil {
		tlog.Errorw("Failed to read the graph status", "error", err)
		return nil, model.NewAppError("graph.status_failed", http.StatusInternalServerError)
	}

	job, err := a.Store.Jobs.GetLatestByType(model.JobTypeGraphRebuild)
	if err != nil {
		tlog.Errorw("Failed to read the last graph rebuild", "error", err)
		return nil, model.NewAppError("graph.status_failed", http.StatusInternalServerError)
	}

	status := &model.GraphRebuildStatus{
		Built:  built == "true",
		Status: model.GraphRebuildNeverRun,
	}

	if job == nil {
		return status, nil
	}

	result, err := job.GraphRebuildPayload()
	if err != nil {
		tlog.Errorw("Failed to read the graph rebuild result", "job_id", job.ID, "error", err)
		return nil, model.NewAppError("graph.status_failed", http.StatusInternalServerError)
	}

	status.Status = job.Status
	status.Result = *result
	status.StartedAt = job.StartedAt
	status.FinishedAt = job.FinishedAt

	if job.Error != nil {
		status.Reason = *job.Error
	}

	return status, nil
}

func (a *App) graphChartSaved(chart model.Chart) {
	if !a.graphEnabled() {
		return
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	ids, err := a.loadGraphBuiltInIDs(ctx)
	if err != nil {
		tlog.Errorw("Failed to load the built-in graph types", "error", err)
		return
	}

	_, _, err = a.updateChartGraphRelations(ctx, ids, chart)
	if err != nil {
		tlog.Errorw("Failed to update the chart's graph relations", "chart_id", chart.ID, "error", err)
	}
}

func (a *App) graphChartDeleted(chartID string) {
	if !a.graphEnabled() {
		return
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	ids, err := a.loadGraphBuiltInIDs(ctx)
	if err != nil {
		tlog.Errorw("Failed to load the built-in graph types", "error", err)
		return
	}

	err = a.Store.Graph.DeleteEntityRelations(ctx, ids.chartType, []string{chartID})
	if err != nil {
		tlog.Errorw("Failed to remove the chart's graph relations", "chart_id", chartID, "error", err)
	}
}

func (a *App) graphDataModelChanged(workspaceID string) {
	if !a.graphEnabled() {
		return
	}

	ctx, cancel := a.dbCtx(context.Background())
	defer cancel()

	ids, err := a.loadGraphBuiltInIDs(ctx)
	if err != nil {
		tlog.Errorw("Failed to load the built-in graph types", "error", err)
		return
	}

	_, err = a.rebuildGraphWorkspace(ctx, workspaceID, ids)
	if err != nil {
		tlog.Errorw("Failed to update the graph after a data model change", "workspace_id", workspaceID, "error", err)
	}
}
