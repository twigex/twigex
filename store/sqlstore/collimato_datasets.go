// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/twigex/twigex/model"
)

// The names Cube.js saved things under.
const (
	oldDatasetFileType    = "cube"
	oldDatasetReference   = "{CUBE}"
	newDatasetReference   = "{DATASET}"
	oldViewPartsKey       = "cubes"
	newViewPartsKey       = "parts"
	oldBuilderModelEnding = ".yml"
)

type savedBuilderModel struct {
	id           string
	workspaceID  string
	fileType     string
	name         string
	builderModel string
}

// RenameCubesToDatasets moves what was saved under Cube.js's names to the
// engine's, in one transaction. Running it again changes nothing.
//
//   - files of type cube, and knowledge graph datasets of kind cube, become
//     dataset;
//   - Cube.js's own settings files (cubejs, env) are deleted;
//   - builder models drop the .yml from their name, and their YAML, unless
//     a file of the plain name exists;
//   - in builder models, {CUBE} becomes {DATASET} and "cubes" becomes "parts".
//
// It returns the builder models that kept their .yml name.
func (c *collimatoRepository) RenameCubesToDatasets() ([]string, error) {
	tx, err := c.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE collimato_workspace_files SET file_type = ? WHERE file_type = ?`,
		model.CollimatoFileTypeDataset, oldDatasetFileType)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM collimato_workspace_files WHERE file_type IN ('cubejs', 'env')`)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`UPDATE graph_datasets SET kind = ? WHERE kind = ?`,
		model.GraphDatasetKindDataset, oldDatasetFileType)
	if err != nil {
		return nil, err
	}

	builderModels, err := savedBuilderModels(tx)
	if err != nil {
		return nil, err
	}

	var keptOldName []string

	for _, saved := range builderModels {
		renamed, err := renameBuilderModel(tx, saved)
		if err != nil {
			return nil, err
		}

		if !renamed {
			keptOldName = append(keptOldName, saved.name)
		}
	}

	return keptOldName, tx.Commit()
}

func savedBuilderModels(tx *sql.Tx) ([]savedBuilderModel, error) {
	rows, err := tx.Query(
		`SELECT id, workspace_id, file_type, name, builder_model FROM collimato_workspace_files
		 WHERE builder_model IS NOT NULL`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var builderModels []savedBuilderModel

	for rows.Next() {
		var saved savedBuilderModel

		err := rows.Scan(&saved.id, &saved.workspaceID, &saved.fileType, &saved.name, &saved.builderModel)
		if err != nil {
			return nil, err
		}

		builderModels = append(builderModels, saved)
	}

	return builderModels, rows.Err()
}

// renameBuilderModel updates one builder model. It reports false when the
// model had to keep its .yml name, because a file of the plain name and the
// same type exists: the unique key would refuse the plain name. Such a model
// still works, since a dataset's name comes from its saved model.
func renameBuilderModel(tx *sql.Tx, saved savedBuilderModel) (bool, error) {
	builderModel, err := withDatasetWords(saved.builderModel, saved.fileType)
	if err != nil {
		return false, err
	}

	if !strings.HasSuffix(saved.name, oldBuilderModelEnding) {
		_, err := tx.Exec(
			`UPDATE collimato_workspace_files SET builder_model = ? WHERE id = ?`,
			builderModel, saved.id)

		return true, err
	}

	plainName := strings.TrimSuffix(saved.name, oldBuilderModelEnding)

	// The comparison follows the column's collation, as the unique key does,
	// so a file named Orders clashes with orders.yml too.
	var clashes bool
	err = tx.QueryRow(
		`SELECT EXISTS(
		   SELECT 1 FROM collimato_workspace_files
		   WHERE workspace_id = ? AND file_type = ? AND name = ? AND id <> ?
		 )`,
		saved.workspaceID, saved.fileType, plainName, saved.id).Scan(&clashes)
	if err != nil {
		return false, err
	}

	if clashes {
		_, err := tx.Exec(
			`UPDATE collimato_workspace_files SET builder_model = ? WHERE id = ?`,
			builderModel, saved.id)

		return false, err
	}

	_, err = tx.Exec(
		`UPDATE collimato_workspace_files SET name = ?, content = '', builder_model = ? WHERE id = ?`,
		plainName, builderModel, saved.id)

	return true, err
}

// withDatasetWords rewrites a saved builder model with the query engine's
// words: {DATASET} for {CUBE} in every text value, and a view's parts under
// "parts". A model that is not JSON is left as it is; the engine reports it
// as one it cannot read.
func withDatasetWords(builderModel string, fileType string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(builderModel))
	decoder.UseNumber()

	var parsed any
	if err := decoder.Decode(&parsed); err != nil {
		return builderModel, nil
	}

	parsed = replaceDatasetReference(parsed)

	fields, isObject := parsed.(map[string]any)

	isOldView := isObject && fileType == model.CollimatoFileTypeView && fields[oldViewPartsKey] != nil
	if isOldView {
		fields[newViewPartsKey] = fields[oldViewPartsKey]
		delete(fields, oldViewPartsKey)
	}

	var encoded bytes.Buffer

	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(parsed); err != nil {
		return "", err
	}

	return strings.TrimSuffix(encoded.String(), "\n"), nil
}

func replaceDatasetReference(value any) any {
	switch typed := value.(type) {
	case string:
		return strings.ReplaceAll(typed, oldDatasetReference, newDatasetReference)
	case map[string]any:
		for key, inner := range typed {
			typed[key] = replaceDatasetReference(inner)
		}

		return typed
	case []any:
		for position, inner := range typed {
			typed[position] = replaceDatasetReference(inner)
		}

		return typed
	}

	return value
}
