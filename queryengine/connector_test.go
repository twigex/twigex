// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Passwords that would change a joined address if they were pasted into one.
var awkwardPasswords = []string{
	"p@ss/word",
	"with space",
	"quote'd",
	"back\\slash",
	"tcp(evil:3306)/other",
	"x host=evil.example.com",
}

func TestConnectorMySQLKeepsTheCredentialsItWasGiven(t *testing.T) {
	connector := NewConnector(ConnectorOptions{})

	for _, password := range awkwardPasswords {
		settings := mysqlSettings()
		settings.Password = password

		config, err := connector.mysqlConfig(settings)
		if err != nil {
			t.Errorf("password %q: %v", password, err)
			continue
		}

		// What the driver writes, it must read back the same.
		parsed, err := mysql.ParseDSN(config.FormatDSN())
		if err != nil {
			t.Errorf("password %q: unreadable settings: %v", password, err)
			continue
		}

		if parsed.Passwd != password {
			t.Errorf("password %q came back as %q", password, parsed.Passwd)
		}

		if parsed.Addr != "db:3306" {
			t.Errorf("password %q moved the address to %q", password, parsed.Addr)
		}

		if parsed.DBName != "shop" {
			t.Errorf("password %q changed the database to %q", password, parsed.DBName)
		}

		if parsed.User != "u" {
			t.Errorf("password %q changed the user to %q", password, parsed.User)
		}

		if parsed.MultiStatements || parsed.AllowAllFiles {
			t.Errorf("password %q turned on several statements or local files", password)
		}
	}
}

func TestConnectorPostgresKeepsTheCredentialsItWasGiven(t *testing.T) {
	connector := NewConnector(ConnectorOptions{})

	for _, password := range awkwardPasswords {
		settings := postgresSettings()
		settings.Password = password

		config, err := connector.postgresConfig(settings)
		if err != nil {
			t.Errorf("password %q: %v", password, err)
			continue
		}

		if config.Password != password {
			t.Errorf("password %q came back as %q", password, config.Password)
		}

		if config.Host != "db" || config.Port != 5432 {
			t.Errorf("password %q moved the address to %s:%d", password, config.Host, config.Port)
		}

		if config.Database != "shop" {
			t.Errorf("password %q changed the database to %q", password, config.Database)
		}
	}
}

// The Postgres driver reads PGSSLMODE, PGHOST and so on from the machine,
// and for some values adds extra attempts: without encryption, or to
// localhost. None of them may stay.
func TestConnectorPostgresIgnoresTheMachinesSettings(t *testing.T) {
	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGSSLMODE", "prefer")
	t.Setenv("PGOPTIONS", "-c default_transaction_read_only=off")
	t.Setenv("PGTARGETSESSIONATTRS", "read-write")

	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyFull

	config, err := NewConnector(ConnectorOptions{}).postgresConfig(settings)
	if err != nil {
		t.Fatalf("postgresConfig: %v", err)
	}

	if len(config.Fallbacks) != 0 {
		t.Errorf("want no extra attempts, got %d", len(config.Fallbacks))
	}

	if config.Host != "db" {
		t.Errorf("want host db, got %q", config.Host)
	}

	if config.TLSConfig == nil || config.TLSConfig.ServerName != "db" {
		t.Error("want verify-full encryption for db")
	}

	if config.RuntimeParams["default_transaction_read_only"] != "on" {
		t.Errorf("want read-only sessions, got settings %v", config.RuntimeParams)
	}

	if _, hasOptions := config.RuntimeParams["options"]; hasOptions {
		t.Error("PGOPTIONS from the machine must not reach the session")
	}

	if config.ValidateConnect != nil {
		t.Error("PGTARGETSESSIONATTRS must not add a check that refuses read-only sessions")
	}
}

func TestConnectorPostgresSessionSettings(t *testing.T) {
	connector := NewConnector(ConnectorOptions{QueryTimeLimit: 90 * time.Second})

	config, err := connector.postgresConfig(postgresSettings())
	if err != nil {
		t.Fatalf("postgresConfig: %v", err)
	}

	want := map[string]string{
		"timezone":                      "UTC",
		"default_transaction_read_only": "on",
		"statement_timeout":             "90000",
	}

	for name, value := range want {
		if config.RuntimeParams[name] != value {
			t.Errorf("%s: want %q, got %q", name, value, config.RuntimeParams[name])
		}
	}
}

