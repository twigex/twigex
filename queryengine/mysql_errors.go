// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// MySQL's and MariaDB's error numbers.
const (
	mysqlAccessDenied           = 1045
	mysqlAccessToDatabaseDenied = 1044
	mysqlAccessDeniedNoPassword = 1698
	mysqlUnknownDatabase        = 1049
	mysqlUnknownVariable        = 1193
	mysqlTimeLimitReached       = 3024
	mariadbTimeLimitReached     = 1969
)

func (mysqlDatabase) connectionFailure(err error) ConnectionFailure {
	var mysqlError *mysql.MySQLError
	if !errors.As(err, &mysqlError) {
		return ""
	}

	switch mysqlError.Number {
	case mysqlAccessDenied, mysqlAccessToDatabaseDenied, mysqlAccessDeniedNoPassword:
		return ConnectionFailureCredentials
	case mysqlUnknownDatabase:
		return ConnectionFailureDatabase
	}

	return ""
}

func (mysqlDatabase) stoppedAtTimeLimit(err error) bool {
	var mysqlError *mysql.MySQLError
	if !errors.As(err, &mysqlError) {
		return false
	}

	return mysqlError.Number == mysqlTimeLimitReached || mysqlError.Number == mariadbTimeLimitReached
}

func isUnknownMySQLVariable(err error) bool {
	var mysqlError *mysql.MySQLError

	return errors.As(err, &mysqlError) && mysqlError.Number == mysqlUnknownVariable
}
