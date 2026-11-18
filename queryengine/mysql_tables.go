// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

// listTablesSQL lists the tables of the connection's own database, with the
// full column type, because only that tells tinyint(1), used for yes/no, from
// a small number.
func (mysqlDatabase) listTablesSQL() string {
	return `SELECT table_schema, table_name, column_name, column_type, is_nullable
FROM information_schema.columns
WHERE table_schema = DATABASE()
ORDER BY table_name, ordinal_position`
}

// suggestFieldType reads the type's name, without its size and words such as
// "unsigned": "int(11) unsigned" is an int. Only tinyint(1) needs its size,
// because MySQL uses it for yes/no.
func (mysqlDatabase) suggestFieldType(columnType string) FieldType {
	if strings.HasPrefix(columnType, "tinyint(1)") {
		return FieldTypeYesNo
	}

	name, _, _ := strings.Cut(columnType, "(")
	name, _, _ = strings.Cut(name, " ")

	return mysqlFieldTypes[name]
}

var mysqlFieldTypes = map[string]FieldType{
	"bool":       FieldTypeYesNo,
	"boolean":    FieldTypeYesNo,
	"tinyint":    FieldTypeNumber,
	"smallint":   FieldTypeNumber,
	"mediumint":  FieldTypeNumber,
	"int":        FieldTypeNumber,
	"integer":    FieldTypeNumber,
	"bigint":     FieldTypeNumber,
	"decimal":    FieldTypeNumber,
	"numeric":    FieldTypeNumber,
	"float":      FieldTypeNumber,
	"double":     FieldTypeNumber,
	"real":       FieldTypeNumber,
	"year":       FieldTypeNumber,
	"char":       FieldTypeText,
	"varchar":    FieldTypeText,
	"tinytext":   FieldTypeText,
	"text":       FieldTypeText,
	"mediumtext": FieldTypeText,
	"longtext":   FieldTypeText,
	"enum":       FieldTypeText,
	"set":        FieldTypeText,
	"time":       FieldTypeText,
	"date":       FieldTypeTime,
	"datetime":   FieldTypeTime,
	"timestamp":  FieldTypeTime,
}

// isLongNumberType: BIGINT and DECIMAL can be longer than a browser keeps.
func (mysqlDatabase) isLongNumberType(resultColumnType string) bool {
	return resultColumnType == "BIGINT" || resultColumnType == "DECIMAL"
}
