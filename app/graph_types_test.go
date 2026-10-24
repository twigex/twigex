// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
)

func (f *fakeGraphStore) GetEntityTypeByID(_ context.Context, id string) (*model.GraphEntityType, error) {
	for _, entityType := range f.entityTypes {
		if entityType.ID == id {
			return &entityType, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphStore) CreateEntityType(_ context.Context, entityType model.GraphEntityType) error {
	for _, existing := range f.entityTypes {
		if existing.Key == entityType.Key {
			return model.ErrGraphTypeKeyTaken
		}
	}

	f.entityTypes = append(f.entityTypes, entityType)

	return nil
}

func (f *fakeGraphStore) UpdateEntityType(_ context.Context, entityType model.GraphEntityType) error {
	for i := range f.entityTypes {
		if f.entityTypes[i].ID == entityType.ID {
			f.entityTypes[i] = entityType
		}
	}

	return nil
}

func (f *fakeGraphStore) DeleteEntityType(_ context.Context, id string) error {
	f.entityTypes = slices.DeleteFunc(f.entityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.ID == id
	})

	return nil
}

func (f *fakeGraphStore) GetRelationTypeByID(_ context.Context, id string) (*model.GraphRelationType, error) {
	for _, relationType := range f.relationTypes {
		if relationType.ID == id {
			return &relationType, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeGraphStore) CreateRelationType(_ context.Context, relationType model.GraphRelationType) error {
	for _, existing := range f.relationTypes {
		if existing.Key == relationType.Key {
			return model.ErrGraphTypeKeyTaken
		}
	}

	f.relationTypes = append(f.relationTypes, relationType)

	return nil
}

func (f *fakeGraphStore) UpdateRelationType(_ context.Context, relationType model.GraphRelationType) error {
	for i := range f.relationTypes {
		if f.relationTypes[i].ID == relationType.ID {
			f.relationTypes[i] = relationType
		}
	}

	return nil
}

func (f *fakeGraphStore) DeleteRelationType(_ context.Context, id string) error {
	f.relationTypes = slices.DeleteFunc(f.relationTypes, func(relationType model.GraphRelationType) bool {
		return relationType.ID == id
	})

	return nil
}

var graphTypeAdmin = model.User{ID: "admin", Role: model.SystemAdminRoleId}

var graphTypeMember = model.User{ID: "anna"}

// newGraphTypesTestApp builds a workspace with "companies" and "orders", the
// way Cube describes them: companies.id is the primary key, so Cube leaves it
// out of the field list, and companies.key is a visible copy of it.
func newGraphTypesTestApp() *graphTestApp {
	g := newGraphTestApp(true)

	g.collimato.permissions = graphViewerPermissions

	companies := collimato.Cube{Name: "companies", Type: "cube"}
	companies.Dimensions = []collimato.Dimension{{Name: "companies.key"}, {Name: "companies.name"}}

	orders := collimato.Cube{Name: "orders", Type: "cube"}
	orders.Dimensions = []collimato.Dimension{{Name: "orders.key"}, {Name: "orders.number"}, {Name: "orders.company_id"}}
	orders.Measures = []collimato.Measure{{Name: "orders.count"}}

	g.meta["ws"] = &collimato.CubeCollection{Cubes: []collimato.Cube{companies, orders}}

	g.graph.datasets = []model.GraphDataset{
		{ID: "ds-companies", WorkspaceID: "ws", Name: "companies"},
		{ID: "ds-orders", WorkspaceID: "ws", Name: "orders"},
	}

	g.graph.fields = []model.GraphDatasetField{
		{ID: "f-companies.id", WorkspaceID: "ws", DatasetID: "ds-companies", Name: "companies.id"},
		{ID: "f-companies.key", WorkspaceID: "ws", DatasetID: "ds-companies", Name: "companies.key"},
		{ID: "f-companies.name", WorkspaceID: "ws", DatasetID: "ds-companies", Name: "companies.name"},
		{ID: "f-orders.key", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.key"},
		{ID: "f-orders.number", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.number"},
		{ID: "f-orders.company_id", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.company_id"},
		{ID: "f-orders.count", WorkspaceID: "ws", DatasetID: "ds-orders", Name: "orders.count"},
	}

	return g
}

func companyTypeInput() model.GraphEntityTypeInput {
	return model.GraphEntityTypeInput{
		Key:           "company",
		Name:          "Company",
		SourceID:      "ds-companies",
		KeyProperty:   "f-companies.key",
		TitleProperty: "f-companies.name",
	}
}

func orderTypeInput() model.GraphEntityTypeInput {
	return model.GraphEntityTypeInput{
		Key:           "order",
		Name:          "Order",
		SourceID:      "ds-orders",
		KeyProperty:   "f-orders.key",
		TitleProperty: "f-orders.number",
	}
}

func createCompanyAndOrderTypes(t *testing.T, g *graphTestApp) (*model.GraphEntityType, *model.GraphEntityType) {
	t.Helper()

	company, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, companyTypeInput())
	if appErr != nil {
		t.Fatalf("create company: %v", appErr)
	}

	order, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, orderTypeInput())
	if appErr != nil {
		t.Fatalf("create order: %v", appErr)
	}

	return company, order
}

func placesInput(company, order *model.GraphEntityType) model.GraphRelationTypeInput {
	return model.GraphRelationTypeInput{
		Key:         "company_places_order",
		Name:        "places",
		ReverseName: "placed by",
		FromTypeID:  company.ID,
		ToTypeID:    order.ID,
		Property:    "f-orders.company_id",
	}
}

func TestCreateGraphEntityTypeReadsItsEntitiesFromTheDataset(t *testing.T) {
	g := newGraphTypesTestApp()

	company, _ := createCompanyAndOrderTypes(t, g)

	saved, err := g.graph.GetEntityTypeByID(context.Background(), company.ID)
	if err != nil {
		t.Fatalf("the type was not saved: %v", err)
	}

	if saved.Source != model.GraphSourceDataset || saved.SourceID != "ds-companies" || saved.BuiltIn {
		t.Errorf("want a dataset type on ds-companies, got %+v", saved)
	}
}

func TestCreateGraphEntityTypeRefusesBadInput(t *testing.T) {
	cases := []struct {
		name   string
		user   model.User
		change func(input *model.GraphEntityTypeInput)
		want   string
	}{
		{"a user without the permission", graphTypeMember, func(*model.GraphEntityTypeInput) {}, "graph.forbidden"},
		{"a key with capitals", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.Key = "Company" }, "graph.type_key_invalid"},
		{"a built-in key", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.Key = model.GraphTypeChart }, "graph.type_key_taken"},
		{"no name", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.Name = "  " }, "graph.type_name_required"},
		{"a dataset that does not exist", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.SourceID = "ds-missing" }, "graph.type_dataset_not_found"},
		{"a field of another dataset", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.TitleProperty = "f-orders.number" }, "graph.type_field_not_found"},
		{"the primary key Cube hides", graphTypeAdmin, func(input *model.GraphEntityTypeInput) { input.KeyProperty = "f-companies.id" }, "graph.type_field_not_visible"},
		{"a measure as the name", graphTypeAdmin, func(input *model.GraphEntityTypeInput) {
			input.SourceID = "ds-orders"
			input.KeyProperty = "f-orders.key"
			input.TitleProperty = "f-orders.count"
		}, "graph.type_field_is_measure"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGraphTypesTestApp()

			input := companyTypeInput()
			tc.change(&input)

			_, appErr := g.app.CreateGraphEntityType(context.Background(), tc.user, input)
			if appErr == nil || appErr.ID != tc.want {
				t.Errorf("want %s, got %v", tc.want, appErr)
			}
		})
	}
}

