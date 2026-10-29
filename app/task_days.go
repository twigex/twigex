// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// taskDay turns what a client sent for a start or due date into the day it
// stores, "" for none.
func taskDay(value string) (string, bool) {
	if value == "" || value == "0" || value == "null" {
		return "", true
	}

	if day, err := time.Parse(time.DateOnly, value); err == nil {
		return day.Format(time.DateOnly), true
	}

	if len(value) > len(time.DateOnly) && value[len(time.DateOnly)] == 'T' {
		if day, err := time.Parse(time.DateOnly, value[:len(time.DateOnly)]); err == nil {
			return day.Format(time.DateOnly), true
		}
	}

	// TODO: remove the Unix time below a release after dates became days. Only
	// a tab opened before then still sends one, at its user's midnight, which
	// rounds to the nearest day in UTC.
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds < 0 {
		return "", false
	}

	if seconds > 1e12 {
		seconds /= 1000
	}

	days := math.Round(float64(seconds) / 86400)
	day := time.Unix(int64(days)*86400, 0).UTC()
	if day.Year() > 9999 {
		return "", false
	}

	return day.Format(time.DateOnly), true
}

// checkTaskDatesInOrder refuses a start date after the task's due date, or a
// due date before its start date. A task whose dates are already the wrong
// way round can still be saved until one of them changes.
func (a *App) checkTaskDatesInOrder(tableName, taskID, field, day string) *model.AppError {
	if day == "" {
		return nil
	}

	other := "due_date"
	if field == "due_date" {
		other = "start_date"
	}

	stored, err := a.Store.Workspace.GetTaskFieldValue(tableName, taskID, other)
	if err != nil {
		tlog.Errorw("Failed to read the task's other date", "table", tableName, "task_id", taskID, "field", other, "error", err)
		return model.NewAppError("workspace.task_update_failed", http.StatusInternalServerError)
	}

	otherDay, ok := taskDay(stored)
	if !ok || otherDay == "" {
		return nil
	}

	start, due := day, otherDay
	if field == "due_date" {
		start, due = otherDay, day
	}

	if start > due {
		return model.NewAppError("workspace.task_dates_reversed", http.StatusBadRequest)
	}

	return nil
}
