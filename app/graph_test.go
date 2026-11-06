// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeGraphStore struct {
	store.GraphStore
	entityTypes   []model.GraphEntityType
	relationTypes []model.GraphRelationType
}

func (f *fakeGraphStore) CreateOrUpdateEntityType(_ context.Context, entityType model.GraphEntityType) error {
	entityType.ID = "id-" + entityType.Key
	f.entityTypes = append(f.entityTypes, entityType)

	return nil
}

func (f *fakeGraphStore) GetAllEntityTypes(_ context.Context) ([]model.GraphEntityType, error) {
	return f.entityTypes, nil
}

func (f *fakeGraphStore) CreateOrUpdateRelationType(_ context.Context, relationType model.GraphRelationType) error {
	f.relationTypes = append(f.relationTypes, relationType)

	return nil
}

func graphApp(s *fakeGraphStore) *App {
	return &App{
		Store: store.Store{Graph: s},
		ConfigStore: config.ConfigStore{
			Config: &model.ServerConfig{
				SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
			},
		},
	}
}

func TestEnsureGraphBuiltInTypesRelationsPointToTheirEntityTypes(t *testing.T) {
	s := &fakeGraphStore{}
	a := graphApp(s)

	a.EnsureGraphBuiltInTypes()

	if len(s.entityTypes) != 4 {
		t.Fatalf("want 4 entity types, got %d", len(s.entityTypes))
	}

	if len(s.relationTypes) != 4 {
		t.Fatalf("want 4 relation types, got %d", len(s.relationTypes))
	}

	for _, relationType := range s.relationTypes {
		if !relationType.BuiltIn {
			t.Errorf("%s is not marked built in", relationType.Key)
		}

		if relationType.Key != model.GraphRelationChartUsesDatasetField {
			continue
		}

		if relationType.FromTypeID != "id-chart" {
			t.Errorf("want from type id-chart, got %q", relationType.FromTypeID)
		}

		if relationType.ToTypeID != "id-dataset_field" {
			t.Errorf("want to type id-dataset_field, got %q", relationType.ToTypeID)
		}

		if relationType.Kind != model.GraphRelationKindStored {
			t.Errorf("want kind stored, got %q", relationType.Kind)
		}
	}
}
