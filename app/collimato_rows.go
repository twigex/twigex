// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Cube may hold each request open for a few seconds before answering "Continue
// wait", so the limit is a length of time, not a number of tries.
const collimatoWaitLimit = 30 * time.Second

var collimatoWaitInterval = 500 * time.Millisecond

// emptyDataQuery starts every query with empty lists, because Cube rejects a
// whole query when a list is null, and a Go slice that was never filled is
// sent as null.
func emptyDataQuery() model.DataQuery {
	return model.DataQuery{
		Measures:       make([]string, 0),
		Dimensions:     make([]string, 0),
		Filters:        make([]model.QueryFilter, 0),
		TimeDimensions: make([]model.QueryTimeDimension, 0),
		Order:          make([][]string, 0),
	}
}

// CollimatoGetRows reads rows of a dataset as the user, through the same
// checks as a chart: the user's workspace role, then the data rules, which
// also add the row filters, and only then Cube.
//
// The rows are read ungrouped. Cube groups a query that has fields but no
// measures, which makes the database sort the whole group before returning a
// page: 3.5 s instead of 0.05 s for one company with a million orders.
func (a *App) CollimatoGetRows(ctx context.Context, user model.User, workspaceID string, request model.CollimatoRowsRequest) ([]map[string]any, *model.AppError) {
	// Without a limit Collimato's maximum of 50,000 rows would apply, and a
	// sort field that isn't read is refused by Cube with a vaguer error.
	if request.Limit <= 0 || !slices.Contains(request.Fields, request.OrderBy) {
		tlog.Errorw("A row read needs a limit and a sort field among its fields", "workspace_id", workspaceID)

		return nil, model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
	}

	query := emptyDataQuery()
	query.Dimensions = append(query.Dimensions, request.Fields...)
	query.Filters = append(query.Filters, request.Filters...)
	query.Order = append(query.Order, []string{request.OrderBy, "asc"})
	query.Limit = request.Limit
	query.Ungrouped = true

	// Paging by "after this key" stays fast however far in the page is; an
	// offset makes the database skip every row before it.
	if request.After != "" {
		after := model.QueryFilter{Member: request.OrderBy, Operator: "gt", Values: []string{request.After}}

		query.Filters = append(query.Filters, after)
	}

	data, appErr := a.runCollimatoQuery(ctx, user, workspaceID, query)
	if appErr != nil {
		return nil, appErr
	}

	rows := make([]map[string]any, 0, len(data))
	for _, item := range data {
		row, ok := item.(map[string]any)
		if ok {
			rows = append(rows, row)
		}
	}

	return rows, nil
}

// runCollimatoQuery waits while Cube answers "Continue wait", which it does
// while it is still preparing a query; the chart screens poll the same way.
// Cube reports any other failure inside a normal answer, and reading that as
// "no rows" would make an entity look missing.
func (a *App) runCollimatoQuery(ctx context.Context, user model.User, workspaceID string, query model.DataQuery) ([]any, *model.AppError) {
	started := time.Now()

	for time.Since(started) < collimatoWaitLimit {
		body, appErr := a.queryData(ctx, user, workspaceID, query)
		if appErr != nil {
			return nil, appErr
		}

		message, failed := body["error"]
		if !failed {
			data, _ := body["data"].([]any)

			return data, nil
		}

		if message != "Continue wait" {
			return nil, cubeQueryError(fmt.Sprint(message))
		}

		select {
		case <-ctx.Done():
			return nil, model.NewAppError("collimato.request_failed", http.StatusInternalServerError)
		case <-time.After(collimatoWaitInterval):
		}
	}

	tlog.Errorw("Cube kept asking to wait", "workspace_id", workspaceID)

	return nil, model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
}

func cubeQueryError(message string) *model.AppError {
	id := CubeErrorID(message)
	if id == "" {
		id = "collimato.query_failed"
	}

	return model.NewAppError(id, http.StatusInternalServerError)
}
