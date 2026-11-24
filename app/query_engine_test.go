// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

var loadTranslationsOnce sync.Once

func loadTranslations(t *testing.T) {
	t.Helper()

	loadTranslationsOnce.Do(func() {
		if err := i18n.Init("../i18n"); err != nil {
			t.Fatalf("load translations: %v", err)
		}
	})
}

// fakeEngineStore holds a workspace's saved files and connections.
type fakeEngineStore struct {
	files       map[string][]model.CollimatoWorkspaceFile
	connections []model.Connection
}

func (store *fakeEngineStore) GetWorkspaceFiles(_ string, fileType string) ([]model.CollimatoWorkspaceFile, error) {
	return store.files[fileType], nil
}

func (store *fakeEngineStore) GetConnections(string) ([]model.Connection, error) {
	return store.connections, nil
}

func builderFile(t *testing.T, name string, fileType string, builderModel any) model.CollimatoWorkspaceFile {
	t.Helper()

	encoded, err := json.Marshal(builderModel)
	if err != nil {
		t.Fatal(err)
	}

	text := string(encoded)

	return model.CollimatoWorkspaceFile{Name: name, FileType: fileType, BuilderModel: &text}
}

const engineTestTables = `
CREATE TABLE stores (
	id      INT PRIMARY KEY,
	name    VARCHAR(50) NOT NULL,
	is_open BOOLEAN NOT NULL
);

INSERT INTO stores VALUES (1, 'Riga', TRUE), (2, 'Tallinn', FALSE);

CREATE TABLE orders (
	id         INT PRIMARY KEY,
	store_id   INT NOT NULL,
	amount     DECIMAL(10, 2) NOT NULL,
	ordered_at DATETIME NOT NULL
);

INSERT INTO orders VALUES
	(1, 1, 100.50, '2025-03-02 10:00:00'),
	(2, 1,  50.00, '2025-03-20 12:00:00'),
	(3, 2,  20.25, '2025-04-01 09:00:00');
`

// engineTestDatabase makes a database of its own on the MySQL server of
// TEST_MYSQL_DSN, and the saved connection that reaches it.
func engineTestDatabase(t *testing.T) model.Connection {
	t.Helper()

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set; skipping tests that run queries")
	}

	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("TEST_MYSQL_DSN: %v", err)
	}

	server, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	databaseName := fmt.Sprintf("twigex_builtin_engine_%d", time.Now().UnixNano())

	if _, err := server.Exec("CREATE DATABASE " + databaseName); err != nil {
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		dropper, err := sql.Open("mysql", dsn)
		if err != nil {
			return
		}
		defer dropper.Close()

		if _, err := dropper.Exec("DROP DATABASE " + databaseName); err != nil {
			t.Logf("drop test database %s: %v", databaseName, err)
		}
	})

	if _, err := server.Exec("USE " + databaseName + ";\n" + engineTestTables); err != nil {
		t.Fatalf("create test tables: %v", err)
	}

	host, port, err := net.SplitHostPort(config.Addr)
	if err != nil {
		t.Fatalf("TEST_MYSQL_DSN address: %v", err)
	}

	connection := model.Connection{
		ID:          "connection-1",
		WorkspaceID: "ws",
		Type:        "mysql",
		Host:        host,
		Port:        port,
		Username:    config.User,
		Password:    config.Passwd,
		Database:    databaseName,
		UpdatedAt:   1,
	}

	return connection
}

func engineTestFiles(t *testing.T) map[string][]model.CollimatoWorkspaceFile {
	t.Helper()

	stores := model.Dataset{
		Name:         "stores",
		SqlTable:     "stores",
		ConnectionID: "connection-1",
		Dimensions: []model.DatasetDimension{
			{Name: "id", SQL: "id", Type: "number", PrimaryKey: true},
			{Name: "name", SQL: "name", Type: "string", Title: "Store"},
			{Name: "is_open", SQL: "is_open", Type: "boolean"},
		},
		Measures: []model.DatasetMeasure{{Name: "count", Type: "count", Title: "Stores"}},
	}

	orders := model.Dataset{
		Name:         "orders",
		SqlTable:     "orders",
		ConnectionID: "connection-1",
		Dimensions: []model.DatasetDimension{
			{Name: "id", SQL: "id", Type: "number", PrimaryKey: true},
			{Name: "ordered_at", SQL: "ordered_at", Type: "time", Title: "Order date"},
		},
		Measures: []model.DatasetMeasure{
			{Name: "revenue", Type: "sum", SQL: "amount"},
			{Name: "average_order", Type: "avg", SQL: "amount"},
		},
		Joins: []model.DatasetJoin{{Name: "stores", Relationship: "many_to_one", ThisColumn: "store_id", OtherColumn: "id"}},
	}

	returns := model.Dataset{
		Name:         "returns",
		SqlTable:     "returns",
		ConnectionID: "connection-1",
		Joins:        []model.DatasetJoin{{Name: "missing", Relationship: "many_to_one", ThisColumn: "a", OtherColumn: "b"}},
	}

	writtenByHand := model.CollimatoWorkspaceFile{Name: "legacy.yml", FileType: model.CollimatoFileTypeDataset}

	files := map[string][]model.CollimatoWorkspaceFile{
		model.CollimatoFileTypeDataset: {
			builderFile(t, "stores", model.CollimatoFileTypeDataset, stores),
			builderFile(t, "orders", model.CollimatoFileTypeDataset, orders),
			builderFile(t, "returns", model.CollimatoFileTypeDataset, returns),
			writtenByHand,
		},
	}

	return files
}

