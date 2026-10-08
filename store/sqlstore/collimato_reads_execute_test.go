// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// Without a row, the QueryRow reads return none, Scan never runs, and the
// column list goes unexercised.
func seedCollimatoRow(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("seed: %v", err)
	}
}

// These used to select star and scan positionally. A column name that is wrong
// or misplaced only shows when the statement runs; nothing asserts on results.
func TestCollimatoReadsExecute(t *testing.T) {
	db := requireDB(t)
	cleanTables(t,
		"collimato_workspace_users", "collimato_workspace_roles",
		"collimato_charts", "collimato_dashboard_filters",
		"collimato_dashboards", "collimato_workspaces", "users")
	repo := &collimatoRepository{Db: db}
	ctx := context.Background()

	ws := seedCollimatoWorkspace(t, db)
	user := seedUser(t, db)
	seedCollimatoWorkspaceUser(t, db, ws, user, "admin")

	now := time.Now().Unix()
	chart, dash, filter, role := "chart-1", "dash-1", "f-1", "role-1"

	seedCollimatoRow(t, db, `INSERT INTO collimato_charts
		(id, name, chart_type, configuration, data, owner_id, workspace_id, created_at, updated_at, deleted_at)
		VALUES (?, 'Chart', 'bar', '{}', '{}', ?, ?, ?, ?, 0)`,
		chart, user, ws, now, now)

	seedCollimatoRow(t, db, `INSERT INTO collimato_dashboards
		(id, title, description, owner_id, workspace_id, created_at, updated_at, deleted_at)
		VALUES (?, 'Dash', '', ?, ?, ?, ?, 0)`,
		dash, user, ws, now, now)

	seedCollimatoRow(t, db, `INSERT INTO collimato_dashboard_filters
		(id, dashboard_id, name, table_name, column_name, filter_values, apply_to, created_at, updated_at, deleted_at, operator)
		VALUES (?, ?, 'Filter', 'tbl', 'col', '[]', '[]', ?, ?, 0, 'equals')`,
		filter, dash, now, now)

	seedCollimatoRow(t, db, `INSERT INTO collimato_workspace_roles
		(id, workspace_id, name, displayname, description, permissions, table_level_permissions,
		 column_level_permissions, row_level_permissions, auto_update, created_at, updated_at)
		VALUES (?, ?, 'admin', 'Admin', '', '[]', '[]', '[]', '[]', 0, ?, ?)`,
		role, ws, now, now)

	reads := map[string]func() error{
		"GetWorkspaces":          func() error { _, err := repo.GetWorkspaces(); return err },
		"GetWorkspaceByID":       func() error { _, err := repo.GetWorkspaceByID(ws); return err },
		"GetWorkspacesForUser":   func() error { _, err := repo.GetWorkspacesForUser(user); return err },
		"CountWorkspaces":        func() error { _, err := repo.CountWorkspaces(ctx); return err },
		"GetWorkspaceUsers":      func() error { _, err := repo.GetWorkspaceUsers(ws); return err },
		"GetWorkspaceUserByID":   func() error { _, err := repo.GetWorkspaceUserByUserID(user, ws); return err },
		"GetWorkspaceRoles":      func() error { _, err := repo.GetWorkspaceRoles(ws); return err },
		"GetWorkspaceRoleByID":   func() error { _, err := repo.GetWorkspaceRoleByID(ws, role); return err },
		"GetRolesByName":         func() error { _, err := repo.GetRolesByName([]string{"admin"}, ws); return err },
		"GetCharts":              func() error { _, err := repo.GetCharts(ws); return err },
		"GetChartByID":           func() error { _, err := repo.GetChartByID(chart); return err },
		"GetDashboards":          func() error { _, err := repo.GetDashboards(ws); return err },
		"GetDashboardByID":       func() error { _, err := repo.GetDashboardByID(dash); return err },
		"GetDashboardFilterByID": func() error { _, err := repo.GetDashboardFilterByID(dash, filter); return err },
		"GetWorkspaceGroups":     func() error { _, err := repo.GetWorkspaceGroups(ctx, ws); return err },
		"GetEffectiveRoles":      func() error { _, err := repo.GetEffectiveRolesForUser(ctx, ws, user); return err },
		"GetWorkspaceFiles":      func() error { _, err := repo.GetWorkspaceFiles(ws, "model"); return err },
		"GetConnections":         func() error { _, err := repo.GetConnections(ws); return err },
	}

	for name, run := range reads {
		t.Run(name, func(t *testing.T) {
			if err := run(); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		})
	}
}
