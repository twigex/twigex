// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// fakeRowsEngine answers queries in turn and remembers what it was sent.
type fakeRowsEngine struct {
	QueryEngine
	answers []map[string]any
	queries []model.DataQuery
}

func (f *fakeRowsEngine) Query(_ context.Context, _ string, query model.DataQuery) (map[string]any, error) {
	f.queries = append(f.queries, query)

	answer := f.answers[0]
	if len(f.answers) > 1 {
		f.answers = f.answers[1:]
	}

	return answer, nil
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

func rowsTestApp(answers ...map[string]any) (*App, *fakeRowsEngine) {
	engine := &fakeRowsEngine{answers: answers}

	a := &App{
		Store: store.Store{
			Collimato: &fakeMetaCollimatoStore{permissions: []string{model.CollimatoPermissions.PermissionViewCharts.Id}},
		},
		QueryEngine: engine,
	}

	return a, engine
}

func cubeRows(rows ...map[string]any) map[string]any {
	data := make([]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, row)
	}

	return map[string]any{"data": data}
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

func TestCollimatoGetRowsAsksCubeForUngroupedRowsAfterAKey(t *testing.T) {
	a, engine := rowsTestApp(cubeRows(map[string]any{"orders.key": "o-00000060", "orders.number": "Order #100060"}))

	rows, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr != nil {
		t.Fatalf("CollimatoGetRows: %v", appErr)
	}

	if len(rows) != 1 || rows[0]["orders.key"] != "o-00000060" {
		t.Errorf("want the one row Cube sent, got %v", rows)
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
	a, engine := rowsTestApp(cubeRows())
	a.CollimatoRoles = fakeRowRules{}

	_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr != nil {
		t.Fatalf("CollimatoGetRows: %v", appErr)
	}

	found := slices.ContainsFunc(engine.queries[0].Filters, func(filter model.QueryFilter) bool {
		return filter.Member == "orders.country"
	})

	if !found {
		t.Errorf("want the row rule in the query Cube ran, got %v", engine.queries[0].Filters)
	}
}

func TestCollimatoGetRowsWaitsWhileCubeIsPreparing(t *testing.T) {
	interval := collimatoWaitInterval
	collimatoWaitInterval = time.Millisecond

	t.Cleanup(func() {
		collimatoWaitInterval = interval
	})

	a, engine := rowsTestApp(
		map[string]any{"error": "Continue wait"},
		cubeRows(map[string]any{"orders.key": "o-1"}),
	)

	rows, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr != nil {
		t.Fatalf("CollimatoGetRows: %v", appErr)
	}

	if len(rows) != 1 || len(engine.queries) != 2 {
		t.Errorf("want the row after asking twice, got %d rows after %d queries", len(rows), len(engine.queries))
	}
}

func TestCollimatoGetRowsReportsACubeErrorInsteadOfNoRows(t *testing.T) {
	a, _ := rowsTestApp(map[string]any{"error": "Invalid query format: \"measures\" must be an array"})

	rows, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr == nil || appErr.ID != "collimato.query_failed" {
		t.Errorf("want collimato.query_failed, got rows %v and %v", rows, appErr)
	}
}

func TestCollimatoGetRowsRefusesAUserOutsideTheWorkspace(t *testing.T) {
	a, engine := rowsTestApp(cubeRows())
	a.Store.Collimato = &fakeMetaCollimatoStore{}

	_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", companyOrders())
	if appErr == nil || appErr.ID != "collimato.forbidden" {
		t.Errorf("want collimato.forbidden, got %v", appErr)
	}

	if len(engine.queries) != 0 {
		t.Errorf("want Cube never asked, got %d queries", len(engine.queries))
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
			a, engine := rowsTestApp(cubeRows())

			request := companyOrders()
			tc.change(&request)

			_, appErr := a.CollimatoGetRows(context.Background(), model.User{ID: "u1"}, "ws", request)
			if appErr == nil {
				t.Errorf("want the read refused")
			}

			if len(engine.queries) != 0 {
				t.Errorf("want Cube never asked, got %d queries", len(engine.queries))
			}
		})
	}
}

func TestQueryDataNeverSendsAnUngroupedQuery(t *testing.T) {
	a, engine := rowsTestApp(cubeRows())

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
