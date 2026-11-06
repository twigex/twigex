// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/twigex/twigex/model"
)

type graphRepository struct {
	Db *sql.DB
}

const (
	graphEntityTypeColumns   = "id, `key`, name, source, source_id, key_property, title_property, description, built_in, created_by, created_at, updated_at"
	graphRelationTypeColumns = "id, `key`, name, reverse_name, from_type_id, to_type_id, kind, description, built_in, created_by, created_at, updated_at"
	graphRelationColumns     = "id, relation_type_id, from_type_id, from_id, to_type_id, to_id, origin, created_by, created_at"
	graphDatasetColumns      = "id, workspace_id, name, kind, created_at"
	graphDatasetFieldColumns = "id, workspace_id, dataset_id, name, created_at"
)

const graphInsertBatchSize = 500

func NewGraphRepository(db *sql.DB) *graphRepository {
	return &graphRepository{Db: db}
}

func (g *graphRepository) CreateOrUpdateEntityType(ctx context.Context, entityType model.GraphEntityType) error {
	now := time.Now().Unix()

	_, err := g.Db.ExecContext(ctx,
		"INSERT INTO graph_entity_types ("+graphEntityTypeColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		     name = VALUES(name),
		     source = VALUES(source),
		     source_id = VALUES(source_id),
		     key_property = VALUES(key_property),
		     title_property = VALUES(title_property),
		     description = VALUES(description),
		     built_in = VALUES(built_in),
		     updated_at = VALUES(updated_at)`,
		model.NewID(),
		entityType.Key,
		entityType.Name,
		entityType.Source,
		nullableString(entityType.SourceID),
		nullableString(entityType.KeyProperty),
		nullableString(entityType.TitleProperty),
		nullableString(entityType.Description),
		entityType.BuiltIn,
		nullableString(entityType.CreatedBy),
		now,
		now,
	)

	return err
}

func (g *graphRepository) GetAllEntityTypes(ctx context.Context) ([]model.GraphEntityType, error) {
	rows, err := g.Db.QueryContext(ctx, "SELECT "+graphEntityTypeColumns+" FROM graph_entity_types ORDER BY `key`")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	entityTypes := make([]model.GraphEntityType, 0)
	for rows.Next() {
		entityType, err := scanGraphEntityType(rows)
		if err != nil {
			return nil, err
		}

		entityTypes = append(entityTypes, *entityType)
	}

	return entityTypes, rows.Err()
}

func (g *graphRepository) CreateOrUpdateRelationType(ctx context.Context, relationType model.GraphRelationType) error {
	now := time.Now().Unix()

	_, err := g.Db.ExecContext(ctx,
		"INSERT INTO graph_relation_types ("+graphRelationTypeColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		     name = VALUES(name),
		     reverse_name = VALUES(reverse_name),
		     from_type_id = VALUES(from_type_id),
		     to_type_id = VALUES(to_type_id),
		     kind = VALUES(kind),
		     description = VALUES(description),
		     built_in = VALUES(built_in),
		     updated_at = VALUES(updated_at)`,
		model.NewID(),
		relationType.Key,
		relationType.Name,
		relationType.ReverseName,
		relationType.FromTypeID,
		relationType.ToTypeID,
		relationType.Kind,
		nullableString(relationType.Description),
		relationType.BuiltIn,
		nullableString(relationType.CreatedBy),
		now,
		now,
	)

	return err
}

func (g *graphRepository) GetAllRelationTypes(ctx context.Context) ([]model.GraphRelationType, error) {
	rows, err := g.Db.QueryContext(ctx, "SELECT "+graphRelationTypeColumns+" FROM graph_relation_types ORDER BY `key`")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	relationTypes := make([]model.GraphRelationType, 0)
	for rows.Next() {
		relationType, err := scanGraphRelationType(rows)
		if err != nil {
			return nil, err
		}

		relationTypes = append(relationTypes, *relationType)
	}

	return relationTypes, rows.Err()
}

func (g *graphRepository) AddDatasets(ctx context.Context, datasets []model.GraphDataset) (int64, error) {
	now := time.Now().Unix()

	var added int64
	for start := 0; start < len(datasets); start += graphInsertBatchSize {
		end := min(start+graphInsertBatchSize, len(datasets))
		batch := datasets[start:end]

		args := make([]any, 0, len(batch)*5)
		for _, dataset := range batch {
			args = append(args, model.NewID(), dataset.WorkspaceID, dataset.Name, dataset.Kind, now)
		}

		// A sync running at the same time may have added the same name first,
		// and its row keeps its ID.
		result, err := g.Db.ExecContext(ctx,
			"INSERT INTO graph_datasets ("+graphDatasetColumns+") VALUES "+rowPlaceholders(len(batch), 5)+" ON DUPLICATE KEY UPDATE id = id",
			args...,
		)
		if err != nil {
			return 0, err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}

		added += count
	}

	return added, nil
}

