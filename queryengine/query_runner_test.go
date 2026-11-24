// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

// heldRuns stands in for RunQuery: it counts runs and holds each one until
// the test lets them finish, or until its context is stopped.
type heldRuns struct {
	mutex    sync.Mutex
	count    int
	contexts []context.Context
	finish   chan struct{}
	err      error
}

func newHeldRuns() *heldRuns {
	return &heldRuns{finish: make(chan struct{})}
}

func (held *heldRuns) run(ctx context.Context, _ *sql.DB, query BuiltQuery) (Result, error) {
	held.mutex.Lock()
	held.count++
	held.contexts = append(held.contexts, ctx)
	held.mutex.Unlock()

	select {
	case <-held.finish:
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}

	if held.err != nil {
		return Result{}, held.err
	}

	result := Result{
		Columns: query.Columns,
		Rows:    []map[string]any{{"sales.count": 3.0}},
	}

	return result, nil
}

func (held *heldRuns) runCount() int {
	held.mutex.Lock()
	defer held.mutex.Unlock()

	return held.count
}

func (held *heldRuns) firstContext() context.Context {
	held.mutex.Lock()
	defer held.mutex.Unlock()

	return held.contexts[0]
}

func runnerWith(held *heldRuns) *QueryRunner {
	runner := NewQueryRunner()
	runner.run = held.run

	return runner
}

func (runner *QueryRunner) waitingFor(connection *sql.DB, query BuiltQuery) int {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	run, isRunning := runner.running[runKey(connection, query)]
	if !isRunning {
		return 0
	}

	return run.waiting
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("gave up waiting")
		}

		time.Sleep(time.Millisecond)
	}
}

func countQuery(country string) BuiltQuery {
	query := BuiltQuery{
		SQL:     "SELECT COUNT(*) AS `sales__count` FROM `sales` WHERE `country` = ?",
		Values:  []any{country},
		Columns: []Column{{Key: "sales.count", Title: "Count", Kind: FieldKindMeasure, Type: FieldTypeNumber}},
	}

	return query
}

type runAnswer struct {
	result Result
	err    error
}

// askInBackground runs a query in another goroutine and gives a channel for
// its answer.
func askInBackground(ctx context.Context, runner *QueryRunner, connection *sql.DB, query BuiltQuery) chan runAnswer {
	answers := make(chan runAnswer, 1)

	go func() {
		result, err := runner.Run(ctx, connection, query)
		answers <- runAnswer{result: result, err: err}
	}()

	return answers
}

func TestQueryRunnerRunsIdenticalQueriesOnce(t *testing.T) {
	held := newHeldRuns()
	runner := runnerWith(held)
	connection := new(sql.DB)
	query := countQuery("Latvia")

	var answers []chan runAnswer
	for range 3 {
		answers = append(answers, askInBackground(context.Background(), runner, connection, query))
	}

	waitUntil(t, func() bool { return runner.waitingFor(connection, query) == 3 })
	close(held.finish)

	wantRows := []map[string]any{{"sales.count": 3.0}}

	for _, answer := range answers {
		got := <-answer
		if got.err != nil {
			t.Fatalf("Run: %v", got.err)
		}

		if !reflect.DeepEqual(got.result.Rows, wantRows) {
			t.Errorf("rows = %v, want %v", got.result.Rows, wantRows)
		}
	}

	if held.runCount() != 1 {
		t.Errorf("the query ran %d times, want once", held.runCount())
	}
}

func TestQueryRunnerRunsDifferentQueriesApart(t *testing.T) {
	held := newHeldRuns()
	runner := runnerWith(held)
	connection := new(sql.DB)
	otherConnection := new(sql.DB)

	// The same SQL read with another type, as after a model was saved.
	otherType := countQuery("Latvia")
	otherType.Columns[0].Type = FieldTypeText

	first := askInBackground(context.Background(), runner, connection, countQuery("Latvia"))
	otherValue := askInBackground(context.Background(), runner, connection, countQuery("Estonia"))
	otherPool := askInBackground(context.Background(), runner, otherConnection, countQuery("Latvia"))
	otherTypeAnswer := askInBackground(context.Background(), runner, connection, otherType)

	waitUntil(t, func() bool { return held.runCount() == 4 })
	close(held.finish)

	for _, answer := range []chan runAnswer{first, otherValue, otherPool, otherTypeAnswer} {
		if got := <-answer; got.err != nil {
			t.Fatalf("Run: %v", got.err)
		}
	}
}

