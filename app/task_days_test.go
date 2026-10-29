// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestTaskDayKeepsTheDayItWasSetFor(t *testing.T) {
	for _, tc := range []struct {
		name, value, want string
		ok                bool
	}{
		{"a day", "2026-10-28", "2026-10-28", true},
		{"a day as the server sends it", "2026-10-28T00:00:00Z", "2026-10-28", true},
		{"no date", "", "", true},
		{"no date from an older tab", "0", "", true},
		{"midnight in Riga", "1793138400", "2026-10-28", true},
		{"midnight in Los Angeles", "1793170800", "2026-10-28", true},
		{"midnight in Almaty", "1793127600", "2026-10-28", true},
		{"midnight in Riga in milliseconds", "1793138400000", "2026-10-28", true},
		{"a day written another way", "28.10.2026", "", false},
		{"a negative time", "-1", "", false},
		{"a time past year 9999", "9000000000000000", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := taskDay(tc.value)
			if got != tc.want || ok != tc.ok {
				t.Errorf("taskDay(%q) = %q, %v, want %q, %v", tc.value, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestNewTaskFieldsTakeADayOrAnOlderTabsTime(t *testing.T) {
	var fields model.NewTaskFields
	if err := json.Unmarshal([]byte(`{"start_date": "2026-10-27", "due_date": 1793138400}`), &fields); err != nil {
		t.Fatal(err)
	}

	if fields.StartDate != "2026-10-27" || fields.DueDate != "1793138400" {
		t.Errorf("fields = %+v, want the day and the time as text", fields)
	}
}

func TestTasksSortByDueDayThenByWhenTheyWereCreated(t *testing.T) {
	early := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	items := []map[string]interface{}{
		{"id": "later", "due_date": "2026-10-28"},
		{"id": "undated", "due_date": nil, "created_at": early.Add(48 * time.Hour).Unix()},
		{"id": "first", "due_date": early},
	}

	sortItemsByDueDate(items)

	var order []string
	for _, item := range items {
		order = append(order, item["id"].(string))
	}

	if order[0] != "first" || order[1] != "undated" || order[2] != "later" {
		t.Errorf("order = %v, want first, undated, later", order)
	}
}

func TestAssignedSectionsUseEachUsersOwnToday(t *testing.T) {
	riga, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		t.Skip("no time zone data")
	}

	losAngeles, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Skip("no time zone data")
	}

	now := time.Date(2026, 10, 28, 1, 0, 0, 0, riga)
	for _, tc := range []struct {
		loc   *time.Location
		today string
	}{
		{riga, "2026-10-28"},
		{losAngeles, "2026-10-27"},
	} {
		sections := assignedSections(now.In(tc.loc))
		if got := sections[1].filter.Args[0]; got != tc.today {
			t.Errorf("today in %s starts on %v, want %s", tc.loc, got, tc.today)
		}
	}
}

func TestEditingADateHandsTheStoreTheDay(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		want        any
	}{
		{"a day", "2026-10-28", "2026-10-28"},
		{"an older tab's Unix time", "1793138400", "2026-10-28"},
		{"a leap day", "2028-02-29", "2028-02-29"},
		{"clearing it", "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			if _, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", "due_date", tc.value, workspaceAdmin, ""); appErr != nil {
				t.Fatal(appErr)
			}

			if ws.writes != 1 || ws.written != tc.want {
				t.Errorf("wrote %v in %d writes, want %v", ws.written, ws.writes, tc.want)
			}
		})
	}

	for _, value := range []string{"28.10.2026", "2026-02-30", "soon"} {
		a, ws := taskWriteApp()
		_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", "due_date", value, workspaceAdmin, "")
		if appErr == nil || appErr.Status != http.StatusBadRequest || ws.writes != 0 {
			t.Errorf("%q: error %v after %d writes, want a bad request and no write", value, appErr, ws.writes)
		}
	}
}

func TestEditingADateKeepsStartBeforeDue(t *testing.T) {
	stored := map[string]string{"start_date": "2026-10-20T00:00:00Z", "due_date": "2026-10-28"}
	for _, tc := range []struct {
		name, field, value string
		stored             map[string]string
		want               int
	}{
		{"a due date after the start", "due_date", "2026-10-30", stored, http.StatusOK},
		{"a due date on the start day", "due_date", "2026-10-20", stored, http.StatusOK},
		{"a due date before the start", "due_date", "2026-10-19", stored, http.StatusBadRequest},
		{"a start date before the due date", "start_date", "2026-10-01", stored, http.StatusOK},
		{"a start date after the due date", "start_date", "2026-10-29", stored, http.StatusBadRequest},
		{"clearing the start date", "start_date", "", stored, http.StatusOK},
		{"a start date with no due date", "start_date", "2026-12-01", map[string]string{"due_date": ""}, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ws := taskWriteApp()
			ws.stored = tc.stored
			_, appErr := a.UpdateWorkspaceTask(context.Background(), "ws", "t-a", "task-1", tc.field, tc.value, workspaceAdmin, "")

			if tc.want == http.StatusOK && (appErr != nil || ws.writes != 1) {
				t.Fatalf("error = %v after %d writes, want the date saved", appErr, ws.writes)
			}

			if tc.want != http.StatusOK && (appErr == nil || appErr.ID != "workspace.task_dates_reversed" || ws.writes != 0) {
				t.Fatalf("error = %v after %d writes, want the dates refused as reversed", appErr, ws.writes)
			}
		})
	}
}
