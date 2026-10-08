// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/twigex/twigex/model"
)

// taskCursor marks the last task a page showed, so the next page starts
// after it rather than skipping over every task before it. Value is the
// task's value in the column the page is sorted by. Saved is its place in a
// Kanban column's saved order, when it has one.
type taskCursor struct {
	ID    string `json:"i"`
	Value *int64 `json:"v,omitempty"`
	Saved *int   `json:"s,omitempty"`
}

func (c taskCursor) encode() string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

// decodeTaskCursor reads a cursor a page returned. An empty one is the
// first page, and gives nil.
func decodeTaskCursor(s string) (*taskCursor, *model.AppError) {
	if s == "" {
		return nil, nil
	}

	var c taskCursor
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		err = json.Unmarshal(b, &c)
	}

	if err == nil && c.ID == "" {
		err = errors.New("the cursor names no task")
	}

	if err != nil {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	return &c, nil
}

func (c taskCursor) after(column string) model.SQLFilter {
	return model.RowsAfter(column, c.Value, c.ID)
}

// cursorValue reads a whole number column from a scanned row, which holds it
// as a number or as text depending on how the query was sent.
func cursorValue(v any) *int64 {
	var n int64
	switch v := v.(type) {
	case int64:
		n = v
	case int:
		n = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil
		}

		n = parsed
	default:
		return nil
	}

	return &n
}

func notInFilter(column string, values []string) model.SQLFilter {
	if len(values) == 0 {
		return model.SQLFilter{}
	}

	args := make([]any, len(values))
	for i, v := range values {
		args[i] = v
	}

	return model.SQLFilter{SQL: column + " NOT IN (" + sqlPlaceholderList(len(values)) + ")", Args: args}
}
