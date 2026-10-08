// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// taskExportChunk is how many rows are filled in and written at a time.
const taskExportChunk = 1000

// taskReportColumns are the task report's own columns, in its order, with
// the labels used when the page does not send its own.
var taskReportColumns = []struct{ name, label string }{
	{"name", "Name"},
	{"start_date", "Start date"},
	{"due_date", "Due date"},
	{"assignee", "Assignee"},
	{"status", "Status"},
	{"updated_at", "Updated at"},
	{"workspace_name", "Workspace"},
	{"table_name", "Table"},
	{"link_to_table", "Link"},
}

// taskReportSortable are the report's columns that every task table has.
var taskReportSortable = map[string]bool{
	"name":       true,
	"start_date": true,
	"due_date":   true,
	"assignee":   true,
	"status":     true,
	"updated_at": true,
}

// taskExportSkipped are columns that mean nothing outside the app.
var taskExportSkipped = map[string]bool{"id": true, "deleted_at": true, "parent_task_id": true}

type taskExportTable struct {
	table   model.WorkspaceTable
	name    string
	link    string
	headers map[string]model.WorkspaceHeaders
	list    []model.WorkspaceHeaders
	filter  model.SQLFilter
}

// TaskExport is a prepared CSV export of the task report. Preparing it checks
// access and reads the tables' structure, so problems are reported before any
// of the file is sent.
type TaskExport struct {
	a          *App
	tables     []taskExportTable
	columns    []string
	labels     []string
	orderSQL   string
	loc        *time.Location
	dateLayout string
	tableNames map[string]string
}

// PrepareTaskExport readies the task report's rows for the given workspaces
// and filters: the report's columns first, labelled from labels where it has
// them, then every other field. Dates are in timezone, laid out the way clock
// ("12h" or "24h") asks for.
func (a *App) PrepareTaskExport(ctx context.Context, workspaceIDs []string, user model.User, filters model.FilterPayload, timezone, clock string, labels map[string]string) (*TaskExport, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	metas, err := a.Store.Workspace.GetTableMetas(workspaceIDs)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace tables for export", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("workspace.tasks_retrieval_failed", http.StatusInternalServerError)
	}

	filters, appErr := a.widenStatusFilters(filters, workspaceIDs)
	if appErr != nil {
		return nil, appErr
	}

	if timezone == "" {
		timezone = a.resolveUserTimezone(user.ID)
	}

	loc, lerr := time.LoadLocation(timezone)
	if lerr != nil || timezone == "" {
		loc = time.UTC
	}

	export := &TaskExport{a: a, loc: loc, dateLayout: "02.01.2006", tableNames: map[string]string{}}
	if clock == "12h" {
		export.dateLayout = "01/02/2006"
	}
	// One order runs on every table in the report, so only the columns every
	// task table has can be sorted on.
	export.orderSQL = buildSortSQL(filters.Sort, func(name string) bool { return taskReportSortable[name] })
	if export.orderSQL == buildSortSQL(nil, nil) {
		export.orderSQL = "ORDER BY (main.due_date IS NULL OR main.due_date = 0), main.due_date, main.created_at, main.id"
	}

	seen := map[string]bool{}
	for _, c := range taskReportColumns {
		seen[c.name] = true
		export.columns = append(export.columns, c.name)
		label := labels[c.name]
		if label == "" {
			label = c.label
		}

		export.labels = append(export.labels, label)
	}

	workspaceByTable := make(map[string]string, len(metas))
	for _, t := range metas {
		workspaceByTable[t.ID] = t.WorkspaceID
	}

	viewIDs, err := a.Store.Workspace.GetMainViewIDs(workspaceByTable)
	if err != nil {
		tlog.Warnw("Failed to load main views for the task export", "user_id", user.ID, "error", err)
	}

	siteURL := ""
	if cfg := a.ConfigStore.Config; cfg != nil && cfg.ServerSettings.SiteURL != nil {
		siteURL = strings.TrimSuffix(*cfg.ServerSettings.SiteURL, "/")
	}

	for _, t := range metas {
		colTypes, err := a.Store.Workspace.GetTableColumnTypes(t.Name)
		if err != nil {
			tlog.Warnw("Skipping a table in the task export", "table_id", t.ID, "error", err)
			continue
		}

		list := a.buildTableHeaders(t.ID, colTypes, true, true)

		et := taskExportTable{
			table:   t,
			name:    exportTableName(t),
			headers: map[string]model.WorkspaceHeaders{},
			list:    list,
			filter:  a.filterSQLForHeaders(t.ID, list, filters, loc),
		}
		if siteURL != "" {
			et.link = fmt.Sprintf("%s/projects/%s/grid/%s/view/%s", siteURL, url.PathEscape(t.WorkspaceID), url.PathEscape(t.ID), url.PathEscape(viewIDs[t.ID]))
		}

		for _, h := range list {
			et.headers[h.Name] = h
			if seen[h.Name] || taskExportSkipped[h.Name] || h.HeaderUsage == "calculations" {
				continue
			}

			seen[h.Name] = true
			export.columns = append(export.columns, h.Name)
			label := h.DisplayName
			if label == "" {
				label = h.Name
			}

			export.labels = append(export.labels, label)
		}

		export.tables = append(export.tables, et)
	}

	return export, nil
}

