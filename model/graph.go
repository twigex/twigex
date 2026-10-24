// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"errors"
	"slices"
	"strings"
)

var ErrGraphTypeKeyTaken = errors.New("another graph type has this key")

const (
	GraphSourceBuiltIn = "builtin"
	GraphSourceDataset = "dataset"
)

const GraphBuiltSetting = "graph.built"

const (
	GraphRelationKindTable    = "table"
	GraphRelationKindStored   = "stored"
	GraphRelationKindProperty = "property"
)

const (
	GraphRelationOriginStructured = "structured"
	GraphRelationOriginManual     = "manual"
)

const (
	GraphStatusExperimental = "experimental"
	GraphStatusActive       = "active"
	GraphStatusDeprecated   = "deprecated"
)

var GraphStatuses = []string{GraphStatusExperimental, GraphStatusActive, GraphStatusDeprecated}

const (
	GraphOneToOne   = "one_to_one"
	GraphOneToMany  = "one_to_many"
	GraphManyToOne  = "many_to_one"
	GraphManyToMany = "many_to_many"
)

var GraphCardinalities = []string{GraphOneToOne, GraphOneToMany, GraphManyToOne, GraphManyToMany}

// GraphColors are the colours a type may have. They are names, not colour
// codes, because the screen styles each one with its own fixed classes.
var GraphColors = []string{
	"indigo", "emerald", "amber", "pink", "cyan", "violet", "orange",
	"green", "rose", "blue", "yellow", "teal", "slate",
}

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
	PluralName    string `json:"plural_name,omitempty"`
	Icon          string `json:"icon,omitempty"`
	Color         string `json:"color,omitempty"`
	Status        string `json:"status"`
	BuiltIn       bool   `json:"built_in"`
	CreatedBy     string `json:"created_by,omitempty"`
	UpdatedBy     string `json:"updated_by,omitempty"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`

	// Names for the screen, looked up from the IDs above when the types are
	// listed for someone who manages them; never saved.
	SourceName        string `json:"source_name,omitempty"`
	KeyPropertyName   string `json:"key_property_name,omitempty"`
	TitlePropertyName string `json:"title_property_name,omitempty"`
}

type GraphRelationType struct {
	ID          string                   `json:"id"`
	Key         string                   `json:"key"`
	Name        string                   `json:"name"`
	ReverseName string                   `json:"reverse_name"`
	FromTypeID  string                   `json:"from_type_id"`
	ToTypeID    string                   `json:"to_type_id"`
	Kind        string                   `json:"kind"`
	KindConfig  *GraphRelationKindConfig `json:"kind_config,omitempty"`
	Cardinality string                   `json:"cardinality,omitempty"`
	Description string                   `json:"description,omitempty"`
	BuiltIn     bool                     `json:"built_in"`
	CreatedBy   string                   `json:"created_by,omitempty"`
	UpdatedBy   string                   `json:"updated_by,omitempty"`
	CreatedAt   int64                    `json:"created_at"`
	UpdatedAt   int64                    `json:"updated_at"`
}

// GraphRelationKindConfig holds field IDs from graph_dataset_fields. For a
// property relation, Property is the field on the "to" type that holds the "from"
// entity's key, or its FromProperty when one is set.
type GraphRelationKindConfig struct {
	Property     string `json:"property"`
	FromProperty string `json:"from_property,omitempty"`

	// Names for the screen, as on GraphEntityType; never saved.
	PropertyName     string `json:"property_name,omitempty"`
	FromPropertyName string `json:"from_property_name,omitempty"`
}

