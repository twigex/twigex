// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// fakeRowsEngine answers every query with the same rows, or fails with err,
// and remembers what it was sent.
type fakeRowsEngine struct {
	QueryEngine
	rows    []map[string]any
	err     error
	queries []model.DataQuery
}

func (f *fakeRowsEngine) Query(_ context.Context, _ string, query model.DataQuery) (*collimato.Answer, error) {
	f.queries = append(f.queries, query)

	if f.err != nil {
		return nil, f.err
	}

	return &collimato.Answer{Rows: f.rows}, nil
}

// fakeRowRules adds a row rule to every query, the way the enterprise data
// rules add "country = LV" for a user limited to Latvia.
type fakeRowRules struct {
	interfaces.CollimatoRoles
}

func (fakeRowRules) HasPermissionToFields(_ []model.CollimatoRole, query *model.DataQuery) (bool, error) {
	rule := model.QueryFilter{Member: "orders.country", Operator: "equals", Values: []string{"LV"}}

	query.Filters = append(query.Filters, rule)

	return true, nil
}

func rowsTestApp(rows ...map[string]any) (*App, *fakeRowsEngine) {
	engine := &fakeRowsEngine{rows: rows}

	a := &App{
		Store: store.Store{
			Collimato: &fakeFieldListCollimatoStore{permissions: []string{model.CollimatoPermissions.PermissionViewCharts.Id}},
		},
		QueryEngine: engine,
	}

	return a, engine
}

func companyOrders() model.CollimatoRowsRequest {
	return model.CollimatoRowsRequest{
		Fields:  []string{"orders.key", "orders.number"},
		Filters: []model.QueryFilter{{Member: "orders.company_id", Operator: "equals", Values: []string{"c-01"}}},
		OrderBy: "orders.key",
		After:   "o-00000050",
		Limit:   50,
	}
}

func TestCollimatoGetRowsAsksForUngroupedRowsAfterAKey(t *testing.T) {
	a, engine := rowsTestApp(map[string]any{"orders.key": "o-00000060", "orders.number": "Order #100060"})

	rows, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr != nil {
		t.Fatalf("CollimatoGetRows: %v", appErr)
	}

	if len(rows) != 1 || rows[0]["orders.key"] != "o-00000060" {
		t.Errorf("want the one row the engine sent, got %v", rows)
	}

	sent := engine.queries[0]

	if !sent.Ungrouped {
		t.Errorf("want an ungrouped query")
	}

	if !slices.Equal(sent.Order[0], []string{"orders.key", "asc"}) {
		t.Errorf("want rows sorted by the key, got %v", sent.Order)
	}

	after := model.QueryFilter{Member: "orders.key", Operator: "gt", Values: []string{"o-00000050"}}

	found := slices.ContainsFunc(sent.Filters, func(filter model.QueryFilter) bool {
		return filter.Member == after.Member && filter.Operator == after.Operator && slices.Equal(filter.Values, after.Values)
	})

	if !found {
		t.Errorf("want the rows after o-00000050, got %v", sent.Filters)
	}

	raw, err := json.Marshal(sent)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	if !strings.Contains(string(raw), `"measures":[]`) || !strings.Contains(string(raw), `"timeDimensions":[]`) {
		t.Errorf("want empty lists sent as [], got %s", raw)
	}
}

func TestCollimatoGetRowsKeepsTheRowRules(t *testing.T) {
	a, engine := rowsTestApp()
	a.CollimatoRoles = fakeRowRules{}

	_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr != nil {
		t.Fatalf("CollimatoGetRows: %v", appErr)
	}

	found := slices.ContainsFunc(engine.queries[0].Filters, func(filter model.QueryFilter) bool {
		return filter.Member == "orders.country"
	})

	if !found {
		t.Errorf("want the row rule in the query that ran, got %v", engine.queries[0].Filters)
	}
}

func TestCollimatoGetRowsReportsAFailureInsteadOfNoRows(t *testing.T) {
	a, engine := rowsTestApp()
	engine.err = newQueryError(queryErrorTimeLimit, "")

	rows, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr == nil || appErr.ID != queryErrorTimeLimit {
		t.Errorf("want %s, got rows %v and %v", queryErrorTimeLimit, rows, appErr)
	}
}

func TestCollimatoGetRowsRefusesAUserOutsideTheWorkspace(t *testing.T) {
	a, engine := rowsTestApp()
	a.Store.Collimato = &fakeFieldListCollimatoStore{}

	_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr == nil || appErr.ID != "collimato.forbidden" {
		t.Errorf("want collimato.forbidden, got %v", appErr)
	}

	if len(engine.queries) != 0 {
		t.Errorf("want no query run, got %d queries", len(engine.queries))
	}
}

func TestCollimatoGetRowsNeedsALimitAndASortFieldItReads(t *testing.T) {
	cases := []struct {
		name   string
		change func(request *model.CollimatoRowsRequest)
	}{
		{"no limit", func(request *model.CollimatoRowsRequest) { request.Limit = 0 }},
		{"no sort field", func(request *model.CollimatoRowsRequest) { request.OrderBy = "" }},
		{"a sort field it doesn't read", func(request *model.CollimatoRowsRequest) { request.OrderBy = "orders.amount" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, engine := rowsTestApp()

			request := companyOrders()
			tc.change(&request)

			_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", request)
			if appErr == nil {
				t.Errorf("want the read refused")
			}

			if len(engine.queries) != 0 {
				t.Errorf("want no query run, got %d queries", len(engine.queries))
			}
		})
	}
}

func TestQueryDataNeverSendsAnUngroupedQuery(t *testing.T) {
	a, engine := rowsTestApp()

	query := emptyDataQuery()
	query.Measures = []string{"salaries.average"}
	query.Ungrouped = true

	_, appErr := a.QueryData(context.Background(), model.User{ID: "u1"}, "ws", query)
	if appErr != nil {
		t.Fatalf("QueryData: %v", appErr)
	}

	if engine.queries[0].Ungrouped {
		t.Errorf("want ungrouped switched off for a query from the browser")
	}
}

// The graph reads an error below 500 as "you may not see this", so an engine
// refusal must not reach it as a 400.
func TestRowReadErrorsAreFailuresNotRefusals(t *testing.T) {
	cases := []struct {
		name   string
		appErr *model.AppError
		wantID string
	}{
		{"a model error", newQueryError(queryErrorModelError, "orders").appError(), "collimato.query_failed"},
		{"the database's message", model.NewAppErrorRaw("Error 1054", http.StatusBadGateway), "collimato.query_failed"},
		{"a refusal", model.NewAppError("collimato.permission_denied", http.StatusForbidden), "collimato.permission_denied"},
		{"the time limit", newQueryError(queryErrorTimeLimit, "").appError(), queryErrorTimeLimit},
		{"a connection failure", newQueryError("connection.verification.unreachable", "").appError(), "connection.verification.unreachable"},
	}

	for _, testCase := range cases {
		got := rowReadError("ws", testCase.appErr)

		if got.ID != testCase.wantID {
			t.Errorf("%s: want %s, got %s", testCase.name, testCase.wantID, got.ID)
		}

		if got.Status < http.StatusInternalServerError && got.Status != http.StatusForbidden {
			t.Errorf("%s: a failure must be 500 or above, got %d", testCase.name, got.Status)
		}
	}
}