// FileName names the export after its workspace when it covers only one, with
// the day it was made.
func (e *TaskExport) FileName() string {
	name := "Task report"
	if len(e.tables) > 0 {
		name = e.tables[0].table.WorkspaceName
		for _, t := range e.tables[1:] {
			if t.table.WorkspaceID != e.tables[0].table.WorkspaceID {
				name = "Task report"
				break
			}
		}
	}

	return name + " " + time.Now().In(e.loc).Format("2006-01-02") + ".csv"
}

func exportTableName(t model.WorkspaceTable) string {
	if t.DisplayName.Valid && t.DisplayName.String != "" {
		return t.DisplayName.String
	}

	if idx := strings.Index(t.Name, "_"); idx >= 0 {
		return t.Name[idx+1:]
	}

	return t.Name
}

// WriteCSV writes the export to w, table by table, filling in and writing a
// chunk of rows at a time so memory does not grow with the number of rows.
// flush, when set, is called after each chunk so the rows reach the client.
func (e *TaskExport) WriteCSV(ctx context.Context, w io.Writer, flush func()) error {
	if _, err := io.WriteString(w, "\xEF\xBB\xBF"); err != nil {
		return err
	}

	out := csv.NewWriter(w)
	if err := out.Write(e.labels); err != nil {
		return err
	}

	for _, t := range e.tables {
		chunk := make([]map[string]interface{}, 0, taskExportChunk)
		send := func() error {
			if len(chunk) == 0 {
				return nil
			}

			if err := e.writeChunk(ctx, out, t, chunk); err != nil {
				return err
			}

			out.Flush()
			if err := out.Error(); err != nil {
				return err
			}

			if flush != nil {
				flush()
			}

			chunk = chunk[:0]
			return nil
		}

		err := e.a.Store.Workspace.StreamTableRows(ctx, t.table.Name, t.filter, e.orderSQL, func(row map[string]interface{}) error {
			chunk = append(chunk, row)
			if len(chunk) < taskExportChunk {
				return nil
			}

			return send()
		})
		if err == nil {
			err = send()
		}

		if err != nil {
			return fmt.Errorf("export table %s: %w", t.table.ID, err)
		}
	}

	out.Flush()
	return out.Error()
}

// exportLookups are the names a chunk of rows refers to by id.
type exportLookups struct {
	users map[string]string
	files map[string]map[string][]string // field name to task id to file names
}

func (e *TaskExport) writeChunk(ctx context.Context, out *csv.Writer, t taskExportTable, rows []map[string]interface{}) error {
	e.a.hydrateLinkedAndSingleSelect(ctx, rows, t.list)
	lookups, err := e.lookups(ctx, t, rows)
	if err != nil {
		return err
	}

	record := make([]string, len(e.columns))
	for _, row := range rows {
		for i, name := range e.columns {
			switch name {
			case "workspace_name":
				record[i] = t.table.WorkspaceName
			case "table_name":
				record[i] = t.name
			case "link_to_table":
				record[i] = t.link
			default:
				h, ok := t.headers[name]
				if !ok {
					record[i] = ""
					continue
				}

				record[i] = e.cell(h, row, lookups)
			}

			record[i] = inertCell(record[i])
		}

		if err := out.Write(record); err != nil {
			return err
		}
	}

	return ctx.Err()
}

