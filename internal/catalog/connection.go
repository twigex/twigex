// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type Connection struct {
	Type     string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	Config   map[string]string
}

func (c *Connection) mode() string {
	if c.SSLMode == "" {
		return SSLModeDisable
	}

	return c.SSLMode
}

func (c *Connection) Open() (*sql.DB, error) {
	switch c.Type {
	case Mysql:
		return c.openMysql()
	case Postgre:
		return c.openPostgres()
	default:
		return nil, fmt.Errorf("catalog: unsupported store %s", c.Type)
	}
}

func (c *Connection) mysqlConfig() (*mysql.Config, error) {
	tlsParam, err := registerMysqlTLS(c)
	if err != nil {
		return nil, err
	}

	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.Host, c.Port)
	cfg.User = c.User
	cfg.Passwd = c.Password
	cfg.DBName = c.Database
	cfg.TLSConfig = tlsParam

	return cfg, nil
}

func (c *Connection) openMysql() (*sql.DB, error) {
	cfg, err := c.mysqlConfig()
	if err != nil {
		return nil, err
	}

	return sql.Open(Mysql, cfg.FormatDSN())
}

func (c *Connection) postgresConfig() (*pgx.ConnConfig, error) {
	cfg, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(c.Port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("catalog: bad port %s", c.Port)
	}

	cfg.Host = c.Host
	cfg.Port = uint16(port)
	cfg.User = c.User
	cfg.Password = c.Password
	cfg.Database = c.Database

	cfg.TLSConfig, err = c.tlsConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Connection) openPostgres() (*sql.DB, error) {
	cfg, err := c.postgresConfig()
	if err != nil {
		return nil, err
	}

	return sql.OpenDB(stdlib.GetConnector(*cfg)), nil
}
