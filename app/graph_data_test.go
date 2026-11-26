// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
)

// fakeTableEngine stands in for the query engine over small in-memory tables. It filters,
// sorts and limits rows the way the row reads need, so the tests check real
// answers rather than which queries were sent.
type fakeTableEngine struct {
	QueryEngine
	tables  map[string][]map[string]any
	queries int
}

func (f *fakeTableEngine) Query(_ context.Context, _ string, query model.DataQuery) (*collimato.Answer, error) {
	f.queries++

	dataset := strings.Split(query.Dimensions[0], ".")[0]

	matching := make([]map[string]any, 0)
	for _, row := range f.tables[dataset] {
		if fakeRowMatches(row, query.Filters) {
			matching = append(matching, row)
		}
	}

	orderBy := query.Order[0][0]

	slices.SortFunc(matching, func(first, second map[string]any) int {
		return strings.Compare(fmt.Sprint(first[orderBy]), fmt.Sprint(second[orderBy]))
	})

	if query.Ungrouped && len(matching) > query.Limit {
		matching = matching[:query.Limit]
	}

	rows := make([]map[string]any, 0, len(matching))
	seen := make(map[string]bool)

	for _, row := range matching {
		picked := make(map[string]any, len(query.Dimensions))
		for _, field := range query.Dimensions {
			picked[field] = row[field]
		}

		// A grouped read gives each set of values once, as GROUP BY does.
		values := fmt.Sprint(picked)
		if !query.Ungrouped && seen[values] {
			continue
		}

		seen[values] = true
		rows = append(rows, picked)
	}

	if len(rows) > query.Limit {
		rows = rows[:query.Limit]
	}

	return &collimato.Answer{Rows: rows}, nil
}

// FieldList describes each table with every column its rows hold visible.
func (f *fakeTableEngine) FieldList(_ context.Context, _ string) (*collimato.FieldList, error) {
	fieldList := &collimato.FieldList{}

	for name, rows := range f.tables {
		dataset := collimato.DatasetFields{Name: name}
		seen := make(map[string]bool)

		for _, row := range rows {
			for column := range row {
				if seen[column] {
					continue
				}

				seen[column] = true
				dataset.Fields = append(dataset.Fields, collimato.Field{Key: column, Kind: collimato.FieldKindDimension})
			}
		}

		fieldList.Datasets = append(fieldList.Datasets, dataset)
	}

	return fieldList, nil
}

func fakeRowMatches(row map[string]any, filters []model.QueryFilter) bool {
	for _, filter := range filters {
		value := fmt.Sprint(row[filter.Member])

		switch filter.Operator {
		case "equals":
			if !slices.Contains(filter.Values, value) {
				return false
			}
		case "gt":
			if value <= filter.Values[0] {
				return false
			}
		case "set":
			if row[filter.Member] == nil {
				return false
			}
		case "contains":
			if !strings.Contains(strings.ToLower(value), strings.ToLower(filter.Values[0])) {
				return false
			}
		}
	}

	return true
}

func fakeOrder(key, company, supplier, forOrder string) map[string]any {
	return map[string]any{
		"orders.key":          key,
		"orders.number":       "Order " + key,
		"orders.company_id":   company,
		"orders.supplier_id":  supplier,
		"orders.for_order_id": forOrder,
	}
}

func fakeCompany(key, name string) map[string]any {
	return map[string]any{"companies.key": key, "companies.name": name}
}

