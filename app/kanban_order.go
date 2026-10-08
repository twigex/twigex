// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// A Kanban view's item_order holds its columns and, in each column's
// "order", the cards placed by hand, first to last. The rest of a column's
// cards follow by when they were created. Only the server changes the
// placed cards, one move at a time: a client holds one page of each column,
// so an order it sent whole would drop every card it had not loaded.

var errUnknownKanbanColumn = errors.New("the view has no such column")

// kanbanCardsSQL matches the tasks a board shows as cards: subtasks are
// shown under their parent instead.
const kanbanCardsSQL = model.TopLevelTaskSQL

func kanbanOrderID(v any) string {
	if v == nil {
		return ""
	}

	return fmt.Sprint(v)
}

func kanbanColumns(order map[string]any) []map[string]any {
	raw, _ := order["fields"].([]any)
	columns := make([]map[string]any, 0, len(raw))
	for _, c := range raw {
		if column, ok := c.(map[string]any); ok {
			columns = append(columns, column)
		}
	}

	return columns
}

func placedCards(column map[string]any) []any {
	placed, _ := column["order"].([]any)
	return placed
}

func placedIDs(column map[string]any) []string {
	placed := placedCards(column)
	ids := make([]string, 0, len(placed))
	for _, p := range placed {
		if card, ok := p.(map[string]any); ok {
			ids = append(ids, kanbanOrderID(card["id"]))
		}
	}

	return ids
}

// moveKanbanCard places taskID in column after afterID, or first when
// afterID is empty. When afterID has no place yet, the column's cards up to
// it are placed first, in the order the board shows them, which rest gives.
func moveKanbanCard(orderJSON, taskID, columnID, afterID string, rest func(placed []string) ([]string, error)) (string, error) {
	var order map[string]any
	if err := json.Unmarshal([]byte(orderJSON), &order); err != nil {
		return "", err
	}

	section, _ := order["section"].(string)
	if section == "" {
		section = "single_select"
	}

	var target map[string]any
	for _, column := range kanbanColumns(order) {
		if kanbanOrderID(column["id"]) == columnID {
			target = column
		}

		kept := make([]any, 0, len(placedCards(column)))
		for _, p := range placedCards(column) {
			if card, ok := p.(map[string]any); !ok || kanbanOrderID(card["id"]) != taskID {
				kept = append(kept, p)
			}
		}

		column["order"] = kept
	}

	if target == nil {
		return "", errUnknownKanbanColumn
	}

	placed := placedCards(target)
	ids := placedIDs(target)
	at := 0
	if afterID != "" {
		i := slices.Index(ids, afterID)
		if i < 0 {
			more, err := rest(ids)
			if err != nil {
				return "", err
			}

			// Placing the rest only helps when it reaches afterID; otherwise
			// every card would be saved as placed for nothing.
			if slices.Contains(more, afterID) {
				for _, id := range more {
					if id == taskID {
						continue
					}

					placed = append(placed, map[string]any{"id": id, section: columnID})
					ids = append(ids, id)
					if id == afterID {
						break
					}
				}
			}

			i = slices.Index(ids, afterID)
		}

		at = i + 1
		if i < 0 {
			at = len(placed)
		}
	}

	card := map[string]any{"id": taskID, section: columnID}
	placed = slices.Insert(placed, at, any(card))
	if len(placed) > kanbanSavedOrderLimit {
		at = min(at, kanbanSavedOrderLimit-1)
		kept := append(slices.Clone(placed[:at]), card)
		for _, p := range placed[at:] {
			if len(kept) == kanbanSavedOrderLimit {
				break
			}

			if c, ok := p.(map[string]any); !ok || kanbanOrderID(c["id"]) != taskID {
				kept = append(kept, p)
			}
		}

		placed = kept
	}

	target["order"] = placed

	out, err := json.Marshal(order)
	return string(out), err
}

