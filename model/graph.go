// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"slices"
	"strings"
)

const GraphSourceBuiltIn = "builtin"

const (
	GraphRelationKindTable  = "table"
	GraphRelationKindStored = "stored"
)

const (
	GraphRelationOriginStructured = "structured"
	GraphRelationOriginManual     = "manual"
)

const (
	GraphDatasetKindCube = "cube"
	GraphDatasetKindView = "view"
)

const (
	GraphTypeDataset      = "dataset"
	GraphTypeDatasetField = "dataset_field"
	GraphTypeChart        = "chart"
	GraphTypeDashboard    = "dashboard"
)

const (
	GraphRelationDashboardContainsChart       = "dashboard_contains_chart"
	GraphRelationChartUsesDatasetField        = "chart_uses_dataset_field"
	GraphRelationDashboardUsesDatasetField    = "dashboard_uses_dataset_field"
	GraphRelationDatasetFieldBelongsToDataset = "dataset_field_belongs_to_dataset"
)

type GraphEntityType struct {
	ID            string `json:"id"`
	Key           string `json:"key"`
	Name          string `json:"name"`
	Source        string `json:"source"`
	SourceID      string `json:"source_id,omitempty"`
	KeyProperty   string `json:"key_property,omitempty"`
	TitleProperty string `json:"title_property,omitempty"`
	Description   string `json:"description,omitempty"`
	BuiltIn       bool   `json:"built_in"`
	CreatedBy     string `json:"created_by,omitempty"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
}

type GraphRelationType struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	ReverseName string `json:"reverse_name"`
	FromTypeID  string `json:"from_type_id"`
	ToTypeID    string `json:"to_type_id"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
	BuiltIn     bool   `json:"built_in"`
	CreatedBy   string `json:"created_by,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type GraphRelation struct {
	ID             string `json:"id"`
	RelationTypeID string `json:"relation_type_id"`
	FromTypeID     string `json:"from_type_id"`
	FromID         string `json:"from_id"`
	ToTypeID       string `json:"to_type_id"`
	ToID           string `json:"to_id"`
	Origin         string `json:"origin"`
	CreatedBy      string `json:"created_by,omitempty"`
	CreatedAt      int64  `json:"created_at"`
}

type GraphDataset struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	CreatedAt   int64  `json:"created_at"`
}

type GraphDatasetField struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	DatasetID   string `json:"dataset_id"`
	Name        string `json:"name"`
	CreatedAt   int64  `json:"created_at"`
}

type GraphBuiltInRelationType struct {
	Key         string
	Name        string
	ReverseName string
	FromType    string
	ToType      string
	Kind        string
}

func GraphBuiltInEntityTypes() []GraphEntityType {
	return []GraphEntityType{
		{Key: GraphTypeDataset, Name: "Dataset", Source: GraphSourceBuiltIn, BuiltIn: true},
		{Key: GraphTypeDatasetField, Name: "Dataset field", Source: GraphSourceBuiltIn, BuiltIn: true},
		{Key: GraphTypeChart, Name: "Chart", Source: GraphSourceBuiltIn, BuiltIn: true},
		{Key: GraphTypeDashboard, Name: "Dashboard", Source: GraphSourceBuiltIn, BuiltIn: true},
	}
}

func GraphBuiltInRelationTypes() []GraphBuiltInRelationType {
	return []GraphBuiltInRelationType{
		{
			Key:         GraphRelationDashboardContainsChart,
			Name:        "contains",
			ReverseName: "contained in",
			FromType:    GraphTypeDashboard,
			ToType:      GraphTypeChart,
			Kind:        GraphRelationKindTable,
		},
		{
			Key:         GraphRelationChartUsesDatasetField,
			Name:        "uses",
			ReverseName: "used by",
			FromType:    GraphTypeChart,
			ToType:      GraphTypeDatasetField,
			Kind:        GraphRelationKindStored,
		},
		{
			Key:         GraphRelationDashboardUsesDatasetField,
			Name:        "uses",
			ReverseName: "used by",
			FromType:    GraphTypeDashboard,
			ToType:      GraphTypeDatasetField,
			Kind:        GraphRelationKindTable,
		},
		{
			Key:         GraphRelationDatasetFieldBelongsToDataset,
			Name:        "belongs to",
			ReverseName: "has",
			FromType:    GraphTypeDatasetField,
			ToType:      GraphTypeDataset,
			Kind:        GraphRelationKindTable,
		},
	}
}

// GraphQueryFields lists every field a chart's query reads, including the
// fields it only filters, groups by time or sorts on.
func GraphQueryFields(query DataQuery) []string {
	members := make([]string, 0)
	members = append(members, query.Measures...)
	members = append(members, query.Dimensions...)

	for _, timeDimension := range query.TimeDimensions {
		members = append(members, timeDimension.Dimension)
	}

	for _, filter := range query.Filters {
		filterMembers := graphFilterMembers(filter)
		members = append(members, filterMembers...)
	}

	for _, order := range query.Order {
		if len(order) > 0 {
			members = append(members, order[0])
		}
	}

	fields := make([]string, 0, len(members))
	for _, member := range members {
		field := graphFieldOf(member)
		if field == "" {
			continue
		}

		if slices.Contains(fields, field) {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func graphFilterMembers(filter QueryFilter) []string {
	members := make([]string, 0)

	if filter.Member != "" {
		members = append(members, filter.Member)
	}

	for _, inner := range filter.And {
		innerMembers := graphFilterMembers(inner)
		members = append(members, innerMembers...)
	}

	for _, inner := range filter.Or {
		innerMembers := graphFilterMembers(inner)
		members = append(members, innerMembers...)
	}

	return members
}

func graphFieldOf(member string) string {
	parts := strings.SplitN(member, ".", 3)
	if len(parts) < 3 {
		return member
	}

	return parts[0] + "." + parts[1]
}