// newGraphDataTestApp saves two types, Company and Order, and three relation
// types: "places" (orders.company_id), "supplies" (orders.supplier_id) and
// "is fulfilled by" between two orders (orders.for_order_id). Tallinn Tech
// buys from Baltic Foods (o-1264), and Baltic Foods buys from Riga Steel for
// that order (o-1268).
func newGraphDataTestApp() (*graphTestApp, *fakeTableEngine) {
	g := newGraphTestApp(true)

	g.collimato.permissions = graphViewerPermissions

	engine := &fakeTableEngine{tables: map[string][]map[string]any{
		"companies": {
			fakeCompany("c-03", "Baltic Foods"),
			fakeCompany("c-04", "Riga Steel"),
			fakeCompany("c-05", "Tallinn Tech"),
		},
		"orders": {
			fakeOrder("o-1264", "c-05", "c-03", ""),
			fakeOrder("o-1268", "c-03", "c-04", "o-1264"),
		},
	}}
	g.app.QueryEngine = engine

	g.graph.datasets = []model.GraphDataset{
		{ID: "ds-companies", WorkspaceID: "ws", Name: "companies"},
		{ID: "ds-orders", WorkspaceID: "ws", Name: "orders"},
	}

	for _, name := range []string{"companies.key", "companies.name"} {
		g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-" + name, WorkspaceID: "ws", DatasetID: "ds-companies", Name: name})
	}

	for _, name := range []string{"orders.key", "orders.number", "orders.company_id", "orders.supplier_id", "orders.for_order_id"} {
		g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-" + name, WorkspaceID: "ws", DatasetID: "ds-orders", Name: name})
	}

	g.graph.entityTypes = append(g.graph.entityTypes,
		model.GraphEntityType{ID: "t-company", Key: "company", Source: model.GraphSourceDataset, SourceID: "ds-companies", KeyProperty: "f-companies.key", TitleProperty: "f-companies.name"},
		model.GraphEntityType{ID: "t-order", Key: "order", Source: model.GraphSourceDataset, SourceID: "ds-orders", KeyProperty: "f-orders.key", TitleProperty: "f-orders.number"},
	)

	propertyRelation := func(key, from, to, property string) model.GraphRelationType {
		return model.GraphRelationType{
			ID:         "l-" + key,
			Key:        key,
			FromTypeID: from,
			ToTypeID:   to,
			Kind:       model.GraphRelationKindProperty,
			KindConfig: &model.GraphRelationKindConfig{Property: property},
		}
	}

	g.graph.relationTypes = append(g.graph.relationTypes,
		propertyRelation("company_places_order", "t-company", "t-order", "f-orders.company_id"),
		propertyRelation("company_supplies_order", "t-company", "t-order", "f-orders.supplier_id"),
		propertyRelation("order_fulfilled_by_order", "t-order", "t-order", "f-orders.for_order_id"),
	)

	return g, engine
}

func dataGroup(relations *model.GraphEntityRelations, relationType, direction string) *model.GraphRelationGroup {
	for _, group := range relations.Groups {
		if group.RelationType == relationType && group.Direction == direction {
			return &group
		}
	}

	return nil
}

func TestDataEntityRelationsFollowTheProperties(t *testing.T) {
	g, _ := newGraphDataTestApp()

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "order", "o-1264")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	if relations.Entity.Title != "Order o-1264" {
		t.Errorf("want the order's own name, got %+v", relations.Entity)
	}

	placedBy := dataGroup(relations, "company_places_order", model.GraphDirectionIn)
	if !slices.Equal(graphItemIDs(placedBy), []string{"c-05"}) {
		t.Errorf("want placed by Tallinn Tech, got %+v", placedBy)
	}

	suppliedBy := dataGroup(relations, "company_supplies_order", model.GraphDirectionIn)
	if !slices.Equal(graphItemIDs(suppliedBy), []string{"c-03"}) {
		t.Errorf("want supplied by Baltic Foods, got %+v", suppliedBy)
	}

	fulfilledBy := dataGroup(relations, "order_fulfilled_by_order", model.GraphDirectionOut)
	if !slices.Equal(graphItemIDs(fulfilledBy), []string{"o-1268"}) {
		t.Errorf("want fulfilled by o-1268, got %+v", fulfilledBy)
	}

	if dataGroup(relations, "order_fulfilled_by_order", model.GraphDirectionIn) != nil {
		t.Errorf("want no 'fulfils' group: o-1264 was not placed for another order")
	}
}