func TestQueryRunnerRunsAgainAfterFinishing(t *testing.T) {
	held := newHeldRuns()
	close(held.finish)

	runner := runnerWith(held)
	connection := new(sql.DB)
	query := countQuery("Latvia")

	for range 2 {
		if _, err := runner.Run(context.Background(), connection, query); err != nil {
			t.Fatalf("Run: %v", err)
		}
	}

	if held.runCount() != 2 {
		t.Errorf("the query ran %d times, want twice: nothing is kept", held.runCount())
	}
}

func TestQueryRunnerKeepsRunningWhileSomeoneWaits(t *testing.T) {
	held := newHeldRuns()
	runner := runnerWith(held)
	connection := new(sql.DB)
	query := countQuery("Latvia")

	firstContext, stopFirst := context.WithCancel(context.Background())

	first := askInBackground(firstContext, runner, connection, query)
	second := askInBackground(context.Background(), runner, connection, query)

	waitUntil(t, func() bool { return runner.waitingFor(connection, query) == 2 })
	stopFirst()

	if got := <-first; !errors.Is(got.err, context.Canceled) {
		t.Fatalf("first asker: error %v, want context.Canceled", got.err)
	}

	if held.firstContext().Err() != nil {
		t.Fatal("the query was stopped although someone still waits for it")
	}

	close(held.finish)

	if got := <-second; got.err != nil {
		t.Fatalf("second asker: %v", got.err)
	}
}

func TestQueryRunnerStopsWhenEveryoneGivesUp(t *testing.T) {
	held := newHeldRuns()
	runner := runnerWith(held)
	connection := new(sql.DB)
	query := countQuery("Latvia")

	askerContext, stopAsker := context.WithCancel(context.Background())

	asker := askInBackground(askerContext, runner, connection, query)

	waitUntil(t, func() bool { return held.runCount() == 1 })
	stopAsker()
	<-asker

	waitUntil(t, func() bool { return held.firstContext().Err() != nil })

	close(held.finish)

	if _, err := runner.Run(context.Background(), connection, query); err != nil {
		t.Fatalf("a later asker joined the stopped run: %v", err)
	}

	if held.runCount() != 2 {
		t.Errorf("the query ran %d times, want a new run after the stopped one", held.runCount())
	}
}

func TestQueryRunnerGivesEveryoneTheError(t *testing.T) {
	held := newHeldRuns()
	held.err = errors.New("the database went away")

	runner := runnerWith(held)
	connection := new(sql.DB)
	query := countQuery("Latvia")

	first := askInBackground(context.Background(), runner, connection, query)
	second := askInBackground(context.Background(), runner, connection, query)

	waitUntil(t, func() bool { return runner.waitingFor(connection, query) == 2 })
	close(held.finish)

	for _, answer := range []chan runAnswer{first, second} {
		if got := <-answer; !errors.Is(got.err, held.err) {
			t.Errorf("error = %v, want %v", got.err, held.err)
		}
	}
}

func TestQueryRunnerTurnsAPanicIntoAnError(t *testing.T) {
	runner := NewQueryRunner()
	runner.run = func(context.Context, *sql.DB, BuiltQuery) (Result, error) {
		panic("a value no conversion expected")
	}

	_, err := runner.Run(context.Background(), new(sql.DB), countQuery("Latvia"))
	if err == nil {
		t.Fatal("Run gave no error for a run that panicked")
	}
}

func TestQueryRunnerOnEachDatabase(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		RowLimit:   100,
	}

	runner := NewQueryRunner()

	eachTestDatabase(t, func(t *testing.T, database testDatabase) {
		models := salesModels()
		models.DatabaseTypes = map[string]DatabaseType{"connection-1": database.databaseType}

		query, err := BuildQuery(models, request)
		if err != nil {
			t.Fatalf("BuildQuery: %v", err)
		}

		wantResult, err := RunQuery(context.Background(), database.connection, query)
		if err != nil {
			t.Fatalf("RunQuery: %v", err)
		}

		gotResult, err := runner.Run(context.Background(), database.connection, query)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}

		if !reflect.DeepEqual(gotResult, wantResult) {
			t.Errorf("Run gave\n%v\nRunQuery gave\n%v", gotResult, wantResult)
		}
	})
}