func TestConnectorMySQLCarriesTheEncryptionNameIntoItsSettings(t *testing.T) {
	settings := mysqlSettings()
	settings.SSLMode = SSLModeVerifyFull

	config, err := NewConnector(ConnectorOptions{}).mysqlConfig(settings)
	if err != nil {
		t.Fatalf("mysqlConfig: %v", err)
	}

	parsed, err := mysql.ParseDSN(config.FormatDSN())
	if err != nil {
		t.Fatalf("unreadable settings: %v", err)
	}

	if parsed.TLSConfig == "" || parsed.TLSConfig == "false" {
		t.Errorf("want named encryption settings, got %q", parsed.TLSConfig)
	}
}

func TestConnectorRefusesABadPortAndAnUnknownDatabaseType(t *testing.T) {
	connector := NewConnector(ConnectorOptions{})

	badPort := postgresSettings()
	badPort.Port = "not a port"

	if _, err := connector.Open(badPort); err == nil {
		t.Error("want an error rather than a connection to port 0")
	}

	unknownType := postgresSettings()
	unknownType.DatabaseType = "bigquery"

	if _, err := connector.Open(unknownType); err == nil {
		t.Error("want an error for an unknown database type")
	}
}

func TestConnectorRefusesAPostgresSocketPath(t *testing.T) {
	settings := postgresSettings()
	settings.Host = "/var/run/postgresql"

	if _, err := NewConnector(ConnectorOptions{}).Open(settings); err == nil {
		t.Error("want an error rather than a connection to a socket on this machine")
	}
}

// The following run on the test databases.

func openWithConnector(t *testing.T, connector *Connector, settings ConnectionSettings) *sql.DB {
	t.Helper()

	connection, err := connector.Open(settings)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() {
		connection.Close()
	})

	return connection
}

func TestConnectorSessionsAreInUTC(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		connection := openWithConnector(t, NewConnector(ConnectorOptions{}), database.settings)

		question := "SELECT @@session.time_zone"
		want := "+00:00"

		if database.databaseType == DatabasePostgres {
			question = "SHOW timezone"
			want = "UTC"
		}

		var got string
		if err := connection.QueryRow(question).Scan(&got); err != nil {
			t.Fatalf("ask the time zone: %v", err)
		}

		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})
}

func TestConnectorSessionsCannotWrite(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		connection := openWithConnector(t, NewConnector(ConnectorOptions{}), database.settings)

		statements := []string{
			"INSERT INTO regions (id, name) VALUES (99, 'Written')",
			"UPDATE regions SET name = 'Changed'",
			"CREATE TABLE written_by_engine (id INT)",
		}

		for _, statement := range statements {
			if _, err := connection.Exec(statement); err == nil {
				t.Errorf("want %q refused", statement)
			}
		}

		var count int
		if err := connection.QueryRow("SELECT COUNT(*) FROM regions WHERE name IN ('Written', 'Changed')").Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}

		if count != 0 {
			t.Errorf("want nothing written, found %d rows", count)
		}
	})
}

// MariaDB and Postgres stop a query with an error at the limit. MySQL ends
// SLEEP early instead, returning 1. Either way it must not run its whole
// second.
func TestConnectorStopsQueriesAtTheTimeLimit(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		connector := NewConnector(ConnectorOptions{QueryTimeLimit: 200 * time.Millisecond})
		connection := openWithConnector(t, connector, database.settings)

		sleep := "SELECT SLEEP(1)"
		if database.databaseType == DatabasePostgres {
			sleep = "SELECT pg_sleep(1)"
		}

		started := time.Now()
		rows, err := connection.Query(sleep)
		if err == nil {
			for rows.Next() {
			}

			err = rows.Err()
			rows.Close()
		}

		took := time.Since(started)

		if took > 800*time.Millisecond {
			t.Errorf("want the query stopped at about 200 ms, it took %v (error %v)", took, err)
		}

		// MySQL ends SLEEP early without an error; the others report one.
		if err != nil && !QueryReachedTimeLimit(err) {
			t.Errorf("the error is not recognised as the time limit: %v", err)
		}
	})
}

// A query the app stops is not one the database stopped at the time limit,
// though Postgres uses one code for both: the driver gives the context's
// error for the app's.
func TestConnectorQueryStoppedByTheAppIsNotTheTimeLimit(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		sleep := "SELECT SLEEP(2)"
		if database.databaseType == DatabasePostgres {
			sleep = "SELECT pg_sleep(2)"
		}

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		rows, err := database.connection.QueryContext(ctx, sleep)
		if err == nil {
			for rows.Next() {
			}

			err = rows.Err()
			rows.Close()
		}

		if err == nil {
			t.Fatal("want the stopped query to give an error")
		}

		if QueryReachedTimeLimit(err) {
			t.Errorf("the app's own stop is taken for the time limit: %v", err)
		}
	})
}