func TestDataEntityRelationsSayWhichFieldTheyAreReadFrom(t *testing.T) {
	g, _ := newGraphDataTestApp()

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "order", "o-1264")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	placedBy := dataGroup(relations, "company_places_order", model.GraphDirectionIn)
	if placedBy == nil {
		t.Fatalf("want a placed by group, got %+v", relations.Groups)
	}

	item := placedBy.Items[0]
	if item.Origin != model.GraphRelationOriginStructured || item.Source != "orders.company_id" {
		t.Errorf("want a structured relation read from orders.company_id, got %+v", item)
	}
}

func TestDataEntityRelationsBetweenTwoOrdersWorkBothWays(t *testing.T) {
	g, _ := newGraphDataTestApp()

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "order", "o-1268")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	fulfils := dataGroup(relations, "order_fulfilled_by_order", model.GraphDirectionIn)
	if !slices.Equal(graphItemIDs(fulfils), []string{"o-1264"}) {
		t.Errorf("want o-1268 to fulfil o-1264, got %+v", fulfils)
	}
}

func TestDataEntityRelationsCountABigGroupAndCapIt(t *testing.T) {
	cases := []struct {
		name       string
		orders     int
		wantCount  int
		wantCapped bool
	}{
		{"a group too big to list", graphGroupItemLimit + 10, graphGroupItemLimit + 10, false},
		{"a group over the cap", graphCountCap + 200, graphCountCap, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := newGraphDataTestApp()

			for i := 0; i < tc.orders; i++ {
				key := fmt.Sprintf("o-%05d", i)
				engine.tables["orders"] = append(engine.tables["orders"], fakeOrder(key, "c-04", "c-03", ""))
			}

			relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "company", "c-04")
			if appErr != nil {
				t.Fatalf("GetGraphEntityRelations: %v", appErr)
			}

			placed := dataGroup(relations, "company_places_order", model.GraphDirectionOut)
			if placed == nil {
				t.Fatalf("want a group of placed orders")
			}

			if placed.Count != tc.wantCount || placed.CountCapped != tc.wantCapped || len(placed.Items) != 0 {
				t.Errorf("want %d (capped %v) with no items, got %d (capped %v) with %d items",
					tc.wantCount, tc.wantCapped, placed.Count, placed.CountCapped, len(placed.Items))
			}
		})
	}
}

func TestGraphRelationPageFollowsTheKeyOrder(t *testing.T) {
	g, engine := newGraphDataTestApp()

	for i := 0; i < 120; i++ {
		key := fmt.Sprintf("o-%05d", i)
		engine.tables["orders"] = append(engine.tables["orders"], fakeOrder(key, "c-04", "c-03", ""))
	}

	pages := make([]*model.GraphRelationPage, 0)
	after := ""

	for len(pages) < 5 {
		page, appErr := g.app.GetGraphRelationPage(context.Background(), model.User{ID: "u1"}, "company", "c-04", "company_places_order", model.GraphDirectionOut, after)
		if appErr != nil {
			t.Fatalf("GetGraphRelationPage: %v", appErr)
		}

		pages = append(pages, page)

		if page.Next == "" {
			break
		}

		after = page.Next
	}

	if len(pages) != 3 {
		t.Fatalf("want 3 pages for 120 orders, got %d", len(pages))
	}

	if len(pages[0].Items) != 50 || len(pages[1].Items) != 50 || len(pages[2].Items) != 20 {
		t.Errorf("want 50, 50 and 20 orders, got %d, %d and %d", len(pages[0].Items), len(pages[1].Items), len(pages[2].Items))
	}

	if pages[1].Items[0].Entity.ID != "o-00050" {
		t.Errorf("want page 2 to start right after page 1, got %s", pages[1].Items[0].Entity.ID)
	}
}

