// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Tests that run queries use a fresh database made for this test run on each
// server they are given: MySQL or MariaDB from TEST_MYSQL_DSN, Postgres from
// TEST_POSTGRES_DSN (a postgres:// address). Each such test runs on every
// database it has, and skips itself when it has none, like the store tests.

// testDatabase is one database the tests made, opened with its session in
// UTC, as Twigex will open connections.
type testDatabase struct {
	name         string
	databaseType DatabaseType
	connection   *sql.DB

	// settings reach the same database through the engine's own Connector.
	settings ConnectionSettings

	// openInTimeZone opens the database again with the session a number of
	// hours east of UTC, to check that dates are still read in UTC.
	openInTimeZone func(hoursEastOfUTC int) (*sql.DB, error)

	drop func()
}

var testDatabases []testDatabase

func TestMain(m *testing.M) {
	databaseName := fmt.Sprintf("queryengine_test_%d", time.Now().UnixNano())

	if err := makeTestDatabases(databaseName); err != nil {
		fmt.Fprintf(os.Stderr, "test database setup failed: %v\n", err)
		dropTestDatabases()
		os.Exit(1)
	}

	code := m.Run()

	dropTestDatabases()
	os.Exit(code)
}

func makeTestDatabases(databaseName string) error {
	mysqlLink := os.Getenv("TEST_MYSQL_DSN")
	if mysqlLink != "" {
		database, err := makeMySQLTestDatabase(mysqlLink, databaseName)
		if err != nil {
			return fmt.Errorf("MySQL: %w", err)
		}

		testDatabases = append(testDatabases, database)
	}

	postgresLink := os.Getenv("TEST_POSTGRES_DSN")
	if postgresLink != "" {
		database, err := makePostgresTestDatabase(postgresLink, databaseName)
		if err != nil {
			return fmt.Errorf("Postgres: %w", err)
		}

		testDatabases = append(testDatabases, database)
	}

	return nil
}

func dropTestDatabases() {
	for _, database := range testDatabases {
		database.drop()
	}
}

func makeMySQLTestDatabase(rootLink string, databaseName string) (testDatabase, error) {
	settings, err := mysql.ParseDSN(rootLink)
	if err != nil {
		return testDatabase{}, err
	}

	settings.MultiStatements = true
	settings.DBName = ""
	serverLink := settings.FormatDSN()

	root, err := sql.Open("mysql", serverLink)
	if err != nil {
		return testDatabase{}, err
	}
	defer root.Close()

	quotedName := mysqlDatabase{}.quoteName(databaseName)

	if _, err := root.Exec("CREATE DATABASE " + quotedName); err != nil {
		return testDatabase{}, err
	}

	settings.DBName = databaseName

	host, port, err := net.SplitHostPort(settings.Addr)
	if err != nil {
		return testDatabase{}, err
	}

	connectionSettings := ConnectionSettings{
		DatabaseType: DatabaseMySQL,
		Host:         host,
		Port:         port,
		Database:     databaseName,
		User:         settings.User,
		Password:     settings.Passwd,
	}

	openInTimeZone := func(hoursEastOfUTC int) (*sql.DB, error) {
		zoneSettings := *settings
		zoneSettings.Params = map[string]string{"time_zone": fmt.Sprintf("'%+03d:00'", hoursEastOfUTC)}

		return sql.Open("mysql", zoneSettings.FormatDSN())
	}

	database := testDatabase{
		name:           "MySQL",
		databaseType:   DatabaseMySQL,
		settings:       connectionSettings,
		openInTimeZone: openInTimeZone,
		drop: func() {
			dropDatabase("mysql", serverLink, "DROP DATABASE "+quotedName)
		},
	}

	connection, err := openInTimeZone(0)
	if err != nil {
		return database, err
	}

	database.connection = connection

	if _, err := connection.Exec(testTablesSQL(DatabaseMySQL)); err != nil {
		return database, err
	}

	return database, nil
}

func makePostgresTestDatabase(rootLink string, databaseName string) (testDatabase, error) {
	root, err := sql.Open("pgx", rootLink)
	if err != nil {
		return testDatabase{}, err
	}
	defer root.Close()

	quotedName := postgresDatabase{}.quoteName(databaseName)

	if _, err := root.Exec("CREATE DATABASE " + quotedName); err != nil {
		return testDatabase{}, err
	}

	address, err := url.Parse(rootLink)
	if err != nil {
		return testDatabase{}, err
	}

	address.Path = "/" + databaseName

	password, _ := address.User.Password()

	connectionSettings := ConnectionSettings{
		DatabaseType: DatabasePostgres,
		Host:         address.Hostname(),
		Port:         address.Port(),
		Database:     databaseName,
		User:         address.User.Username(),
		Password:     password,
	}

	// Postgres reads "+03:00" as three hours west of UTC; Etc/GMT-3 is three
	// hours east.
	openInTimeZone := func(hoursEastOfUTC int) (*sql.DB, error) {
		zoneAddress := *address
		query := zoneAddress.Query()
		query.Set("timezone", fmt.Sprintf("Etc/GMT%+d", -hoursEastOfUTC))
		zoneAddress.RawQuery = query.Encode()

		return sql.Open("pgx", zoneAddress.String())
	}

	database := testDatabase{
		name:           "Postgres",
		databaseType:   DatabasePostgres,
		settings:       connectionSettings,
		openInTimeZone: openInTimeZone,
		drop: func() {
			dropDatabase("pgx", rootLink, "DROP DATABASE "+quotedName+" WITH (FORCE)")
		},
	}

	connection, err := openInTimeZone(0)
	if err != nil {
		return database, err
	}

	database.connection = connection

	if _, err := connection.Exec(testTablesSQL(DatabasePostgres)); err != nil {
		return database, err
	}

	// A table with the same name in another schema, which listing tables must
	// keep apart.
	if _, err := connection.Exec(archiveSchemaSQL); err != nil {
		return database, err
	}

	return database, nil
}