func TestCreateGraphEntityTypeRefusesAFieldHiddenFromTheCreator(t *testing.T) {
	g := newGraphTypesTestApp()

	g.app.CollimatoRoles = fakeColumnRules{hidden: []string{"companies.name"}}

	_, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, companyTypeInput())
	if appErr == nil || appErr.ID != "graph.type_field_not_visible" {
		t.Errorf("want graph.type_field_not_visible, got %v", appErr)
	}
}

func TestCreateGraphRelationTypeChecksThePropertyOnTheToType(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	relationType, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	if relationType.Kind != model.GraphRelationKindProperty || relationType.KindConfig.Property != "f-orders.company_id" {
		t.Errorf("want a property relation on orders.company_id, got %+v", relationType)
	}

	wrongSide := placesInput(company, order)
	wrongSide.Key = "company_places_order_again"
	wrongSide.Property = "f-companies.key"

	_, appErr = g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, wrongSide)
	if appErr == nil || appErr.ID != "graph.type_field_not_found" {
		t.Errorf("want graph.type_field_not_found for a property on the from type, got %v", appErr)
	}
}

func TestCreateGraphRelationTypeRefusesATakenKey(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	_, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	_, appErr = g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr == nil || appErr.ID != "graph.type_key_taken" || appErr.Status != http.StatusConflict {
		t.Errorf("want graph.type_key_taken, got %v", appErr)
	}
}