func TestConnectorUsesTheGivenDial(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		var dialed []string
		var lock sync.Mutex

		dial := func(ctx context.Context, network string, address string) (net.Conn, error) {
			lock.Lock()
			dialed = append(dialed, address)
			lock.Unlock()

			var dialer net.Dialer

			return dialer.DialContext(ctx, network, address)
		}

		connector := NewConnector(ConnectorOptions{Dial: dial})

		if err := connector.Check(context.Background(), database.settings); err != nil {
			t.Fatalf("Check: %v", err)
		}

		lock.Lock()
		defer lock.Unlock()

		if len(dialed) == 0 {
			t.Fatal("want the connection made through the given dial")
		}

		_, port, _ := net.SplitHostPort(dialed[0])
		if port != database.settings.Port {
			t.Errorf("want port %s dialed, got %q", database.settings.Port, dialed[0])
		}
	})
}

var errRefusedByHost = errors.New("this address is not allowed")

func TestConnectorADialThatRefusesStopsTheConnection(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		refuse := func(ctx context.Context, network string, address string) (net.Conn, error) {
			return nil, errRefusedByHost
		}

		connector := NewConnector(ConnectorOptions{Dial: refuse})

		err := connector.Check(context.Background(), database.settings)
		if err == nil {
			t.Fatal("want the connection refused")
		}

		// The MySQL driver keeps only the message of a dial error.
		keepsReason := errors.Is(err, errRefusedByHost) || strings.Contains(err.Error(), errRefusedByHost.Error())
		if !keepsReason {
			t.Errorf("want the host's reason in the error, got %v", err)
		}
	})
}

func TestConnectorNamesWrongPasswordsAndDatabases(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		connector := NewConnector(ConnectorOptions{})

		wrongPassword := database.settings
		wrongPassword.Password = "not the password"

		// A server set to trust local connections accepts any password, and
		// there is nothing to check.
		err := connector.Check(context.Background(), wrongPassword)

		switch got := ConnectionFailureOf(err); {
		case err == nil:
			t.Log("the server accepts any password from here; not checked")
		case got != ConnectionFailureCredentials:
			t.Errorf("wrong password: want %q, got %q: %v", ConnectionFailureCredentials, got, err)
		}

		wrongDatabase := database.settings
		wrongDatabase.Database = "no_such_database_here"

		err = connector.Check(context.Background(), wrongDatabase)
		if got := ConnectionFailureOf(err); got != ConnectionFailureDatabase {
			t.Errorf("wrong database: want %q, got %q: %v", ConnectionFailureDatabase, got, err)
		}
	})
}

// SQL in a model can change Postgres session settings inside a SELECT. The
// next query on the same pooled connection must have them back.
func TestConnectorPutsPostgresSessionSettingsBackBeforeReuse(t *testing.T) {
	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		if database.databaseType != DatabasePostgres {
			t.Skip("a MySQL SELECT cannot change session settings")
		}

		connector := NewConnector(ConnectorOptions{QueryTimeLimit: time.Minute, MaxOpenConnections: 1})
		connection := openWithConnector(t, connector, database.settings)

		changeSettings := "SELECT set_config('default_transaction_read_only', 'off', false), " +
			"set_config('statement_timeout', '0', false), set_config('timezone', 'Europe/Riga', false)"

		if _, err := connection.Exec(changeSettings); err != nil {
			t.Fatalf("change settings: %v", err)
		}

		if _, err := connection.Exec("CREATE TABLE written_after_reset (id INT)"); err == nil {
			t.Error("want writing refused on the reused connection")
		}

		var readOnly, timeLimit, timeZone string

		err := connection.QueryRow("SELECT current_setting('default_transaction_read_only'), "+
			"current_setting('statement_timeout'), current_setting('timezone')").Scan(&readOnly, &timeLimit, &timeZone)
		if err != nil {
			t.Fatalf("read settings: %v", err)
		}

		if readOnly != "on" || timeLimit != "1min" || timeZone != "UTC" {
			t.Errorf("want on, 1min, UTC; got %s, %s, %s", readOnly, timeLimit, timeZone)
		}
	})
}