func dropDatabase(driverName string, rootLink string, dropSQL string) {
	root, err := sql.Open(driverName, rootLink)
	if err != nil {
		return
	}
	defer root.Close()

	if _, err := root.Exec(dropSQL); err != nil {
		fmt.Fprintf(os.Stderr, "could not drop test database: %v\n", err)
	}
}

// eachTestDatabase runs a test once on every test database, and skips it when
// there is none.
func eachTestDatabase(t *testing.T, run func(t *testing.T, database testDatabase)) {
	t.Helper()

	if len(testDatabases) == 0 {
		t.Skip("neither TEST_MYSQL_DSN nor TEST_POSTGRES_DSN is set; skipping tests that run queries")
	}

	for _, database := range testDatabases {
		t.Run(database.name, func(t *testing.T) {
			run(t, database)
		})
	}
}

// databaseResult is a request's result on one test database.
type databaseResult struct {
	database testDatabase
	result   Result
}

// runOnEachDatabase builds and runs a request on every test database, with
// its session in UTC.
func runOnEachDatabase(t *testing.T, models Models, request Request) []databaseResult {
	t.Helper()

	return runOnEachDatabaseInTimeZone(t, models, request, 0)
}

func runOnEachDatabaseInTimeZone(t *testing.T, models Models, request Request, hoursEastOfUTC int) []databaseResult {
	t.Helper()

	if len(testDatabases) == 0 {
		t.Skip("neither TEST_MYSQL_DSN nor TEST_POSTGRES_DSN is set; skipping tests that run queries")
	}

	var results []databaseResult

	for _, database := range testDatabases {
		connection := database.connection

		if hoursEastOfUTC != 0 {
			zoneConnection, err := database.openInTimeZone(hoursEastOfUTC)
			if err != nil {
				t.Fatalf("%s: open in time zone: %v", database.name, err)
			}
			defer zoneConnection.Close()

			connection = zoneConnection
		}

		result := runOn(t, database, connection, models, request)
		results = append(results, databaseResult{database: database, result: result})
	}

	return results
}

func runOn(t *testing.T, database testDatabase, connection *sql.DB, models Models, request Request) Result {
	t.Helper()

	models.DatabaseTypes = map[string]DatabaseType{"connection-1": database.databaseType}

	query, err := BuildQuery(models, request)
	if err != nil {
		t.Fatalf("%s: BuildQuery: %v", database.name, err)
	}

	result, err := RunQuery(context.Background(), connection, query)
	if err != nil {
		t.Fatalf("%s: RunQuery: %v\n%s", database.name, err, query.SQL)
	}

	return result
}

// testDatabaseTypes says the tests' connection is MySQL, for tests that only
// build SQL.
func testDatabaseTypes() map[string]DatabaseType {
	return map[string]DatabaseType{"connection-1": DatabaseMySQL}
}

func salesDataset() Dataset {
	return Dataset{
		Name:         "sales",
		Table:        "sales",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "country", Title: "Country", Type: FieldTypeText, SQL: "country"},
			{Name: "channel", Type: FieldTypeText, SQL: "channel"},
			{Name: "price", Type: FieldTypeNumber, SQL: "amount"},
			{Name: "quantity", Type: FieldTypeNumber, SQL: "quantity"},
			{Name: "is_paid", Type: FieldTypeYesNo, SQL: "is_paid"},
			{Name: "big_code", Type: FieldTypeNumber, SQL: "big_code"},
			{Name: "note", Type: FieldTypeText, SQL: "note"},
			{Name: "rating", Type: FieldTypeNumber, SQL: "rating"},
		},
		Measures: []Measure{
			{Name: "count", Title: "Sales", Type: MeasureCount},
			{Name: "countries", Type: MeasureCountDistinct, SQL: "country"},
			{Name: "revenue", Type: MeasureSum, SQL: "amount"},
			{Name: "average_sale", Type: MeasureAverage, SQL: "amount"},
			{Name: "smallest_sale", Type: MeasureMinimum, SQL: "amount"},
			{Name: "largest_sale", Type: MeasureMaximum, SQL: "amount"},
			{Name: "profit", Type: MeasureSum, SQL: "amount - cost"},
			{Name: "margin", Type: MeasureFormula, SQL: "ROUND(100 * SUM({DATASET}.amount - {DATASET}.cost) / NULLIF(SUM({DATASET}.amount), 0), 1)"},
		},
	}
}

func salesModels() Models {
	return Models{Datasets: []Dataset{salesDataset()}, DatabaseTypes: testDatabaseTypes()}
}

func eventsDataset() Dataset {
	return Dataset{
		Name:         "events",
		Table:        "events",
		ConnectionID: "connection-1",
		Dimensions: []Dimension{
			{Name: "id", Type: FieldTypeNumber, SQL: "id", IsPrimaryKey: true, IsHidden: true},
			{Name: "happened_at", Title: "Happened", Type: FieldTypeTime, SQL: "happened_at"},
		},
		Measures: []Measure{
			{Name: "count", Title: "Events", Type: MeasureCount},
		},
	}
}

func eventsModels() Models {
	return Models{Datasets: []Dataset{eventsDataset(), salesDataset()}, DatabaseTypes: testDatabaseTypes()}
}