func TestCreateGraphRelationTypeNeedsDatasetTypesAtBothEnds(t *testing.T) {
	g := newGraphTypesTestApp()

	_, order := createCompanyAndOrderTypes(t, g)

	input := model.GraphRelationTypeInput{
		Key:         "chart_shows_order",
		Name:        "shows",
		ReverseName: "shown in",
		FromTypeID:  "id-" + model.GraphTypeChart,
		ToTypeID:    order.ID,
		Property:    "f-orders.company_id",
	}

	_, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, input)
	if appErr == nil || appErr.ID != "graph.relation_type_needs_dataset_types" {
		t.Errorf("want graph.relation_type_needs_dataset_types, got %v", appErr)
	}
}

func TestDeleteGraphEntityTypeRefusesATypeARelationUses(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	relationType, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	appErr = g.app.DeleteGraphEntityType(context.Background(), graphTypeAdmin, company.ID)
	if appErr == nil || appErr.ID != "graph.type_in_use" {
		t.Fatalf("want graph.type_in_use, got %v", appErr)
	}

	appErr = g.app.DeleteGraphRelationType(context.Background(), graphTypeAdmin, relationType.ID)
	if appErr != nil {
		t.Fatalf("DeleteGraphRelationType: %v", appErr)
	}

	appErr = g.app.DeleteGraphEntityType(context.Background(), graphTypeAdmin, company.ID)
	if appErr != nil {
		t.Errorf("want the unused type deleted, got %v", appErr)
	}
}

func TestBuiltInGraphTypesCannotBeChanged(t *testing.T) {
	g := newGraphTypesTestApp()

	appErr := g.app.DeleteGraphEntityType(context.Background(), graphTypeAdmin, "id-"+model.GraphTypeChart)
	if appErr == nil || appErr.ID != "graph.type_built_in" {
		t.Errorf("want graph.type_built_in for an entity type, got %v", appErr)
	}

	appErr = g.app.DeleteGraphRelationType(context.Background(), graphTypeAdmin, "id-"+model.GraphRelationChartUsesDatasetField)
	if appErr == nil || appErr.ID != "graph.type_built_in" {
		t.Errorf("want graph.type_built_in for a relation type, got %v", appErr)
	}
}

func TestUpdateGraphRelationTypeKeepsItsKeyAndTypes(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	relationType, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	input := placesInput(company, order)
	input.Key = "renamed"
	input.Name = "orders"
	input.FromTypeID = order.ID

	updated, appErr := g.app.UpdateGraphRelationType(context.Background(), graphTypeAdmin, relationType.ID, input)
	if appErr != nil {
		t.Fatalf("UpdateGraphRelationType: %v", appErr)
	}

	if updated.Key != "company_places_order" || updated.FromTypeID != company.ID || updated.Name != "orders" {
		t.Errorf("want the name changed and the key and types kept, got %+v", updated)
	}
}

