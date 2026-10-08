// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

type collimatoRepository struct {
	Db *sql.DB
}

const (
	collimatoWorkspaceUserColumns = `id, workspace_id, user_id, role, created_at, updated_at, deleted_at`
	collimatoWorkspaceRoleColumns = `id, workspace_id, name, displayname, description, permissions, ` +
		`table_level_permissions, column_level_permissions, row_level_permissions, auto_update, created_at, updated_at`
	collimatoChartColumns = `id, name, chart_type, configuration, data, owner_id, ` +
		`workspace_id, created_at, updated_at, deleted_at`
	collimatoDashboardColumns       = `id, title, description, owner_id, workspace_id, created_at, updated_at, deleted_at`
	collimatoDashboardFilterColumns = `id, dashboard_id, name, table_name, column_name, filter_values, ` +
		`apply_to, created_at, updated_at, deleted_at, operator`
	collimatoWorkspaceColumns = `id, name, description, status, server_id, created_by, created_at, ` +
		`updated_at, deleted_at, secret`
	collimatoDashboardChartColumns = `id, dashboard_id, chart_id, created_by, position, created_at, updated_at, deleted_at`
)

func NewCollimatoRepository(db *sql.DB) (*collimatoRepository, error) {
	return &collimatoRepository{Db: db}, nil
}

