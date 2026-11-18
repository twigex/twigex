// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"time"
)

// ConnectorOptions are the host app's choices for every connection.
type ConnectorOptions struct {
	// Dial, when set, makes every network connection to a database. The
	// engine decides nothing about which addresses may be reached; a host app
	// that does checks the address here, when it is actually connected to.
	Dial func(ctx context.Context, network string, address string) (net.Conn, error)

	// QueryTimeLimit is the longest a query may run. The database itself
	// stops it, so it stops even if the app does not. Zero means no limit.
	QueryTimeLimit time.Duration

	// MaxOpenConnections caps how many connections one database's pool
	// keeps open. Zero means defaultMaxOpenConnections.
	MaxOpenConnections int
}

const (
	defaultMaxOpenConnections = 4
	maxIdleConnections        = 2
	idleConnectionLifetime    = 5 * time.Minute
	connectionLifetime        = 30 * time.Minute
)

// Connector opens connections to customer databases. The host app makes one
// with its options and keeps it.
type Connector struct {
	options ConnectorOptions

	// mysqlNetwork is the name this connector's dial function is given to the
	// MySQL driver under, which is how this driver version takes one.
	mysqlNetwork string
}

func NewConnector(options ConnectorOptions) *Connector {
	connector := &Connector{
		options:      options,
		mysqlNetwork: registerMySQLDial(options.Dial),
	}

	return connector
}

// Open gives a pool of connections to a database. Every session in it reads
// times in UTC, cannot write, and is stopped by the database at the query
// time limit. Open does not reach the database; the first query or Check
// does.
func (connector *Connector) Open(settings ConnectionSettings) (*sql.DB, error) {
	if _, err := settings.portNumber(); err != nil {
		return nil, err
	}

	sessionConnector, err := connector.sessionConnector(settings)
	if err != nil {
		return nil, err
	}

	database := sql.OpenDB(sessionConnector)

	maxOpen := connector.options.MaxOpenConnections
	if maxOpen <= 0 {
		maxOpen = defaultMaxOpenConnections
	}

	database.SetMaxOpenConns(maxOpen)
	database.SetMaxIdleConns(min(maxIdleConnections, maxOpen))
	database.SetConnMaxIdleTime(idleConnectionLifetime)
	database.SetConnMaxLifetime(connectionLifetime)

	return database, nil
}

// Check opens a connection and reaches the database with it, for "Test
// connection". ConnectionFailureOf names the reason when it fails.
func (connector *Connector) Check(ctx context.Context, settings ConnectionSettings) error {
	database, err := connector.Open(settings)
	if err != nil {
		return err
	}
	defer database.Close()

	return database.PingContext(ctx)
}

func (connector *Connector) sessionConnector(settings ConnectionSettings) (driver.Connector, error) {
	database, isSupported := databaseFor(settings.DatabaseType)
	if !isSupported {
		return nil, fmt.Errorf("query engine: unknown database type %q", settings.DatabaseType)
	}

	return database.openConnector(connector, settings)
}