func (e *TaskExport) lookups(ctx context.Context, t taskExportTable, rows []map[string]interface{}) (exportLookups, error) {
	l := exportLookups{users: map[string]string{}, files: map[string]map[string][]string{}}
	var userIDs, taskIDs, tableIDs []string
	for _, row := range rows {
		if id, _ := row["id"].(string); id != "" {
			taskIDs = append(taskIDs, id)
		}
	}

	for name, h := range t.headers {
		for _, row := range rows {
			v, _ := row[name].(string)
			if v == "" {
				continue
			}

			switch {
			case isUserHeader(h):
				if _, known := l.users[v]; !known {
					l.users[v] = v
					userIDs = append(userIDs, v)
				}
			case h.HeaderUsage == "master link":
				if _, known := e.tableNames[v]; !known {
					e.tableNames[v] = ""
					tableIDs = append(tableIDs, v)
				}
			}
		}

		if h.HeaderUsage == "file" && len(taskIDs) > 0 {
			names, err := e.a.Store.Workspace.GetAttachmentNames(ctx, t.table.ID, h.ID, taskIDs)
			if err != nil {
				return l, err
			}

			l.files[name] = names
		}
	}

	if len(userIDs) > 0 {
		users, err := e.a.Store.User.GetByIDs(userIDs)
		if err != nil {
			return l, err
		}

		for _, u := range users {
			l.users[u.ID] = strings.TrimSpace(u.Name + " " + u.LastName)
		}
	}

	if len(tableIDs) > 0 {
		tables, err := e.a.Store.Workspace.GetTablesByIDs(tableIDs)
		if err != nil {
			return l, err
		}

		for _, lt := range tables {
			e.tableNames[lt.ID] = exportTableName(lt)
		}
	}

	return l, nil
}

func isUserHeader(h model.WorkspaceHeaders) bool {
	return h.HeaderUsage == "assignee" || h.HeaderUsage == "default_assignee" || h.Name == "assignee" || h.Name == "created_by"
}

// inertCell keeps a spreadsheet from running a typed value as a formula, as
// Excel does with a cell starting =, +, - or @. A number is left as it is.
func inertCell(s string) string {
	if s == "" || !strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return s
	}

	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s
	}

	return "'" + s
}

func (e *TaskExport) cell(h model.WorkspaceHeaders, row map[string]interface{}, l exportLookups) string {
	if h.HeaderUsage == "file" {
		id, _ := row["id"].(string)
		return strings.Join(l.files[h.Name][id], ", ")
	}

	switch v := row[h.Name].(type) {
	case nil:
		return ""
	case *model.TaskOrderField:
		if v == nil {
			return ""
		}

		return v.Name
	case []model.LinkedItem:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, item.Name)
		}

		return strings.Join(parts, ", ")
	case time.Time:
		if v.IsZero() {
			return ""
		}

		return v.Format(e.dateLayout)
	}

	raw := fmt.Sprint(row[h.Name])
	switch {
	case isUserHeader(h):
		return l.users[raw]
	case h.HeaderUsage == "master link":
		return e.tableNames[raw]
	case h.HeaderType == "DATE":
		if d, err := time.Parse("2006-01-02", raw); err == nil {
			return d.Format(e.dateLayout)
		}

		return raw
	case isDateHeader(h):
		return e.formatDate(raw)
	case h.HeaderType == "TINYINT":
		return strconv.FormatBool(raw == "1" || raw == "true")
	}

	return raw
}

func isDateHeader(h model.WorkspaceHeaders) bool {
	switch h.Name {
	case "start_date", "due_date", "updated_at", "created_at":
		return true
	}

	return h.HeaderUsage == "default_date"
}

// formatDate formats a unix time in seconds or milliseconds the way the task
// report shows dates.
func (e *TaskExport) formatDate(raw string) string {
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n == 0 {
		return ""
	}

	if n > 1e12 {
		return time.UnixMilli(n).In(e.loc).Format(e.dateLayout)
	}

	return time.Unix(n, 0).In(e.loc).Format(e.dateLayout)
}
