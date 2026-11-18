// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
	"github.com/twigex/twigex/tlog"
)

// QueryEngine is everything Twigex needs to answer charts. Tests replace it.
type QueryEngine interface {
	FieldList(ctx context.Context, workspaceID string) (*collimato.FieldList, error)
	SQL(ctx context.Context, workspaceID string, query model.DataQuery) (string, error)
	Query(ctx context.Context, workspaceID string, query model.DataQuery) (*collimato.Answer, error)
}

// engineAdapter answers with the query engine. It reads the workspace's
// builder models and connections for every request, so a saved change is used
// by the next one: nothing is compiled, so there is nothing to rebuild.
type engineAdapter struct {
	store           engineAdapterStore
	decryptPassword func(ciphertext string) (string, error)
	pools           *connectionPools
}

// engineAdapterStore is what the engine reads from Twigex's database.
type engineAdapterStore interface {
	GetWorkspaceFiles(workspaceID, fileType string) ([]model.CollimatoWorkspaceFile, error)
	GetConnections(workspace string) ([]model.Connection, error)
}

// engineWorkspace is a workspace's models as the engine reads them.
type engineWorkspace struct {
	models      queryengine.Models
	unsupported []collimato.UnsupportedFile
	connections []model.Connection
}

func newEngineAdapter(store engineAdapterStore, decryptPassword func(string) (string, error), connector *queryengine.Connector) *engineAdapter {
	return &engineAdapter{
		store:           store,
		decryptPassword: decryptPassword,
		pools:           newConnectionPools(connector),
	}
}

func (engine *engineAdapter) readWorkspace(workspaceID string) (engineWorkspace, error) {
	var files []model.CollimatoWorkspaceFile

	for _, fileType := range []string{model.CollimatoFileTypeDataset, model.CollimatoFileTypeView} {
		filesOfType, err := engine.store.GetWorkspaceFiles(workspaceID, fileType)
		if err != nil {
			return engineWorkspace{}, err
		}

		files = append(files, filesOfType...)
	}

	listedAt := time.Now()

	connections, err := engine.store.GetConnections(workspaceID)
	if err != nil {
		return engineWorkspace{}, err
	}

	engine.pools.closeDeleted(workspaceID, connections, listedAt)

	models, unsupported := collimato.EngineModels(files)
	models.DatabaseTypes = collimato.EngineDatabaseTypes(connections)

	workspace := engineWorkspace{
		models:      models,
		unsupported: unsupported,
		connections: connections,
	}

	return workspace, nil
}

func (engine *engineAdapter) FieldList(_ context.Context, workspaceID string) (*collimato.FieldList, error) {
	workspace, err := engine.readWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	engineFieldList := queryengine.ListFields(workspace.models)

	fieldList := collimato.FieldListFor(workspace.models, engineFieldList)
	fieldList.ModelErrors = collimato.ModelErrors(workspace.unsupported, engineFieldList)

	return &fieldList, nil
}

func (engine *engineAdapter) SQL(_ context.Context, workspaceID string, query model.DataQuery) (string, error) {
	workspace, err := engine.readWorkspace(workspaceID)
	if err != nil {
		return "", err
	}

	builtQuery, err := workspace.buildQuery(query)
	if err != nil {
		return "", err
	}

	return builtQuery.SQL, nil
}

func (engine *engineAdapter) Query(ctx context.Context, workspaceID string, query model.DataQuery) (*collimato.Answer, error) {
	workspace, err := engine.readWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}

	builtQuery, err := workspace.buildQuery(query)
	if err != nil {
		return nil, err
	}

	connection, found := workspace.connection(builtQuery.ConnectionID)
	if !found {
		return nil, newQueryError(queryErrorConnectionMissing, "")
	}

	database, err := engine.openConnection(connection)
	if err != nil {
		return nil, databaseError(workspaceID, err)
	}

	result, err := queryengine.RunQuery(ctx, database, builtQuery)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, databaseError(workspaceID, err)
	}

	answer := collimato.AnswerFor(result, workspace.models)

	return &answer, nil
}

