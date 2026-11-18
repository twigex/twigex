// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

// listTablesSQL lists every schema the user can see, except Postgres's own
// system schemas.
func (postgresDatabase) listTablesSQL() string {
	return `SELECT table_schema, table_name, column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_schema NOT IN ('pg_catalog', 'information_schema') AND table_schema NOT LIKE 'pg\_toast%'
ORDER BY table_schema, table_name, ordinal_position`
}

func (postgresDatabase) suggestFieldType(columnType string) FieldType {
	return postgresFieldTypes[columnType]
}

var postgresFieldTypes = map[string]FieldType{
	"boolean":                     FieldTypeYesNo,
	"smallint":                    FieldTypeNumber,
	"integer":                     FieldTypeNumber,
	"bigint":                      FieldTypeNumber,
	"numeric":                     FieldTypeNumber,
	"real":                        FieldTypeNumber,
	"double precision":            FieldTypeNumber,
	"character varying":           FieldTypeText,
	"character":                   FieldTypeText,
	"text":                        FieldTypeText,
	"uuid":                        FieldTypeText,
	"name":                        FieldTypeText,
	"time without time zone":      FieldTypeText,
	"time with time zone":         FieldTypeText,
	"date":                        FieldTypeTime,
	"timestamp without time zone": FieldTypeTime,
	"timestamp with time zone":    FieldTypeTime,
}

// isLongNumberType: INT8 (bigint) and NUMERIC can be longer than a browser
// keeps.
func (postgresDatabase) isLongNumberType(resultColumnType string) bool {
	return resultColumnType == "INT8" || resultColumnType == "NUMERIC"
}
