// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql/driver"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func (postgresDatabase) openConnector(connector *Connector, settings ConnectionSettings) (driver.Connector, error) {
	return connector.postgresConnector(settings)
}

// postgresConfig builds the driver's settings field by field. The settings
// the driver read from this machine's environment (PGHOST, PGSSLMODE, …) are
// replaced: its own extra attempts, which can retry without encryption or on
// another host, are removed, and the session settings are the engine's.
func (connector *Connector) postgresConfig(settings ConnectionSettings) (*pgx.ConnConfig, error) {
	config, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	port, err := settings.portNumber()
	if err != nil {
		return nil, err
	}

	err = settings.checkHost()
	if err != nil {
		return nil, err
	}

	config.Host = settings.Host
	config.Port = port
	config.User = settings.User
	config.Password = settings.Password
	config.Database = settings.Database
	config.Fallbacks = nil

	// PGTARGETSESSIONATTRS can add a check that refuses read-only sessions,
	// which every engine session is.
	config.ValidateConnect = nil

	config.TLSConfig, err = tlsConfig(settings)
	if err != nil {
		return nil, err
	}

	config.RuntimeParams = map[string]string{
		"timezone":                      "UTC",
		"default_transaction_read_only": "on",
	}

	timeLimit := connector.options.QueryTimeLimit
	if timeLimit > 0 {
		config.RuntimeParams["statement_timeout"] = strconv.FormatInt(timeLimit.Milliseconds(), 10)
	}

	if connector.options.Dial != nil {
		config.DialFunc = connector.options.Dial
	}

	return config, nil
}

func (connector *Connector) postgresConnector(settings ConnectionSettings) (driver.Connector, error) {
	config, err := connector.postgresConfig(settings)
	if err != nil {
		return nil, err
	}

	resetSession := stdlib.OptionResetSession(resetPostgresSession)

	return stdlib.GetConnector(*config, resetSession), nil
}

// resetPostgresSession runs before a pooled connection is used again. SQL in
// a model can change session settings, for example
// set_config('default_transaction_read_only', 'off', false) inside a SELECT,
// and the change would stay with the connection. RESET ALL puts every
// setting back to what the connection started with: read-only, the time
// limit and UTC.
func resetPostgresSession(ctx context.Context, connection *pgx.Conn) error {
	_, err := connection.Exec(ctx, "RESET ALL")

	return err
}