func (c *collimatoRepository) CreateConnection(conn model.NewConnection) (*model.Connection, error) {
	id := model.NewID()
	t := time.Now().Unix()

	_, err := c.Db.Exec(`INSERT INTO collimato_workspace_connections
		(id, workspace_id, display_name, type, host, port, database_name,
		username, password, ssl_mode, config, encrypted_config, is_default,
		created_by, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, conn.WorkspaceID, conn.DisplayName, conn.Type, conn.Host, conn.Port,
		conn.Database, conn.Username, conn.Password, conn.SSLMode, conn.Config,
		conn.EncryptedConfig, conn.Default, conn.CreatedBy, t, t, 0)
	if err != nil {
		return nil, err
	}

	return &model.Connection{
		ID:          id,
		WorkspaceID: conn.WorkspaceID,
		Type:        conn.Type,
		DisplayName: conn.DisplayName,
		Host:        conn.Host,
		Port:        conn.Port,
		Username:    conn.Username,
		Password:    conn.Password,
		Database:    conn.Database,
		SSLMode:     conn.SSLMode,
		Config:      conn.Config,
		Default:     conn.Default,
		CreatedBy:   conn.CreatedBy,
		CreatedAt:   t,
		UpdatedAt:   t,
		DeletedAt:   0,
	}, nil
}

func (c *collimatoRepository) GetConnections(workspace string) ([]model.Connection, error) {
	rows, err := c.Db.Query(`SELECT id, workspace_id, display_name, type, host, port, database_name,
		username, password, ssl_mode, config, encrypted_config, is_default,
		created_by, created_at, updated_at, deleted_at
		FROM collimato_workspace_connections WHERE workspace_id = ?`, workspace)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	conns := make([]model.Connection, 0)
	for rows.Next() {
		var conn model.Connection
		err = rows.Scan(&conn.ID, &conn.WorkspaceID, &conn.DisplayName, &conn.Type,
			&conn.Host, &conn.Port, &conn.Database, &conn.Username, &conn.Password,
			&conn.SSLMode, &conn.Config, &conn.EncryptedConfig, &conn.Default,
			&conn.CreatedBy, &conn.CreatedAt, &conn.UpdatedAt, &conn.DeletedAt)
		if err != nil {
			return nil, err
		}

		conns = append(conns, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conns, nil
}

func (c *collimatoRepository) GetConnectionByID(id string) (*model.Connection, error) {
	conn := &model.Connection{}
	err := c.Db.QueryRow(`SELECT id, workspace_id, display_name, type, host, port, database_name,
		username, password, ssl_mode, config, encrypted_config, is_default,
		created_by, created_at, updated_at, deleted_at
		FROM collimato_workspace_connections WHERE id = ?`, id).Scan(&conn.ID, &conn.WorkspaceID, &conn.DisplayName, &conn.Type,
		&conn.Host, &conn.Port, &conn.Database, &conn.Username, &conn.Password,
		&conn.SSLMode, &conn.Config, &conn.EncryptedConfig, &conn.Default,
		&conn.CreatedBy, &conn.CreatedAt, &conn.UpdatedAt, &conn.DeletedAt)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *collimatoRepository) UpdateConnection(conn model.Connection) error {
	_, err := c.Db.Exec(`UPDATE collimato_workspace_connections
		SET display_name = ?, type = ?, host = ?, port = ?, database_name = ?,
			username = ?, password = ?, ssl_mode = ?, config = ?,
			encrypted_config = ?, is_default = ?, updated_at = ?
		WHERE id = ?`,
		conn.DisplayName, conn.Type, conn.Host, conn.Port, conn.Database,
		conn.Username, conn.Password, conn.SSLMode, conn.Config,
		conn.EncryptedConfig, conn.Default, time.Now().Unix(), conn.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) DeleteConnection(id string) error {
	_, err := c.Db.Exec("DELETE FROM collimato_workspace_connections WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) AddWorkspaceUsers(workspaceID string, users []string, roles []string) error {
	if len(users) == 0 {
		return nil
	}

	t := time.Now().Unix()
	rolesStr := strings.Join(roles, ",")

	args := make([]any, 0, len(users)*7)
	for _, user := range users {
		args = append(args, model.NewID(), workspaceID, user, rolesStr, t, t, 0)
	}

	_, err := c.Db.Exec(
		`INSERT INTO collimato_workspace_users (`+collimatoWorkspaceUserColumns+`)
		 VALUES `+rowPlaceholders(len(users), 7)+`
		 ON DUPLICATE KEY UPDATE
		     role = IF(deleted_at = 0, role, VALUES(role)),
		     updated_at = VALUES(updated_at),
		     deleted_at = 0`, args...)

	return err
}

func (c *collimatoRepository) RemoveWorkspaceUser(workspaceID string, userID string) error {
	_, err := c.Db.Exec("UPDATE collimato_workspace_users SET deleted_at = ? WHERE workspace_id = ? AND user_id = ?", time.Now().Unix(), workspaceID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) GetWorkspaceUsers(workspaceID string) ([]model.CollimatoWorkspaceUser, error) {
	rows, err := c.Db.Query(`SELECT `+userColumns("users")+`,
	                                collimato_workspace_users.id, collimato_workspace_users.workspace_id,
	                                collimato_workspace_users.user_id, collimato_workspace_users.role,
	                                collimato_workspace_users.created_at, collimato_workspace_users.updated_at,
	                                collimato_workspace_users.deleted_at
	                         FROM users, collimato_workspace_users
	                         WHERE collimato_workspace_users.workspace_id = ?
	                           AND users.id = collimato_workspace_users.user_id
	                           AND collimato_workspace_users.deleted_at = 0`, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.CollimatoWorkspaceUser, 0)
	for rows.Next() {
		user := &model.User{}
		var workspaceUser model.CollimatoWorkspaceUser
		var timezone []byte
		err = rows.Scan(&user.ID, &user.Email, &user.Role,
			&user.Password, &user.AuthService, &user.Username,
			&user.Name, &user.LastName, &user.StorageLimit, &user.Photo,
			&timezone, &user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData,
			&workspaceUser.ID, &workspaceUser.WorkspaceID, &workspaceUser.UserID, &workspaceUser.Role,
			&workspaceUser.CreatedAt, &workspaceUser.UpdatedAt, &workspaceUser.DeletedAt)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
			return nil, err
		}

		workspaceUser.UserInfo = *user

		users = append(users, workspaceUser)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *collimatoRepository) GetWorkspaceUserByUserID(userID string, workspaceID string) (*model.CollimatoWorkspaceUser, error) {
	workspaceUser := &model.CollimatoWorkspaceUser{}
	err := c.Db.QueryRow("SELECT "+collimatoWorkspaceUserColumns+" FROM collimato_workspace_users WHERE user_id =? AND deleted_at = 0 AND workspace_id =? ", userID, workspaceID).Scan(&workspaceUser.ID, &workspaceUser.WorkspaceID, &workspaceUser.UserID, &workspaceUser.Role, &workspaceUser.CreatedAt, &workspaceUser.UpdatedAt, &workspaceUser.DeletedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return workspaceUser, nil
}

func (c *collimatoRepository) UpdateUserRoles(workspaceID, userID string, roles []string) error {
	rolesStr := strings.Join(roles, ",")

	_, err := c.Db.Exec("UPDATE collimato_workspace_users SET role = ? WHERE workspace_id = ? AND id = ?", rolesStr, workspaceID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) GetRolesByName(roleNames []string, workspaceID string) ([]model.CollimatoRole, error) {
	if len(roleNames) == 0 {
		return []model.CollimatoRole{}, nil
	}

	args := make([]interface{}, len(roleNames)+1)
	for i, name := range roleNames {
		args[i] = name
	}

	args[len(roleNames)] = workspaceID

	stmt := `SELECT ` + collimatoWorkspaceRoleColumns + ` FROM collimato_workspace_roles WHERE name IN (?` + strings.Repeat(",?", len(roleNames)-1) + `) AND workspace_id = ?`

	rows, err := c.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	roles := make([]model.CollimatoRole, 0)
	for rows.Next() {
		var role model.CollimatoRole
		var permissions string
		var table_permissions []byte
		var column_permissions []byte
		var row_level_permissions []byte

		err = rows.Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.DisplayName, &role.Description,
			&permissions, &table_permissions, &column_permissions, &row_level_permissions,
			&role.AutoUpdate, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}

		arrayOfPermissions := strings.Split(permissions, " ")

		role.Permissions = arrayOfPermissions

		err = json.Unmarshal(table_permissions, &role.TablePermissions)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(column_permissions, &role.ColumnPermissions)
		if err != nil {
			return nil, err
		}

		err := json.Unmarshal(row_level_permissions, &role.RowPermissions)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *collimatoRepository) CreateChart(chart model.NewChart) (*model.Chart, error) {
	t := time.Now().Unix()
	id := model.NewID()

	config, err := json.Marshal(chart.Configuration)
	if err != nil {
		return nil, err
	}

	data := map[string]any{}
	data["model"] = chart.Model
	data["query"] = chart.Query

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	tx, err := c.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO collimato_charts (` + collimatoChartColumns + `) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, chart.Name, chart.ChartType, config, dataBytes, chart.OwnerID, chart.WorkspaceID, t, t, 0)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &model.Chart{
		ID:            id,
		Name:          chart.Name,
		ChartType:     chart.ChartType,
		Configuration: chart.Configuration,
		Data:          data,
		OwnerID:       chart.OwnerID,
		WorkspaceID:   chart.WorkspaceID,
		CreatedAt:     t,
		UpdatedAt:     t,
		DeletedAt:     0,
	}, nil
}

// GetCharts returns all collimato_charts with their queries. This should be used only by the owner of the charts.
func (c *collimatoRepository) GetCharts(workspaceID string) ([]model.Chart, error) {
	rows, err := c.Db.Query("SELECT "+collimatoChartColumns+" FROM collimato_charts WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	charts := make([]model.Chart, 0)
	for rows.Next() {
		var chart model.Chart
		var config string
		var data string
		err = rows.Scan(&chart.ID, &chart.Name, &chart.ChartType, &config, &data, &chart.OwnerID, &chart.WorkspaceID, &chart.CreatedAt, &chart.UpdatedAt, &chart.DeletedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(config), &chart.Configuration)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(data), &chart.Data)
		if err != nil {
			return nil, err
		}

		charts = append(charts, chart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return charts, nil
}

func (c *collimatoRepository) GetChartByID(id string) (*model.Chart, error) {
	chart := &model.Chart{}
	var config string
	var data string
	err := c.Db.QueryRow("SELECT "+collimatoChartColumns+" FROM collimato_charts WHERE collimato_charts.id = ?",
		id).Scan(&chart.ID, &chart.Name, &chart.ChartType, &config, &data, &chart.OwnerID, &chart.WorkspaceID, &chart.CreatedAt, &chart.UpdatedAt, &chart.DeletedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(config), &chart.Configuration)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(data), &chart.Data)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func (c *collimatoRepository) UpdateChart(id string, chart model.Chart) (*model.Chart, error) {
	config, err := json.Marshal(chart.Configuration)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(chart.Data)
	if err != nil {
		return nil, err
	}

	t := time.Now().Unix()

	tx, err := c.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec("UPDATE collimato_charts SET name = ?, chart_type = ?, configuration = ?, data =?, updated_at = ? WHERE id = ?", chart.Name, chart.ChartType, config, data, t, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &model.Chart{
		ID:            id,
		Name:          chart.Name,
		ChartType:     chart.ChartType,
		Configuration: chart.Configuration,
		Data:          chart.Data,
		OwnerID:       chart.OwnerID,
		WorkspaceID:   chart.WorkspaceID,
		CreatedAt:     chart.CreatedAt,
		UpdatedAt:     t,
		DeletedAt:     chart.DeletedAt,
	}, nil
}

func (c *collimatoRepository) DeleteChart(id string) error {
	tx, err := c.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM collimato_dashboard_charts WHERE chart_id = ?", id); err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM collimato_charts WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func (c *collimatoRepository) GetDashboards(workspaceID string) ([]model.Dashboard, error) {
	rows, err := c.Db.Query("SELECT "+collimatoDashboardColumns+" FROM collimato_dashboards WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dashboards := make([]model.Dashboard, 0)
	for rows.Next() {
		var dashboard model.Dashboard
		err = rows.Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description, &dashboard.OwnerID, &dashboard.WorkspaceID, &dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.DeletedAt)
		if err != nil {
			return nil, err
		}

		dashboards = append(dashboards, dashboard)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dashboards, nil
}

func (c *collimatoRepository) GetDashboardByID(id string) (*model.Dashboard, error) {
	dashboard := &model.Dashboard{}

	tx, err := c.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = tx.QueryRow("SELECT "+collimatoDashboardColumns+" FROM collimato_dashboards WHERE id = ?", id).
		Scan(&dashboard.ID, &dashboard.Title, &dashboard.Description, &dashboard.OwnerID, &dashboard.WorkspaceID, &dashboard.CreatedAt, &dashboard.UpdatedAt, &dashboard.DeletedAt)
	if err != nil {
		return nil, err
	}

	dashboard.Charts = make([]model.Chart, 0)
	var position []byte
	results, err := tx.Query(`
		SELECT collimato_charts.*, collimato_dashboard_charts.position
		FROM collimato_charts
		INNER JOIN collimato_dashboard_charts ON collimato_dashboard_charts.chart_id = collimato_charts.id
		WHERE collimato_charts.deleted_at = 0 AND collimato_dashboard_charts.dashboard_id = ?`, id)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	for results.Next() {
		chart := model.Chart{}
		var configuration, data []byte

		err = results.Scan(&chart.ID, &chart.Name, &chart.ChartType, &configuration, &data, &chart.OwnerID,
			&chart.WorkspaceID, &chart.CreatedAt, &chart.UpdatedAt, &chart.DeletedAt, &position)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(configuration, &chart.Configuration); err != nil {
			return nil, err
		}

		if err = json.Unmarshal(data, &chart.Data); err != nil {
			return nil, err
		}

		if err = json.Unmarshal(position, &chart.Position); err != nil {
			return nil, err
		}

		dashboard.Charts = append(dashboard.Charts, chart)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	dashboard.Filters = make([]model.DashboardFilter, 0)
	filtersResults, err := tx.Query(`SELECT `+collimatoDashboardFilterColumns+` FROM collimato_dashboard_filters WHERE dashboard_id = ?`, id)
	if err != nil {
		return nil, err
	}

	defer filtersResults.Close()

	for filtersResults.Next() {
		filter := model.DashboardFilter{}
		var values, applyTo []byte

		err = filtersResults.Scan(&filter.ID, &filter.DashboardID, &filter.Name, &filter.Table, &filter.Column, &values, &applyTo,
			&filter.CreatedAt, &filter.UpdatedAt, &filter.DeletedAt, &filter.Operator)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(values, &filter.Values); err != nil {
			return nil, err
		}

		if err = json.Unmarshal(applyTo, &filter.ApplyTo); err != nil {
			return nil, err
		}

		dashboard.Filters = append(dashboard.Filters, filter)
	}

	if err := filtersResults.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return dashboard, nil
}

func (c *collimatoRepository) CreateDashboard(dashboard model.Dashboard) (*model.Dashboard, error) {
	t := time.Now().Unix()
	id := model.NewID()

	_, err := c.Db.Exec("INSERT INTO collimato_dashboards ("+collimatoDashboardColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?)", id, dashboard.Title, dashboard.Description, dashboard.OwnerID, dashboard.WorkspaceID, t, t, 0)
	if err != nil {
		return nil, err
	}

	return &model.Dashboard{
		ID:          id,
		Title:       dashboard.Title,
		Description: dashboard.Description,
		OwnerID:     dashboard.OwnerID,
		WorkspaceID: dashboard.WorkspaceID,
		CreatedAt:   t,
		UpdatedAt:   t,
		DeletedAt:   0,
	}, nil
}

func (c *collimatoRepository) UpdateDashboard(dashboard model.Dashboard) (*model.Dashboard, error) {
	_, err := c.Db.Exec("UPDATE collimato_dashboards SET title = ?, description = ?, updated_at = ? WHERE id = ?", dashboard.Title, dashboard.Description, time.Now().Unix(), dashboard.ID)
	if err != nil {
		return nil, err
	}

	return &dashboard, nil
}

func (c *collimatoRepository) AddDashboardFilter(dashboardID string, filter model.DashboardFilter) (*model.DashboardFilter, error) {
	t := time.Now().Unix()
	id := model.NewID()

	values, err := json.Marshal(filter.Values)
	if err != nil {
		return nil, err
	}

	applyTo, err := json.Marshal(filter.ApplyTo)
	if err != nil {
		return nil, err
	}

	_, err = c.Db.Exec("INSERT INTO collimato_dashboard_filters ("+collimatoDashboardFilterColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, dashboardID, filter.Name, filter.Table, filter.Column, values, applyTo, t, t, 0, filter.Operator)
	if err != nil {
		return nil, err
	}

	return &model.DashboardFilter{
		ID:          id,
		DashboardID: dashboardID,
		Name:        filter.Name,
		Table:       filter.Table,
		Column:      filter.Column,
		Operator:    filter.Operator,
		Values:      filter.Values,
		ApplyTo:     filter.ApplyTo,
		CreatedAt:   t,
		UpdatedAt:   t,
		DeletedAt:   0,
	}, nil
}

func (c *collimatoRepository) UpdateDashboardFilter(filter model.DashboardFilter) error {
	t := time.Now().Unix()

	values, err := json.Marshal(filter.Values)
	if err != nil {
		return err
	}

	applyTo, err := json.Marshal(filter.ApplyTo)
	if err != nil {
		return err
	}

	_, err = c.Db.Exec("UPDATE collimato_dashboard_filters SET name = ?, table_name = ?, column_name = ?, filter_values = ?, apply_to = ?, updated_at = ?, operator = ? WHERE id = ?",
		filter.Name, filter.Table, filter.Column, values, applyTo, t, filter.Operator, filter.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) GetDashboardFilterByID(dashboardID string, filterID string) (*model.DashboardFilter, error) {
	rows := c.Db.QueryRow("SELECT "+collimatoDashboardFilterColumns+" FROM collimato_dashboard_filters WHERE dashboard_id = ? AND id = ?", dashboardID, filterID)

	filter := model.DashboardFilter{}
	var values, applyTo []byte

	err := rows.Scan(&filter.ID, &filter.DashboardID, &filter.Name, &filter.Table, &filter.Column, &values, &applyTo, &filter.CreatedAt, &filter.UpdatedAt, &filter.DeletedAt, &filter.Operator)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(values, &filter.Values); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(applyTo, &filter.ApplyTo); err != nil {
		return nil, err
	}

	return &filter, nil
}

func (c *collimatoRepository) DeleteDashboardFilter(dashboardID, filterID string) error {
	_, err := c.Db.Exec("DELETE FROM collimato_dashboard_filters WHERE dashboard_id = ? AND id = ?", dashboardID, filterID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) DeleteDashboard(id string) error {
	_, err := c.Db.Exec("DELETE FROM collimato_dashboards WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) AddDashboardCharts(userID string, dashboardID string, charts []model.DashboardChartPatch) error {
	t := time.Now().Unix()

	// Begin a transaction
	tx, err := c.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Delete all charts from the dashboard
	_, err = tx.Exec("DELETE FROM collimato_dashboard_charts WHERE dashboard_id = ?", dashboardID)
	if err != nil {
		return err
	}

	if len(charts) == 0 {
		return tx.Commit()
	}

	args := make([]any, 0, len(charts)*8)
	for _, item := range charts {
		position, err := json.Marshal(item.Position)
		if err != nil {
			return err
		}

		args = append(args, model.NewID(), dashboardID, item.ID, userID, position, t, t, 0)
	}

	_, err = tx.Exec(
		`INSERT INTO collimato_dashboard_charts (`+collimatoDashboardChartColumns+`)
		 VALUES `+rowPlaceholders(len(charts), 8)+`
		 ON DUPLICATE KEY UPDATE
		     position = VALUES(position), updated_at = VALUES(updated_at)`, args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *collimatoRepository) CreateWorkspace(userID, name, description, status, secret string) (*model.CollimatoWorkspace, error) {
	id := model.NewID()
	t := time.Now().Unix()
	serverID := model.NewID()

	tx, err := c.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO collimato_workspaces ("+collimatoWorkspaceColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, name, description, status, serverID, userID, t, t, 0, secret)
	if err != nil {
		return nil, err
	}

	roles := model.MakeDefaultCollimatoWorkspaceRoles()

	for _, role := range roles {
		permissions := strings.Join(role.Permissions, " ")

		tablePermissions, err := json.Marshal(role.TablePermissions)
		if err != nil {
			return nil, err
		}

		colLevelPermissions, err := json.Marshal(role.ColumnPermissions)
		if err != nil {
			return nil, err
		}

		rowLevelPermissions, err := json.Marshal(role.RowPermissions)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec("INSERT INTO collimato_workspace_roles ("+collimatoWorkspaceRoleColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			model.NewID(), id, role.Name, role.DisplayName, role.Description, permissions, tablePermissions, colLevelPermissions, rowLevelPermissions, role.AutoUpdate, t, t)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &model.CollimatoWorkspace{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      status,
		ServerID:    serverID,
		CreatedBy:   userID,
		CreatedAt:   t,
		UpdatedAt:   t,
		DeletedAt:   0,
	}, nil
}

func (c *collimatoRepository) DeleteWorkspace(workspaceID string) error {
	tx, err := c.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM collimato_workspace_users WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_workspace_roles WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_workspace_connections WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM collimato_dashboards WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	defer rows.Close()

	var dashboardIDs []string
	for rows.Next() {
		var dashboardID string
		if err := rows.Scan(&dashboardID); err != nil {
			return err
		}

		dashboardIDs = append(dashboardIDs, dashboardID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for _, dashboardID := range dashboardIDs {
		_, err = tx.Exec("DELETE FROM collimato_dashboard_charts WHERE dashboard_id = ?", dashboardID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("DELETE FROM collimato_dashboard_filters WHERE dashboard_id = ?", dashboardID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM collimato_charts WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_dashboards WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_workspace_groups WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_workspace_files WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM collimato_workspaces WHERE id = ?", workspaceID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) UpdateWorkspace(workspace model.CollimatoWorkspace) (*model.CollimatoWorkspace, error) {
	_, err := c.Db.Exec("UPDATE collimato_workspaces SET name = ?, description = ?, status = ?, updated_at = ?, secret = ? WHERE id = ?",
		workspace.Name, workspace.Description, workspace.Status, time.Now().Unix(), workspace.Secret, workspace.ID)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (c *collimatoRepository) CountWorkspaces(ctx context.Context) (int, error) {
	var count int
	err := c.Db.QueryRowContext(ctx, "SELECT COUNT(*) FROM collimato_workspaces WHERE deleted_at = 0").Scan(&count)

	return count, err
}

func (c *collimatoRepository) GetWorkspaces() ([]model.CollimatoWorkspace, error) {
	rows, err := c.Db.Query("SELECT " + collimatoWorkspaceColumns + " FROM collimato_workspaces")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workspaces := make([]model.CollimatoWorkspace, 0)
	for rows.Next() {
		var workspace model.CollimatoWorkspace
		err = rows.Scan(&workspace.ID, &workspace.Name, &workspace.Description, &workspace.Status,
			&workspace.ServerID, &workspace.CreatedBy, &workspace.CreatedAt,
			&workspace.UpdatedAt, &workspace.DeletedAt, &workspace.Secret)
		if err != nil {
			return nil, err
		}

		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (c *collimatoRepository) GetWorkspacesForUser(userID string) ([]model.CollimatoWorkspace, error) {
	// A user reaches a workspace either directly (collimato_workspace_users) or
	// through a group attached to it (collimato_workspace_groups + group_members).
	// UNION de-duplicates a user who has both paths.
	rows, err := c.Db.Query(`SELECT cw.* FROM collimato_workspaces cw
							 JOIN collimato_workspace_users cwu
							   ON cwu.workspace_id = cw.id AND cwu.user_id = ? AND cwu.deleted_at = 0
							 UNION
							 SELECT cw.* FROM collimato_workspaces cw
							 JOIN collimato_workspace_groups cwg
							   ON cwg.workspace_id = cw.id AND cwg.deleted_at = 0
							 JOIN group_members gm
							   ON gm.group_id = cwg.group_id AND gm.user_id = ?
							 JOIN user_groups ug
							   ON ug.id = cwg.group_id AND ug.deleted_at = 0`, userID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workspaces := make([]model.CollimatoWorkspace, 0)
	for rows.Next() {
		var workspace model.CollimatoWorkspace
		err = rows.Scan(&workspace.ID, &workspace.Name, &workspace.Description,
			&workspace.Status, &workspace.ServerID, &workspace.CreatedBy,
			&workspace.CreatedAt, &workspace.UpdatedAt, &workspace.DeletedAt, &workspace.Secret)
		if err != nil {
			return nil, err
		}

		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (c *collimatoRepository) GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error) {
	row := c.Db.QueryRow("SELECT "+collimatoWorkspaceColumns+" FROM collimato_workspaces WHERE id = ?", id)

	workspace := &model.CollimatoWorkspace{}
	err := row.Scan(&workspace.ID, &workspace.Name, &workspace.Description, &workspace.Status,
		&workspace.ServerID, &workspace.CreatedBy, &workspace.CreatedAt,
		&workspace.UpdatedAt, &workspace.DeletedAt, &workspace.Secret)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

func (c *collimatoRepository) CreateWorkspaceRole(role model.CollimatoRole) (*model.CollimatoRole, error) {
	id := model.NewID()
	t := time.Now().Unix()
	permissions := strings.Join(role.Permissions, " ")

	tablePermissions, err := json.Marshal(role.TablePermissions)
	if err != nil {
		return nil, err
	}

	colLevelPermissions, err := json.Marshal(role.ColumnPermissions)
	if err != nil {
		return nil, err
	}

	rowLevelPermissions, err := json.Marshal(role.RowPermissions)
	if err != nil {
		return nil, err
	}

	_, err = c.Db.Exec("INSERT INTO collimato_workspace_roles ("+collimatoWorkspaceRoleColumns+") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ,?)",
		id, role.WorkspaceID, role.Name, role.DisplayName, role.Description, permissions, tablePermissions, colLevelPermissions, rowLevelPermissions, role.AutoUpdate, t, t)
	if err != nil {
		return nil, err
	}

	return &model.CollimatoRole{
		ID:                id,
		WorkspaceID:       role.WorkspaceID,
		Name:              role.Name,
		DisplayName:       role.DisplayName,
		Description:       role.Description,
		Permissions:       role.Permissions,
		TablePermissions:  role.TablePermissions,
		ColumnPermissions: role.ColumnPermissions,
		RowPermissions:    role.RowPermissions,
		CreatedAt:         t,
		UpdatedAt:         t,
	}, nil
}

func (c *collimatoRepository) GetWorkspaceRoles(workspaceID string) ([]model.CollimatoRole, error) {
	results, err := c.Db.Query("SELECT "+collimatoWorkspaceRoleColumns+" FROM collimato_workspace_roles WHERE workspace_id =?", workspaceID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	roles := make([]model.CollimatoRole, 0)

	for results.Next() {
		role := model.CollimatoRole{}
		var permissions string
		var table_permissions []byte
		var col_permissions []byte
		var row_permissions []byte
		err = results.Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.DisplayName, &role.Description, &permissions, &table_permissions, &col_permissions, &row_permissions, &role.AutoUpdate, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}

		role.Permissions = strings.Split(permissions, " ")

		err = json.Unmarshal(table_permissions, &role.TablePermissions)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(col_permissions, &role.ColumnPermissions)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(row_permissions, &role.RowPermissions)
		if err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *collimatoRepository) GetWorkspaceRoleByID(workspaceID, roleID string) (*model.CollimatoRole, error) {
	row := c.Db.QueryRow("SELECT "+collimatoWorkspaceRoleColumns+" FROM collimato_workspace_roles WHERE workspace_id = ? AND id = ?", workspaceID, roleID)

	role := model.CollimatoRole{}
	var permissions string
	var table_permissions []byte
	var col_permissions []byte
	var row_permissions []byte
	err := row.Scan(&role.ID, &role.WorkspaceID, &role.Name, &role.DisplayName, &role.Description, &permissions, &table_permissions, &col_permissions, &row_permissions, &role.AutoUpdate, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	role.Permissions = strings.Split(permissions, " ")

	err = json.Unmarshal(table_permissions, &role.TablePermissions)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(col_permissions, &role.ColumnPermissions)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(row_permissions, &role.RowPermissions)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (c *collimatoRepository) UpdateWorkspaceRole(workspaceID string, role model.CollimatoRole) error {
	permissions := strings.Join(role.Permissions, " ")

	tablePermissions, err := json.Marshal(role.TablePermissions)
	if err != nil {
		return err
	}

	colLevelPermissions, err := json.Marshal(role.ColumnPermissions)
	if err != nil {
		return err
	}

	rowLevelPermissions, err := json.Marshal(role.RowPermissions)
	if err != nil {
		return err
	}

	_, err = c.Db.Exec("UPDATE collimato_workspace_roles SET name = ?, displayname = ?, description = ?, permissions = ?, table_level_permissions = ?, column_level_permissions = ?, row_level_permissions = ?, updated_at = ? WHERE workspace_id = ? AND id = ?",
		role.Name, role.DisplayName, role.Description, permissions, tablePermissions, colLevelPermissions, rowLevelPermissions, time.Now().Unix(), workspaceID, role.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) DeleteWorkspaceRole(workspaceID, roleID string) error {
	_, err := c.Db.Exec("DELETE FROM collimato_workspace_roles WHERE workspace_id = ? AND id = ?", workspaceID, roleID)
	if err != nil {
		return err
	}

	return nil
}

func (c *collimatoRepository) GetWorkspaceFiles(workspaceID, fileType string) ([]model.CollimatoWorkspaceFile, error) {
	rows, err := c.Db.Query(
		`SELECT id, workspace_id, name, file_type, content, encrypted, created_at, updated_at, builder_model
		 FROM collimato_workspace_files WHERE workspace_id = ? AND file_type = ?`,
		workspaceID, fileType)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files := make([]model.CollimatoWorkspaceFile, 0)
	for rows.Next() {
		var f model.CollimatoWorkspaceFile
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.FileType, &f.Content, &f.Encrypted, &f.CreatedAt, &f.UpdatedAt, &f.BuilderModel); err != nil {
			return nil, err
		}

		files = append(files, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (c *collimatoRepository) GetWorkspaceFile(workspaceID, name, fileType string) (*model.CollimatoWorkspaceFile, error) {
	f := &model.CollimatoWorkspaceFile{}
	err := c.Db.QueryRow(
		`SELECT id, workspace_id, name, file_type, content, encrypted, created_at, updated_at, builder_model
		 FROM collimato_workspace_files WHERE workspace_id = ? AND file_type = ? AND name = ?`,
		workspaceID, fileType, name).
		Scan(&f.ID, &f.WorkspaceID, &f.Name, &f.FileType, &f.Content, &f.Encrypted, &f.CreatedAt, &f.UpdatedAt, &f.BuilderModel)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c *collimatoRepository) UpdateWorkspaceFile(file model.CollimatoWorkspaceFile) (*model.CollimatoWorkspaceFile, error) {
	t := time.Now().Unix()

	existing, err := c.GetWorkspaceFile(file.WorkspaceID, file.Name, file.FileType)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		_, err = c.Db.Exec(
			`UPDATE collimato_workspace_files SET content = ?, encrypted = ?, updated_at = ?, builder_model = ?
			 WHERE workspace_id = ? AND file_type = ? AND name = ?`,
			file.Content, file.Encrypted, t, file.BuilderModel, file.WorkspaceID, file.FileType, file.Name)
		if err != nil {
			return nil, err
		}

		file.ID = existing.ID
		file.CreatedAt = existing.CreatedAt
		file.UpdatedAt = t
		return &file, nil
	}

	id := model.NewID()
	_, err = c.Db.Exec(
		`INSERT INTO collimato_workspace_files (id, workspace_id, name, file_type, content, encrypted, created_at, updated_at, builder_model)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, file.WorkspaceID, file.Name, file.FileType, file.Content, file.Encrypted, t, t, file.BuilderModel)
	if err != nil {
		return nil, err
	}

	file.ID = id
	file.CreatedAt = t
	file.UpdatedAt = t
	return &file, nil
}

func (c *collimatoRepository) DeleteWorkspaceFile(workspaceID, name, fileType string) error {
	_, err := c.Db.Exec(
		`DELETE FROM collimato_workspace_files WHERE workspace_id = ? AND file_type = ? AND name = ?`,
		workspaceID, fileType, name)
	return err
}
