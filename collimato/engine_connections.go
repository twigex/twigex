// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

var engineDatabaseTypes = map[string]queryengine.DatabaseType{
	"mysql":    queryengine.DatabaseMySQL,
	"postgres": queryengine.DatabasePostgres,
}

// EngineDatabaseTypes gives the engine the kind of database of each saved
// connection, by connection ID. A connection of another type is left out, and
// the engine then refuses queries on it.
func EngineDatabaseTypes(connections []model.Connection) map[string]queryengine.DatabaseType {
	databaseTypes := make(map[string]queryengine.DatabaseType, len(connections))

	for _, connection := range connections {
		databaseType, known := engineDatabaseTypes[connection.Type]
		if !known {
			continue
		}

		databaseTypes[connection.ID] = databaseType
	}

	return databaseTypes
}

// EngineConnectionSettings gives the engine what it needs to reach a saved
// connection. The password is passed separately because the saved one is
// encrypted. A type the engine does not know is passed on as it is, and the
// engine refuses it when opening.
func EngineConnectionSettings(connection model.Connection, password string) queryengine.ConnectionSettings {
	databaseType, known := engineDatabaseTypes[connection.Type]
	if !known {
		databaseType = queryengine.DatabaseType(connection.Type)
	}

	settings := queryengine.ConnectionSettings{
		DatabaseType:  databaseType,
		Host:          connection.Host,
		Port:          connection.Port,
		Database:      connection.Database,
		User:          connection.Username,
		Password:      password,
		SSLMode:       queryengine.SSLMode(connection.SSLMode),
		CACertificate: connection.Config[model.ConnectionCACertKey],
	}

	return settings
}

// BuilderTableName is how a dataset names a table. A MySQL connection is to
// one database, so the name alone. In Postgres a table in the default schema,
// public, is named alone too, and one in another schema with its schema in
// front: archive.regions.
func BuilderTableName(connectionType string, table queryengine.Table) string {
	isOtherPostgresSchema := connectionType == "postgres" && table.Schema != "public"
	if isOtherPostgresSchema {
		return table.Schema + "." + table.Name
	}

	return table.Name
}

// DatasetNameForTable is the name a dataset made from a table gets. A table
// in another Postgres schema has a dot in its name, which a dataset name
// cannot have: archive.regions becomes archive_regions.
func DatasetNameForTable(tableName string) string {
	return strings.ReplaceAll(tableName, ".", "_")
}

var builderFieldTypes = map[queryengine.FieldType]string{
	queryengine.FieldTypeText:   "string",
	queryengine.FieldTypeNumber: "number",
	queryengine.FieldTypeTime:   "time",
	queryengine.FieldTypeYesNo:  "boolean",
}

// SuggestBuilderType suggests the builder's type for a database column, or
// empty when the column cannot be a field.
func SuggestBuilderType(connectionType string, columnType string) string {
	databaseType := engineDatabaseTypes[connectionType]
	fieldType := queryengine.SuggestFieldType(databaseType, columnType)

	return builderFieldTypes[fieldType]
}