func (g *graphRepository) GetAllDatasetsForWorkspace(ctx context.Context, workspaceID string) ([]model.GraphDataset, error) {
	rows, err := g.Db.QueryContext(ctx, "SELECT "+graphDatasetColumns+" FROM graph_datasets WHERE workspace_id = ? ORDER BY name", workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	datasets := make([]model.GraphDataset, 0)
	for rows.Next() {
		var dataset model.GraphDataset
		err := rows.Scan(&dataset.ID, &dataset.WorkspaceID, &dataset.Name, &dataset.Kind, &dataset.CreatedAt)
		if err != nil {
			return nil, err
		}

		datasets = append(datasets, dataset)
	}

	return datasets, rows.Err()
}

func (g *graphRepository) DeleteDatasets(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	args := toInterfaceSlice(ids)

	tx, err := g.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM graph_dataset_fields WHERE dataset_id IN ("+sqlPlaceholders(len(ids))+")", args...)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM graph_datasets WHERE id IN ("+sqlPlaceholders(len(ids))+")", args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (g *graphRepository) AddDatasetFields(ctx context.Context, fields []model.GraphDatasetField) (int64, error) {
	now := time.Now().Unix()

	var added int64
	for start := 0; start < len(fields); start += graphInsertBatchSize {
		end := min(start+graphInsertBatchSize, len(fields))
		batch := fields[start:end]

		args := make([]any, 0, len(batch)*5)
		for _, field := range batch {
			args = append(args, model.NewID(), field.WorkspaceID, field.DatasetID, field.Name, now)
		}

		result, err := g.Db.ExecContext(ctx,
			"INSERT INTO graph_dataset_fields ("+graphDatasetFieldColumns+") VALUES "+rowPlaceholders(len(batch), 5)+" ON DUPLICATE KEY UPDATE id = id",
			args...,
		)
		if err != nil {
			return 0, err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}

		added += count
	}

	return added, nil
}

func (g *graphRepository) GetAllDatasetFieldsForWorkspace(ctx context.Context, workspaceID string) ([]model.GraphDatasetField, error) {
	rows, err := g.Db.QueryContext(ctx, "SELECT "+graphDatasetFieldColumns+" FROM graph_dataset_fields WHERE workspace_id = ? ORDER BY name", workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanGraphDatasetFields(rows)
}

func (g *graphRepository) GetDatasetFieldsByNames(ctx context.Context, workspaceID string, names []string) ([]model.GraphDatasetField, error) {
	if len(names) == 0 {
		return make([]model.GraphDatasetField, 0), nil
	}

	args := []any{workspaceID}
	args = append(args, toInterfaceSlice(names)...)

	rows, err := g.Db.QueryContext(ctx,
		"SELECT "+graphDatasetFieldColumns+" FROM graph_dataset_fields WHERE workspace_id = ? AND name IN ("+sqlPlaceholders(len(names))+") ORDER BY name",
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanGraphDatasetFields(rows)
}

func (g *graphRepository) DeleteDatasetFields(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	args := toInterfaceSlice(ids)

	_, err := g.Db.ExecContext(ctx, "DELETE FROM graph_dataset_fields WHERE id IN ("+sqlPlaceholders(len(ids))+")", args...)

	return err
}

func (g *graphRepository) UpdateStructuredRelations(ctx context.Context, relationTypeID, fromTypeID, fromID, toTypeID string, toIDs []string) (int64, int64, error) {
	tx, err := g.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}

	defer tx.Rollback()

	deleteQuery := "DELETE FROM graph_relations WHERE relation_type_id = ? AND from_type_id = ? AND from_id = ? AND origin = ?"
	deleteArgs := []any{relationTypeID, fromTypeID, fromID, model.GraphRelationOriginStructured}

	if len(toIDs) > 0 {
		deleteQuery += " AND to_id NOT IN (" + sqlPlaceholders(len(toIDs)) + ")"
		deleteArgs = append(deleteArgs, toInterfaceSlice(toIDs)...)
	}

	deleted, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		return 0, 0, err
	}

	removed, err := deleted.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	var added int64
	if len(toIDs) > 0 {
		now := time.Now().Unix()

		args := make([]any, 0, len(toIDs)*9)
		for _, toID := range toIDs {
			args = append(args, model.NewID(), relationTypeID, fromTypeID, fromID, toTypeID, toID, model.GraphRelationOriginStructured, nil, now)
		}

		// A relation already there keeps its row. That includes a manual relation
		// between the same two entities, which stays in place of the structured one.
		inserted, err := tx.ExecContext(ctx,
			"INSERT INTO graph_relations ("+graphRelationColumns+") VALUES "+rowPlaceholders(len(toIDs), 9)+" ON DUPLICATE KEY UPDATE id = id",
			args...,
		)
		if err != nil {
			return 0, 0, err
		}

		added, err = inserted.RowsAffected()
		if err != nil {
			return 0, 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	return added, removed, nil
}

func (g *graphRepository) GetRelationsFrom(ctx context.Context, typeID, entityID string) ([]model.GraphRelation, error) {
	rows, err := g.Db.QueryContext(ctx,
		"SELECT "+graphRelationColumns+" FROM graph_relations WHERE from_type_id = ? AND from_id = ? ORDER BY relation_type_id, to_id",
		typeID, entityID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanGraphRelations(rows)
}

func (g *graphRepository) GetRelationsTo(ctx context.Context, typeID, entityID string) ([]model.GraphRelation, error) {
	rows, err := g.Db.QueryContext(ctx,
		"SELECT "+graphRelationColumns+" FROM graph_relations WHERE to_type_id = ? AND to_id = ? ORDER BY relation_type_id, from_id",
		typeID, entityID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanGraphRelations(rows)
}

func (g *graphRepository) DeleteEntityRelations(ctx context.Context, typeID string, entityIDs []string) error {
	if len(entityIDs) == 0 {
		return nil
	}

	in := sqlPlaceholders(len(entityIDs))

	idArgs := toInterfaceSlice(entityIDs)

	args := []any{typeID}
	args = append(args, idArgs...)
	args = append(args, typeID)
	args = append(args, idArgs...)

	_, err := g.Db.ExecContext(ctx,
		"DELETE FROM graph_relations WHERE (from_type_id = ? AND from_id IN ("+in+")) OR (to_type_id = ? AND to_id IN ("+in+"))",
		args...,
	)

	return err
}

type graphRowScanner interface {
	Scan(dest ...any) error
}

func scanGraphEntityType(row graphRowScanner) (*model.GraphEntityType, error) {
	var entityType model.GraphEntityType
	var sourceID, keyProperty, titleProperty, description, createdBy sql.NullString

	err := row.Scan(
		&entityType.ID,
		&entityType.Key,
		&entityType.Name,
		&entityType.Source,
		&sourceID,
		&keyProperty,
		&titleProperty,
		&description,
		&entityType.BuiltIn,
		&createdBy,
		&entityType.CreatedAt,
		&entityType.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	entityType.SourceID = sourceID.String
	entityType.KeyProperty = keyProperty.String
	entityType.TitleProperty = titleProperty.String
	entityType.Description = description.String
	entityType.CreatedBy = createdBy.String

	return &entityType, nil
}

func scanGraphRelationType(row graphRowScanner) (*model.GraphRelationType, error) {
	var relationType model.GraphRelationType
	var description, createdBy sql.NullString

	err := row.Scan(
		&relationType.ID,
		&relationType.Key,
		&relationType.Name,
		&relationType.ReverseName,
		&relationType.FromTypeID,
		&relationType.ToTypeID,
		&relationType.Kind,
		&description,
		&relationType.BuiltIn,
		&createdBy,
		&relationType.CreatedAt,
		&relationType.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	relationType.Description = description.String
	relationType.CreatedBy = createdBy.String

	return &relationType, nil
}

func scanGraphDatasetFields(rows *sql.Rows) ([]model.GraphDatasetField, error) {
	fields := make([]model.GraphDatasetField, 0)
	for rows.Next() {
		var field model.GraphDatasetField
		err := rows.Scan(&field.ID, &field.WorkspaceID, &field.DatasetID, &field.Name, &field.CreatedAt)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, rows.Err()
}

func scanGraphRelations(rows *sql.Rows) ([]model.GraphRelation, error) {
	relations := make([]model.GraphRelation, 0)
	for rows.Next() {
		var relation model.GraphRelation
		var createdBy sql.NullString

		err := rows.Scan(
			&relation.ID,
			&relation.RelationTypeID,
			&relation.FromTypeID,
			&relation.FromID,
			&relation.ToTypeID,
			&relation.ToID,
			&relation.Origin,
			&createdBy,
			&relation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		relation.CreatedBy = createdBy.String
		relations = append(relations, relation)
	}

	return relations, rows.Err()
}