// buildQuery turns a chart's request into SQL. A request the engine cannot
// answer gives a QueryError saying why.
func (workspace engineWorkspace) buildQuery(query model.DataQuery) (queryengine.BuiltQuery, error) {
	request, err := collimato.EngineRequest(query)
	if err != nil {
		return queryengine.BuiltQuery{}, newQueryError(queryErrorInvalidRequest, err.Error())
	}

	builtQuery, err := queryengine.BuildQuery(workspace.models, request)
	if err != nil {
		return queryengine.BuiltQuery{}, workspace.requestError(err)
	}

	return builtQuery, nil
}

func (workspace engineWorkspace) connection(connectionID string) (model.Connection, bool) {
	for _, connection := range workspace.connections {
		if connection.ID == connectionID {
			return connection, true
		}
	}

	return model.Connection{}, false
}

func (engine *engineAdapter) openConnection(connection model.Connection) (*sql.DB, error) {
	password, err := engine.decryptPassword(connection.Password)
	if err != nil {
		return nil, errors.New("the connection's password cannot be decrypted")
	}

	return engine.pools.open(connection, password)
}

// connectionPools keeps one pool of database connections per saved
// connection, so queries reuse open connections. A connection whose settings
// changed since its pool was opened gets a new pool with the new settings,
// and a deleted connection's pool is closed when its workspace is next read.
// A pool nobody uses holds no open connections after a while
// (Connector.Open), so pools are not closed otherwise.
type connectionPools struct {
	connector *queryengine.Connector

	mutex sync.Mutex
	pools map[string]connectionPool
}

type connectionPool struct {
	database    *sql.DB
	workspaceID string
	openedAt    time.Time

	// settings are the ones the pool was opened with. They are compared, not
	// the save time, which is in whole seconds and misses quick edits.
	settings queryengine.ConnectionSettings
}

func newConnectionPools(connector *queryengine.Connector) *connectionPools {
	return &connectionPools{
		connector: connector,
		pools:     make(map[string]connectionPool),
	}
}

func (pools *connectionPools) open(connection model.Connection, password string) (*sql.DB, error) {
	pools.mutex.Lock()
	defer pools.mutex.Unlock()

	settings := collimato.EngineConnectionSettings(connection, password)

	pool, found := pools.pools[connection.ID]
	if found && pool.settings == settings {
		return pool.database, nil
	}

	database, err := pools.connector.Open(settings)
	if err != nil {
		return nil, err
	}

	if found {
		// Closing waits for queries still running on the old pool, so it
		// happens without holding up other queries.
		go closeOldPool(connection.ID, pool.database)
	}

	pools.pools[connection.ID] = connectionPool{database: database, workspaceID: connection.WorkspaceID, openedAt: time.Now(), settings: settings}

	return database, nil
}

// closeDeleted closes the pools of a workspace's connections that are no
// longer saved. Comparing with what is saved also catches deletes made on
// another Twigex server. A pool opened after the list was read may be newer
// than the list, so it is kept.
func (pools *connectionPools) closeDeleted(workspaceID string, saved []model.Connection, listedAt time.Time) {
	pools.mutex.Lock()
	defer pools.mutex.Unlock()

	isSaved := make(map[string]bool)
	for _, connection := range saved {
		isSaved[connection.ID] = true
	}

	for connectionID, pool := range pools.pools {
		if pool.workspaceID != workspaceID || isSaved[connectionID] || pool.openedAt.After(listedAt) {
			continue
		}

		go closeOldPool(connectionID, pool.database)
		delete(pools.pools, connectionID)
	}
}

func closeOldPool(connectionID string, database *sql.DB) {
	err := database.Close()
	if err != nil {
		tlog.Warnw("Failed to close the connections of an edited connection",
			"connection_id", connectionID,
			"error", err,
		)
	}
}