func TestGetGraphTypesHidesWhereTheDataComesFrom(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	_, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	forMember, appErr := g.app.GetGraphTypes(context.Background(), graphTypeMember)
	if appErr != nil {
		t.Fatalf("GetGraphTypes for a member: %v", appErr)
	}

	for _, entityType := range forMember.EntityTypes {
		if entityType.SourceID != "" || entityType.KeyProperty != "" || entityType.TitleProperty != "" {
			t.Errorf("want no source settings for a member, got %+v", entityType)
		}
	}

	for _, relationType := range forMember.RelationTypes {
		if relationType.KindConfig != nil {
			t.Errorf("want no relation settings for a member, got %+v", relationType)
		}
	}

	forAdmin, appErr := g.app.GetGraphTypes(context.Background(), graphTypeAdmin)
	if appErr != nil {
		t.Fatalf("GetGraphTypes for an admin: %v", appErr)
	}

	found := slices.ContainsFunc(forAdmin.EntityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.ID == company.ID && entityType.SourceID == "ds-companies"
	})

	if !found {
		t.Errorf("want the source settings for an admin")
	}
}

func TestEnsureGraphBuiltInTypesLeavesAUserTypeWithTheSameKeyAlone(t *testing.T) {
	g := newGraphTestApp(true)

	userChart := model.GraphEntityType{ID: "user-chart", Key: model.GraphTypeChart, Name: "My charts", Source: model.GraphSourceDataset}
	g.graph.entityTypes = []model.GraphEntityType{userChart}
	g.graph.relationTypes = nil

	g.app.EnsureGraphBuiltInTypes()

	charts := 0
	for _, entityType := range g.graph.entityTypes {
		if entityType.Key == model.GraphTypeChart {
			charts++

			if entityType.ID != "user-chart" || entityType.BuiltIn {
				t.Errorf("want the user's type unchanged, got %+v", entityType)
			}
		}
	}

	if charts != 1 {
		t.Errorf("want one type with the chart key, got %d", charts)
	}

	_, err := g.app.loadGraphBuiltInIDs(context.Background())
	if err == nil {
		t.Errorf("want the missing built-in reported, not the user's type used in its place")
	}
}

func TestGetGraphTypesNamesTheSourcesForAdminsOnly(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	_, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	forAdmin, appErr := g.app.GetGraphTypes(context.Background(), graphTypeAdmin)
	if appErr != nil {
		t.Fatalf("GetGraphTypes: %v", appErr)
	}

	companyIndex := slices.IndexFunc(forAdmin.EntityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.ID == company.ID
	})
	named := forAdmin.EntityTypes[companyIndex]

	if named.SourceName != "companies" || named.KeyPropertyName != "companies.key" || named.TitlePropertyName != "companies.name" {
		t.Errorf("want the dataset and field names, got %+v", named)
	}

	placesIndex := slices.IndexFunc(forAdmin.RelationTypes, func(relationType model.GraphRelationType) bool {
		return relationType.Key == "company_places_order"
	})

	if forAdmin.RelationTypes[placesIndex].KindConfig.PropertyName != "orders.company_id" {
		t.Errorf("want the property name, got %+v", forAdmin.RelationTypes[placesIndex].KindConfig)
	}

	for _, stored := range g.graph.relationTypes {
		if stored.KindConfig != nil && stored.KindConfig.PropertyName != "" {
			t.Errorf("want the stored relation type left without names, got %+v", stored.KindConfig)
		}
	}

	forMember, appErr := g.app.GetGraphTypes(context.Background(), graphTypeMember)
	if appErr != nil {
		t.Fatalf("GetGraphTypes for a member: %v", appErr)
	}

	for _, entityType := range forMember.EntityTypes {
		if entityType.SourceName != "" || entityType.KeyPropertyName != "" {
			t.Errorf("want no names for a member, got %+v", entityType)
		}
	}
}