func noDecryption(password string) (string, error) {
	return password, nil
}

func engineForTest(t *testing.T) (*engineAdapter, *fakeEngineStore) {
	t.Helper()

	loadTranslations(t)

	store := &fakeEngineStore{
		files:       engineTestFiles(t),
		connections: []model.Connection{engineTestDatabase(t)},
	}

	connector := queryengine.NewConnector(queryengine.ConnectorOptions{QueryTimeLimit: time.Minute})
	engine := newEngineAdapter(store, noDecryption, connector)

	return engine, store
}

func TestEngineAdapterListsFieldsAndModelErrors(t *testing.T) {
	loadTranslations(t)

	store := &fakeEngineStore{files: engineTestFiles(t)}
	engine := newEngineAdapter(store, noDecryption, engineConnector)

	fieldList, err := engine.FieldList(context.Background(), "ws")
	if err != nil {
		t.Fatalf("FieldList: %v", err)
	}

	names := []string{}
	for _, dataset := range fieldList.Datasets {
		names = append(names, dataset.Name)
	}

	if !reflect.DeepEqual(names, []string{"orders", "stores"}) {
		t.Errorf("want orders and stores, got %v", names)
	}

	var revenue collimato.Field
	for _, field := range fieldList.Datasets[0].Fields {
		if field.Key == "orders.revenue" {
			revenue = field
		}
	}

	if revenue.FullTitle != "Orders Revenue" || revenue.MeasureType != "sum" {
		t.Errorf("want revenue with its titles and total type, got %+v", revenue)
	}

	wantErrors := []collimato.ModelError{
		{Name: "legacy.yml", Reason: collimato.UnsupportedWrittenByHand},
		{Name: "returns", Reason: collimato.UnsupportedHasErrors, Detail: "unknown_join_target: missing"},
	}

	if !reflect.DeepEqual(fieldList.ModelErrors, wantErrors) {
		t.Errorf("want model errors\n%+v\ngot\n%+v", wantErrors, fieldList.ModelErrors)
	}

	TranslateModelErrors(fieldList, "en")

	wantMessage := "legacy.yml was written by hand for the previous analytics engine and is no longer supported. Recreate it in the data model builder."
	if fieldList.ModelErrors[0].Message != wantMessage {
		t.Errorf("want message %q, got %q", wantMessage, fieldList.ModelErrors[0].Message)
	}

	if !fieldList.HasModelsThatMayComeBack() {
		t.Error("a dataset with errors may come back, so the graph sync must wait")
	}
}

func TestEngineAdapterAnswersWithTheEnginesValues(t *testing.T) {
	engine, _ := engineForTest(t)

	query := model.DataQuery{
		Measures:   []string{"orders.revenue", "orders.average_order"},
		Dimensions: []string{"stores.name", "stores.is_open"},
		TimeDimensions: []model.QueryTimeDimension{{
			Dimension:   "orders.ordered_at",
			Granularity: "month",
			DateRange:   model.QueryDateRange{IsArray: true, ArrayValue: []string{"2025-01-01", "2025-12-31"}},
		}},
		Order: [][]string{{"orders.ordered_at", "asc"}},
		Limit: 100,
	}

	answer, err := engine.Query(context.Background(), "ws", query)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	encoded, err := json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	checkAnswerJSON(t, "rows", decoded["rows"], `[
		{"stores.name": "Riga", "stores.is_open": true, "orders.ordered_at": "2025-03-01T00:00:00Z", "orders.revenue": 150.5, "orders.average_order": 75.25},
		{"stores.name": "Tallinn", "stores.is_open": false, "orders.ordered_at": "2025-04-01T00:00:00Z", "orders.revenue": 20.25, "orders.average_order": 20.25}
	]`)

	checkAnswerJSON(t, "date ranges", decoded["dateRanges"],
		`[{"field": "orders.ordered_at", "start": "2025-01-01T00:00:00Z", "end": "2025-12-31T23:59:59.999Z"}]`)

	checkAnswerJSON(t, "date column", answerColumn(t, decoded, "orders.ordered_at"),
		`{"key": "orders.ordered_at", "title": "Order date", "fullTitle": "Orders Order date", "kind": "dimension", "type": "time", "dateStep": "month"}`)
}

