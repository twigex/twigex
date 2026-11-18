// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "database/sql/driver"

// DatabaseType is the kind of database a connection is to.
type DatabaseType string

const (
	DatabaseMySQL    DatabaseType = "mysql"
	DatabasePostgres DatabaseType = "postgres"
)

// database is everything the engine does differently for one kind of
// database. Each kind has four files named after it: _syntax.go writes its
// SQL, _connection.go opens its connections, _tables.go reads its column
// types and _errors.go reads its driver's errors. The rest of the engine is
// the same for every database.
type database interface {
	sqlSyntax

	// openConnector gives the driver's connector for one connection. Every
	// session it opens must read and write times in UTC, refuse to change
	// data, and be stopped by the database itself after the connector's
	// QueryTimeLimit.
	openConnector(connector *Connector, settings ConnectionSettings) (driver.Connector, error)

	// listTablesSQL lists every column of every table the connection may
	// read, as five columns: schema, table, column, column type, and "YES"
	// when the column can be empty. Rows come in each table's column order.
	listTablesSQL() string

	// suggestFieldType gives the field type for a column type as
	// listTablesSQL names it, in lower case, or empty when none fits.
	suggestFieldType(columnType string) FieldType

	// isLongNumberType says whether a column type, as the driver names it in
	// results, can hold numbers longer than the 15 to 16 digits a browser
	// keeps exactly. Fields of these types are sent as text.
	isLongNumberType(resultColumnType string) bool

	// connectionFailure names why a connection failed, from the driver's own
	// error, or gives "" when the error is not one it knows.
	connectionFailure(err error) ConnectionFailure

	// stoppedAtTimeLimit says whether the driver's error means the database
	// stopped a query because it ran past the time limit.
	stoppedAtTimeLimit(err error) bool
}

// supportedDatabases is every kind of database the engine can use.
var supportedDatabases = map[DatabaseType]database{
	DatabaseMySQL:    mysqlDatabase{},
	DatabasePostgres: postgresDatabase{},
}

func databaseFor(databaseType DatabaseType) (database, bool) {
	found, isSupported := supportedDatabases[databaseType]

	return found, isSupported
}
