// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres's error codes.
const (
	postgresBadLogin        = "28000"
	postgresBadPassword     = "28P01"
	postgresUnknownDatabase = "3D000"
	postgresQueryCanceled   = "57014"
)

func (postgresDatabase) connectionFailure(err error) ConnectionFailure {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return ""
	}

	switch postgresError.Code {
	case postgresBadLogin, postgresBadPassword:
		return ConnectionFailureCredentials
	case postgresUnknownDatabase:
		return ConnectionFailureDatabase
	}

	return ""
}

// stoppedAtTimeLimit goes by the code alone, as the message is in the
// server's language. When the app cancels a query, the driver gives the
// context's error instead, so the code means the time limit, or the rare
// query someone cancelled on the server by hand.
func (postgresDatabase) stoppedAtTimeLimit(err error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}

	return postgresError.Code == postgresQueryCanceled
}
