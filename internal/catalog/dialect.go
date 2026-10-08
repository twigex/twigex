// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

type SqlDialect struct {
	Query string
}

var mysqlDialect = SqlDialect{
	Query: `
	SELECT table_schema, table_name, column_name, data_type,
       is_nullable, ordinal_position
FROM information_schema.columns
WHERE table_schema = DATABASE()
ORDER BY table_name, ordinal_position;`,
}

var psqlDialect = SqlDialect{
	Query: `
SELECT table_schema, table_name, column_name, data_type,
       is_nullable, ordinal_position
FROM information_schema.columns
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY table_schema, table_name, ordinal_position;`,
}
