// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"slices"
	"testing"

	"github.com/twigex/twigex/model"
)

func graphTypeKeys(types *model.GraphTypes) ([]string, []string) {
	entityKeys := make([]string, 0, len(types.EntityTypes))
	for _, entityType := range types.EntityTypes {
		entityKeys = append(entityKeys, entityType.Key)
	}

	relationKeys := make([]string, 0, len(types.RelationTypes))
	for _, relationType := range types.RelationTypes {
		relationKeys = append(relationKeys, relationType.Key)
	}

	return entityKeys, relationKeys
}

// Every graph endpoint must answer for a hidden type as if it did not exist.
// A new endpoint belongs in this test.
func TestAGraphTypeOnHiddenDataIsGoneFromEveryEndpoint(t *testing.T) {
	g, _ := newGraphDataTestApp()

	g.app.CollimatoRoles = fakeColumnRules{hidden: []string{"companies.name"}}

	ctx := context.Background()
	user := model.User{ID: "u1"}

	types, appErr := g.app.GetGraphTypes(ctx, user)
	if appErr != nil {
		t.Fatalf("GetGraphTypes: %v", appErr)
	}

	entityKeys, relationKeys := graphTypeKeys(types)

	if slices.Contains(entityKeys, "company") || !slices.Contains(entityKeys, "order") {
		t.Errorf("want company left out and order kept, got %v", entityKeys)
	}

	if slices.Contains(relationKeys, "company_places_order") || !slices.Contains(relationKeys, "order_fulfilled_by_order") {
		t.Errorf("want the relations to company left out and the others kept, got %v", relationKeys)
	}

	_, appErr = g.app.GetGraphEntityRelations(ctx, user, "company", "c-05")
	if appErr == nil || appErr.ID != "graph.entity_not_found" {
		t.Errorf("relations: want graph.entity_not_found, got %v", appErr)
	}

	_, appErr = g.app.GetGraphRelationPage(ctx, user, "company", "c-05", "company_places_order", model.GraphDirectionOut, "")
	if appErr == nil || appErr.ID != "graph.entity_not_found" {
		t.Errorf("relation page: want graph.entity_not_found, got %v", appErr)
	}

	_, appErr = g.app.SearchGraphEntities(ctx, user, "company", "", "")
	if appErr == nil || appErr.ID != "graph.entity_not_found" {
		t.Errorf("search: want graph.entity_not_found, got %v", appErr)
	}

	order, appErr := g.app.GetGraphEntityRelations(ctx, user, "order", "o-1264")
	if appErr != nil {
		t.Fatalf("order relations: %v", appErr)
	}

	for _, group := range order.Groups {
		if group.EntityType == "company" {
			t.Errorf("want no group leading to the hidden type, got %+v", group)
		}
	}
}

func TestAGraphManagerOutsideTheWorkspaceCannotSeeOrDeleteItsTypes(t *testing.T) {
	g, _ := newGraphDataTestApp()

	ctx := context.Background()
	outsider := model.User{ID: "outsider", Role: model.SystemAdminRoleId}
	member := model.User{ID: "u1", Role: model.SystemAdminRoleId}

	types, appErr := g.app.GetGraphTypes(ctx, outsider)
	if appErr != nil {
		t.Fatalf("GetGraphTypes: %v", appErr)
	}

	entityKeys, _ := graphTypeKeys(types)
	if slices.Contains(entityKeys, "company") || slices.Contains(entityKeys, "order") {
		t.Errorf("want the workspace's types left out, got %v", entityKeys)
	}

	appErr = g.app.DeleteGraphRelationType(ctx, outsider, "l-company_places_order")
	if appErr == nil || appErr.ID != "graph.type_not_found" {
		t.Errorf("want graph.type_not_found, got %v", appErr)
	}

	types, appErr = g.app.GetGraphTypes(ctx, member)
	if appErr != nil {
		t.Fatalf("GetGraphTypes for a member: %v", appErr)
	}

	index := slices.IndexFunc(types.EntityTypes, func(entityType model.GraphEntityType) bool {
		return entityType.Key == "company"
	})

	if index == -1 || types.EntityTypes[index].SourceName != "companies" {
		t.Errorf("want a manager in the workspace to see the type and its dataset")
	}
}

// A type whose dataset was removed names nothing any more, and managers need
// to see it to delete it.
func TestAGraphTypeWhoseDatasetWasRemovedShowsOnlyToManagers(t *testing.T) {
	g, _ := newGraphDataTestApp()

	g.graph.datasets = slices.DeleteFunc(g.graph.datasets, func(dataset model.GraphDataset) bool {
		return dataset.ID == "ds-companies"
	})

	ctx := context.Background()

	forManager, appErr := g.app.GetGraphTypes(ctx, model.User{ID: "u1", Role: model.SystemAdminRoleId})
	if appErr != nil {
		t.Fatalf("GetGraphTypes for a manager: %v", appErr)
	}

	managerKeys, _ := graphTypeKeys(forManager)
	if !slices.Contains(managerKeys, "company") {
		t.Errorf("want the broken type shown to a manager, got %v", managerKeys)
	}

	forUser, appErr := g.app.GetGraphTypes(ctx, model.User{ID: "u1"})
	if appErr != nil {
		t.Fatalf("GetGraphTypes for a user: %v", appErr)
	}

	userKeys, _ := graphTypeKeys(forUser)
	if slices.Contains(userKeys, "company") {
		t.Errorf("want the broken type left out for a user, got %v", userKeys)
	}
}