// mergeKanbanColumns takes a view's columns, their names, colours, widths,
// visibility and order, and the fields shown on cards from what a client
// sent, and keeps the cards placed in each column as they are stored.
func mergeKanbanColumns(storedJSON, incomingJSON string) (string, error) {
	var incoming map[string]any
	if err := json.Unmarshal([]byte(incomingJSON), &incoming); err != nil {
		return "", err
	}

	var stored map[string]any
	if storedJSON != "" {
		_ = json.Unmarshal([]byte(storedJSON), &stored)
	}

	placed := map[string][]any{}
	for _, column := range kanbanColumns(stored) {
		placed[kanbanOrderID(column["id"])] = placedCards(column)
	}

	columns := kanbanColumns(incoming)
	fields := make([]any, 0, len(columns))
	for _, column := range columns {
		column["order"] = placed[kanbanOrderID(column["id"])]
		if column["order"] == nil {
			column["order"] = []any{}
		}

		fields = append(fields, column)
	}

	incoming["fields"] = fields

	out, err := json.Marshal(incoming)
	return string(out), err
}

// withoutPlacedCards returns a Kanban order with its columns but not the
// cards placed in them, for clients: the server orders the cards when it
// sends them, and the lists can hold thousands of ids.
func withoutPlacedCards(orderJSON string) string {
	var order map[string]any
	if err := json.Unmarshal([]byte(orderJSON), &order); err != nil {
		return orderJSON
	}

	columns := kanbanColumns(order)
	if len(columns) == 0 {
		return orderJSON
	}

	for _, column := range columns {
		delete(column, "order")
	}

	out, err := json.Marshal(order)
	if err != nil {
		return orderJSON
	}

	return string(out)
}

// MoveKanbanCard places a card in a Kanban column after another card, or
// first when afterID is empty, and tells the view's other viewers.
func (a *App) MoveKanbanCard(ctx context.Context, workspaceID, tableID, viewID, taskID, columnID, afterID string, user model.User) *model.AppError {
	if appErr := a.requireViewAccess(ctx, user.ID, viewID); appErr != nil {
		return appErr
	}

	if !a.canReadTable(ctx, user, workspaceID, tableID) || !a.ProjectWorkspaceTableHasPermission(user, workspaceID, tableID, model.PermissionUpdateTask) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.requireTaskAccess(user, workspaceID, tableID, taskID); appErr != nil {
		return appErr
	}

	invalid := model.NewAppError("request.invalid", http.StatusBadRequest)
	if !validLinkID(taskID) || columnID == "" || (afterID != "" && !validLinkID(afterID)) || afterID == taskID {
		return invalid
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	view, err := a.Store.Workspace.GetView(ctx, viewID)
	if err != nil || view.TableID != tableID || view.ViewType != "kanban" {
		return model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	}

	// The grouping field is written into the column filters' SQL, so it must
	// be one of the table's columns.
	var grouping kanbanViewOrder
	if err := json.Unmarshal([]byte(view.TaskOrder), &grouping); err != nil {
		return invalid
	}

	if has, err := a.Store.Workspace.TableHasColumn(tableName, grouping.Section); err != nil || !has {
		return invalid
	}

	_, err = a.Store.Workspace.ChangeViewOrder(ctx, viewID, workspaceID, tableID, func(stored string) (string, error) {
		var shape kanbanViewOrder
		if err := json.Unmarshal([]byte(stored), &shape); err != nil {
			return "", err
		}

		rest := func(placed []string) ([]string, error) {
			groups := kanbanColumnFilters(shape)
			for i, field := range shape.Fields {
				if field.ID != columnID {
					continue
				}

				cards := model.SQLFilter{SQL: kanbanCardsSQL}.And(groups[i]).And(notInFilter("main.id", placed))
				return a.Store.Workspace.GetKanbanRestIDs(ctx, tableName, cards, afterID, kanbanSavedOrderLimit)
			}

			return nil, errUnknownKanbanColumn
		}

		return moveKanbanCard(stored, taskID, columnID, afterID, rest)
	})
	switch {
	case errors.Is(err, errUnknownKanbanColumn):
		return invalid
	case errors.Is(err, sql.ErrNoRows):
		return model.NewAppError("workspace.view_or_field_not_found", http.StatusNotFound)
	case err != nil:
		tlog.Errorw("Failed to move a Kanban card", "view_id", viewID, "task_id", taskID, "error", err)
		return model.NewAppError("workspace.view_update_failed", http.StatusInternalServerError)
	}

	a.publishViewChange(ctx, viewID, projectChange{kind: "KANBAN_CARD_MOVED", workspaceID: workspaceID, tableID: tableID,
		data: map[string]any{"view_id": viewID, "task_id": taskID, "column": columnID, "after_id": afterID}})
	return nil
}
