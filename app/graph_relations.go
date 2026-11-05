// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

const graphGroupItemLimit = 50

type graphVisibleMeta struct {
	datasets   map[string]bool
	fields     map[string]bool
	dimensions map[string]bool
	titles     map[string]string
}

func graphEntityNotFound() *model.AppError {
	return model.NewAppError("graph.entity_not_found", http.StatusNotFound)
}

func graphRelationsFailed(entityID string, err error) *model.AppError {
	tlog.Errorw("Failed to read the graph relations", "entity_id", entityID, "error", err)

	return model.NewAppError("graph.relations_failed", http.StatusInternalServerError)
}

// graphCheckRefused reports whether Collimato refused a check rather than
// failing. A refused entity is answered as not found, because saying "no
// access" would reveal that it exists.
func graphCheckRefused(appErr *model.AppError) bool {
	return appErr.Status < http.StatusInternalServerError
}

// GetGraphEntityRelations lists an entity's neighbours for one user. The graph
// decides nothing itself: every entity, at both ends of every relation, passes
// the same Collimato check that guards it on its own page.
func (a *App) GetGraphEntityRelations(ctx context.Context, user model.User, entityType, entityID string) (*model.GraphEntityRelations, *model.AppError) {
	if !a.graphEnabled() {
		return nil, model.NewAppError("graph.disabled", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	source, found, err := a.graphSourceFor(ctx, user, entityType)
	if err != nil {
		return nil, graphRelationsFailed(entityID, err)
	}

	if !found {
		return nil, graphEntityNotFound()
	}

	relations, appErr := source.relations(ctx, user, entityType, entityID)
	if appErr != nil {
		return nil, appErr
	}

	for i := range relations.Groups {
		if relations.Groups[i].Count > graphGroupItemLimit {
			relations.Groups[i].Items = make([]model.GraphRelationItem, 0)
		}
	}

	return relations, nil
}

// graphGroupPage cuts one page out of a group that was read whole.
func graphGroupPage(relations *model.GraphEntityRelations, relationTypeKey, direction, after string) (*model.GraphRelationPage, *model.AppError) {
	index := slices.IndexFunc(relations.Groups, func(group model.GraphRelationGroup) bool {
		return group.RelationType == relationTypeKey && group.Direction == direction
	})

	if index == -1 {
		return &model.GraphRelationPage{Items: make([]model.GraphRelationItem, 0)}, nil
	}

	items := relations.Groups[index].Items

	start := 0

	if after != "" {
		last := slices.IndexFunc(items, func(item model.GraphRelationItem) bool {
			return item.Entity.ID == after
		})

		if last == -1 {
			return &model.GraphRelationPage{Items: make([]model.GraphRelationItem, 0)}, nil
		}

		start = last + 1
	}

	end := min(start+graphGroupItemLimit, len(items))
	page := &model.GraphRelationPage{Items: items[start:end]}

	if end < len(items) {
		page.Next = items[end-1].Entity.ID
	}

	return page, nil
}

func (a *App) chartGraphRelations(ctx context.Context, user model.User, ids *graphBuiltInIDs, chartID string) (*model.GraphEntityRelations, *model.AppError) {
	stored, err := a.Store.Collimato.GetChartByID(chartID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, graphEntityNotFound()
	}

	if err != nil {
		return nil, graphRelationsFailed(chartID, err)
	}

	workspaceID := stored.WorkspaceID

	chart, appErr := a.GetChartByID(user, workspaceID, chartID)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return nil, graphEntityNotFound()
		}

		return nil, appErr
	}

	meta, appErr := a.graphVisibleMeta(ctx, user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	relations, err := a.Store.Graph.GetRelationsFrom(ctx, ids.chartType, chartID)
	if err != nil {
		return nil, graphRelationsFailed(chartID, err)
	}

	workspaceFields, err := a.Store.Graph.GetAllDatasetFieldsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, graphRelationsFailed(chartID, err)
	}

	fieldsByID := make(map[string]model.GraphDatasetField, len(workspaceFields))
	for _, field := range workspaceFields {
		fieldsByID[field.ID] = field
	}

	fieldItems := make([]model.GraphRelationItem, 0)
	for _, relation := range relations {
		if relation.RelationTypeID != ids.chartUsesFieldRelation {
			continue
		}

		field, found := fieldsByID[relation.ToID]
		if !found {
			continue
		}

		if !meta.fields[field.Name] {
			continue
		}

		item := model.GraphRelationItem{Entity: graphFieldRef(field), Origin: relation.Origin}
		fieldItems = append(fieldItems, item)
	}

	dashboardIDs, err := a.Store.Collimato.GetDashboardIDsForChart(chartID)
	if err != nil {
		return nil, graphRelationsFailed(chartID, err)
	}

	dashboards, appErr := a.graphVisibleDashboards(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	dashboardItems := graphStructuredItems(graphDashboardRefs(dashboardIDs, dashboards))

	result := &model.GraphEntityRelations{
		Entity: model.GraphEntityRef{Type: model.GraphTypeChart, ID: chart.ID, Title: chart.Name},
		Groups: make([]model.GraphRelationGroup, 0),
	}

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationChartUsesDatasetField,
		Direction:    model.GraphDirectionOut,
		EntityType:   model.GraphTypeDatasetField,
	}, fieldItems)

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDashboardContainsChart,
		Direction:    model.GraphDirectionIn,
		EntityType:   model.GraphTypeDashboard,
	}, dashboardItems)

	return result, nil
}