type GraphEntityTypeInput struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	PluralName    string `json:"plural_name"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	SourceID      string `json:"source_id"`
	KeyProperty   string `json:"key_property"`
	TitleProperty string `json:"title_property"`
}

type GraphRelationTypeInput struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	ReverseName  string `json:"reverse_name"`
	Description  string `json:"description"`
	FromTypeID   string `json:"from_type_id"`
	ToTypeID     string `json:"to_type_id"`
	Property     string `json:"property"`
	FromProperty string `json:"from_property"`
	Cardinality  string `json:"cardinality"`
}

// GraphDatasetChoice is a dataset someone setting up types may choose, with
// the fields of it they can see and may use as a key, name or property.
type GraphDatasetChoice struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	WorkspaceID   string             `json:"workspace_id"`
	WorkspaceName string             `json:"workspace_name"`
	Fields        []GraphFieldChoice `json:"fields"`
}

type GraphFieldChoice struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

type GraphTypes struct {
	EntityTypes   []GraphEntityType   `json:"entity_types"`
	RelationTypes []GraphRelationType `json:"relation_types"`
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

type GraphRebuildResult struct {
	Workspaces       int   `json:"workspaces"`
	FailedWorkspaces int   `json:"failed_workspaces"`
	DatasetsAdded    int64 `json:"datasets_added"`
	DatasetsRemoved  int64 `json:"datasets_removed"`
	FieldsAdded      int64 `json:"fields_added"`
	FieldsRemoved    int64 `json:"fields_removed"`
	RelationsAdded   int64 `json:"relations_added"`
	RelationsRemoved int64 `json:"relations_removed"`
}

func (r *GraphRebuildResult) Add(other GraphRebuildResult) {
	r.DatasetsAdded += other.DatasetsAdded
	r.DatasetsRemoved += other.DatasetsRemoved
	r.FieldsAdded += other.FieldsAdded
	r.FieldsRemoved += other.FieldsRemoved
	r.RelationsAdded += other.RelationsAdded
	r.RelationsRemoved += other.RelationsRemoved
}

const GraphRebuildNeverRun = "never"

type GraphRebuildStatus struct {
	Built      bool               `json:"built"`
	Status     string             `json:"status"`
	Reason     string             `json:"reason,omitempty"`
	Result     GraphRebuildResult `json:"result"`
	StartedAt  int64              `json:"started_at"`
	FinishedAt int64              `json:"finished_at"`
}

const (
	GraphDirectionOut = "out"
	GraphDirectionIn  = "in"
)

type GraphEntityRef struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Title string `json:"title"`
}

// GraphRelationItem is one relation in a group: the entity at the other end,
// and where the relation comes from.
type GraphRelationItem struct {
	Entity GraphEntityRef `json:"entity"`
	Origin string         `json:"origin"`
	// Source is the field the relation is read from, for a relation found in
	// the data, such as orders.company_id.
	Source string `json:"source,omitempty"`
}

// GraphRelationGroup lists the neighbours behind one relation type. A group larger
// than its item limit carries only the count, which the screen shows as a box.
type GraphRelationGroup struct {
	RelationType string `json:"relation_type"`
	Direction    string `json:"direction"`
	EntityType   string `json:"entity_type"`
	Count        int    `json:"count"`
	// CountCapped means there are more than Count; the screen shows "1000+".
	CountCapped bool                `json:"count_capped,omitempty"`
	Items       []GraphRelationItem `json:"items"`
}

// GraphRelationPage is one page of a group too large to send with its entity.
// Next is the key to ask for the following page with; empty on the last page.
type GraphRelationPage struct {
	Items []GraphRelationItem `json:"items"`
	Next  string              `json:"next,omitempty"`
}

// GraphEntityPage is one page of a type's entities, read the same way.
type GraphEntityPage struct {
	Items []GraphEntityRef `json:"items"`
	Next  string           `json:"next,omitempty"`
}

type GraphEntityRelations struct {
	Entity GraphEntityRef       `json:"entity"`
	Groups []GraphRelationGroup `json:"groups"`
}

type GraphBuiltInRelationType struct {
	Key         string
	Name        string
	ReverseName string
	FromType    string
	ToType      string
	Kind        string
	Cardinality string
}

func GraphBuiltInEntityTypes() []GraphEntityType {
	return []GraphEntityType{
		{
			Key:        GraphTypeDataset,
			Name:       "Dataset",
			PluralName: "Datasets",
			Icon:       "circle-stack",
			Color:      "amber",
			Status:     GraphStatusActive,
			Source:     GraphSourceBuiltIn,
			BuiltIn:    true,
		},
		{
			Key:        GraphTypeDatasetField,
			Name:       "Dataset field",
			PluralName: "Dataset fields",
			Icon:       "hashtag",
			Color:      "pink",
			Status:     GraphStatusActive,
			Source:     GraphSourceBuiltIn,
			BuiltIn:    true,
		},
		{
			Key:        GraphTypeChart,
			Name:       "Chart",
			PluralName: "Charts",
			Icon:       "chart-bar",
			Color:      "indigo",
			Status:     GraphStatusActive,
			Source:     GraphSourceBuiltIn,
			BuiltIn:    true,
		},
		{
			Key:        GraphTypeDashboard,
			Name:       "Dashboard",
			PluralName: "Dashboards",
			Icon:       "squares-2x2",
			Color:      "emerald",
			Status:     GraphStatusActive,
			Source:     GraphSourceBuiltIn,
			BuiltIn:    true,
		},
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
			Cardinality: GraphManyToMany,
		},
		{
			Key:         GraphRelationChartUsesDatasetField,
			Name:        "uses",
			ReverseName: "used by",
			FromType:    GraphTypeChart,
			ToType:      GraphTypeDatasetField,
			Kind:        GraphRelationKindStored,
			Cardinality: GraphManyToMany,
		},
		{
			Key:         GraphRelationDashboardUsesDatasetField,
			Name:        "uses",
			ReverseName: "used by",
			FromType:    GraphTypeDashboard,
			ToType:      GraphTypeDatasetField,
			Kind:        GraphRelationKindTable,
			Cardinality: GraphManyToMany,
		},
		{
			Key:         GraphRelationDatasetFieldBelongsToDataset,
			Name:        "belongs to",
			ReverseName: "has",
			FromType:    GraphTypeDatasetField,
			ToType:      GraphTypeDataset,
			Kind:        GraphRelationKindTable,
			Cardinality: GraphManyToOne,
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

const ChartTypeMap = "map"

type GraphMapPoint struct {
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type GraphMapPopover struct {
	Fields []string `json:"fields"`
}

type GraphMapLayerConfig struct {
	Coordinates      GraphMapPoint   `json:"coordinates"`
	Source           GraphMapPoint   `json:"source"`
	Target           GraphMapPoint   `json:"target"`
	RadiusField      string          `json:"radius_field"`
	FillColorField   string          `json:"fill_color_field"`
	WidthField       string          `json:"width_field"`
	SourceColorField string          `json:"source_color_field"`
	TargetColorField string          `json:"target_color_field"`
	GeojsonField     string          `json:"geojson_field"`
	Weight           string          `json:"weight"`
	Popover          GraphMapPopover `json:"popover"`
}

type GraphMapLayer struct {
	Model   string              `json:"model"`
	Filters []QueryFilter       `json:"filters"`
	Config  GraphMapLayerConfig `json:"config"`
}

type GraphMapConfiguration struct {
	Layers []GraphMapLayer `json:"layers"`
}

// GraphMapFields follows layerQuery in mapQuery.js. A map chart keeps no query
// of its own; each layer asks for a count of its model plus the fields its
// settings point at. A layer's config is reset when its type changes, so every
// field key can be read without checking the type.
func GraphMapFields(configuration GraphMapConfiguration) []string {
	query := DataQuery{}

	for _, layer := range configuration.Layers {
		config := layer.Config

		if layer.Model != "" {
			query.Measures = append(query.Measures, layer.Model+".count")
		}

		query.Dimensions = append(query.Dimensions,
			config.Coordinates.Lat,
			config.Coordinates.Long,
			config.Source.Lat,
			config.Source.Long,
			config.Target.Lat,
			config.Target.Long,
			config.RadiusField,
			config.FillColorField,
			config.WidthField,
			config.SourceColorField,
			config.TargetColorField,
			config.GeojsonField,
			config.Weight,
		)
		query.Dimensions = append(query.Dimensions, config.Popover.Fields...)

		query.Filters = append(query.Filters, layer.Filters...)
	}

	return GraphQueryFields(query)
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