func TestGraphTypeLookIsCheckedAndDefaulted(t *testing.T) {
	cases := []struct {
		name   string
		change func(input *model.GraphEntityTypeInput)
		want   string
	}{
		{"an unknown status", func(input *model.GraphEntityTypeInput) { input.Status = "retired" }, "graph.type_status_invalid"},
		{"a colour code instead of a name", func(input *model.GraphEntityTypeInput) { input.Color = "#ff0000" }, "graph.type_color_invalid"},
		{"an icon name with spaces", func(input *model.GraphEntityTypeInput) { input.Icon = "building office" }, "graph.type_icon_invalid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGraphTypesTestApp()

			input := companyTypeInput()
			tc.change(&input)

			_, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, input)
			if appErr == nil || appErr.ID != tc.want {
				t.Errorf("want %s, got %v", tc.want, appErr)
			}
		})
	}

	g := newGraphTypesTestApp()

	input := companyTypeInput()
	input.PluralName = " Companies "
	input.Icon = "building-office"
	input.Color = "cyan"

	company, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, input)
	if appErr != nil {
		t.Fatalf("CreateGraphEntityType: %v", appErr)
	}

	if company.PluralName != "Companies" || company.Icon != "building-office" || company.Color != "cyan" {
		t.Errorf("want the look saved, got %+v", company)
	}

	input.Status = model.GraphStatusDeprecated

	updated, appErr := g.app.UpdateGraphEntityType(context.Background(), graphTypeAdmin, company.ID, input)
	if appErr != nil {
		t.Fatalf("UpdateGraphEntityType: %v", appErr)
	}

	if updated.Status != model.GraphStatusDeprecated || updated.UpdatedBy != graphTypeAdmin.ID {
		t.Errorf("want the status changed and the editor recorded, got %+v", updated)
	}
}

func TestGraphRelationTypeCardinalityDefaultsToOneToMany(t *testing.T) {
	g := newGraphTypesTestApp()

	company, order := createCompanyAndOrderTypes(t, g)

	relationType, appErr := g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, placesInput(company, order))
	if appErr != nil {
		t.Fatalf("CreateGraphRelationType: %v", appErr)
	}

	if relationType.Cardinality != model.GraphOneToMany {
		t.Errorf("want one to many, got %q", relationType.Cardinality)
	}

	wrong := placesInput(company, order)
	wrong.Key = "company_places_order_too"
	wrong.Cardinality = "several"

	_, appErr = g.app.CreateGraphRelationType(context.Background(), graphTypeAdmin, wrong)
	if appErr == nil || appErr.ID != "graph.relation_type_cardinality_invalid" {
		t.Errorf("want graph.relation_type_cardinality_invalid, got %v", appErr)
	}
}

func TestGraphEntityTypeWithoutStatusIsReportedActive(t *testing.T) {
	g := newGraphTypesTestApp()

	company, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, companyTypeInput())
	if appErr != nil {
		t.Fatalf("CreateGraphEntityType: %v", appErr)
	}

	if company.Status != model.GraphStatusActive {
		t.Errorf("want active, got %q", company.Status)
	}
}

func TestGraphTypeNamesHaveALengthLimit(t *testing.T) {
	g := newGraphTypesTestApp()

	tooLong := strings.Repeat("a", 256)

	input := companyTypeInput()
	input.Name = tooLong

	_, appErr := g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, input)
	if appErr == nil || appErr.ID != "graph.type_name_too_long" {
		t.Errorf("want graph.type_name_too_long for the name, got %v", appErr)
	}

	input = companyTypeInput()
	input.PluralName = tooLong

	_, appErr = g.app.CreateGraphEntityType(context.Background(), graphTypeAdmin, input)
	if appErr == nil || appErr.ID != "graph.type_name_too_long" {
		t.Errorf("want graph.type_name_too_long for the plural name, got %v", appErr)
	}
}