func TestDataEntityRelationsHideOnlyTheRelationWhosePropertyIsHidden(t *testing.T) {
	g, _ := newGraphDataTestApp()

	g.app.CollimatoRoles = fakeColumnRules{hidden: []string{"orders.supplier_id"}}

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "order", "o-1264")
	if appErr != nil {
		t.Fatalf("want the order still visible, got %v", appErr)
	}

	if dataGroup(relations, "company_supplies_order", model.GraphDirectionIn) != nil {
		t.Errorf("want no 'supplied by' group while orders.supplier_id is hidden")
	}

	if dataGroup(relations, "company_places_order", model.GraphDirectionIn) == nil {
		t.Errorf("want the 'placed by' group kept")
	}
}

func TestDataEntityRelationsSkipARelationTypeWhoseFieldWasRemoved(t *testing.T) {
	g, _ := newGraphDataTestApp()

	g.graph.fields = slices.DeleteFunc(g.graph.fields, func(field model.GraphDatasetField) bool {
		return field.Name == "orders.supplier_id"
	})

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "order", "o-1264")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	if dataGroup(relations, "company_supplies_order", model.GraphDirectionIn) != nil {
		t.Errorf("want the broken relation type skipped")
	}

	if dataGroup(relations, "company_places_order", model.GraphDirectionIn) == nil {
		t.Errorf("want the other relation types kept")
	}
}

func TestDataEntityRelationsAnswerNotFoundForAMissingOrHiddenEntity(t *testing.T) {
	cases := []struct {
		name     string
		typeKey  string
		id       string
		outsider bool
	}{
		{"a key that does not exist", "company", "c-99", false},
		{"a type that does not exist", "invoice", "c-04", false},
		{"a user outside the workspace", "company", "c-04", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := newGraphDataTestApp()

			user := model.User{ID: "u1"}
			if tc.outsider {
				user.ID = "outsider"
			}

			_, appErr := g.app.GetGraphEntityRelations(context.Background(), user, tc.typeKey, tc.id)
			if appErr == nil || appErr.ID != "graph.entity_not_found" {
				t.Errorf("want graph.entity_not_found, got %v", appErr)
			}
		})
	}
}

func TestSearchGraphEntitiesFindsByName(t *testing.T) {
	g, _ := newGraphDataTestApp()

	found, appErr := g.app.SearchGraphEntities(context.Background(), model.User{ID: "u1"}, "company", "tallinn", "")
	if appErr != nil {
		t.Fatalf("SearchGraphEntities: %v", appErr)
	}

	if len(found.Items) != 1 || found.Items[0].ID != "c-05" || found.Items[0].Title != "Tallinn Tech" {
		t.Errorf("want Tallinn Tech, got %+v", found.Items)
	}

	if found.Next != "" {
		t.Errorf("want no next page, got %q", found.Next)
	}

	_, appErr = g.app.SearchGraphEntities(context.Background(), model.User{ID: "outsider"}, "company", "", "")
	if appErr == nil || appErr.ID != "graph.entity_not_found" {
		t.Errorf("want the type unknown to a user outside the workspace, got %v", appErr)
	}
}

