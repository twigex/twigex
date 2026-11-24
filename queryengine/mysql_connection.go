// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
)

func (mysqlDatabase) openConnector(connector *Connector, settings ConnectionSettings) (driver.Connector, error) {
	return connector.mysqlConnector(settings)
}

var mysqlDialCount atomic.Int64

// registerMySQLDial gives the MySQL driver the host app's dial function under
// a network name of its own, which is how this driver version takes one, and
// returns that name. Without a dial function the driver dials itself, "tcp".
func registerMySQLDial(dial func(ctx context.Context, network string, address string) (net.Conn, error)) string {
	if dial == nil {
		return "tcp"
	}

	networkName := fmt.Sprintf("queryengine-%d", mysqlDialCount.Add(1))

	mysql.RegisterDialContext(networkName, func(ctx context.Context, address string) (net.Conn, error) {
		return dial(ctx, "tcp", address)
	})

	return networkName
}

// mysqlConfig builds the driver's settings field by field, never from a
// joined address, so nothing in a name or password can change them.
// Several statements at once and loading local files stay off.
func (connector *Connector) mysqlConfig(settings ConnectionSettings) (*mysql.Config, error) {
	tlsName, err := registerMySQLTLS(settings)
	if err != nil {
		return nil, err
	}

	config := mysql.NewConfig()
	config.Net = connector.mysqlNetwork
	config.Addr = net.JoinHostPort(settings.Host, settings.Port)
	config.User = settings.User
	config.Passwd = settings.Password
	config.DBName = settings.Database
	config.TLSConfig = tlsName

	return config, nil
}

func (connector *Connector) mysqlConnector(settings ConnectionSettings) (driver.Connector, error) {
	config, err := connector.mysqlConfig(settings)
	if err != nil {
		return nil, err
	}

	driverConnector, err := mysql.NewConnector(config)
	if err != nil {
		releaseMySQLTLS(config.TLSConfig)
		return nil, err
	}

	sessionConnector := mysqlSessionConnector{
		connector: driverConnector,
		timeLimit: connector.options.QueryTimeLimit,
		tlsName:   config.TLSConfig,
	}

	return sessionConnector, nil
}

// mysqlSessionConnector sets up every new MySQL session before it is used.
// The driver can only set plain variables itself, and the read-only setting
// and the time limit are not plain variables on every server.
type mysqlSessionConnector struct {
	connector driver.Connector
	timeLimit time.Duration
	tlsName   string
}

// Close is called when the pool is closed, and gives up the pool's use of
// its encryption settings.
func (sessionConnector mysqlSessionConnector) Close() error {
	releaseMySQLTLS(sessionConnector.tlsName)

	return nil
}

func (sessionConnector mysqlSessionConnector) Connect(ctx context.Context) (driver.Conn, error) {
	connection, err := sessionConnector.connector.Connect(ctx)
	if err != nil {
		return nil, err
	}

	if err := setUpMySQLSession(ctx, connection, sessionConnector.timeLimit); err != nil {
		connection.Close()
		return nil, err
	}

	return connection, nil
}

func (sessionConnector mysqlSessionConnector) Driver() driver.Driver {
	return sessionConnector.connector.Driver()
}

func setUpMySQLSession(ctx context.Context, connection driver.Conn, timeLimit time.Duration) error {
	executor, canExecute := connection.(driver.ExecerContext)
	if !canExecute {
		return errors.New("query engine: the MySQL driver cannot run statements")
	}

	run := func(statement string) error {
		_, err := executor.ExecContext(ctx, statement, nil)
		return err
	}

	if err := run("SET time_zone = '+00:00'"); err != nil {
		return err
	}

	if err := run("SET SESSION TRANSACTION READ ONLY"); err != nil {
		return err
	}

	if timeLimit <= 0 {
		return nil
	}

	// MySQL calls the limit max_execution_time, in milliseconds; MariaDB
	// does not have it and calls its own max_statement_time, in seconds. A
	// server older than both (MySQL 5.7.8, MariaDB 10.1) has neither; its
	// queries are then stopped only by the app's own deadline.
	err := run(fmt.Sprintf("SET SESSION max_execution_time = %d", timeLimit.Milliseconds()))
	if !isUnknownMySQLVariable(err) {
		return err
	}

	err = run(fmt.Sprintf("SET SESSION max_statement_time = %.3f", timeLimit.Seconds()))
	if isUnknownMySQLVariable(err) {
		return nil
	}

	return err
}

// mysqlTLSUsers counts the open pools using each name registerMySQLTLS gave,
// so that a name is removed from the driver when the last of them closes.
var mysqlTLSUsers = struct {
	mutex sync.Mutex
	count map[string]int
}{count: make(map[string]int)}

// registerMySQLTLS gives the MySQL driver the encryption under a name, which
// is how this driver version takes it. The name comes from the settings, so
// the same settings reuse one entry and a changed certificate gets its own.
// Each name given must be released with releaseMySQLTLS.
func registerMySQLTLS(settings ConnectionSettings) (string, error) {
	config, err := tlsConfig(settings)
	if err != nil {
		return "", err
	}

	if config == nil {
		return "false", nil
	}

	encryptionSettings := []string{string(settings.sslMode()), settings.Host, settings.CACertificate}
	sum := sha256.Sum256([]byte(strings.Join(encryptionSettings, "\x00")))
	name := "queryengine-" + hex.EncodeToString(sum[:8])

	mysqlTLSUsers.mutex.Lock()
	defer mysqlTLSUsers.mutex.Unlock()

	if mysqlTLSUsers.count[name] == 0 {
		if err := mysql.RegisterTLSConfig(name, config); err != nil {
			return "", err
		}
	}

	mysqlTLSUsers.count[name]++

	return name, nil
}

// releaseMySQLTLS gives up one use of a name registerMySQLTLS gave, and
// removes it from the driver when nothing uses it any more.
func releaseMySQLTLS(name string) {
	mysqlTLSUsers.mutex.Lock()
	defer mysqlTLSUsers.mutex.Unlock()

	if mysqlTLSUsers.count[name] == 0 {
		return
	}

	mysqlTLSUsers.count[name]--

	if mysqlTLSUsers.count[name] == 0 {
		delete(mysqlTLSUsers.count, name)
		mysql.DeregisterTLSConfig(name)
	}
}
