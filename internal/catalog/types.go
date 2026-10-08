// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import "database/sql"

var (
	Mysql   = "mysql"
	Postgre = "postgres"
)

const (
	SSLModeDisable    = "disable"
	SSLModeRequire    = "require"
	SSLModeVerifyCA   = "verify-ca"
	SSLModeVerifyFull = "verify-full"
)

type ColumnRow struct {
	Schema   string
	Table    string
	Column   string
	DataType string
	Nullable sql.NullString
	Ordinal  int
}

type TableName struct {
	Name    string         `json:"name"`
	Columns []ColumnDetail `json:"columns"`
}

type ColumnDetail struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