// The Entities table pages through a type 50 at a time, and a search keeps
// its filter on every page.
func TestSearchGraphEntitiesPagesInKeyOrder(t *testing.T) {
	g, engine := newGraphDataTestApp()

	for i := 0; i < 120; i++ {
		key := fmt.Sprintf("c-1%04d", i)
		name := fmt.Sprintf("Company %d", i)

		if i%2 == 0 {
			name = fmt.Sprintf("Shop %d", i)
		}

		engine.tables["companies"] = append(engine.tables["companies"], fakeCompany(key, name))
	}

	sizes := make([]int, 0)
	after := ""

	for len(sizes) < 5 {
		page, appErr := g.app.SearchGraphEntities(context.Background(), model.User{ID: "u1"}, "company", "shop", after)
		if appErr != nil {
			t.Fatalf("SearchGraphEntities: %v", appErr)
		}

		for _, item := range page.Items {
			if !strings.HasPrefix(item.Title, "Shop") {
				t.Fatalf("want only shops, got %q", item.Title)
			}
		}

		if after != "" && page.Items[0].ID <= after {
			t.Errorf("want the page to start after %s, got %s", after, page.Items[0].ID)
		}

		sizes = append(sizes, len(page.Items))

		if page.Next == "" {
			break
		}

		after = page.Next
	}

	if !slices.Equal(sizes, []int{50, 10}) {
		t.Errorf("want pages of 50 and 10 shops, got %v", sizes)
	}
}

// A "same as" relation matches on another field than the key: a company and a
// registry entry with the same registration number.
func TestDataEntityRelationsCanMatchOnAFromProperty(t *testing.T) {
	g, engine := newGraphDataTestApp()

	engine.tables["companies"][2]["companies.reg_no"] = "40003000001"
	engine.tables["registry"] = []map[string]any{
		{"registry.key": "r-1", "registry.name": "TALLINN TECH OU", "registry.reg_no": "40003000001"},
		{"registry.key": "r-2", "registry.name": "OTHER OU", "registry.reg_no": "40003999999"},
	}

	g.graph.datasets = append(g.graph.datasets, model.GraphDataset{ID: "ds-registry", WorkspaceID: "ws", Name: "registry"})

	g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-companies.reg_no", WorkspaceID: "ws", DatasetID: "ds-companies", Name: "companies.reg_no"})

	for _, name := range []string{"registry.key", "registry.name", "registry.reg_no"} {
		g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-" + name, WorkspaceID: "ws", DatasetID: "ds-registry", Name: name})
	}

	g.graph.entityTypes = append(g.graph.entityTypes, model.GraphEntityType{
		ID: "t-registration", Key: "registration", Source: model.GraphSourceDataset,
		SourceID: "ds-registry", KeyProperty: "f-registry.key", TitleProperty: "f-registry.name",
	})

	g.graph.relationTypes = append(g.graph.relationTypes, model.GraphRelationType{
		ID: "l-same", Key: "company_same_as_registration", FromTypeID: "t-company", ToTypeID: "t-registration",
		Kind:       model.GraphRelationKindProperty,
		KindConfig: &model.GraphRelationKindConfig{Property: "f-registry.reg_no", FromProperty: "f-companies.reg_no"},
	})

	fromCompany, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "company", "c-05")
	if appErr != nil {
		t.Fatalf("company relations: %v", appErr)
	}

	sameAs := dataGroup(fromCompany, "company_same_as_registration", model.GraphDirectionOut)
	if !slices.Equal(graphItemIDs(sameAs), []string{"r-1"}) {
		t.Errorf("want the registry entry with the same number, got %+v", sameAs)
	}

	fromRegistry, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "registration", "r-1")
	if appErr != nil {
		t.Fatalf("registry relations: %v", appErr)
	}

	back := dataGroup(fromRegistry, "company_same_as_registration", model.GraphDirectionIn)
	if !slices.Equal(graphItemIDs(back), []string{"c-05"}) {
		t.Errorf("want Tallinn Tech back from the registry entry, got %+v", back)
	}
}