func (a *App) datasetFieldGraphRelations(ctx context.Context, user model.User, ids *graphBuiltInIDs, fieldID string) (*model.GraphEntityRelations, *model.AppError) {
	field, err := a.Store.Graph.GetDatasetFieldByID(ctx, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, graphEntityNotFound()
	}

	if err != nil {
		return nil, graphRelationsFailed(fieldID, err)
	}

	workspaceID := field.WorkspaceID

	meta, appErr := a.graphVisibleMeta(ctx, user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	if !meta.fields[field.Name] {
		return nil, graphEntityNotFound()
	}

	relations, err := a.Store.Graph.GetRelationsTo(ctx, ids.datasetFieldType, fieldID)
	if err != nil {
		return nil, graphRelationsFailed(fieldID, err)
	}

	charts, appErr := a.graphVisibleCharts(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	chartItems := make([]model.GraphRelationItem, 0)
	for _, relation := range relations {
		if relation.RelationTypeID != ids.chartUsesFieldRelation {
			continue
		}

		chart, found := charts[relation.FromID]
		if !found {
			continue
		}

		ref := model.GraphEntityRef{Type: model.GraphTypeChart, ID: chart.ID, Title: chart.Name}
		chartItems = append(chartItems, model.GraphRelationItem{Entity: ref, Origin: relation.Origin})
	}

	dashboardItems := make([]model.GraphEntityRef, 0)
	if a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionViewDashboardFilters) {
		dashboardIDs, err := a.Store.Collimato.GetDashboardIDsForFilterColumn(workspaceID, field.Name)
		if err != nil {
			return nil, graphRelationsFailed(fieldID, err)
		}

		dashboards, appErr := a.graphVisibleDashboards(user, workspaceID)
		if appErr != nil {
			return nil, appErr
		}

		dashboardItems = graphDashboardRefs(dashboardIDs, dashboards)
	}

	datasetItems := make([]model.GraphEntityRef, 0)

	dataset, err := a.Store.Graph.GetDatasetByID(ctx, field.DatasetID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, graphRelationsFailed(fieldID, err)
	}

	if dataset != nil && meta.datasets[dataset.Name] {
		datasetItems = append(datasetItems, graphDatasetRef(*dataset))
	}

	result := &model.GraphEntityRelations{
		Entity: graphFieldRef(*field),
		Groups: make([]model.GraphRelationGroup, 0),
	}

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationChartUsesDatasetField,
		Direction:    model.GraphDirectionIn,
		EntityType:   model.GraphTypeChart,
	}, chartItems)

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDashboardUsesDatasetField,
		Direction:    model.GraphDirectionIn,
		EntityType:   model.GraphTypeDashboard,
	}, graphStructuredItems(dashboardItems))

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDatasetFieldBelongsToDataset,
		Direction:    model.GraphDirectionOut,
		EntityType:   model.GraphTypeDataset,
	}, graphStructuredItems(datasetItems))

	return result, nil
}

