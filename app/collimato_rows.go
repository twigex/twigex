// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"slices"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// emptyDataQuery starts every query with empty lists, so the query reads the
// same as one the browser sends.
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
// also add the row filters, and only then the query.
//
// The rows are read ungrouped: grouping makes the database sort the whole
// group before returning a page, 3.5 s instead of 0.05 s for a million rows.
func (a *App) CollimatoGetRows(ctx context.Context, user model.User, workspaceID string, request model.CollimatoRowsRequest) ([]map[string]any, *model.AppError) {
	// Without a limit Collimato's maximum of 50,000 rows would apply, and a
	// sort field that isn't read is refused with a vaguer error.
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

	answer, appErr := a.queryData(ctx, user, workspaceID, query)
	if appErr != nil {
		return nil, rowReadError(workspaceID, appErr)
	}

	return answer.Rows, nil
}

// rowReadError keeps a refusal (the user may not read these rows), a
// connection failure and the time limit as they are. Anything else the
// engine or the database refused is a failed read, not a refusal: the graph
// reads an error below 500 as "you may not see this" and would quietly show
// nothing.
func rowReadError(workspaceID string, appErr *model.AppError) *model.AppError {
	isRequestRefused := appErr.Status == http.StatusBadRequest
	isDatabaseMessage := appErr.Raw

	if !isRequestRefused && !isDatabaseMessage {
		return appErr
	}

	tlog.Errorw("A row read failed",
		"workspace_id", workspaceID,
		"error", appErr.Message,
	)

	return model.NewAppError("collimato.query_failed", http.StatusInternalServerError)
}
