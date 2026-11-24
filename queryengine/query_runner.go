// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sync"
)

// QueryRunner runs queries, and runs a query only once when the same query is
// asked for again while it is still running: everyone asking gets that one
// answer. Nothing is kept after it finishes, so answers are never old. Many
// people asking for the same queries at once cost the database each query
// once, not once per person. The host app makes one and keeps it.
type QueryRunner struct {
	mutex   sync.Mutex
	running map[string]*sharedRun

	// run is RunQuery; tests replace it to count and hold runs.
	run func(ctx context.Context, connection *sql.DB, query BuiltQuery) (Result, error)
}

// sharedRun is one query running for everyone who asked for it.
type sharedRun struct {
	finished chan struct{}
	result   Result
	err      error

	// waiting is how many askers still wait. When the last one gives up, the
	// query is stopped.
	waiting int
	stop    context.CancelFunc
}

func NewQueryRunner() *QueryRunner {
	runner := &QueryRunner{
		running: map[string]*sharedRun{},
		run:     RunQuery,
	}

	return runner
}

// Run is RunQuery, sharing a run with identical queries already running on
// the same connection pool. Cancelling ctx stops waiting; the query itself
// stops only when everyone waiting for it has stopped.
func (runner *QueryRunner) Run(ctx context.Context, connection *sql.DB, query BuiltQuery) (Result, error) {
	key := runKey(connection, query)

	runner.mutex.Lock()

	run, isRunning := runner.running[key]
	if !isRunning {
		run = runner.start(ctx, key, connection, query)
	}

	run.waiting++

	runner.mutex.Unlock()

	select {
	case <-run.finished:
		if run.err != nil {
			return Result{}, run.err
		}

		return resultFor(query, run.result), nil

	case <-ctx.Done():
		runner.stopWaiting(key, run)
		return Result{}, ctx.Err()
	}
}

// runKey names a query by its connection pool (the database and login), SQL,
// values and column types. The types decide how rows are read, and a saved
// model can change them without changing the SQL.
func runKey(connection *sql.DB, query BuiltQuery) string {
	var columnTypes []string

	for _, column := range query.Columns {
		columnType := fmt.Sprintf("%s/%s/%s", column.Kind, column.Type, column.MeasureType)
		columnTypes = append(columnTypes, columnType)
	}

	return fmt.Sprintf("%p\x00%s\x00%#v\x00%v", connection, query.SQL, query.Values, columnTypes)
}

// start begins a run. Its query does not stop with the first asker's ctx,
// only when nobody waits any more. The mutex is held.
func (runner *QueryRunner) start(ctx context.Context, key string, connection *sql.DB, query BuiltQuery) *sharedRun {
	runContext, stop := context.WithCancel(context.WithoutCancel(ctx))

	run := &sharedRun{
		finished: make(chan struct{}),
		stop:     stop,
	}

	runner.running[key] = run

	go func() {
		defer stop()

		result, err := runner.runCatchingPanic(runContext, connection, query)

		runner.mutex.Lock()
		runner.forget(key, run)
		runner.mutex.Unlock()

		run.result = result
		run.err = err
		close(run.finished)
	}()

	return run
}

// runCatchingPanic runs the query, turning a panic into an error. The run has
// a goroutine of its own, where a panic would stop the whole program rather
// than fail one request, as it would in the request's own goroutine.
func (runner *QueryRunner) runCatchingPanic(ctx context.Context, connection *sql.DB, query BuiltQuery) (result Result, err error) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			result = Result{}
			err = fmt.Errorf("query engine: running the query failed: %v", recovered)
		}
	}()

	return runner.run(ctx, connection, query)
}

// stopWaiting is one asker giving up. When it was the last one, the query is
// stopped and forgotten, so a later asker starts a new run rather than join
// a stopped one.
func (runner *QueryRunner) stopWaiting(key string, run *sharedRun) {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	run.waiting--

	if run.waiting > 0 {
		return
	}

	run.stop()
	runner.forget(key, run)
}

// forget removes a run, unless a newer run of the same query has taken its
// place. The mutex is held.
func (runner *QueryRunner) forget(key string, run *sharedRun) {
	if runner.running[key] == run {
		delete(runner.running, key)
	}
}

// resultFor gives an asker the shared rows with its own query's columns and
// date ranges. The SQL and column types are the same, so only titles could
// differ, if a model was saved between two identical requests. The rows are
// shared, so whoever reads them must not change them.
func resultFor(query BuiltQuery, shared Result) Result {
	columns := slices.Clone(query.Columns)

	for position := range columns {
		if position < len(shared.Columns) {
			columns[position].KeptAsText = shared.Columns[position].KeptAsText
		}
	}

	result := Result{
		Columns:    columns,
		Rows:       shared.Rows,
		DateRanges: query.DateRanges,
	}

	return result
}