func (a *App) datasetGraphRelations(ctx context.Context, user model.User, datasetID string) (*model.GraphEntityRelations, *model.AppError) {
	dataset, err := a.Store.Graph.GetDatasetByID(ctx, datasetID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, graphEntityNotFound()
	}

	if err != nil {
		return nil, graphRelationsFailed(datasetID, err)
	}

	meta, appErr := a.graphVisibleMeta(ctx, user, dataset.WorkspaceID)
	if appErr != nil {
		return nil, appErr
	}

	if !meta.datasets[dataset.Name] {
		return nil, graphEntityNotFound()
	}

	workspaceFields, err := a.Store.Graph.GetAllDatasetFieldsForWorkspace(ctx, dataset.WorkspaceID)
	if err != nil {
		return nil, graphRelationsFailed(datasetID, err)
	}

	fieldItems := make([]model.GraphEntityRef, 0)
	for _, field := range workspaceFields {
		if field.DatasetID != datasetID {
			continue
		}

		if !meta.fields[field.Name] {
			continue
		}

		fieldItems = append(fieldItems, graphFieldRef(field))
	}

	result := &model.GraphEntityRelations{
		Entity: graphDatasetRef(*dataset),
		Groups: make([]model.GraphRelationGroup, 0),
	}

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDatasetFieldBelongsToDataset,
		Direction:    model.GraphDirectionIn,
		EntityType:   model.GraphTypeDatasetField,
	}, graphStructuredItems(fieldItems))

	return result, nil
}

func (a *App) dashboardGraphRelations(ctx context.Context, user model.User, dashboardID string) (*model.GraphEntityRelations, *model.AppError) {
	stored, err := a.Store.Collimato.GetDashboardByID(dashboardID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, graphEntityNotFound()
	}

	if err != nil {
		return nil, graphRelationsFailed(dashboardID, err)
	}

	workspaceID := stored.WorkspaceID

	// GetDashboardByID leaves the filters out for users who may not see them.
	dashboard, appErr := a.GetDashboardByID(user, workspaceID, dashboardID)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return nil, graphEntityNotFound()
		}

		return nil, appErr
	}

	charts, appErr := a.graphVisibleCharts(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	chartItems := make([]model.GraphEntityRef, 0)
	for _, dashboardChart := range dashboard.Charts {
		chart, found := charts[dashboardChart.ID]
		if !found {
			continue
		}

		chartItems = append(chartItems, model.GraphEntityRef{Type: model.GraphTypeChart, ID: chart.ID, Title: chart.Name})
	}

	filterColumns := make([]string, 0, len(dashboard.Filters))
	for _, filter := range dashboard.Filters {
		filterColumns = append(filterColumns, filter.Column)
	}

	filterFields, err := a.Store.Graph.GetDatasetFieldsByNames(ctx, workspaceID, filterColumns)
	if err != nil {
		return nil, graphRelationsFailed(dashboardID, err)
	}

	meta, appErr := a.graphVisibleMeta(ctx, user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	fieldItems := make([]model.GraphEntityRef, 0)
	for _, field := range filterFields {
		if !meta.fields[field.Name] {
			continue
		}

		fieldItems = append(fieldItems, graphFieldRef(field))
	}

	result := &model.GraphEntityRelations{
		Entity: model.GraphEntityRef{Type: model.GraphTypeDashboard, ID: dashboard.ID, Title: dashboard.Title},
		Groups: make([]model.GraphRelationGroup, 0),
	}

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDashboardContainsChart,
		Direction:    model.GraphDirectionOut,
		EntityType:   model.GraphTypeChart,
	}, graphStructuredItems(chartItems))

	result.Groups = addGraphGroup(result.Groups, model.GraphRelationGroup{
		RelationType: model.GraphRelationDashboardUsesDatasetField,
		Direction:    model.GraphDirectionOut,
		EntityType:   model.GraphTypeDatasetField,
	}, graphStructuredItems(fieldItems))

	return result, nil
}