func TestEngineAdapterSQL(t *testing.T) {
	loadTranslations(t)

	store := &fakeEngineStore{
		files:       engineTestFiles(t),
		connections: []model.Connection{{ID: "connection-1", Type: "mysql"}},
	}

	engine := newEngineAdapter(store, noDecryption, engineConnector)

	sqlText, err := engine.SQL(context.Background(), "ws", model.DataQuery{Measures: []string{"orders.revenue"}, Limit: 10})
	if err != nil {
		t.Fatalf("SQL: %v", err)
	}

	wantSQL := "SELECT SUM(`orders`.`amount`) AS `orders__revenue`\nFROM `orders` AS `orders`\nLIMIT 10"
	if sqlText != wantSQL {
		t.Errorf("want SQL\n%s\ngot\n%s", wantSQL, sqlText)
	}
}

func queryErrorOf(t *testing.T, err error) *QueryError {
	t.Helper()

	var queryError *QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("want a QueryError, got %v", err)
	}

	return queryError
}

func TestEngineAdapterNamesWhyItCannotAnswer(t *testing.T) {
	loadTranslations(t)

	store := &fakeEngineStore{
		files:       engineTestFiles(t),
		connections: []model.Connection{{ID: "connection-1", Type: "mysql"}},
	}

	engine := newEngineAdapter(store, noDecryption, engineConnector)

	cases := []struct {
		name        string
		query       model.DataQuery
		wantID      string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "a field the model does not have",
			query:       model.DataQuery{Measures: []string{"orders.profit"}, Limit: 10},
			wantID:      queryErrorUnknownField,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "This chart uses orders.profit, which the data model no longer has.",
		},
		{
			name:        "a field from a file written by hand",
			query:       model.DataQuery{Measures: []string{"legacy.count"}, Limit: 10},
			wantID:      queryErrorUnsupportedModel,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "This chart uses legacy.count, from a data model written for the previous analytics engine, which is no longer supported. Recreate the data model in the data model builder.",
		},
		{
			name:        "a dataset with errors",
			query:       model.DataQuery{Dimensions: []string{"returns.id"}, Limit: 10},
			wantID:      queryErrorModelError,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "The data model returns has an error. Open it in the data model builder to fix it.",
		},
		{
			name:       "a request that cannot be read",
			query:      model.DataQuery{Measures: []string{"orders.revenue"}, Order: [][]string{{"orders.revenue"}}, Limit: 10},
			wantID:     queryErrorInvalidRequest,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range cases {
		_, err := engine.SQL(context.Background(), "ws", testCase.query)

		queryError := queryErrorOf(t, err)
		if queryError.ID != testCase.wantID {
			t.Errorf("%s: want %s, got %+v", testCase.name, testCase.wantID, queryError)
			continue
		}

		appError := queryError.appError()
		if appError.Status != testCase.wantStatus {
			t.Errorf("%s: want status %d, got %d", testCase.name, testCase.wantStatus, appError.Status)
		}

		message := appError.Translate("en")
		if testCase.wantMessage != "" && message != testCase.wantMessage {
			t.Errorf("%s: want message %q, got %q", testCase.name, testCase.wantMessage, message)
		}
	}
}

func TestQueryErrorInTheReadersLanguage(t *testing.T) {
	loadTranslations(t)

	appError := newQueryError(queryErrorTimeLimit, "").appError()

	if appError.Status != http.StatusGatewayTimeout {
		t.Errorf("want status 504, got %d", appError.Status)
	}

	message := appError.Translate("lv")
	if message != "Vaicājums ilga vairāk nekā 10 minūtes un tika apturēts." {
		t.Errorf("want the Latvian message, got %q", message)
	}
}

func TestEngineAdapterGivesTheDatabasesOwnMessage(t *testing.T) {
	engine, store := engineForTest(t)

	broken := model.Dataset{
		Name:         "orders",
		SqlTable:     "no_such_table",
		ConnectionID: "connection-1",
		Measures:     []model.DatasetMeasure{{Name: "count", Type: "count"}},
	}

	store.files[model.CollimatoFileTypeDataset] = []model.CollimatoWorkspaceFile{
		builderFile(t, "orders", model.CollimatoFileTypeDataset, broken),
	}

	_, err := engine.Query(context.Background(), "ws", model.DataQuery{Measures: []string{"orders.count"}, Limit: 10})

	queryError := queryErrorOf(t, err)
	if queryError.ID != "" || queryError.DatabaseMessage == "" {
		t.Errorf("want the database's own message, got %+v", queryError)
	}

	appError := queryError.appError()
	if appError.Status != http.StatusBadGateway || appError.Translate("en") != queryError.DatabaseMessage {
		t.Errorf("want the database's message with status 502, got %d %q", appError.Status, appError.Translate("en"))
	}
}

func TestEngineAdapterOpensANewPoolWhenTheConnectionIsEdited(t *testing.T) {
	engine, store := engineForTest(t)

	query := model.DataQuery{Measures: []string{"orders.revenue"}, Limit: 10}

	if _, err := engine.Query(context.Background(), "ws", query); err != nil {
		t.Fatalf("Query: %v", err)
	}

	firstPool := engine.pools.pools["connection-1"].database

	if _, err := engine.Query(context.Background(), "ws", query); err != nil {
		t.Fatalf("Query: %v", err)
	}

	if engine.pools.pools["connection-1"].database != firstPool {
		t.Error("an unchanged connection got a new pool")
	}

	// Saved within the same second, so the time it was saved is unchanged.
	store.connections[0].Password = "wrong"

	_, err := engine.Query(context.Background(), "ws", query)

	queryError := queryErrorOf(t, err)
	if queryError.ID != "connection.verification.credentials" {
		t.Errorf("want the edited, wrong password used and named, got %+v", queryError)
	}
}

// A deleted connection's pool is closed when its workspace is next read; a
// pool of another workspace is kept.
func TestEngineAdapterClosesThePoolOfADeletedConnection(t *testing.T) {
	engine, store := engineForTest(t)

	query := model.DataQuery{Measures: []string{"orders.revenue"}, Limit: 10}

	if _, err := engine.Query(context.Background(), "ws", query); err != nil {
		t.Fatalf("Query: %v", err)
	}

	otherWorkspacePool := connectionPool{database: new(sql.DB), workspaceID: "other-ws"}
	engine.pools.pools["connection-2"] = otherWorkspacePool

	store.connections = nil

	if _, err := engine.FieldList(context.Background(), "ws"); err != nil {
		t.Fatalf("FieldList: %v", err)
	}

	if _, isOpen := engine.pools.pools["connection-1"]; isOpen {
		t.Error("the deleted connection's pool is still open")
	}

	if _, isOpen := engine.pools.pools["connection-2"]; !isOpen {
		t.Error("another workspace's pool was closed")
	}
}

// A list read before a connection was saved does not close the pool another
// request has just opened for it.
func TestEngineAdapterKeepsAPoolNewerThanTheConnectionList(t *testing.T) {
	pools := newConnectionPools(engineConnector)

	listedAt := time.Now()
	pools.pools["connection-new"] = connectionPool{database: new(sql.DB), workspaceID: "ws", openedAt: listedAt.Add(time.Millisecond)}

	pools.closeDeleted("ws", nil, listedAt)

	if _, isOpen := pools.pools["connection-new"]; !isOpen {
		t.Error("a pool opened after the list was read was closed")
	}
}

func TestEngineAdapterRefusesAConnectionOfAnotherWorkspace(t *testing.T) {
	engine, store := engineForTest(t)

	store.connections[0].ID = "connection-of-this-workspace"

	_, err := engine.Query(context.Background(), "ws", model.DataQuery{Measures: []string{"orders.revenue"}, Limit: 10})

	queryError := queryErrorOf(t, err)
	if queryError.ID != queryErrorConnectionMissing {
		t.Errorf("want the query refused, got %+v", queryError)
	}
}

func answerColumn(t *testing.T, decoded map[string]any, key string) any {
	t.Helper()

	for _, column := range decoded["columns"].([]any) {
		if column.(map[string]any)["key"] == key {
			return column
		}
	}

	t.Fatalf("no column %s", key)

	return nil
}

func checkAnswerJSON(t *testing.T, name string, got any, wantJSON string) {
	t.Helper()

	var want any
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatalf("bad test JSON: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.Marshal(got)
		t.Errorf("%s:\nwant %s\ngot  %s", name, wantJSON, gotJSON)
	}
}