// newGraphTripsTestApp adds trips whose zone repeats, and a zone type made
// straight from that column, the way a type is made from a column of a table
// of events rather than from a table of zones.
func newGraphTripsTestApp() (*graphTestApp, *fakeTableEngine) {
	g, engine := newGraphDataTestApp()

	trip := func(key, zone string) map[string]any {
		return map[string]any{"trips.key": key, "trips.zone": zone}
	}

	engine.tables["trips"] = []map[string]any{
		trip("t-1", "z-1"), trip("t-2", "z-1"), trip("t-3", "z-1"),
		trip("t-4", "z-2"), trip("t-5", "z-2"),
	}

	g.graph.datasets = append(g.graph.datasets, model.GraphDataset{ID: "ds-trips", WorkspaceID: "ws", Name: "trips"})

	for _, name := range []string{"trips.key", "trips.zone"} {
		g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-" + name, WorkspaceID: "ws", DatasetID: "ds-trips", Name: name})
	}

	g.graph.entityTypes = append(g.graph.entityTypes,
		model.GraphEntityType{ID: "t-zone", Key: "zone", Source: model.GraphSourceDataset, SourceID: "ds-trips", KeyProperty: "f-trips.zone", TitleProperty: "f-trips.zone"},
		model.GraphEntityType{ID: "t-trip", Key: "trip", Source: model.GraphSourceDataset, SourceID: "ds-trips", KeyProperty: "f-trips.key", TitleProperty: "f-trips.key"},
	)

	g.graph.relationTypes = append(g.graph.relationTypes, model.GraphRelationType{
		ID:         "l-zone_has_trip",
		Key:        "zone_has_trip",
		FromTypeID: "t-zone",
		ToTypeID:   "t-trip",
		Kind:       model.GraphRelationKindProperty,
		KindConfig: &model.GraphRelationKindConfig{Property: "f-trips.zone"},
	})

	return g, engine
}

func TestATypeFromARepeatingColumnListsEachEntityOnce(t *testing.T) {
	g, _ := newGraphTripsTestApp()

	found, appErr := g.app.SearchGraphEntities(context.Background(), model.User{ID: "u1"}, "zone", "", "")
	if appErr != nil {
		t.Fatalf("SearchGraphEntities: %v", appErr)
	}

	ids := []string{}
	for _, item := range found.Items {
		ids = append(ids, item.ID)
	}

	if !slices.Equal(ids, []string{"z-1", "z-2"}) {
		t.Errorf("want each zone once, got %v", ids)
	}
}

func TestARelationToARepeatingColumnShowsTheEntityOnce(t *testing.T) {
	g, _ := newGraphTripsTestApp()

	relations, appErr := g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "trip", "t-1")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	zones := dataGroup(relations, "zone_has_trip", model.GraphDirectionIn)
	if zones == nil || zones.Count != 1 || len(zones.Items) != 1 || zones.Items[0].Entity.ID != "z-1" {
		t.Errorf("want trip t-1 in zone z-1 once, got %+v", zones)
	}

	relations, appErr = g.app.GetGraphEntityRelations(context.Background(), model.User{ID: "u1"}, "zone", "z-1")
	if appErr != nil {
		t.Fatalf("GetGraphEntityRelations: %v", appErr)
	}

	trips := dataGroup(relations, "zone_has_trip", model.GraphDirectionOut)
	if trips == nil || trips.Count != 3 {
		t.Errorf("want zone z-1 to have 3 trips, got %+v", trips)
	}
}

func TestATypeWithAUniqueKeyIsReadOnce(t *testing.T) {
	g, engine := newGraphDataTestApp()

	_, appErr := g.app.SearchGraphEntities(context.Background(), model.User{ID: "u1"}, "company", "", "")
	if appErr != nil {
		t.Fatalf("SearchGraphEntities: %v", appErr)
	}

	if engine.queries != 1 {
		t.Errorf("want one fast read for unique keys, got %d queries", engine.queries)
	}
}

// zoneTrips gives each zone one trip for each of its names, the way a zone
// type named by a field that varies from trip to trip sees its data.
func zoneTrips(zones []string, namesPerZone int) []map[string]any {
	trips := make([]map[string]any, 0, len(zones)*namesPerZone)

	for _, zone := range zones {
		for name := 1; name <= namesPerZone; name++ {
			trip := map[string]any{
				"trips.key":    fmt.Sprintf("t-%s-%04d", zone, name),
				"trips.zone":   zone,
				"trips.vendor": fmt.Sprintf("Vendor %04d", name),
			}

			trips = append(trips, trip)
		}
	}

	return trips
}