// graphVisibleMeta reads the datasets and fields the user's roles allow,
// with hidden columns already removed. A user who may not query the
// workspace sees none of them.
func (a *App) graphVisibleMeta(ctx context.Context, user model.User, workspaceID string) (*graphVisibleMeta, *model.AppError) {
	visible := &graphVisibleMeta{
		datasets:   make(map[string]bool),
		fields:     make(map[string]bool),
		dimensions: make(map[string]bool),
		titles:     make(map[string]string),
	}

	collection, appErr := a.GetWorkspaceMeta(ctx, user, workspaceID)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return visible, nil
		}

		return nil, appErr
	}

	for _, cube := range collection.Cubes {
		visible.datasets[cube.Name] = true

		for _, dimension := range cube.Dimensions {
			visible.fields[dimension.Name] = true
			visible.dimensions[dimension.Name] = true
			visible.titles[dimension.Name] = dimension.Title
		}

		for _, measure := range cube.Measures {
			visible.fields[measure.Name] = true
		}
	}

	return visible, nil
}

func (a *App) graphVisibleCharts(user model.User, workspaceID string) (map[string]model.Chart, *model.AppError) {
	visible := make(map[string]model.Chart)

	charts, appErr := a.GetCharts(user, workspaceID)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return visible, nil
		}

		return nil, appErr
	}

	for _, chart := range charts {
		visible[chart.ID] = chart
	}

	return visible, nil
}

func (a *App) graphVisibleDashboards(user model.User, workspaceID string) (map[string]model.Dashboard, *model.AppError) {
	visible := make(map[string]model.Dashboard)

	dashboards, appErr := a.GetDashboards(user, workspaceID)
	if appErr != nil {
		if graphCheckRefused(appErr) {
			return visible, nil
		}

		return nil, appErr
	}

	for _, dashboard := range dashboards {
		visible[dashboard.ID] = dashboard
	}

	return visible, nil
}

func graphFieldRef(field model.GraphDatasetField) model.GraphEntityRef {
	return model.GraphEntityRef{Type: model.GraphTypeDatasetField, ID: field.ID, Title: field.Name}
}

func graphDatasetRef(dataset model.GraphDataset) model.GraphEntityRef {
	return model.GraphEntityRef{Type: model.GraphTypeDataset, ID: dataset.ID, Title: dataset.Name}
}

func graphDashboardRefs(dashboardIDs []string, visible map[string]model.Dashboard) []model.GraphEntityRef {
	refs := make([]model.GraphEntityRef, 0, len(dashboardIDs))

	for _, id := range dashboardIDs {
		dashboard, found := visible[id]
		if !found {
			continue
		}

		refs = append(refs, model.GraphEntityRef{Type: model.GraphTypeDashboard, ID: dashboard.ID, Title: dashboard.Title})
	}

	return refs
}

// graphStructuredItems makes relations read from Collimato's own tables, which
// all come straight from the data.
func graphStructuredItems(refs []model.GraphEntityRef) []model.GraphRelationItem {
	items := make([]model.GraphRelationItem, 0, len(refs))

	for _, ref := range refs {
		items = append(items, model.GraphRelationItem{Entity: ref, Origin: model.GraphRelationOriginStructured})
	}

	return items
}

// addGraphGroup leaves out an empty group, so nothing hints that hidden
// neighbours exist.
func addGraphGroup(groups []model.GraphRelationGroup, group model.GraphRelationGroup, items []model.GraphRelationItem) []model.GraphRelationGroup {
	if len(items) == 0 {
		return groups
	}

	slices.SortFunc(items, func(first, second model.GraphRelationItem) int {
		return compareGraphRefs(first.Entity, second.Entity)
	})

	group.Count = len(items)
	group.Items = items

	return append(groups, group)
}

func compareGraphRefs(first, second model.GraphEntityRef) int {
	byTitle := strings.Compare(first.Title, second.Title)
	if byTitle != 0 {
		return byTitle
	}

	return strings.Compare(first.ID, second.ID)
}
