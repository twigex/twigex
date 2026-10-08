// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"strings"
	"testing"
)

// Running the statements is not enough: two columns of the same type swapped in
// a list still execute, still scan, and put each value in the other's field.
// Comparing against information_schema is what catches that.
func TestColumnListsMatchTheSchema(t *testing.T) {
	db := requireDB(t)

	// The file and post attachment lists are excluded: they carry a derived
	// expression rather than only column names.
	lists := map[string]string{
		"password_resets":             passwordResetColumns,
		"channels":                    channelColumns,
		"status":                      statusColumns,
		"jobs":                        jobColumns,
		"notifications":               notificationColumns,
		"file_metadata":               fileMetadataColumns,
		"posts":                       postColumns,
		"collimato_workspace_users":   collimatoWorkspaceUserColumns,
		"collimato_workspace_roles":   collimatoWorkspaceRoleColumns,
		"collimato_charts":            collimatoChartColumns,
		"collimato_dashboards":        collimatoDashboardColumns,
		"collimato_dashboard_filters": collimatoDashboardFilterColumns,
		"collimato_workspaces":        collimatoWorkspaceColumns,
		"collimato_dashboard_charts":  collimatoDashboardChartColumns,
		"activity":                    activityColumns,
		"licenses":                    licenseColumns,
		"app_migrations":              appMigrationColumns,
		"file_metadata_entries":       fileMetadataEntryColumns,
	}

	for table, list := range lists {
		t.Run(table, func(t *testing.T) {
			rows, err := db.Query(
				`SELECT COLUMN_NAME FROM information_schema.COLUMNS
				 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
				 ORDER BY ORDINAL_POSITION`, table)
			if err != nil {
				t.Fatalf("read schema: %v", err)
			}
			defer rows.Close()

			actual := make([]string, 0)
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err != nil {
					t.Fatalf("scan: %v", err)
				}
				actual = append(actual, name)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("read schema: %v", err)
			}
			if len(actual) == 0 {
				t.Fatalf("no such table")
			}

			declared := make([]string, 0)
			for _, c := range strings.Split(list, ",") {
				declared = append(declared, strings.TrimSpace(c))
			}

			if strings.Join(declared, ",") != strings.Join(actual, ",") {
				t.Errorf("column list does not match the table\n  declared: %s\n  actual:   %s",
					strings.Join(declared, ", "), strings.Join(actual, ", "))
			}
		})
	}
}