func numberedZones(count int) []string {
	zones := make([]string, 0, count)

	for zone := 1; zone <= count; zone++ {
		zones = append(zones, fmt.Sprintf("z-%03d", zone))
	}

	return zones
}

func newZonesByVendorTestApp(trips []map[string]any) (*graphTestApp, *fakeTableEngine) {
	g, engine := newGraphTripsTestApp()

	engine.tables["trips"] = trips

	g.graph.fields = append(g.graph.fields, model.GraphDatasetField{ID: "f-trips.vendor", WorkspaceID: "ws", DatasetID: "ds-trips", Name: "trips.vendor"})
	g.graph.entityTypes = append(g.graph.entityTypes, model.GraphEntityType{ID: "t-zone-by-vendor", Key: "zone_by_vendor", Source: model.GraphSourceDataset, SourceID: "ds-trips", KeyProperty: "f-trips.zone", TitleProperty: "f-trips.vendor"})

	return g, engine
}

// listAll pages through a type's entities the way the Entities table does.
func listAll(t *testing.T, g *graphTestApp, typeKey string) []model.GraphEntityRef {
	t.Helper()

	all := []model.GraphEntityRef{}
	after := ""

	for range 20 {
		page, appErr := g.app.SearchGraphEntities(context.Background(), model.User{ID: "u1"}, typeKey, "", after)
		if appErr != nil {
			t.Fatalf("SearchGraphEntities: %v", appErr)
		}

		all = append(all, page.Items...)

		after = page.Next
		if after == "" {
			return all
		}
	}

	t.Fatal("want the list to end")

	return nil
}

func TestABlankKeyIsLeftOut(t *testing.T) {
	trips := zoneTrips([]string{"", "z-1", "z-2"}, 60)
	trips = append(trips, map[string]any{"trips.key": "t-none", "trips.zone": nil, "trips.vendor": "Vendor X"})

	g, _ := newZonesByVendorTestApp(trips)

	ids := []string{}
	for _, entity := range listAll(t, g, "zone_by_vendor") {
		ids = append(ids, entity.ID)
	}

	if !slices.Equal(ids, []string{"z-1", "z-2"}) {
		t.Errorf("want only the zones with a key, got %v", ids)
	}
}

func TestAKeyWithTwoTitlesDoesNotEndTheListEarly(t *testing.T) {
	zones := numberedZones(60)

	g, _ := newZonesByVendorTestApp(zoneTrips(zones, 2))

	ids := []string{}
	for _, entity := range listAll(t, g, "zone_by_vendor") {
		ids = append(ids, entity.ID)
	}

	if !slices.Equal(ids, zones) {
		t.Errorf("want all 60 zones once and in order, got %d: %v", len(ids), ids)
	}
}

func TestAKeyWithManyNamesStillFillsThePage(t *testing.T) {
	zones := numberedZones(60)

	g, engine := newZonesByVendorTestApp(zoneTrips(zones, 100))

	entities := listAll(t, g, "zone_by_vendor")

	// At most three reads a page, two pages here: a quick read, the keys,
	// then their names.
	if engine.queries > 6 {
		t.Errorf("want at most three reads a page, got %d for two pages", engine.queries)
	}

	ids := []string{}
	for _, entity := range entities {
		ids = append(ids, entity.ID)

		named := strings.HasPrefix(entity.Title, "Vendor ") || entity.Title == entity.ID
		if !named {
			t.Errorf("want %s named by one of its vendors or by its key, got %q", entity.ID, entity.Title)
		}
	}

	if !slices.Equal(ids, zones) {
		t.Errorf("want all 60 zones once and in order, got %d: %v", len(ids), ids)
	}
}
