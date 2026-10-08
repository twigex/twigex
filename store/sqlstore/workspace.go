// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand"

	"crypto/md5"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

type workspaceRepository struct {
	Db *sql.DB
}

func NewWorkspaceRepository(Db *sql.DB) (*workspaceRepository, error) {
	repo := &workspaceRepository{}
	repo.Db = Db
	return repo, nil
}

// RenameTaskAndKanbanSectionTx atomically renames the task row in its linked table and
// updates all kanban view JSON blobs that reference the old name.
// Reads (linkedTableName, oldName) must be done by the caller before invoking this.
func (w *workspaceRepository) RenameTaskAndKanbanSectionTx(workspaceID, tableID, taskID, field, newName, oldName, linkedTableName, linkedTableID string) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = w.UpdateTaskNameTx(tx, linkedTableName, taskID, newName); err != nil {
		return err
	}

	if _, err = w.UpdateKanbanSectionsFieldNameTx(tx, workspaceID, tableID, taskID, field, newName, oldName, linkedTableID); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateTask inserts a task and sets fields, columns the caller has checked,
// in the same transaction, so the task is never seen without them.
func (w *workspaceRepository) CreateTask(workspaceID, tableID, name, userID, parentTaskID, taskID string, timestamp int64, fields map[string]any) (*map[string]interface{}, error) {
	var tableName string
	if err := w.Db.QueryRow(`SELECT name FROM workspace_tables WHERE workspace_id = ? AND id = ?`, workspaceID, tableID).Scan(&tableName); err != nil {
		return nil, err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	columns := slices.Sorted(maps.Keys(fields))
	set := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns)+1)
	for _, column := range columns {
		quoted, err := quoteIdent(column)
		if err != nil {
			return nil, err
		}

		set = append(set, quoted+" = ?")
		args = append(args, fields[column])
	}

	args = append(args, taskID)

	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = w.InsertTaskTx(tx, tableName, taskID, name, userID, parentTaskID, timestamp); err != nil {
		return nil, err
	}

	if len(fields) > 0 {
		if _, err = tx.Exec(`UPDATE `+table+` SET `+strings.Join(set, ", ")+
			`, updated_at = UNIX_TIMESTAMP() WHERE id = ?`, args...); err != nil {
			return nil, err
		}
	}

	task, err := w.SelectTaskByIDTx(tx, tableName, taskID)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, sql.ErrNoRows
	}

	err = tx.Commit()
	return task, err
}

// SyncTaskKanban updates kanban views and sections for a newly created
// top-level task; selects gives the option chosen for each single-select field.
func (w *workspaceRepository) SyncTaskKanban(workspaceID, tableID, taskID string, selects map[string]string, task map[string]interface{}) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = w.UpdateKanbanViewsTx(tx, workspaceID, tableID, task); err != nil {
		return err
	}

	for _, field := range slices.Sorted(maps.Keys(selects)) {
		if _, err = w.UpdateKanbanSectionsTx(tx, workspaceID, tableID, taskID, field, selects[field]); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

// CreateTx creates a workspace, its default roles, and the creator as admin member
// in a single transaction. Returns the new workspace ID.
func (w *workspaceRepository) CreateTx(workspaceID, userID, name, description, prefix string, adminRole, userRole model.ProjectWorkspaceRole, member model.WorkspaceMember) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = w.InsertWorkspaceTx(tx, workspaceID, userID, name, description, prefix); err != nil {
		return err
	}

	if err = w.createProjectWorkspaceRoleTx(tx, adminRole); err != nil {
		return err
	}

	if err = w.createProjectWorkspaceRoleTx(tx, userRole); err != nil {
		return err
	}

	if err = w.CreateWorkspaceMemberTx(tx, &member); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateTaskTx updates a task field and conditionally syncs kanban sections/colors in one transaction.
func (w *workspaceRepository) UpdateTaskTx(workspaceID, tableID, taskID, field, value string, typedValue interface{}, hasUpdatedAt, isKanbanField, syncColors bool) (*map[string]interface{}, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := w.UpdateTaskFieldTx(tx, workspaceID, tableID, taskID, field, typedValue, hasUpdatedAt)
	if err != nil {
		return nil, err
	}

	if isKanbanField {
		if _, err = w.UpdateKanbanSectionsTx(tx, workspaceID, tableID, taskID, field, value); err != nil {
			return nil, err
		}
	}

	if syncColors {
		if err = w.UpdateKanbanOptionColorsTx(tx, workspaceID, tableID, taskID, value); err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	return result, err
}

func (w *workspaceRepository) UpdateTaskFieldTx(tx *sql.Tx, workspaceID, tableID, taskID, field string, typedValue interface{}, hasUpdatedAt bool) (*map[string]interface{}, error) {
	var tableName string
	err := tx.QueryRow(`SELECT name FROM workspace_tables WHERE workspace_id = ? AND id = ?`, workspaceID, tableID).Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	escapedTable, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	escapedField, err := quoteIdent(field)
	if err != nil {
		return nil, err
	}

	var updateQuery string
	var args []interface{}
	if hasUpdatedAt {
		updateQuery = fmt.Sprintf(`UPDATE %s SET %s = ?, updated_at = UNIX_TIMESTAMP() WHERE id = ?`, escapedTable, escapedField)
		args = []interface{}{typedValue, taskID}
	} else {
		updateQuery = fmt.Sprintf(`UPDATE %s SET %s = ? WHERE id = ?`, escapedTable, escapedField)
		args = []interface{}{typedValue, taskID}
	}

	if _, err = tx.Exec(updateQuery, args...); err != nil {
		return nil, err
	}

	return w.SelectTaskByIDTx(tx, tableName, taskID)
}

func (w *workspaceRepository) InsertTaskTx(tx *sql.Tx, tableName, newID, name, userID, parentTaskID string, timestamp int64) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	// A task without a parent holds NULL, never "", so that one indexed
	// condition finds the tasks that head the grid's branches.
	var parent any
	if parentTaskID != "" && parentTaskID != newID {
		parent = parentTaskID
	}

	query := `INSERT INTO ` + table + ` (id, name, created_at, created_by, parent_task_id) VALUES (?, ?, ?, ?, ?)`
	_, err = tx.Exec(query, newID, name, timestamp, userID, parent)
	return err
}

func (w *workspaceRepository) SelectTaskByIDTx(tx *sql.Tx, tableName, taskID string) (*map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(`SELECT * FROM `+table+` WHERE id = ?`, taskID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return nil, nil
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	// Text columns come back as bytes, which JSON would send as base64.
	task := make(map[string]interface{})
	for i, col := range columns {
		if b, ok := values[i].([]byte); ok {
			task[col] = string(b)
		} else {
			task[col] = values[i]
		}
	}

	return &task, nil
}

func shortIndexName(table, column string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(table+"_"+column)))
	maxLength := 64 - len("idx__") - len(hash[:6]) // two underscores

	if len(column) > maxLength {
		column = column[:maxLength]
	}

	return fmt.Sprintf("idx_%s_%s", column, hash[:6])
}

// kanbanViewRow holds a view ID and its item_order JSON, used by kanban update helpers.
type kanbanViewRow struct {
	viewID    string
	tableID   string
	itemOrder string
}

// scanKanbanViewRowsTx reads id, item_order rows into a slice and closes the cursor.
// Set withTableID=true when the query also selects table_id as the second column.
func scanKanbanViewRowsTx(rows *sql.Rows, withTableID bool) ([]kanbanViewRow, error) {
	var views []kanbanViewRow
	for rows.Next() {
		var v kanbanViewRow
		var err error
		if withTableID {
			err = rows.Scan(&v.viewID, &v.tableID, &v.itemOrder)
		} else {
			err = rows.Scan(&v.viewID, &v.itemOrder)
		}

		if err != nil {
			rows.Close()
			return nil, err
		}

		views = append(views, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return views, rows.Close()
}

// updateColorInKanbanOrder finds the option with optionID inside the "fields" array of the
// named section and sets its color to newColor. Returns the updated JSON, whether a change
// was made, and any error.
func updateColorInKanbanOrder(itemOrder, section, optionID, newColor string) (string, bool, error) {
	if itemOrder == "" || itemOrder == "null" {
		return itemOrder, false, nil
	}

	var itemOrderMap map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrder), &itemOrderMap); err != nil {
		return itemOrder, false, nil
	}

	if itemOrderMap["section"] != section {
		return itemOrder, false, nil
	}

	fields, ok := itemOrderMap["fields"].([]interface{})
	if !ok {
		return itemOrder, false, nil
	}

	updated := false
	for i, f := range fields {
		fMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		if fMap["id"] == optionID {
			fMap["color"] = newColor
			fields[i] = fMap
			updated = true
			break
		}
	}

	if !updated {
		return itemOrder, false, nil
	}

	itemOrderMap["fields"] = fields
	b, err := json.Marshal(itemOrderMap)
	if err != nil {
		return "", false, err
	}

	return string(b), true, nil
}

func (w *workspaceRepository) UpdateView(itemID string, workspaceID string, tableID string, order string, name string, viewType string) (*model.WorkspaceView, error) {
	_, err := w.Db.Exec(`
	UPDATE workspace_view
	SET name = ?, item_order = ?, updated_at = UNIX_TIMESTAMP()
	WHERE id = ? AND workspace_id = ? AND table_id = ?
	`, name, order, itemID, workspaceID, tableID)
	if err != nil {
		return nil, err
	}

	return &model.WorkspaceView{
		ID:          itemID,
		WorkspaceID: workspaceID,
		TableID:     tableID,
		ViewType:    viewType,
		Name:        name,
		TaskOrder:   order,
	}, nil
}

func (w *workspaceRepository) UpdateViewName(itemID string, workspaceID string, tableID string, name string) error {
	_, err := w.Db.Exec(`UPDATE workspace_view SET name = ?, updated_at = UNIX_TIMESTAMP()
		WHERE id = ? AND workspace_id = ? AND table_id = ?`, name, itemID, workspaceID, tableID)
	return err
}

// ChangeViewOrder rewrites a view's item_order with change, holding the
// view's row so two changes to it wait for each other rather than one
// overwriting the other. It returns sql.ErrNoRows for a view not in the
// table and workspace.
func (w *workspaceRepository) ChangeViewOrder(ctx context.Context, viewID, workspaceID, tableID string, change func(order string) (string, error)) (string, error) {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", err
	}

	defer tx.Rollback()

	var stored sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT item_order FROM workspace_view
		WHERE id = ? AND workspace_id = ? AND table_id = ? FOR UPDATE`,
		viewID, workspaceID, tableID).Scan(&stored); err != nil {
		return "", err
	}

	order, err := change(stored.String)
	if err != nil {
		return "", err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE workspace_view SET item_order = ?, updated_at = UNIX_TIMESTAMP() WHERE id = ?`,
		order, viewID); err != nil {
		return "", err
	}

	return order, tx.Commit()
}

// GetKanbanRestIDs returns, in the order a board shows the cards it has no
// place for, the ids of a table's cards matching filter up to and including
// upTo, at most limit of them.
func (w *workspaceRepository) GetKanbanRestIDs(ctx context.Context, tableName string, filter model.SQLFilter, upTo string, limit int) ([]string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	where := taskSourceWhere
	if filter.SQL != "" {
		where += " AND " + filter.SQL
	}

	rows, err := w.Db.QueryContext(ctx, "SELECT main.id FROM "+table+" AS main WHERE "+where+
		" ORDER BY main.created_at, main.id LIMIT ?", append(slices.Clone(filter.Args), limit)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
		if id == upTo {
			break
		}
	}

	return ids, rows.Err()
}

func (w *workspaceRepository) CreateView(workspaceID string, tableID string, order string, name string, viewType string, userID string, parentTableID string, viewID string) (*model.WorkspaceView, error) {
	_, err := w.Db.Exec(`
		INSERT INTO workspace_view (id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)
	`, viewID, workspaceID, tableID, viewType, parentTableID, name, order, userID)
	if err != nil {
		return nil, err
	}

	return &model.WorkspaceView{
		ID:          viewID,
		WorkspaceID: workspaceID,
		TableID:     tableID,
		ViewType:    viewType,
		Name:        name,
		TaskOrder:   order,
	}, nil
}

func (w *workspaceRepository) ViewExistsByName(tableID, name string) (bool, error) {
	var id string
	err := w.Db.QueryRow(`
		SELECT id FROM workspace_view
		WHERE table_id = ? AND name = ? AND deleted_at = 0
	`, tableID, name).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// GetTableWorkspaceID returns the workspace a table belongs to, or "" when
// there is no such table.
func (w *workspaceRepository) GetTableWorkspaceID(ctx context.Context, tableID string) (string, error) {
	var workspaceID string
	err := w.Db.QueryRowContext(ctx, `SELECT workspace_id FROM workspace_tables WHERE id = ?`, tableID).Scan(&workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return workspaceID, err
}

func (w *workspaceRepository) GetTableName(tableID string) (string, error) {
	var name string
	err := w.Db.QueryRow(`SELECT name FROM workspace_tables WHERE id = ?`, tableID).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (w *workspaceRepository) GetLinkedTableID(tableID, fieldName string) (string, error) {
	var linkedTableID string
	err := w.Db.QueryRow(`
		SELECT linked_table_id FROM workspace_relationships
		WHERE table_id = ? AND table_name = ?
	`, tableID, fieldName).Scan(&linkedTableID)
	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return linkedTableID, nil
}

// IsOptionTableOf reports whether optionTableID holds the options of one of
// tableID's single-select or status fields.
func (w *workspaceRepository) IsOptionTableOf(ctx context.Context, tableID, optionTableID string) (bool, error) {
	var count int
	err := w.Db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workspace_relationships r
		JOIN workspace_tables o ON o.id = r.linked_table_id
		WHERE r.table_id = ? AND r.linked_table_id = ? AND o.single_select = 1 AND o.deleted_at = 0`,
		tableID, optionTableID).Scan(&count)
	return count > 0, err
}

func (w *workspaceRepository) IsTableSingleSelect(tableID string) (bool, error) {
	var singleSelect bool
	err := w.Db.QueryRow("SELECT single_select FROM workspace_tables WHERE id = ?", tableID).Scan(&singleSelect)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return singleSelect, nil
}

func (w *workspaceRepository) GetKanbanStatusOptions(tableName string) ([]model.KanbanStatusOption, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.Query(fmt.Sprintf(`SELECT id, name, color FROM %s`, table))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var options []model.KanbanStatusOption
	for rows.Next() {
		var opt model.KanbanStatusOption
		if err := rows.Scan(&opt.ID, &opt.Name, &opt.Color); err != nil {
			return nil, err
		}

		options = append(options, opt)
	}

	return options, rows.Err()
}

func (w *workspaceRepository) GetTaskFieldValue(tableName, taskID, fieldName string) (string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return "", err
	}

	field, err := quoteIdent(fieldName)
	if err != nil {
		return "", err
	}

	var value string
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", field, table)
	err = w.Db.QueryRow(query, taskID).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

// GetColumnNames returns a table's column names as the table spells them.
func (w *workspaceRepository) GetColumnNames(ctx context.Context, tableName string) ([]string, error) {
	rows, err := w.Db.QueryContext(ctx, `SELECT COLUMN_NAME FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`, tableName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		names = append(names, name)
	}

	return names, rows.Err()
}

func (w *workspaceRepository) TableHasColumn(tableName, column string) (bool, error) {
	var count int
	err := w.Db.QueryRow(`
		SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, tableName, column).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (w *workspaceRepository) CreateMemberToTask(workspaceID string, tableID string, taskID string, userID string, fieldName string) error {
	var tableName string
	if err := w.Db.QueryRow(
		"SELECT name FROM workspace_tables WHERE workspace_id = ? AND id = ?",
		workspaceID, tableID).Scan(&tableName); err != nil {
		return err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	field, err := quoteIdent(fieldName)
	if err != nil {
		return err
	}

	_, err = w.Db.Exec(
		fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", table, field),
		userID, taskID)
	return err
}

// appendFieldToGridItemOrder adds field to the grid view item_order JSON array if it is not already present.
// Returns the updated JSON, whether a modification was made, and any error.
func appendFieldToGridItemOrder(itemOrderJSON string, field model.WorkspaceHeaders) (string, bool, error) {
	if itemOrderJSON == "" || itemOrderJSON == "null" {
		itemOrderJSON = "[]"
	}

	var taskOrder []model.TaskOrderField
	if err := json.Unmarshal([]byte(itemOrderJSON), &taskOrder); err != nil {
		return "", false, err
	}

	for _, f := range taskOrder {
		if f.Name == field.Name {
			return itemOrderJSON, false, nil
		}
	}

	taskOrder = append(taskOrder, model.TaskOrderField{
		Name:        field.Name,
		DisplayName: field.DisplayName,
		Width:       "100",
		Visible:     true,
	})

	updated, err := json.Marshal(taskOrder)
	if err != nil {
		return "", false, err
	}

	return string(updated), true, nil
}

func (w *workspaceRepository) UpdateTaskOrderGridField(workspaceID string, tableID string, field model.WorkspaceHeaders) error {
	rows, err := w.Db.Query(`
		SELECT id, item_order
		FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND view_type = 'grid'
	`, workspaceID, tableID)
	if err != nil {
		return err
	}

	defer rows.Close()

	type viewData struct {
		ID        string
		ItemOrder string
	}

	var views []viewData

	for rows.Next() {
		var view viewData
		if err := rows.Scan(&view.ID, &view.ItemOrder); err != nil {
			return err
		}

		views = append(views, view)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	for _, view := range views {
		updatedJSON, modified, err := appendFieldToGridItemOrder(view.ItemOrder, field)
		if err != nil {
			continue
		}

		if !modified {
			continue
		}

		if _, err := w.Db.Exec(`UPDATE workspace_view SET item_order = ? WHERE id = ?`, updatedJSON, view.ID); err != nil {
			return err
		}
	}

	return nil
}

// fieldInfo is used to read workspace_fields rows before soft-deleting them.
type fieldInfo struct {
	id            string
	parentFieldID sql.NullString
}

// fieldLinkSetup holds the link-related IDs produced when creating link or single-select fields.
type fieldLinkSetup struct {
	linkedID          string
	secondLinkedID    string
	parentLinkTableID string
	createdTable      *model.WorkspaceTable
}

// createLinkFieldTx sets up a link-type field: creates the junction table, inserts the
// relationship row, and optionally mirrors the link in the opposite direction.
// linkTableID, linkTablePhysicalName, secondLinkID, secondLinkTablePhysicalName, linkViewID, and
// secondLinkViewID must be pre-generated by the caller.
func (w *workspaceRepository) createLinkFieldTx(tx *sql.Tx, workspaceID, tableID, linkedTableID, fieldName, fieldNameInSecondTable, userID, linkTableID, secondLinkID, linkTablePhysicalName, secondLinkTablePhysicalName, linkViewID, secondLinkViewID string, linkBothDirections bool) (fieldLinkSetup, string, error) {
	workSpace, err := w.createWorkspaceTableWithTX(tx, linkTableID, workspaceID, linkTablePhysicalName, true, linkedTableID, userID, false, linkBothDirections, "not-set", linkViewID)
	if err != nil {
		return fieldLinkSetup{}, fieldNameInSecondTable, err
	}

	if _, err = tx.Exec(
		`INSERT INTO workspace_relationships (id, workspace_id, table_id, linked_table_id, table_name, created_at, updated_at, deleted_at) VALUES (UUID(), ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)`,
		workspaceID, tableID, workSpace.ID, fieldName,
	); err != nil {
		return fieldLinkSetup{}, fieldNameInSecondTable, err
	}

	setup := fieldLinkSetup{linkedID: workSpace.ID}

	if !linkBothDirections {
		return setup, fieldNameInSecondTable, nil
	}

	secondWorkSpace, err := w.createWorkspaceTableWithTX(tx, secondLinkID, workspaceID, secondLinkTablePhysicalName, true, tableID, userID, false, linkBothDirections, workSpace.ID, secondLinkViewID)
	if err != nil {
		return fieldLinkSetup{}, fieldNameInSecondTable, err
	}

	if _, err = tx.Exec(
		`INSERT INTO workspace_relationships (id, workspace_id, table_id, linked_table_id, table_name, created_at, updated_at, deleted_at) VALUES (UUID(), ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)`,
		workspaceID, linkedTableID, secondWorkSpace.ID, fieldNameInSecondTable,
	); err != nil {
		return fieldLinkSetup{}, fieldNameInSecondTable, err
	}

	if _, err = tx.Exec(`UPDATE workspace_tables SET second_table_id = ? WHERE id = ?`, secondWorkSpace.ID, workSpace.ID); err != nil {
		return fieldLinkSetup{}, fieldNameInSecondTable, err
	}

	setup.secondLinkedID = secondWorkSpace.ID
	setup.parentLinkTableID = workSpace.ParentTableID
	return setup, fieldNameInSecondTable, nil
}

// createSingleSelectFieldTx creates the single-select option table and relationship row.
// singleSelectTableID and singleSelectPhysicalName must be pre-generated by the caller.
func (w *workspaceRepository) createSingleSelectFieldTx(tx *sql.Tx, workspaceID, tableID, fieldName, linkedTableID, singleSelectTableID, singleSelectPhysicalName string) (fieldLinkSetup, error) {
	workSpace, err := w.createSingleSelectTable(workspaceID, singleSelectTableID, singleSelectPhysicalName, false, linkedTableID, true)
	if err != nil {
		return fieldLinkSetup{}, err
	}

	if _, err = tx.Exec(
		`INSERT INTO workspace_relationships (id, workspace_id, table_id, linked_table_id, table_name, created_at, updated_at, deleted_at) VALUES (UUID(), ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)`,
		workspaceID, tableID, workSpace.ID, fieldName,
	); err != nil {
		return fieldLinkSetup{}, err
	}

	return fieldLinkSetup{linkedID: workSpace.ID, createdTable: workSpace}, nil
}

// storeFormulaTx persists the formula JSON for a field. Caller is responsible
// for only calling this when fieldType is "calculations" and formula is non-nil.
func (w *workspaceRepository) storeFormulaTx(tx *sql.Tx, fieldID string, formula *model.FormulaSpec) error {
	b, err := json.Marshal(formula)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE workspace_fields SET formula = ? WHERE id = ?`, string(b), fieldID)
	return err
}

// UpdateFieldFormula replaces the stored formula JSON for a calculations field.
func (w *workspaceRepository) UpdateFieldFormula(ctx context.Context, workspaceID, tableID, fieldName string, formula *model.FormulaSpec) error {
	b, err := json.Marshal(formula)
	if err != nil {
		return err
	}

	_, err = w.Db.ExecContext(
		ctx,
		`UPDATE workspace_fields SET formula = ? WHERE workspace_id = ? AND table_id = ? AND field_name = ?`,
		string(b), workspaceID, tableID, fieldName,
	)
	return err
}

// setupBidirectionalColumnTx adds the mirror column on the linked table, creates its field
// record, and appends it to that table's grid order.
func (w *workspaceRepository) setupBidirectionalColumnTx(tx *sql.Tx, workspaceID, linkedTableID, fieldNameInSecondTable, fieldNameInSecondTableDisplay, selectedType, fieldType, sqlFieldType, secondFieldID, mainFieldID string, setup fieldLinkSetup, linkBothDirections bool) (string, error) {
	var secondTableName string
	if err := tx.QueryRow(`SELECT name FROM workspace_tables WHERE id = ?`, linkedTableID).Scan(&secondTableName); err != nil {
		return "", err
	}

	secondTable, err := quoteIdent(secondTableName)
	if err != nil {
		return "", err
	}

	secondColumn, err := quoteIdent(fieldNameInSecondTable)
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, secondTable, secondColumn, sqlFieldType)); err != nil {
		return "", err
	}

	fieldStoreDataLinked, err := w.CreateWorkspaceTableFieldStoreTx(tx, secondFieldID, workspaceID, linkedTableID, fieldNameInSecondTable, selectedType, mainFieldID, fieldNameInSecondTableDisplay)
	if err != nil {
		return "", err
	}

	linkedHeader := model.WorkspaceHeaders{
		ID:                 fieldStoreDataLinked.ID,
		Name:               fieldNameInSecondTable,
		HeaderType:         fieldType,
		LinkedID:           setup.secondLinkedID,
		HeaderUsage:        selectedType,
		ParentTableID:      setup.parentLinkTableID,
		LinkBothDirections: linkBothDirections,
		DisplayName:        fieldNameInSecondTableDisplay,
	}

	if err = w.UpdateTaskOrderGridField(workspaceID, linkedTableID, linkedHeader); err != nil {
		return "", err
	}

	return fieldStoreDataLinked.ID, nil
}

func (w *workspaceRepository) CreateTableField(p model.CreateFieldParams) (*model.WorkspaceHeaders, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var tableName string
	if err = tx.QueryRow(`SELECT name FROM workspace_tables WHERE id = ?`, p.TableID).Scan(&tableName); err != nil {
		return nil, err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	column, err := quoteIdent(p.FieldName)
	if err != nil {
		return nil, err
	}

	var setup fieldLinkSetup

	switch p.FieldType {
	case "link":
		setup, p.FieldNameInSecondTable, err = w.createLinkFieldTx(tx, p.WorkspaceID, p.TableID, p.LinkedTableID, p.FieldName, p.FieldNameInSecondTable, p.UserID, p.LinkTableID, p.SecondLinkID, p.LinkTablePhysicalName, p.SecondLinkTablePhysicalName, p.LinkViewID, p.SecondLinkViewID, p.LinkBothDirections)
	case "single select":
		setup, err = w.createSingleSelectFieldTx(tx, p.WorkspaceID, p.TableID, p.FieldName, p.LinkedTableID, p.SingleSelectTableID, p.SingleSelectPhysicalName)
	}

	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, p.SQLFieldType)); err != nil {
		return nil, err
	}

	parentFieldID := ""
	if p.LinkBothDirections {
		parentFieldID = p.SecondFieldID
	}

	fieldStoreData, err := w.CreateWorkspaceTableFieldStoreTx(tx, p.MainFieldID, p.WorkspaceID, p.TableID, p.FieldName, p.SelectedType, parentFieldID, p.FieldNameDisplay)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(p.FieldType, "calculations") && p.Formula != nil {
		if err = w.storeFormulaTx(tx, fieldStoreData.ID, p.Formula); err != nil {
			return nil, err
		}
	}

	var headerLinkID string
	if p.LinkBothDirections {
		headerLinkID, err = w.setupBidirectionalColumnTx(tx, p.WorkspaceID, p.LinkedTableID, p.FieldNameInSecondTable, p.FieldNameInSecondTableDisplay, p.SelectedType, p.FieldType, p.SQLFieldType, p.SecondFieldID, fieldStoreData.ID, setup, p.LinkBothDirections)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	header := &model.WorkspaceHeaders{
		ID:                 fieldStoreData.ID,
		Name:               p.FieldName,
		HeaderType:         p.FieldType,
		LinkedID:           setup.linkedID,
		HeaderUsage:        p.SelectedType,
		ParentTableID:      setup.parentLinkTableID,
		ParentLinkTableID:  setup.secondLinkedID,
		LinkBothDirections: p.LinkBothDirections,
		LinkedHeaderID:     headerLinkID,
		DisplayName:        p.FieldNameDisplay,
		TableData:          setup.createdTable,
	}
	if strings.EqualFold(p.FieldType, "calculations") && p.Formula != nil {
		header.Formula = p.Formula
	}

	return header, nil
}

// CreateWorkspaceViewTx inserts a view row. workspaceView.ID must be set by the caller.
func (w *workspaceRepository) CreateWorkspaceViewTx(tx *sql.Tx, workspaceView *model.WorkspaceView) (*model.WorkspaceView, error) {
	query := `
	INSERT INTO workspace_view
	(id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, main_view, created_at, updated_at, deleted_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)
	`

	_, err := tx.Exec(query,
		workspaceView.ID,
		workspaceView.WorkspaceID,
		workspaceView.TableID,
		workspaceView.ViewType,
		workspaceView.ParentTableID,
		workspaceView.Name,
		workspaceView.TaskOrder,
		workspaceView.CreatedBy,
		workspaceView.MainView,
	)

	if err != nil {
		return nil, err
	}

	return workspaceView, nil
}

func (w *workspaceRepository) GenerateUniquePrefix() (string, error) {
	for {
		prefix := generateUniquePrefix()
		var count int
		if err := w.Db.QueryRow("SELECT COUNT(*) FROM workspaces WHERE pre_fix = ?", prefix+"_").Scan(&count); err != nil {
			return "", err
		}

		if count == 0 {
			return prefix, nil
		}
	}
}

func (w *workspaceRepository) InsertWorkspaceTx(tx *sql.Tx, workspaceID, userID, name, description, prefix string) error {
	_, err := tx.Exec(`
		INSERT INTO workspaces (id, title, description, start_date, end_date, pre_fix, created_at, updated_at, deleted_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), ?, ?, ?)
	`, workspaceID, name, description, 0, 0, prefix+"_", 0, 0, userID)
	return err
}

func (w *workspaceRepository) createProjectWorkspaceRoleTx(tx *sql.Tx, role model.ProjectWorkspaceRole) error {
	permissionsStr := strings.Join(role.Permissions, " ")
	role.Name = strings.ReplaceAll(role.Name, " ", "_")

	perTableMode := 0
	if role.PerTableMode {
		perTableMode = 1
	}

	_, err := tx.Exec(`
		INSERT INTO workspace_roles
		(id, name, displayname, description, permissions, workspace_id, per_table_mode, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
	`, role.ID, role.Name, role.DisplayName, role.Description, permissionsStr, role.WorkspaceID, perTableMode)

	return err
}

func (w *workspaceRepository) GetRoleByID(workspaceID, roleID string) (*model.ProjectWorkspaceRole, error) {
	var role model.ProjectWorkspaceRole
	var permissionsString string

	err := w.Db.QueryRow(`
		SELECT id, name, displayname, description, permissions, workspace_id, per_table_mode, created_at, updated_at
		FROM workspace_roles
		WHERE id = ? AND workspace_id = ?
	`, roleID, workspaceID).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&permissionsString,
		&role.WorkspaceID,
		&role.PerTableMode,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if permissionsString != "" {
		role.Permissions = strings.Split(permissionsString, " ")
	} else {
		role.Permissions = []string{}
	}

	return &role, nil
}

func (w *workspaceRepository) DeleteRole(ctx context.Context, workspaceID, roleID, roleName, fallback string) error {
	tx, err := w.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM workspace_roles WHERE id = ? AND workspace_id = ?`, roleID, workspaceID)
	if err != nil {
		return err
	}

	err = unassignRoleTx(ctx, tx,
		`SELECT user_id, role FROM workspace_members WHERE workspace_id = ?`,
		`UPDATE workspace_members SET role = ? WHERE workspace_id = ? AND user_id = ?`,
		workspaceID, roleName, fallback,
	)
	if err != nil {
		return err
	}

	err = unassignRoleTx(ctx, tx,
		`SELECT group_id, roles FROM workspace_groups WHERE workspace_id = ?`,
		`UPDATE workspace_groups SET roles = ? WHERE workspace_id = ? AND group_id = ?`,
		workspaceID, roleName, fallback,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func unassignRoleTx(ctx context.Context, tx *sql.Tx, selectQuery, updateQuery, workspaceID, roleName, fallback string) error {
	type holder struct {
		id    string
		roles string
	}

	rows, err := tx.QueryContext(ctx, selectQuery, workspaceID)
	if err != nil {
		return err
	}

	var holders []holder
	for rows.Next() {
		var h holder
		if err := rows.Scan(&h.id, &h.roles); err != nil {
			rows.Close()
			return err
		}

		holders = append(holders, h)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	rows.Close()

	for _, h := range holders {
		current := strings.Fields(h.roles)

		remaining := slices.DeleteFunc(slices.Clone(current), func(r string) bool { return r == roleName })
		if len(remaining) == len(current) {
			continue
		}

		if len(remaining) == 0 {
			remaining = []string{fallback}
		}

		if _, err := tx.ExecContext(ctx, updateQuery, strings.Join(remaining, " "), workspaceID, h.id); err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) UpdateRole(workspaceID string, role *model.ProjectWorkspaceRole) error {
	permissionsString := strings.Join(role.Permissions, " ")

	perTableMode := 0
	if role.PerTableMode {
		perTableMode = 1
	}

	_, err := w.Db.Exec(`
		UPDATE workspace_roles
		SET name = ?, displayname = ?, description = ?, permissions = ?, per_table_mode = ?, updated_at = UNIX_TIMESTAMP()
		WHERE id = ? AND workspace_id = ?
	`, role.Name, role.DisplayName, role.Description, permissionsString, perTableMode, role.ID, workspaceID)
	if err != nil {
		return err
	}

	return nil
}

func (w *workspaceRepository) CreateWorkspaceMemberTx(tx *sql.Tx, workspaceMember *model.WorkspaceMember) error {
	query := `
	INSERT INTO workspace_members 
	(id, user_id, workspace_id, role, date_joined) 
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := tx.Exec(query,
		workspaceMember.ID,
		workspaceMember.UserID,
		workspaceMember.WorkspaceID,
		workspaceMember.Role,
		workspaceMember.DateJoined,
	)

	return err
}

func (w *workspaceRepository) CreateWorkspaceTableFieldStoreTx(tx *sql.Tx, id string, workspaceID string, tableID string, fieldName string, selectedType string, parentID string, field_display_name string) (*model.WorkspaceFieldData, error) {
	workspaceTableField := model.WorkspaceFieldData{
		ID:            id,
		WorkspaceID:   workspaceID,
		TableID:       tableID,
		Name:          fieldName,
		DisplayName:   field_display_name,
		FieldType:     selectedType,
		ParentFieldID: parentID,
	}

	query := `
	INSERT INTO workspace_fields
	(id, workspace_id, table_id, field_name, field_display_name, field_type, parent_field_id, created_at, updated_at, deleted_at, formula)
	VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0, ?)
	`

	_, err := tx.Exec(query,
		workspaceTableField.ID,
		workspaceTableField.WorkspaceID,
		workspaceTableField.TableID,
		workspaceTableField.Name,
		workspaceTableField.DisplayName,
		workspaceTableField.FieldType,
		workspaceTableField.ParentFieldID,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &workspaceTableField, nil
}

func generateUniquePrefix() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func (w *workspaceRepository) GetRow(workspaceID string) (*model.Workspace, error) {
	var workspace model.Workspace
	err := w.Db.QueryRow(`
        SELECT id, title, COALESCE(description, ''), COALESCE(start_date, 0), COALESCE(end_date, 0), pre_fix, created_at, COALESCE(updated_at, 0), COALESCE(deleted_at, 0)
        FROM workspaces
        WHERE id = ? AND (deleted_at IS NULL OR deleted_at = 0)
    `, workspaceID).Scan(
		&workspace.ID,
		&workspace.Title,
		&workspace.Description,
		&workspace.StartDate,
		&workspace.EndDate,
		&workspace.Prefix,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
		&workspace.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (w *workspaceRepository) GetFolders(workspaceID string) ([]model.WorkspaceFolder, error) {
	rows, err := w.Db.Query(`
        SELECT id, workspace_id, COALESCE(parent_folder_id, '') AS parent_folder_id, name, description,
               created_at, updated_at, deleted_at
        FROM workspace_folders
        WHERE workspace_id = ? AND (deleted_at IS NULL OR deleted_at = 0)
    `, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var folders []model.WorkspaceFolder
	for rows.Next() {
		var folder model.WorkspaceFolder
		err := rows.Scan(
			&folder.ID,
			&folder.WorkspaceID,
			&folder.ParentFolderID,
			&folder.Name,
			&folder.Description,
			&folder.CreatedAt,
			&folder.UpdatedAt,
			&folder.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		folder.IsFolder = true
		folder.Tables = []model.WorkspaceTable{}
		folder.Children = []model.WorkspaceFolder{}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *workspaceRepository) GetUserTimezone(userID string) (*model.UserTimezone, error) {
	var timezoneJSON []byte
	err := r.Db.QueryRow(`SELECT timezone FROM users WHERE id = ?`, userID).Scan(&timezoneJSON)
	if err != nil {
		return nil, err
	}

	var tz model.UserTimezone
	if err := json.Unmarshal(timezoneJSON, &tz); err != nil {
		return nil, err
	}

	return &tz, nil
}

func resolveDateOption(option string, loc *time.Location) (start int64, end int64, isRange bool) {
	if loc == nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	year, month, day := now.Date()
	startOfToday := time.Date(year, month, day, 0, 0, 0, 0, loc)

	switch option {
	case "Today":
		return startOfToday.Unix(), startOfToday.Add(24 * time.Hour).Unix(), true
	case "Yesterday":
		y := startOfToday.Add(-24 * time.Hour)
		return y.Unix(), startOfToday.Unix(), true
	case "Tomorrow":
		t := startOfToday.Add(24 * time.Hour)
		return t.Unix(), t.Add(24 * time.Hour).Unix(), true
	case "Next 7 days":
		return now.Unix(), now.AddDate(0, 0, 7).Unix(), true
	case "Last 7 days":
		return now.AddDate(0, 0, -7).Unix(), now.Unix(), true
	case "This week":
		offset := int(time.Monday - startOfToday.Weekday())
		if offset > 0 {
			offset = -6
		}

		start := startOfToday.AddDate(0, 0, offset)
		end := start.AddDate(0, 0, 7)
		return start.Unix(), end.Unix(), true
	case "Last week":
		offset := int(time.Monday - startOfToday.Weekday())
		if offset > 0 {
			offset = -6
		}

		start := startOfToday.AddDate(0, 0, offset-7)
		end := startOfToday.AddDate(0, 0, offset)
		return start.Unix(), end.Unix(), true
	case "Next week":
		offset := int(time.Monday - startOfToday.Weekday())
		if offset > 0 {
			offset = -6
		}

		start := startOfToday.AddDate(0, 0, offset+7)
		end := start.AddDate(0, 0, 7)
		return start.Unix(), end.Unix(), true
	case "This month":
		firstDay := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		lastDay := firstDay.AddDate(0, 1, 0)
		return firstDay.Unix(), lastDay.Unix(), true
	case "Last month":
		firstDay := time.Date(year, month-1, 1, 0, 0, 0, 0, loc)
		lastDay := firstDay.AddDate(0, 1, 0)
		return firstDay.Unix(), lastDay.Unix(), true
	case "Next month":
		firstDay := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
		lastDay := firstDay.AddDate(0, 1, 0)
		return firstDay.Unix(), lastDay.Unix(), true
	case "This year":
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
		end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc)
		return start.Unix(), end.Unix(), true
	case "Last year":
		start := time.Date(year-1, time.January, 1, 0, 0, 0, 0, loc)
		end := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
		return start.Unix(), end.Unix(), true
	case "Next year":
		start := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc)
		end := time.Date(year+2, time.January, 1, 0, 0, 0, 0, loc)
		return start.Unix(), end.Unix(), true
	case "Today & Earlier":
		return 0, startOfToday.Add(24 * time.Hour).Unix(), true
	case "Overdue":
		return 0, startOfToday.Unix(), true
	case "Later than Today":
		return startOfToday.Add(24 * time.Hour).Unix(), 0, true
	}

	return 0, 0, false
}

// filterHeader returns the column a filter names, or nil when the table has no
// such column. Field names reach SQL only after passing through here.
func filterHeader(headers []model.WorkspaceHeaders, field string) *model.WorkspaceHeaders {
	for i := range headers {
		if headers[i].Name == field {
			return &headers[i]
		}
	}

	return nil
}

func joinFilters(parts []model.SQLFilter, op string) model.SQLFilter {
	var joined model.SQLFilter
	sqls := make([]string, 0, len(parts))
	for _, p := range parts {
		sqls = append(sqls, p.SQL)
		joined.Args = append(joined.Args, p.Args...)
	}

	joined.SQL = strings.Join(sqls, " "+op+" ")
	return joined
}

// combineFilterConditions ORs each run of conditions and ANDs the runs, where
// a filter whose OperatorBetween is AND starts a new run.
func combineFilterConditions(filters []model.Filter, condition func(model.Filter, *model.WorkspaceHeaders) model.SQLFilter, headers []model.WorkspaceHeaders) model.SQLFilter {
	var groups, current []model.SQLFilter
	for i, f := range filters {
		if header := filterHeader(headers, f.Field); header != nil && f.Operator != "" {
			if c := condition(f, header); c.SQL != "" {
				current = append(current, c)
			}
		}

		if (i == len(filters)-1 || strings.EqualFold(filters[i+1].OperatorBetween, "AND")) && len(current) > 0 {
			group := joinFilters(current, "OR")
			group.SQL = "(" + group.SQL + ")"
			groups = append(groups, group)
			current = nil
		}
	}

	return joinFilters(groups, "AND")
}

func combineFilterGroups(groups []model.FilterGroup, build func([]model.Filter) model.SQLFilter) model.SQLFilter {
	var combined model.SQLFilter
	relation := ""
	for _, group := range groups {
		clause := build(group.Filters)
		if clause.SQL == "" {
			continue
		}

		if relation == "" {
			combined = clause
		} else {
			combined = joinFilters([]model.SQLFilter{combined, clause}, relation)
		}

		relation = "AND"
		if group.RelationToNext != nil && strings.EqualFold(*group.RelationToNext, "OR") {
			relation = "OR"
		}
	}

	if combined.SQL != "" {
		combined.SQL = "(" + combined.SQL + ")"
	}

	return combined
}

func dateBoundsCondition(col string, start, end int64) model.SQLFilter {
	switch {
	case end == 0:
		return model.SQLFilter{SQL: col + " >= ?", Args: []any{start}}
	case start == 0:
		return model.SQLFilter{SQL: col + " < ?", Args: []any{end}}
	}

	return model.SQLFilter{SQL: "(" + col + " >= ? AND " + col + " < ?)", Args: []any{start, end}}
}

func equalityCondition(col, op, value string) model.SQLFilter {
	switch op {
	case "is":
		return model.SQLFilter{SQL: col + " = ?", Args: []any{value}}
	case "is_not":
		return model.SQLFilter{SQL: "(" + col + " != ? OR " + col + " IS NULL)", Args: []any{value}}
	}

	return model.SQLFilter{}
}

func setCondition(col, op string) model.SQLFilter {
	switch op {
	case "is_set":
		return model.SQLFilter{SQL: "(" + col + " IS NOT NULL AND " + col + " != '')"}
	case "is_not_set":
		return model.SQLFilter{SQL: "(" + col + " IS NULL OR " + col + " = '')"}
	}

	return model.SQLFilter{}
}

func (w *workspaceRepository) buildFlatFilterSQLInline(filters []model.Filter, headers []model.WorkspaceHeaders, loc *time.Location, tableNames map[string]string) model.SQLFilter {
	if loc == nil {
		loc = time.UTC
	}

	return combineFilterConditions(filters, func(f model.Filter, header *model.WorkspaceHeaders) model.SQLFilter {
		return inlineFilterCondition(f, header, loc, tableNames)
	}, headers)
}

func inlineFilterCondition(f model.Filter, header *model.WorkspaceHeaders, loc *time.Location, tableNames map[string]string) model.SQLFilter {
	// A name that would not quote was not made by the server, so the
	// condition matches nothing rather than being left out.
	noMatch := model.SQLFilter{SQL: "1 = 0"}
	col, err := quoteIdent(header.Name)
	if err != nil {
		return noMatch
	}

	opU := strings.ToLower(f.Operator)

	if header.ParentTableID != "" && !header.SingleSelect {
		linkedTableName := tableNames[header.LinkedID]
		if linkedTableName == "" {
			return model.SQLFilter{}
		}

		linkTable, err := quoteIdent(linkedTableName)
		if err != nil {
			return noMatch
		}

		sidePred := "link.table_item_id = main.id"
		if header.LinkBothDirections {
			sidePred = "(link.table_item_id = main.id OR link.parent_table_item_id = main.id)"
		}

		switch opU {
		case "is_set":
			return model.SQLFilter{SQL: fmt.Sprintf(`EXISTS (
				SELECT 1
				FROM %s AS link
				WHERE %s
			)`, linkTable, sidePred)}

		case "is_not_set":
			return model.SQLFilter{SQL: fmt.Sprintf(`NOT EXISTS (
				SELECT 1
				FROM %s AS link
				WHERE %s
			)`, linkTable, sidePred)}

		case "is", "is_not":
			parentTableName := tableNames[header.ParentTableID]
			if parentTableName == "" {
				return model.SQLFilter{}
			}

			parentTable, err := quoteIdent(parentTableName)
			if err != nil {
				return noMatch
			}

			exists := "EXISTS"
			if opU == "is_not" {
				exists = "NOT EXISTS"
			}

			return model.SQLFilter{SQL: fmt.Sprintf(`%s (
				SELECT 1
				FROM %s AS link
				JOIN %s AS parent ON parent.id = link.parent_table_item_id
				WHERE %s
				AND parent.id = ?
			)`, exists, linkTable, parentTable, sidePred), Args: []any{f.Value}}
		}

		return model.SQLFilter{}
	}

	if f.Operator == "is_set" || f.Operator == "is_not_set" {
		return setCondition(col, f.Operator)
	}

	if header.Name == "start_date" || header.Name == "due_date" {
		if start, end, ok := inlineDateBounds(f, loc); ok {
			return dateBoundsCondition(col, start, end)
		}

		return model.SQLFilter{SQL: col + " = ?", Args: []any{f.Value}}
	}

	if header.Name == "status" && len(f.Values) > 0 {
		in := "(" + sqlPlaceholders(len(f.Values)) + ")"
		args := toInterfaceSlice(f.Values)
		switch opU {
		case "is":
			return model.SQLFilter{SQL: col + " IN " + in, Args: args}
		case "is_not":
			return model.SQLFilter{SQL: "(" + col + " NOT IN " + in + " OR " + col + " IS NULL)", Args: args}
		}
	}

	return equalityCondition(col, f.Operator, f.Value)
}

func inlineDateBounds(f model.Filter, loc *time.Location) (int64, int64, bool) {
	start, end, ok := resolveDateOption(f.Value, loc)

	if f.Date != "" {
		parsed, err := time.Parse("2006-01-02", f.Date)
		if err == nil {
			if loc != nil {
				parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc)
			}

			switch f.Value {
			case "Exact date":
				start = parsed.Unix()
				end = parsed.Add(24 * time.Hour).Unix()
				ok = true
			case "After date":
				start = parsed.Add(24 * time.Hour).Unix()
				end = 0
				ok = true
			case "Before date":
				start = 0
				end = parsed.Unix()
				ok = true
			case "Date range":
				if f.Date2 != "" {
					parsed2, err := time.Parse("2006-01-02", f.Date2)
					if err == nil {
						if loc != nil {
							parsed2 = time.Date(parsed2.Year(), parsed2.Month(), parsed2.Day(), 0, 0, 0, 0, loc)
						}

						start = parsed.Unix()
						end = parsed2.Add(24 * time.Hour).Unix()
						ok = true
					}
				}
			}
		}
	}

	return start, end, ok
}

func (w *workspaceRepository) buildGroupFilterSQL(groups []model.FilterGroup, headers []model.WorkspaceHeaders, loc *time.Location, tableNames map[string]string) model.SQLFilter {
	return combineFilterGroups(groups, func(filters []model.Filter) model.SQLFilter {
		return w.buildFlatFilterSQLInline(filters, headers, loc, tableNames)
	})
}

func (w *workspaceRepository) GetAll(ctx context.Context) ([]model.Workspace, error) {
	rows, err := w.Db.QueryContext(ctx, `
        SELECT id, title, COALESCE(description, ''), COALESCE(start_date, 0), COALESCE(end_date, 0), pre_fix, created_at, COALESCE(updated_at, 0), COALESCE(deleted_at, 0)
        FROM workspaces
        WHERE deleted_at IS NULL OR deleted_at = 0
        ORDER BY title ASC, id ASC
    `)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workspaces := make([]model.Workspace, 0)
	for rows.Next() {
		var workspace model.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Title,
			&workspace.Description,
			&workspace.StartDate,
			&workspace.EndDate,
			&workspace.Prefix,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&workspace.DeletedAt,
		)
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

func (w *workspaceRepository) GetAllForUser(userID string) ([]model.Workspace, error) {
	workspaces := make([]model.Workspace, 0)

	rows, err := w.Db.Query(`
        SELECT w.id, w.title, COALESCE(w.description, ''), COALESCE(w.start_date, 0), COALESCE(w.end_date, 0), w.pre_fix, w.created_at, COALESCE(w.updated_at, 0), COALESCE(w.deleted_at, 0)
        FROM (
          SELECT wm.workspace_id FROM workspace_members wm WHERE wm.user_id = ?
          UNION
          SELECT wg.workspace_id
          FROM group_members gm
          JOIN user_groups ug ON ug.id = gm.group_id AND ug.deleted_at = 0
          JOIN workspace_groups wg ON wg.group_id = gm.group_id
          WHERE gm.user_id = ?
        ) access
        JOIN workspaces w ON w.id = access.workspace_id
        WHERE (w.deleted_at IS NULL OR w.deleted_at = 0)
        ORDER BY w.title ASC
    `, userID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var workspace model.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.Title,
			&workspace.Description,
			&workspace.StartDate,
			&workspace.EndDate,
			&workspace.Prefix,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
			&workspace.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

// GetCustomFields returns a table's fields by column and label, the label
// being the column only for a field saved without one.
func (w *workspaceRepository) GetCustomFields(workspaceID, tableID string) ([]model.TaskOrderField, error) {
	rows, err := w.Db.Query(`
		SELECT field_name, COALESCE(NULLIF(field_display_name, ''), field_name) FROM workspace_fields
		WHERE workspace_id = ? AND table_id = ? AND deleted_at = 0
	`, workspaceID, tableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var names []model.TaskOrderField
	for rows.Next() {
		var field model.TaskOrderField
		if err := rows.Scan(&field.Name, &field.DisplayName); err != nil {
			return nil, err
		}

		names = append(names, field)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func (w *workspaceRepository) UpdateFieldDisplayName(workspaceID, tableID, fieldName, displayName string) error {
	_, err := w.Db.Exec(`
		UPDATE workspace_fields
		SET field_display_name = ?
		WHERE workspace_id = ? AND table_id = ? AND field_name = ?
	`, displayName, workspaceID, tableID, fieldName)
	return err
}

// setColumnWidthInOrder finds the column matching columnName in a grid view item_order JSON array
// and updates its width. Returns the updated JSON, whether the column was found, and any error.
func setColumnWidthInOrder(itemOrderJSON, columnName string, width int) (string, bool, error) {
	var columns []map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrderJSON), &columns); err != nil {
		return "", false, err
	}

	for i, col := range columns {
		if name, ok := col["name"].(string); ok && name == columnName {
			col["width"] = strconv.Itoa(width)
			columns[i] = col
			updated, err := json.Marshal(columns)
			if err != nil {
				return "", false, err
			}

			return string(updated), true, nil
		}
	}

	return itemOrderJSON, false, nil
}

func (w *workspaceRepository) UpdateColumnWidth(workspaceID, tableID, viewID, columnName string, width int) error {
	var currentOrderJSON string
	err := w.Db.QueryRow(`SELECT item_order FROM workspace_view WHERE workspace_id = ? AND table_id = ? AND view_type = 'grid' AND id = ?`,
		workspaceID, tableID, viewID).Scan(&currentOrderJSON)
	if err != nil {
		return err
	}

	updatedJSON, found, err := setColumnWidthInOrder(currentOrderJSON, columnName, width)
	if err != nil {
		return err
	}

	if !found {
		return sql.ErrNoRows
	}

	_, err = w.Db.Exec(`UPDATE workspace_view SET item_order = ? WHERE workspace_id = ? AND table_id = ? AND view_type = 'grid' AND id = ?`,
		updatedJSON, workspaceID, tableID, viewID)
	return err
}

type tableViewKey struct{ tableID, viewID string }

// GetTableViewMetadata loads what hydrating a project needs for each table
// besides its rows, in a fixed number of queries however many tables and
// views there are.
func (w *workspaceRepository) GetTableViewMetadata(workspaceID, userID string, tableIDs []string) (map[string]*model.WorkspaceTableViewMetadata, error) {
	out := make(map[string]*model.WorkspaceTableViewMetadata, len(tableIDs))
	for _, id := range tableIDs {
		out[id] = &model.WorkspaceTableViewMetadata{Views: []model.WorkspaceView{}, GridSettings: []model.WorkspaceView{}}
	}

	if len(tableIDs) == 0 {
		return out, nil
	}

	views, err := w.viewsByTable(workspaceID, tableIDs)
	if err != nil {
		return nil, err
	}

	grid, err := w.gridSettingsByTable(workspaceID, tableIDs)
	if err != nil {
		return nil, err
	}

	saved, err := w.savedFiltersByTable(workspaceID, userID, tableIDs)
	if err != nil {
		return nil, err
	}

	var viewIDs []string
	for _, byTable := range []map[string][]model.WorkspaceView{views, grid} {
		for _, list := range byTable {
			for _, v := range list {
				viewIDs = append(viewIDs, v.ID)
			}
		}
	}

	filters, err := w.activeFiltersByView(workspaceID, userID, viewIDs)
	if err != nil {
		return nil, err
	}

	sorts, err := w.sortsByView(workspaceID, viewIDs)
	if err != nil {
		return nil, err
	}

	for _, id := range tableIDs {
		m := out[id]
		m.Views = append(m.Views, attachFilterAndSort(id, views[id], filters, sorts)...)
		m.GridSettings = append(m.GridSettings, attachFilterAndSort(id, grid[id], filters, sorts)...)
		m.SavedFilters = saved[id]
	}

	return out, nil
}

func attachFilterAndSort(tableID string, views []model.WorkspaceView, filters map[tableViewKey]*model.FilterPayload, sorts map[tableViewKey][]map[string]interface{}) []model.WorkspaceView {
	for i := range views {
		key := tableViewKey{tableID, views[i].ID}
		views[i].Filter = filters[key]
		views[i].Sort = sorts[key]
	}

	return views
}

// The ORDER BY is the order the (table_id, view_type) index returned views in
// when they were read one table at a time. Form views were removed but their
// rows were kept, so they are left out here.
func (w *workspaceRepository) viewsByTable(workspaceID string, tableIDs []string) (map[string][]model.WorkspaceView, error) {
	args := append([]any{workspaceID}, toInterfaceSlice(tableIDs)...)
	rows, err := w.Db.Query(`
		SELECT id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, created_at, updated_at, deleted_at, COALESCE(is_public, 1)
		FROM workspace_view
		WHERE workspace_id = ? AND table_id IN (`+sqlPlaceholders(len(tableIDs))+`) AND main_view = 0 AND view_type <> 'form'
		ORDER BY table_id, view_type, id
	`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	views := make(map[string][]model.WorkspaceView)
	for rows.Next() {
		var v model.WorkspaceView
		if err := rows.Scan(
			&v.ID,
			&v.WorkspaceID,
			&v.TableID,
			&v.ViewType,
			&v.ParentTableID,
			&v.Name,
			&v.TaskOrder,
			&v.CreatedBy,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.DeletedAt,
			&v.IsPublic,
		); err != nil {
			return nil, err
		}

		views[v.TableID] = append(views[v.TableID], v)
	}

	return views, rows.Err()
}

func (w *workspaceRepository) gridSettingsByTable(workspaceID string, tableIDs []string) (map[string][]model.WorkspaceView, error) {
	args := append([]any{workspaceID, "grid"}, toInterfaceSlice(tableIDs)...)
	rows, err := w.Db.Query(`
		SELECT id, workspace_id, table_id, view_type, name, item_order, created_by, created_at, updated_at, deleted_at, COALESCE(is_public, 1)
		FROM workspace_view
		WHERE workspace_id = ? AND view_type = ? AND table_id IN (`+sqlPlaceholders(len(tableIDs))+`) AND main_view = 1
		ORDER BY table_id, id
	`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	views := make(map[string][]model.WorkspaceView)
	for rows.Next() {
		var v model.WorkspaceView
		if err := rows.Scan(
			&v.ID,
			&v.WorkspaceID,
			&v.TableID,
			&v.ViewType,
			&v.Name,
			&v.TaskOrder,
			&v.CreatedBy,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.DeletedAt,
			&v.IsPublic,
		); err != nil {
			return nil, err
		}

		views[v.TableID] = append(views[v.TableID], v)
	}

	return views, rows.Err()
}

func (w *workspaceRepository) savedFiltersByTable(workspaceID, userID string, tableIDs []string) (map[string][]model.SavedFilterWithPayload, error) {
	args := append([]any{workspaceID}, toInterfaceSlice(tableIDs)...)
	rows, err := w.Db.Query(`
	SELECT
		wsf.id,
		wsf.filter_id,
		wf.name,
		wf.is_private,
		wsf.is_active,
		wsf.created_at,
		wsf.table_id,
		wsf.view_id,
		wf.filter_settings
	FROM workspace_saved_filters wsf
	INNER JOIN workspace_filters wf ON wf.id = wsf.filter_id
	WHERE wsf.workspace_id = ?
		AND wsf.table_id IN (`+sqlPlaceholders(len(tableIDs))+`)
		AND (wf.is_private = 0 OR (wf.is_private = 1 AND wsf.user_id = ?))
	ORDER BY wsf.created_at DESC
	`, append(args, userID)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	filters := make(map[string][]model.SavedFilterWithPayload)
	for rows.Next() {
		var f model.SavedFilterWithPayload
		var viewID sql.NullString
		if err := rows.Scan(
			&f.ID,
			&f.FilterID,
			&f.Name,
			&f.IsPrivate,
			&f.IsActive,
			&f.CreatedAt,
			&f.TableID,
			&viewID,
			&f.FilterSettings,
		); err != nil {
			return nil, err
		}

		f.UserID = userID
		f.ViewID = viewID.String
		filters[f.TableID] = append(filters[f.TableID], f)
	}

	return filters, rows.Err()
}

// activeFiltersByView returns each view's newest active filter saved by the user.
func (w *workspaceRepository) activeFiltersByView(workspaceID, userID string, viewIDs []string) (map[tableViewKey]*model.FilterPayload, error) {
	filters := make(map[tableViewKey]*model.FilterPayload)
	if len(viewIDs) == 0 {
		return filters, nil
	}

	args := append(toInterfaceSlice(viewIDs), userID, workspaceID)
	rows, err := w.Db.Query(`
	SELECT wf.table_id, wsf.view_id, wf.filter_settings
	FROM workspace_saved_filters wsf
	INNER JOIN workspace_filters wf ON wf.id = wsf.filter_id
	WHERE wsf.view_id IN (`+sqlPlaceholders(len(viewIDs))+`)
		AND wsf.user_id = ?
		AND wsf.is_active = 1
		AND wf.workspace_id = ?
	ORDER BY wsf.created_at DESC
	`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	newest := make(map[tableViewKey]string)
	for rows.Next() {
		var key tableViewKey
		var settings string
		if err := rows.Scan(&key.tableID, &key.viewID, &settings); err != nil {
			return nil, err
		}

		if _, seen := newest[key]; !seen {
			newest[key] = settings
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for key, settings := range newest {
		var payload model.FilterPayload
		if err := json.Unmarshal([]byte(settings), &payload); err != nil {
			return nil, err
		}

		filters[key] = &payload
	}

	return filters, nil
}

func (w *workspaceRepository) sortsByView(workspaceID string, viewIDs []string) (map[tableViewKey][]map[string]interface{}, error) {
	sorts := make(map[tableViewKey][]map[string]interface{})
	if len(viewIDs) == 0 {
		return sorts, nil
	}

	args := append(toInterfaceSlice(viewIDs), workspaceID)
	rows, err := w.Db.Query(`
	SELECT table_id, view_id, sort_settings
	FROM workspace_sort
	WHERE view_id IN (`+sqlPlaceholders(len(viewIDs))+`) AND workspace_id = ?
	ORDER BY created_at DESC
	`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	newest := make(map[tableViewKey]string)
	for rows.Next() {
		var key tableViewKey
		var settings string
		if err := rows.Scan(&key.tableID, &key.viewID, &settings); err != nil {
			return nil, err
		}

		if _, seen := newest[key]; !seen {
			newest[key] = settings
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for key, settings := range newest {
		var payload []map[string]interface{}
		if err := json.Unmarshal([]byte(settings), &payload); err != nil {
			return nil, err
		}

		sorts[key] = payload
	}

	return sorts, nil
}

func (w *workspaceRepository) GetNameByID(workspaceID string) (*string, error) {
	var name string
	query := `SELECT title FROM workspaces WHERE id = ?`
	err := w.Db.QueryRow(query, workspaceID).Scan(&name)
	if err != nil {
		return nil, err
	}

	return &name, nil
}

func (w *workspaceRepository) GetTaskNameByID(workspaceID string, tableID string, taskID string) (*string, error) {
	var tableName string
	query := `SELECT name FROM workspace_tables WHERE workspace_id = ? AND id = ?`
	err := w.Db.QueryRow(query, workspaceID, tableID).Scan(&tableName)
	if err != nil {
		return nil, err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf(`SELECT name FROM %s WHERE id = ?`, table)
	var taskName string
	err = w.Db.QueryRow(query, taskID).Scan(&taskName)
	if err != nil {
		return nil, err
	}

	return &taskName, nil
}

func (w *workspaceRepository) GetTaskAssigneeID(tableName, taskID string) (string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return "", err
	}

	var assigneeUserID string
	err = w.Db.QueryRow(fmt.Sprintf(`SELECT assignee FROM %s WHERE id = ?`, table), taskID).Scan(&assigneeUserID)
	if err != nil {
		return "", err
	}

	return assigneeUserID, nil
}

func (w *workspaceRepository) GetMemberByUserID(workspaceID, userID string) (*model.WorkspaceMember, error) {
	var member model.WorkspaceMember
	err := w.Db.QueryRow(`
		SELECT id, user_id, workspace_id, role, date_joined
		FROM workspace_members WHERE user_id = ? AND workspace_id = ?
	`, userID, workspaceID).Scan(&member.ID, &member.UserID, &member.WorkspaceID, &member.Role, &member.DateJoined)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

// addFieldToKanbanTaskOrder appends newField to a kanban view's item_order JSON
// if the view's section matches the given field name. Returns updated JSON and
// whether the view was modified.
func addFieldToKanbanTaskOrder(itemOrderJSON, field string, newField model.TaskOrderField) (updated string, modified bool, err error) {
	var taskOrder model.TaskOrder
	if err = json.Unmarshal([]byte(itemOrderJSON), &taskOrder); err != nil {
		return "", false, err
	}

	if taskOrder.Section != field {
		return itemOrderJSON, false, nil
	}

	taskOrder.Fields = append(taskOrder.Fields, newField)
	b, err := json.Marshal(taskOrder)
	if err != nil {
		return "", false, err
	}

	return string(b), true, nil
}

func (w *workspaceRepository) CreateFieldValue(workspaceID string, tableID string, linkedTableID string, tableName string, name string, field string, optionID string, color string) (*model.TaskOrderField, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf(`INSERT INTO %s (id, name, color) VALUES (?, ?, ?)`, table), optionID, name, color)
	if err != nil {
		return nil, err
	}

	newField := model.TaskOrderField{
		Color:       color,
		ID:          optionID,
		Name:        name,
		DisplayName: "",
		Width:       "300",
		Order:       []map[string]interface{}{},
		Visible:     true,
	}

	rows, err := tx.Query(`
		SELECT id, item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND view_type = 'kanban'`, workspaceID, tableID)
	if err != nil {
		return nil, err
	}

	type viewData struct {
		viewID       string
		itemOrderRaw string
	}

	var viewRows []viewData
	for rows.Next() {
		var vd viewData
		if err := rows.Scan(&vd.viewID, &vd.itemOrderRaw); err != nil {
			rows.Close()
			return nil, err
		}

		viewRows = append(viewRows, vd)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, row := range viewRows {
		updatedJSON, modified, err := addFieldToKanbanTaskOrder(row.itemOrderRaw, field, newField)
		if err != nil {
			return nil, err
		}

		if !modified {
			continue
		}

		if _, err := tx.Exec(`UPDATE workspace_view SET item_order = ? WHERE id = ?`, updatedJSON, row.viewID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &newField, nil
}

// addTaskToUnassignedSection appends task to the "Unassigned" section of a kanban item_order JSON blob.
// Returns the updated JSON, whether a modification was made, and any error.
func addTaskToUnassignedSection(itemOrderJSON string, task map[string]interface{}) (string, bool, error) {
	var itemOrder map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrderJSON), &itemOrder); err != nil {
		return "", false, err
	}

	fields, ok := itemOrder["fields"].([]interface{})
	if !ok {
		return itemOrderJSON, false, nil
	}

	for _, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}

		if fieldMap["name"] == "Unassigned" {
			order, ok := fieldMap["order"].([]interface{})
			if !ok {
				continue
			}

			fieldMap["order"] = append(order, map[string]interface{}{
				"id":            task["id"],
				"single_select": fieldMap["id"],
			})
			updated, err := json.Marshal(itemOrder)
			if err != nil {
				return "", false, err
			}

			return string(updated), true, nil
		}
	}

	return itemOrderJSON, false, nil
}

func (w *workspaceRepository) UpdateKanbanViewsTx(tx *sql.Tx, workspaceID string, tableID string, task map[string]interface{}) error {
	if tx == nil {
		return sql.ErrConnDone
	}

	var ping string
	err := tx.QueryRow(`SELECT 'ping'`).Scan(&ping)
	if err != nil {
		return err
	}

	rows, err := tx.Query(`
        SELECT id, item_order
        FROM workspace_view
        WHERE workspace_id = ? AND table_id = ? AND view_type = 'kanban'
    `, workspaceID, tableID)
	if err != nil {
		return err
	}

	defer rows.Close()

	type viewData struct {
		id       string
		itemJSON string
	}

	var buffered []viewData

	for rows.Next() {
		var vd viewData
		if err := rows.Scan(&vd.id, &vd.itemJSON); err != nil {
			return err
		}

		buffered = append(buffered, vd)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if err := rows.Close(); err != nil {
		return err
	}

	for _, view := range buffered {
		updatedJSON, modified, err := addTaskToUnassignedSection(view.itemJSON, task)
		if err != nil {
			return err
		}

		if !modified {
			continue
		}

		_, err = tx.Exec(`
            UPDATE workspace_view
            SET item_order = ?
            WHERE id = ?
        `, updatedJSON, view.id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) DeleteAttachment(fileID, workspaceID string) error {
	result, err := w.Db.Exec(`
        DELETE FROM workspace_attachments
        WHERE id = ? AND workspace_id = ?
    `, fileID, workspaceID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (w *workspaceRepository) UpdateGridSort(workspaceID, tableID, viewID, sortData, sortID string) (bool, error) {
	_, err := w.Db.Exec(`
		INSERT INTO workspace_sort (id, workspace_id, table_id, view_id, sort_settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
		ON DUPLICATE KEY UPDATE sort_settings = VALUES(sort_settings), updated_at = VALUES(updated_at)
	`, sortID, workspaceID, tableID, viewID, sortData)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (w *workspaceRepository) Delete(workspaceID string) (bool, error) {
	_, err := w.Db.Exec("UPDATE workspaces SET deleted_at = UNIX_TIMESTAMP() WHERE id = ?", workspaceID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (w *workspaceRepository) GetTableList(workspaceID, userID string) ([]model.WorkspaceTable, error) {
	rows, err := w.Db.Query(`
		SELECT wt.id, wt.name, wt.display_name, wt.parent_table_id, wt.linked, wt.single_select, folder_id
		FROM workspace_tables wt
		WHERE wt.workspace_id = ? AND wt.deleted_at = 0
		  AND (
		    EXISTS (SELECT 1 FROM workspace_members wm WHERE wm.workspace_id = wt.workspace_id AND wm.user_id = ?)
		    OR EXISTS (
		      SELECT 1 FROM workspace_groups wg
		      JOIN group_members gm ON gm.group_id = wg.group_id
		      JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		      WHERE wg.workspace_id = wt.workspace_id AND gm.user_id = ?
		    )
		  )
	`, workspaceID, userID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tables []model.WorkspaceTable
	for rows.Next() {
		var t model.WorkspaceTable
		var parentTableID, folderID sql.NullString
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.DisplayName,
			&parentTableID,
			&t.Linked,
			&t.SingleSelect,
			&folderID,
		); err != nil {
			return nil, err
		}

		t.ParentTableID = parentTableID.String
		t.FolderID = folderID.String
		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (w *workspaceRepository) GetAllTablesBasic(workspaceID string) ([]model.WorkspaceTable, error) {
	rows, err := w.Db.Query(`
		SELECT id, name, display_name, parent_table_id, linked, single_select, folder_id
		FROM workspace_tables
		WHERE workspace_id = ? AND deleted_at = 0
	`, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tables []model.WorkspaceTable
	for rows.Next() {
		var t model.WorkspaceTable
		var parentTableID, folderID sql.NullString
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.DisplayName,
			&parentTableID,
			&t.Linked,
			&t.SingleSelect,
			&folderID,
		); err != nil {
			return nil, err
		}

		t.ParentTableID = parentTableID.String
		t.FolderID = folderID.String
		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (w *workspaceRepository) GetTableFilters(workspaceID, tableID, userID string) ([]model.SavedFilterWithPayload, error) {
	byTable, err := w.savedFiltersByTable(workspaceID, userID, []string{tableID})
	if err != nil {
		return nil, err
	}

	return byTable[tableID], nil
}

func removePrefix(name string) string {
	parts := strings.Split(name, "_")
	if len(parts) > 1 {
		return strings.Join(parts[1:], "_")
	}

	return name
}

func (w *workspaceRepository) parentTableID(tableID string) string {
	var parentID string

	query := `
		SELECT parent_table_id
		FROM workspace_tables
		WHERE id = ?`

	err := w.Db.QueryRow(query, tableID).Scan(&parentID)
	if err != nil {
		return ""
	}

	return parentID
}

func (w *workspaceRepository) linkTable(tableID string, tableName string) string {
	var linkedTableID string
	query := `
        SELECT linked_table_id
        FROM workspace_relationships
        WHERE table_id = ? AND table_name = ?`

	err := w.Db.QueryRow(query, tableID, tableName).Scan(&linkedTableID)
	if err != nil {
		return ""
	}

	return linkedTableID
}

// Public wrappers for private single-query helpers

func (w *workspaceRepository) GetTableParentID(tableID string) string {
	return w.parentTableID(tableID)
}

func (w *workspaceRepository) GetLinkedTableIDByName(tableID, name string) string {
	return w.linkTable(tableID, name)
}

// GetTableColumnTypes returns the column types for a table via a schema probe query.
func (w *workspaceRepository) GetTableColumnTypes(tableName string) ([]*sql.ColumnType, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 0", table))
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return rows.ColumnTypes()
}

// scanRowsToMaps is an internal helper that reads all rows into []map[string]interface{}.
func scanRowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range columns {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}

		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetTasksByDateRange returns the tasks whose dates touch [from, to], those
// nearest today first, at most limit of them when limit is above zero, with
// how many there are in all.
func (w *workspaceRepository) GetTasksByDateRange(ctx context.Context, tableName string, from, to int64, filter model.SQLFilter, limit int) ([]map[string]interface{}, int, error) {
	// Include tasks where start_date OR due_date falls in the window so
	// tasks with only one date set still appear in the gantt.
	where := `(deleted_at IS NULL OR deleted_at = 0)
		AND (
			(start_date IS NOT NULL AND start_date != 0 AND start_date BETWEEN ? AND ?)
			OR (due_date IS NOT NULL AND due_date != 0 AND due_date BETWEEN ? AND ?)
			OR (start_date IS NOT NULL AND start_date != 0 AND start_date < ?
			    AND due_date IS NOT NULL AND due_date != 0 AND due_date > ?)
		)`
	if filter.SQL != "" {
		where += " AND (" + filter.SQL + ")"
	}

	args := append([]any{from, to, from, to, from, from}, filter.Args...)

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = w.Db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where), args...).Scan(&total)
	if err != nil {
		// Tables without date columns (linked/status tables) return empty, not an error.
		if strings.Contains(err.Error(), "Unknown column") {
			return []map[string]interface{}{}, 0, nil
		}

		return nil, 0, err
	}

	q := fmt.Sprintf(`
		SELECT id, name, start_date, due_date FROM %s
		WHERE %s
		ORDER BY ABS(CAST(COALESCE(NULLIF(start_date, 0), due_date) AS SIGNED) - UNIX_TIMESTAMP()), id
	`, table, where)
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := w.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	items, err := scanRowsToMaps(rows)
	if items == nil {
		items = []map[string]interface{}{}
	}

	return items, total, err
}

// GetLinkedRecordsLite returns one page of a table's rows as id and name:
// those whose name contains search, or, when ids is given, those rows. Task
// tables list their newest rows first and leave out deleted ones; option
// tables, which have neither column, list by name.
// GetTaskAssignees returns who each of the tasks is assigned to, "" for no
// one.
func (w *workspaceRepository) GetTaskAssignees(ctx context.Context, tableName string, taskIDs []string) (map[string]string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	assignees := map[string]string{}
	for chunk := range slices.Chunk(taskIDs, 1000) {
		rows, err := w.Db.QueryContext(ctx, "SELECT id, COALESCE(assignee, '') FROM "+table+" WHERE id IN ("+sqlPlaceholders(len(chunk))+")", toInterfaceSlice(chunk)...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id, assignee string
			if err := rows.Scan(&id, &assignee); err != nil {
				rows.Close()
				return nil, err
			}

			assignees[id] = assignee
		}

		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return assignees, nil
}

// GetLinkedRecordsLite lists a table's tasks, by id and name, for a picker
// that links to them. access limits them to the tasks the user may see.
func (w *workspaceRepository) GetLinkedRecordsLite(ctx context.Context, tableName, search string, ids []string, limit, offset int, access model.SQLFilter) ([]map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	columns, err := w.GetColumnNames(ctx, tableName)
	if err != nil {
		return nil, err
	}

	where := []string{"1 = 1"}
	var args []any
	if slices.Contains(columns, "deleted_at") {
		where = append(where, "(deleted_at IS NULL OR deleted_at = 0)")
	}

	if len(ids) > 0 {
		where = append(where, "id IN ("+sqlPlaceholders(len(ids))+")")
		args = append(args, toInterfaceSlice(ids)...)
	}

	if search != "" {
		where = append(where, "name LIKE ?")
		args = append(args, "%"+escapeLike(search)+"%")
	}

	if access.SQL != "" {
		where = append(where, access.SQL)
		args = append(args, access.Args...)
	}
	// Browsing lists the newest first, both columns descending so the
	// created_at index can be read backwards. A search is listed by name
	// instead: walking that index for a name that few rows contain reads
	// every row one by one, far slower than one pass over the table.
	order := "name, id"
	if search == "" && slices.Contains(columns, "created_at") {
		order = "created_at DESC, id DESC"
	}

	args = append(args, limit, offset)

	rows, err := w.Db.QueryContext(ctx, "SELECT id, name FROM "+table+" WHERE "+strings.Join(where, " AND ")+
		" ORDER BY "+order+" LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items, err := scanRowsToMaps(rows)
	if items == nil {
		items = []map[string]interface{}{}
	}

	return items, err
}

// GetTasksByCalendarRange returns tasks whose date range overlaps [from, to]:
// starts in window, ends in window, or spans the entire window.
func (w *workspaceRepository) GetTasksByCalendarRange(ctx context.Context, tableName string, from, to int64, filter model.SQLFilter) ([]map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
		SELECT id, name, start_date, due_date FROM %s
		WHERE (deleted_at IS NULL OR deleted_at = 0)
		AND (
			(start_date IS NOT NULL AND start_date != 0 AND start_date BETWEEN ? AND ?)
			OR (due_date  IS NOT NULL AND due_date  != 0 AND due_date  BETWEEN ? AND ?)
			OR (start_date IS NOT NULL AND start_date != 0 AND start_date < ?
				AND due_date IS NOT NULL AND due_date != 0 AND due_date > ?)
		)
	`, table)
	if filter.SQL != "" {
		q += " AND (" + filter.SQL + ")"
	}

	rows, err := w.Db.QueryContext(ctx, q, append([]any{from, to, from, to, from, to}, filter.Args...)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return scanRowsToMaps(rows)
}

func (w *workspaceRepository) GetTableRows(tableName string) ([]map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	deletedAtIdx := -1
	for i, col := range columns {
		if col == "deleted_at" {
			deletedAtIdx = i
			break
		}
	}

	var items []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range columns {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		if deletedAtIdx >= 0 {
			switch v := values[deletedAtIdx].(type) {
			case int64:
				if v != 0 {
					continue
				}
			case []byte:
				if string(v) != "0" && string(v) != "" {
					continue
				}
			}
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}

		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetTableRowsFiltered returns non-deleted rows with an optional extra WHERE clause appended.
// The table is aliased as "main", which the filter builders' conditions refer to.
func (w *workspaceRepository) GetTableRowsFiltered(ctx context.Context, tableName string, filter model.SQLFilter) ([]map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT main.* FROM %s AS main WHERE (main.deleted_at IS NULL OR main.deleted_at = 0)", table)
	if filter.SQL != "" {
		query += " AND " + filter.SQL
	}

	rows, err := w.Db.QueryContext(ctx, query, filter.Args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return scanRowsToMaps(rows)
}

// GetRootTasksPagedWithSubtasks pages by the tasks that head a branch, as
// effectiveRootsFrom picks them, and appends every task below them. access
// limits which tasks the user may see at all; filter is the view's filter.
// The branches are counted only when count is set, since counting reads
// the whole table.
func (w *workspaceRepository) GetRootTasksPagedWithSubtasks(ctx context.Context, tableName string, access, filter model.SQLFilter, sortSQL string, limit, offset int, count bool) ([]map[string]interface{}, int, int, error) {
	deletedWhere := "(main.deleted_at IS NULL OR main.deleted_at = 0)"

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, 0, 0, err
	}

	roots, rootArgs := effectiveRootsFrom(table, access, filter)

	var total int
	if count {
		if total, err = w.CountBranches(ctx, tableName, access, filter); err != nil {
			return nil, 0, 0, err
		}
	}

	q := fmt.Sprintf("SELECT main.* FROM %s %s", roots, sortSQL)
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := w.Db.QueryContext(ctx, q, rootArgs...)
	if err != nil {
		return nil, 0, 0, err
	}

	defer rows.Close()
	all, err := scanRowsToMaps(rows)
	if err != nil {
		return nil, 0, 0, err
	}

	// BFS: collect descendants at every depth level until no more children are found.
	// Filter applies only to root tasks; descendants are included regardless of status.
	// A task that is its own parent, or a parent cycle, would otherwise be
	// found again at every level and never end the walk.
	seen := make(map[string]bool, len(all))
	currentIDs := make([]any, 0, len(all))
	for _, r := range all {
		if id, ok := r["id"]; ok && id != nil {
			key := fmt.Sprintf("%v", id)
			seen[key] = true
			currentIDs = append(currentIDs, key)
		}
	}

	for len(currentIDs) > 0 {
		subQ := fmt.Sprintf(
			"SELECT main.* FROM %s AS main WHERE %s AND main.parent_task_id IN (%s)",
			table, deletedWhere, sqlPlaceholders(len(currentIDs)),
		)
		subRows, err := w.Db.QueryContext(ctx, subQ, currentIDs...)
		if err != nil {
			return nil, 0, 0, err
		}

		children, err := scanRowsToMaps(subRows)
		subRows.Close()
		if err != nil {
			return nil, 0, 0, err
		}

		currentIDs = currentIDs[:0]
		for _, c := range children {
			id, ok := c["id"]
			if !ok || id == nil {
				continue
			}

			key := fmt.Sprintf("%v", id)
			if seen[key] {
				continue
			}

			seen[key] = true
			all = append(all, c)
			currentIDs = append(currentIDs, key)
		}
	}

	return all, total, total, nil
}

// CountBranches returns how many branches GetRootTasksPagedWithSubtasks pages
// by under access and filter.
func (w *workspaceRepository) CountBranches(ctx context.Context, tableName string, access, filter model.SQLFilter) (int, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return 0, err
	}

	roots, args := effectiveRootsFrom(table, access, filter)

	var total int
	err = w.Db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+roots, args...).Scan(&total)
	return total, err
}

// effectiveRootsFrom returns the FROM and WHERE of a query over the tasks
// that head a branch of the grid: those that pass access and filter and have
// no parent the user may see. A subtask whose parent the filter hides is
// hidden with it rather than shown on its own. A deleted task takes its
// subtasks with it, so a task's parent is missing only when access hides it;
// without access having no parent is enough, which the database counts from
// the (parent_task_id, deleted_at) index alone. The joined derived table has
// its own "main" alias, so access is applied to the parent row unchanged.
// table is already quoted.
func effectiveRootsFrom(table string, access, filter model.SQLFilter) (string, []any) {
	visible := "(main.deleted_at IS NULL OR main.deleted_at = 0)"
	if access.SQL != "" {
		visible += " AND " + access.SQL
	}

	shown := visible
	if filter.SQL != "" {
		shown += " AND " + filter.SQL
	}

	if access.SQL == "" {
		return table + " AS main WHERE " + shown + " AND " + model.TopLevelTaskSQL, slices.Clone(filter.Args)
	}

	from := fmt.Sprintf(`%s AS main
		LEFT JOIN (SELECT main.id FROM %s AS main WHERE %s) AS visible_parent ON visible_parent.id = main.parent_task_id
		WHERE %s AND (main.parent_task_id IS NULL OR visible_parent.id IS NULL)`,
		table, table, visible, shown)

	args := append(slices.Clone(access.Args), access.Args...)
	return from, append(args, filter.Args...)
}

// positionIn returns taskID's 1-based position among the rows of from in
// sortSQL order, or 0 when it is not among them. Reading the ids in order
// beat ROW_NUMBER() on MySQL 8 at 200,000 rows.
func (w *workspaceRepository) positionIn(from string, args []any, taskID, sortSQL string) (int, error) {
	rows, err := w.Db.Query("SELECT main.id FROM "+from+" "+sortSQL, args...)
	if err != nil {
		return 0, err
	}

	defer rows.Close()
	pos := 0
	for rows.Next() {
		pos++
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}

		if id == taskID {
			return pos, nil
		}
	}

	return 0, rows.Err()
}

// GetBranchPosition returns the 1-based position, among the branches
// GetRootTasksPagedWithSubtasks pages by, of the branch taskID is in, or 0
// when that branch is not shown. sortSQL is a validated ORDER BY clause.
func (w *workspaceRepository) GetBranchPosition(ctx context.Context, tableName, taskID string, access, filter model.SQLFilter, sortSQL string) (int, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return 0, err
	}

	root, err := w.visibleRoot(ctx, table, taskID, access)
	if err != nil || root == "" {
		return 0, err
	}

	from, args := effectiveRootsFrom(table, access, filter)
	return w.positionIn(from, args, root, sortSQL)
}

// visibleRoot returns the topmost task above taskID, or taskID itself, that
// passes access with every task between them, or "" when taskID does not.
func (w *workspaceRepository) visibleRoot(ctx context.Context, table, taskID string, access model.SQLFilter) (string, error) {
	where := "main.id = ? AND (main.deleted_at IS NULL OR main.deleted_at = 0)"
	if access.SQL != "" {
		where += " AND " + access.SQL
	}

	q := "SELECT COALESCE(main.parent_task_id, '') FROM " + table + " AS main WHERE " + where

	root := ""
	seen := map[string]bool{}
	for id := taskID; id != "" && !seen[id]; {
		seen[id] = true

		var parent string
		err := w.Db.QueryRowContext(ctx, q, append([]any{id}, access.Args...)...).Scan(&parent)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}

		if err != nil {
			return "", err
		}

		root = id
		id = parent
	}

	return root, nil
}

// GetTaskParentID returns the parent_task_id for a task, or "" if it's a root task.
func (w *workspaceRepository) GetTaskParentID(tableName, taskID string) (string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return "", err
	}

	var parentID *string
	q := fmt.Sprintf("SELECT parent_task_id FROM %s WHERE id = ? AND (deleted_at IS NULL OR deleted_at = 0) LIMIT 1", table)
	if err := w.Db.QueryRow(q, taskID).Scan(&parentID); err != nil {
		return "", err
	}

	if parentID == nil {
		return "", nil
	}

	return *parentID, nil
}

// GetSubtaskIDs returns the live subtasks at any depth below taskID that pass
// access, leaving out those whose status is one of skipStatusIDs.
func (w *workspaceRepository) GetSubtaskIDs(ctx context.Context, tableName, taskID string, skipStatusIDs []string, access model.SQLFilter) ([]string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	notDeleted := "(main.deleted_at IS NULL OR main.deleted_at = 0)"
	where := "main.id IN (SELECT id FROM subtree) AND main.id <> ?"
	args := []any{taskID, taskID}
	if len(skipStatusIDs) > 0 {
		where += " AND (main.status IS NULL OR main.status NOT IN (" + sqlPlaceholders(len(skipStatusIDs)) + "))"
		for _, id := range skipStatusIDs {
			args = append(args, id)
		}
	}

	if access.SQL != "" {
		where += " AND " + access.SQL
		args = append(args, access.Args...)
	}

	q := fmt.Sprintf(`WITH RECURSIVE subtree (id) AS (
			SELECT main.id FROM %[1]s AS main WHERE main.parent_task_id = ? AND main.id <> main.parent_task_id AND %[2]s
			UNION
			SELECT main.id FROM %[1]s AS main JOIN subtree ON main.parent_task_id = subtree.id
			WHERE main.id <> main.parent_task_id AND %[2]s
		)
		SELECT main.id FROM %[1]s AS main WHERE %[3]s`, table, notDeleted, where)

	rows, err := w.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetTableRowByID fetches a single row from a dynamic table by its id column.
// Returns nil, nil if the row does not exist.
func (w *workspaceRepository) GetTableRowByID(tableName, itemID string) (map[string]interface{}, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.Query("SELECT * FROM "+table+" WHERE id = ?", itemID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}

		return nil, nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	ptrs := make([]interface{}, len(columns))
	for i := range columns {
		ptrs[i] = &values[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}

	item := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		if b, ok := values[i].([]byte); ok {
			item[col] = string(b)
		} else {
			item[col] = values[i]
		}
	}

	return item, nil
}

// BuildFilterSQLInline builds a WHERE clause fragment from the filter payload.
// Used by GetFilteredTableData (columns are prefixed with "main.").
// tableNames must be pre-fetched by the caller (use GetTableNamesByIDs).
func (w *workspaceRepository) BuildFilterSQLInline(filters model.FilterPayload, headers []model.WorkspaceHeaders, loc *time.Location, tableNames map[string]string) model.SQLFilter {
	if len(filters.FlatFilters) > 0 {
		return w.buildFlatFilterSQLInline(filters.FlatFilters, headers, loc, tableNames)
	}

	if len(filters.Groups) > 0 {
		return w.buildGroupFilterSQL(filters.Groups, headers, loc, tableNames)
	}

	return model.SQLFilter{}
}

// GetTableNamesByIDs fetches physical table names for a set of workspace_tables IDs.
// Missing IDs are silently omitted from the result map.
func (w *workspaceRepository) GetTableNamesByIDs(ids []string) (map[string]string, error) {
	result := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := w.Db.Query("SELECT id, name FROM workspace_tables WHERE id IN ("+placeholders+")", args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		result[id] = name
	}

	return result, rows.Err()
}

func (w *workspaceRepository) DeleteTableView(workspaceID string, tableID string, viewID string) (bool, error) {
	_, err := w.Db.Exec("DELETE FROM workspace_view WHERE workspace_id = ? AND table_id = ? AND id = ?", workspaceID, tableID, viewID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// rebalanceKanbanOnFieldDelete removes the deleted field from a kanban view's
// item_order JSON and moves its tasks to the "Unassigned" column (id "0").
// Returns the updated JSON string.
func rebalanceKanbanOnFieldDelete(itemOrderJSON, fieldID string) (string, error) {
	var itemOrder map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrderJSON), &itemOrder); err != nil {
		return "", err
	}

	fields, ok := itemOrder["fields"].([]interface{})
	if !ok {
		return "", sql.ErrNoRows
	}

	var tasksToMove []interface{}
	unassignedIdx := -1

	for i := 0; i < len(fields); i++ {
		fm, ok := fields[i].(map[string]interface{})
		if !ok {
			continue
		}

		if fm["id"] == "0" {
			unassignedIdx = i
			continue
		}

		if fm["id"] == fieldID {
			if order, ok := fm["order"].([]interface{}); ok {
				tasksToMove = append(tasksToMove, order...)
			}

			fields = append(fields[:i], fields[i+1:]...)
			i--
		}
	}

	if unassignedIdx != -1 && len(tasksToMove) > 0 {
		uf, ok := fields[unassignedIdx].(map[string]interface{})
		if !ok {
			return "", sql.ErrNoRows
		}

		if order, ok := uf["order"].([]interface{}); ok {
			uf["order"] = append(order, tasksToMove...)
		} else {
			uf["order"] = tasksToMove
		}
	}

	itemOrder["fields"] = fields
	b, err := json.Marshal(itemOrder)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (w *workspaceRepository) DeleteTableSingleField(workspaceID string, tableID string, fieldID string, linkedTableID string, linkedTableName string) (bool, error) {
	optionTable, err := quoteIdent(linkedTableName)
	if err != nil {
		return false, err
	}

	tx, err := w.Db.Begin()
	if err != nil {
		return false, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}

		if err != nil {
			tx.Rollback()
		}
	}()

	_, execErr := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", optionTable), fieldID)
	if execErr != nil {
		return false, execErr
	}

	rows, err := tx.Query(`
		SELECT id, item_order
		FROM workspace_view
		WHERE parent_table_id = ? AND view_type = 'kanban'
	`, linkedTableID)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	for rows.Next() {
		var viewID, itemOrderJSON string
		if err := rows.Scan(&viewID, &itemOrderJSON); err != nil {
			return false, err
		}

		updatedJSON, err := rebalanceKanbanOnFieldDelete(itemOrderJSON, fieldID)
		if err != nil {
			return false, err
		}

		if _, err := tx.Exec(`UPDATE workspace_view SET item_order = ? WHERE id = ?`, updatedJSON, viewID); err != nil {
			return false, err
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (w *workspaceRepository) DeleteTableField(workspaceID string, tableID string, fieldID string, parentFieldID string) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("UPDATE workspace_fields SET deleted_at = UNIX_TIMESTAMP() WHERE workspace_id = ? AND table_id = ? AND id = ?", workspaceID, tableID, fieldID)
	if err != nil {
		return err
	}

	if parentFieldID != "" {
		_, err = tx.Exec("UPDATE workspace_fields SET deleted_at = UNIX_TIMESTAMP() WHERE id = ?", parentFieldID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (w *workspaceRepository) GetFieldParentInfo(fieldID string) (parentFieldID string, parentTableID string, err error) {
	err = w.Db.QueryRow("SELECT COALESCE(parent_field_id, '') FROM workspace_fields WHERE id = ?", fieldID).Scan(&parentFieldID)
	if err == sql.ErrNoRows || parentFieldID == "" {
		return "", "", nil
	}

	if err != nil {
		return "", "", err
	}

	err = w.Db.QueryRow("SELECT table_id FROM workspace_fields WHERE id = ?", parentFieldID).Scan(&parentTableID)
	if err != nil {
		return "", "", err
	}

	return parentFieldID, parentTableID, nil
}

// GetTableMeta returns physical table name and single_select flag for a given table ID.
func (w *workspaceRepository) GetTableMeta(tableID string) (tableName string, isSingleSelect bool, err error) {
	err = w.Db.QueryRow(`
		SELECT name, COALESCE(single_select, 0) FROM workspace_tables WHERE id = ?
	`, tableID).Scan(&tableName, &isSingleSelect)
	return
}

// GetSingleSelectCascadeInfo returns the parent table physical name and field column name
// to NULL out when a single-select option is deleted.
func (w *workspaceRepository) GetSingleSelectCascadeInfo(tableID string) (parentTableName, fieldName string, err error) {
	err = w.Db.QueryRow(`
		SELECT pt.name, r.table_name
		FROM workspace_relationships r
		JOIN workspace_tables pt ON pt.id = r.table_id
		WHERE r.linked_table_id = ?
		LIMIT 1
	`, tableID).Scan(&parentTableName, &fieldName)
	return
}

// GetJunctionTablesForLinkedTable returns physical names of junction tables
// where the given table is the "parent" side (parent_table_item_id column).
func (w *workspaceRepository) GetJunctionTablesForLinkedTable(tableID string) ([]string, error) {
	rows, err := w.Db.Query(`
		SELECT name FROM workspace_tables
		WHERE parent_table_id = ? AND linked = true AND single_select = false AND deleted_at = 0
	`, tableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}

		if t != "" {
			tables = append(tables, t)
		}
	}

	return tables, rows.Err()
}

// GetJunctionTablesForMainTable returns physical names of junction tables
// where the given table is the "main" side (table_item_id column).
func (w *workspaceRepository) GetJunctionTablesForMainTable(tableID string) ([]string, error) {
	rows, err := w.Db.Query(`
		SELECT wt.name
		FROM workspace_relationships r
		JOIN workspace_tables wt ON wt.id = r.linked_table_id
		WHERE r.table_id = ? AND wt.linked = true AND wt.single_select = false AND wt.deleted_at = 0
	`, tableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}

		if t != "" {
			tables = append(tables, t)
		}
	}

	return tables, rows.Err()
}

// DeleteTaskTx deletes (or soft-deletes) the item and applies all cascade
// operations inside a single transaction. cascades contains pre-resolved cleanup info.
// DeleteTaskTx deletes the tasks together, a task and its subtasks, and
// clears what pointed at each of them.
func (w *workspaceRepository) DeleteTaskTx(tableName string, isSingleSelect bool, taskIDs []string, cascades []model.DeleteCascade) (bool, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	table, err := quoteIdent(tableName)
	if err != nil {
		return false, err
	}

	for _, taskID := range taskIDs {
		if err := deleteTaskTx(tx, table, isSingleSelect, taskID, cascades); err != nil {
			return false, err
		}
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func deleteTaskTx(tx *sql.Tx, table string, isSingleSelect bool, taskID string, cascades []model.DeleteCascade) error {
	var q string
	if isSingleSelect {
		q = fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	} else {
		q = fmt.Sprintf("UPDATE %s SET deleted_at = UNIX_TIMESTAMP() WHERE id = ?", table)
	}

	if _, err := tx.Exec(q, taskID); err != nil {
		return err
	}

	for _, c := range cascades {
		cascadeTable, err := quoteIdent(c.TableName)
		if err != nil {
			return err
		}

		if c.FieldName != "" {
			// Single-select: NULL out the column
			field, err := quoteIdent(c.FieldName)
			if err != nil {
				return err
			}

			q = fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = ?", cascadeTable, field, field)
		} else {
			// Linked table: remove junction rows by the appropriate column
			col := c.JunctionColumn
			if col == "" {
				col = "parent_table_item_id"
			}

			column, err := quoteIdent(col)
			if err != nil {
				return err
			}

			q = fmt.Sprintf("DELETE FROM %s WHERE %s = ?", cascadeTable, column)
		}

		if _, err = tx.Exec(q, taskID); err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) DeleteTable(workspaceID string, tableID string, deletedAt int64) ([]map[string]interface{}, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	linkedFieldIDs, err := w.deleteTableTx(tx, tableID, deletedAt)
	if err != nil {
		return nil, err
	}

	result, err := w.collectParentFieldsTx(tx, linkedFieldIDs, map[string]bool{tableID: true})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// deleteTableTx marks a table and its fields deleted, together with the link
// fields of other tables that point at it, and returns the ids of those.
func (w *workspaceRepository) deleteTableTx(tx *sql.Tx, tableID string, deletedAt int64) ([]string, error) {
	_, err := tx.Exec("UPDATE workspace_tables SET deleted_at = ? WHERE id = ?", deletedAt, tableID)
	if err != nil {
		return nil, err
	}

	// Read all field rows into memory to avoid MySQL "busy buffer" issue
	var fields []fieldInfo

	rows, err := tx.Query("SELECT id, parent_field_id FROM workspace_fields WHERE table_id = ? AND deleted_at = 0", tableID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var f fieldInfo
		if err := rows.Scan(&f.id, &f.parentFieldID); err != nil {
			rows.Close()
			return nil, err
		}

		fields = append(fields, f)
	}

	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}

	parentFieldIDs, err := w.softDeleteFieldsTx(tx, fields, deletedAt)
	if err != nil {
		return nil, err
	}

	incoming, err := incomingLinkFieldsTx(tx, tableID)
	if err != nil {
		return nil, err
	}

	for _, field := range incoming {
		parentFieldIDs = append(parentFieldIDs, field.id)
	}

	seen := map[string]bool{}
	var linked []string
	for _, id := range parentFieldIDs {
		if seen[id] {
			continue
		}

		seen[id] = true
		linked = append(linked, id)

		if _, err := tx.Exec("UPDATE workspace_fields SET deleted_at = ? WHERE id = ?", deletedAt, id); err != nil {
			return nil, err
		}
	}

	return linked, nil
}

// incomingLinkFieldsTx returns the live link fields of other tables that
// point at tableID. A one-way link keeps no field id of its target, so it is
// found through its relationship row and the junction table behind it. That
// row is matched by field name, and a deleted field leaves its row behind, so
// only the latest row for a name counts: a field reusing the name of a
// deleted one belongs to the newer row.
func incomingLinkFieldsTx(tx *sql.Tx, tableID string) ([]linkField, error) {
	rows, err := tx.Query(`SELECT DISTINCT f.id, f.table_id FROM workspace_fields f
		JOIN workspace_relationships r ON r.table_id = f.table_id AND r.table_name = f.field_name AND r.deleted_at = 0
		JOIN workspace_tables j ON j.id = r.linked_table_id
		WHERE j.parent_table_id = ? AND j.linked = 1 AND j.single_select = 0 AND j.deleted_at = 0
		AND f.table_id <> ? AND f.field_type = 'link' AND f.deleted_at = 0
		AND r.created_at = (SELECT MAX(latest.created_at) FROM workspace_relationships latest
			WHERE latest.table_id = r.table_id AND latest.table_name = r.table_name AND latest.deleted_at = 0)`,
		tableID, tableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var fields []linkField
	for rows.Next() {
		var field linkField
		if err := rows.Scan(&field.id, &field.tableID); err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return fields, rows.Err()
}

type linkField struct {
	id      string
	tableID string
}

// softDeleteFieldsTx marks all given fields deleted and returns parent field IDs for linked fields.
func (w *workspaceRepository) softDeleteFieldsTx(tx *sql.Tx, fields []fieldInfo, timestamp int64) ([]string, error) {
	var parentFieldIDs []string
	for _, field := range fields {
		if _, err := tx.Exec("UPDATE workspace_fields SET deleted_at = ? WHERE id = ?", timestamp, field.id); err != nil {
			return nil, err
		}

		if field.parentFieldID.Valid && field.parentFieldID.String != "" {
			parentFieldIDs = append(parentFieldIDs, field.parentFieldID.String)
		}
	}

	return parentFieldIDs, nil
}

// collectParentFieldsTx groups field IDs by their table, leaving out the
// tables in skip, and returns a summary slice.
func (w *workspaceRepository) collectParentFieldsTx(tx *sql.Tx, fieldIDs []string, skip map[string]bool) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	if len(fieldIDs) == 0 {
		return result, nil
	}

	rows, err := tx.Query(`SELECT id, table_id FROM workspace_fields WHERE id IN (`+sqlPlaceholders(len(fieldIDs))+`)`,
		toInterfaceSlice(fieldIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tableMap := make(map[string][]string)
	for rows.Next() {
		var id, tableID string
		if err := rows.Scan(&id, &tableID); err != nil {
			return nil, err
		}

		if !skip[tableID] {
			tableMap[tableID] = append(tableMap[tableID], id)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for tid, fieldIDs := range tableMap {
		result = append(result, map[string]interface{}{
			"table_id":  tid,
			"field_ids": fieldIDs,
		})
	}

	return result, nil
}

func (w *workspaceRepository) createWorkspaceTableWithTX(tx *sql.Tx, tableID string, workspaceID string, name string, linked bool, parent_table_id string, userID string, single_select bool, linkBothDirections bool, secondTableID string, viewID string) (*model.WorkspaceTable, error) {
	workspace, err := w.GetRow(workspaceID)
	if err != nil {
		return nil, err
	}

	tableName := workspace.Prefix + name

	if err = w.createLinkTableDDLTx(tx, tableName); err != nil {
		return nil, err
	}

	if !linkBothDirections {
		secondTableID = ""
	}

	query := `
		INSERT INTO workspace_tables 
		(id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.Exec(query,
		tableID,
		workspaceID,
		tableName,
		"",
		linked,
		single_select,
		parent_table_id,
		linkBothDirections,
		secondTableID,
		"",
		0,
	)
	if err != nil {
		return nil, err
	}

	taskOrder := []model.TaskOrderField{
		{Name: "id", DisplayName: "ID", Width: "150", Visible: true},
		{Name: "table_id", DisplayName: "Table ID", Width: "150", Visible: true},
		{Name: "table_item_id", DisplayName: "Table Item ID", Width: "150", Visible: true},
		{Name: "parent_table_id", DisplayName: "Parent Table ID", Width: "150", Visible: true},
		{Name: "parent_table_item_id", DisplayName: "Parent Table Item ID", Width: "150", Visible: true},
	}

	taskOrderJSON, err := json.Marshal(taskOrder)
	if err != nil {
		return nil, err
	}

	workspaceView := model.WorkspaceView{
		ID:          viewID,
		WorkspaceID: workspaceID,
		ViewType:    "grid",
		TaskOrder:   string(taskOrderJSON),
		CreatedBy:   userID,
		TableID:     tableID,
		MainView:    true,
	}

	_, err = w.CreateWorkspaceViewTx(tx, &workspaceView)
	if err != nil {
		return nil, err
	}

	return &model.WorkspaceTable{
		ID:                tableID,
		WorkspaceID:       workspaceID,
		Name:              name,
		Linked:            linked,
		SingleSelect:      single_select,
		ParentTableID:     parent_table_id,
		BothDirectionLink: linkBothDirections,
		SecondTableID:     secondTableID,
	}, nil
}

func createIndexesTx(tx *sql.Tx, tableName string, layout dynamicTableLayout) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	for _, columns := range layout.indexes {
		name, err := quoteIdent(shortIndexName(tableName, strings.Join(columns, "_")))
		if err != nil {
			return err
		}

		kind := "INDEX"
		if slices.Contains(layout.unique, strings.Join(columns, ",")) {
			kind = "UNIQUE INDEX"
		}

		if _, err := tx.Exec(fmt.Sprintf("CREATE %s %s ON %s (%s);", kind, name, table, strings.Join(columns, ", "))); err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) createLinkTableDDLTx(tx *sql.Tx, tableName string) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS ` + table + ` (
		id VARCHAR(36) NOT NULL PRIMARY KEY,
		table_id VARCHAR(36) NOT NULL,
		table_item_id VARCHAR(36) NOT NULL,
		parent_table_id VARCHAR(36) NOT NULL,
		parent_table_item_id VARCHAR(36) NOT NULL
	);`)
	if err != nil {
		return err
	}

	return createIndexesTx(tx, tableName, linkTableLayout)
}

func (w *workspaceRepository) createTaskTableDDLTx(tx *sql.Tx, tableName string) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS ` + table + ` (
		id VARCHAR(36) NOT NULL PRIMARY KEY,
		name VARCHAR(255),
		start_date BIGINT,
		due_date BIGINT,
		assignee VARCHAR(255),
		status VARCHAR(50),
		description TEXT,
		updated_at BIGINT,
		created_at BIGINT,
		deleted_at BIGINT,
		created_by VARCHAR(255),
		parent_task_id VARCHAR(36) DEFAULT NULL
	);`)
	if err != nil {
		return err
	}

	return createIndexesTx(tx, tableName, taskTableLayout)
}

// CreateTable creates the DDL, metadata row, default grid view, and status table.
// safeName is the physical name without the workspace prefix, which the caller
// generates from the table's id; name is only what the table is called.
func (w *workspaceRepository) CreateTable(tableID string, workspaceID string, name string, safeName string, linked bool, parent_table_id string, userID string, single_select bool, linkBothDirections bool, secondTableID string, folder_id string, taskOrderJSON string, defaultStatuses []model.KanbanStatusOption, statusTableID string, statusTablePhysicalName string, mainViewID string) (*model.WorkspaceTable, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	workspace, err := w.GetRow(workspaceID)
	if err != nil {
		return nil, err
	}

	tableName := workspace.Prefix + safeName

	if linked {
		if err = w.createLinkTableDDLTx(tx, tableName); err != nil {
			return nil, err
		}
	} else {
		if err = w.createTaskTableDDLTx(tx, tableName); err != nil {
			return nil, err
		}
	}

	if !linkBothDirections {
		secondTableID = ""
	}

	_, err = tx.Exec(`
		INSERT INTO workspace_tables
		(id, workspace_id, name, display_name, linked, single_select, parent_table_id,
		 both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableID, workspaceID, tableName, name, linked, single_select, parent_table_id, linkBothDirections, secondTableID, folder_id, 0)
	if err != nil {
		return nil, err
	}

	createdView, err := w.CreateWorkspaceViewTx(tx, &model.WorkspaceView{
		ID:          mainViewID,
		WorkspaceID: workspaceID,
		ViewType:    "grid",
		TaskOrder:   taskOrderJSON,
		CreatedBy:   userID,
		TableID:     tableID,
		MainView:    true,
	})
	if err != nil {
		return nil, err
	}

	statusTableName := workspace.Prefix + statusTablePhysicalName

	if err = w.createSingleSelectTableTx(tx, workspaceID, statusTableID, statusTableName, false, tableID, true); err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO workspace_relationships
		(id, workspace_id, table_id, linked_table_id, table_name, created_at, updated_at, deleted_at)
		VALUES (UUID(), ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)
	`, workspaceID, tableID, statusTableID, "status")
	if err != nil {
		return nil, err
	}

	statusTable, err := quoteIdent(statusTableName)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf(`
		ALTER TABLE %s
		ADD COLUMN status_type VARCHAR(50) DEFAULT 'Not started'
	`, statusTable))
	if err != nil {
		return nil, err
	}

	for _, status := range defaultStatuses {
		_, err = tx.Exec(fmt.Sprintf(`
			INSERT INTO %s (id, name, color, status_type)
			VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE name=name
		`, statusTable), status.ID, status.Name, status.Color, status.StatusType)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &model.WorkspaceTable{
		ID:                tableID,
		WorkspaceID:       workspaceID,
		Name:              name,
		DisplayName:       sql.NullString{String: name, Valid: true},
		Linked:            linked,
		SingleSelect:      single_select,
		ParentTableID:     parent_table_id,
		BothDirectionLink: linkBothDirections,
		SecondTableID:     secondTableID,
		FolderID:          folder_id,
		Views:             []model.WorkspaceView{*createdView},
	}, nil
}

// createSingleSelectTableTx creates the status table DDL and metadata row within an existing TX.
func (w *workspaceRepository) createSingleSelectTableTx(tx *sql.Tx, workspaceID, tableID, tableName string, linked bool, parent_table_id string, single_select bool) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS ` + table + ` (
		id VARCHAR(36) NOT NULL PRIMARY KEY,
		name VARCHAR(255),
		color VARCHAR(255)
	);`); err != nil {
		return err
	}

	if err := createIndexesTx(tx, tableName, selectTableLayout); err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO workspace_tables
		(id, workspace_id, name, display_name, linked, single_select, parent_table_id, both_direction_link, second_table_id, folder_id, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableID, workspaceID, tableName, "", linked, single_select, parent_table_id, false, "", "", 0)
	return err
}

func (w *workspaceRepository) createSingleSelectTable(workspaceID, tableID, physicalName string, linked bool, parent_table_id string, single_select bool) (*model.WorkspaceTable, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	workspace, err := w.GetRow(workspaceID)
	if err != nil {
		return nil, err
	}

	tableName := workspace.Prefix + physicalName

	if err = w.createSingleSelectTableTx(tx, workspaceID, tableID, tableName, linked, parent_table_id, single_select); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &model.WorkspaceTable{
		ID:            tableID,
		WorkspaceID:   workspaceID,
		Name:          physicalName,
		Linked:        linked,
		SingleSelect:  single_select,
		ParentTableID: parent_table_id,
		Headers: []model.WorkspaceHeaders{
			{Name: "id", HeaderType: "VARCHAR", HeaderUsage: "text"},
			{Name: "name", HeaderType: "VARCHAR", HeaderUsage: "text"},
		},
	}, nil
}

// getStatusTypeMap fetches the id→statusType mapping for a table's linked status table.
// Returns an empty map if no status table exists or on error.
func (w *workspaceRepository) getStatusTypeMap(tableID string) map[string]string {
	var statusTableID string
	err := w.Db.QueryRow(`
		SELECT linked_table_id
		FROM workspace_relationships
		WHERE table_id = ? AND table_name = 'status'
	`, tableID).Scan(&statusTableID)
	if err != nil || statusTableID == "" {
		return map[string]string{}
	}

	var statusTableName string
	err = w.Db.QueryRow(`SELECT name FROM workspace_tables WHERE id = ?`, statusTableID).Scan(&statusTableName)
	if err != nil || statusTableName == "" {
		return map[string]string{}
	}

	statusTable, err := quoteIdent(statusTableName)
	if err != nil {
		return map[string]string{}
	}

	m := make(map[string]string)
	rows, err := w.Db.Query(fmt.Sprintf(`SELECT id, status_type FROM %s WHERE status_type IS NOT NULL`, statusTable))
	if err != nil {
		return m
	}

	defer rows.Close()
	for rows.Next() {
		var id, st string
		if err := rows.Scan(&id, &st); err == nil {
			m[id] = st
		}
	}

	if err := rows.Err(); err != nil {
		return m
	}

	return m
}

func (w *workspaceRepository) GetStatusTypeMap(tableID string) (map[string]string, error) {
	return w.getStatusTypeMap(tableID), nil
}

// GetStatusOptions returns all status options (id, name, status_type) for the given table.
func (w *workspaceRepository) GetStatusOptions(ctx context.Context, tableID string) ([]map[string]string, error) {
	var statusTableID string
	err := w.Db.QueryRowContext(ctx, `
		SELECT linked_table_id FROM workspace_relationships
		WHERE table_id = ? AND table_name = 'status'
	`, tableID).Scan(&statusTableID)
	if err != nil || statusTableID == "" {
		return nil, nil
	}

	var statusTableName string
	err = w.Db.QueryRowContext(ctx, `SELECT name FROM workspace_tables WHERE id = ?`, statusTableID).Scan(&statusTableName)
	if err != nil || statusTableName == "" {
		return nil, nil
	}

	statusTable, err := quoteIdent(statusTableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.QueryContext(ctx, fmt.Sprintf(`SELECT id, name, COALESCE(status_type,'') FROM %s`, statusTable))
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []map[string]string
	for rows.Next() {
		var id, name, st string
		if err := rows.Scan(&id, &name, &st); err == nil {
			result = append(result, map[string]string{"id": id, "name": name, "status_type": st})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (w *workspaceRepository) GetAttachment(fileID, workspaceID string) (*model.WorkspaceAttachment, error) {
	query := `SELECT id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, COALESCE(storage_id, ''), created_at, updated_at, deleted_at
              FROM workspace_attachments
              WHERE id = ? AND workspace_id = ? AND deleted_at = 0`

	row := w.Db.QueryRow(query, fileID, workspaceID)

	attachment := &model.WorkspaceAttachment{}
	err := row.Scan(
		&attachment.ID,
		&attachment.WorkspaceID,
		&attachment.TableID,
		&attachment.FieldID,
		&attachment.TaskID,
		&attachment.UserID,
		&attachment.Name,
		&attachment.Size,
		&attachment.MimeType,
		&attachment.Width,
		&attachment.Height,
		&attachment.StorageID,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return attachment, nil
}

func (w *workspaceRepository) UpdateFolder(workspaceID string, folderID string, name string) (bool, error) {
	_, err := w.Db.Exec(`
		UPDATE workspace_folders
		SET name = ?
		WHERE id = ? AND workspace_id = ?
	`, name, folderID, workspaceID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteFolder deletes a folder together with the folders inside it and, when
// withTables is set, the tables in any of them. It returns the ids of those
// tables and, grouped by table, the link fields of tables left standing that
// went with them. It returns sql.ErrNoRows for a folder not live in the
// workspace and model.ErrFolderHasTables when there are tables and
// withTables is not set.
func (w *workspaceRepository) DeleteFolder(ctx context.Context, workspaceID, folderID string, withTables bool) ([]string, []map[string]interface{}, error) {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, nil, err
	}

	defer tx.Rollback()

	if err := lockFolderTreeTx(ctx, tx, workspaceID); err != nil {
		return nil, nil, err
	}

	if err := requireLiveFolderTx(ctx, tx, workspaceID, folderID); err != nil {
		return nil, nil, err
	}

	ids, err := folderSubtreeTx(ctx, tx, workspaceID, folderID)
	if err != nil {
		return nil, nil, err
	}

	tableIDs, err := folderTablesTx(ctx, tx, workspaceID, ids)
	if err != nil {
		return nil, nil, err
	}

	if len(tableIDs) > 0 && !withTables {
		return nil, nil, model.ErrFolderHasTables
	}

	deletedAt := time.Now().Unix()
	deleted := map[string]bool{}
	var linkedFieldIDs []string
	for _, tableID := range tableIDs {
		deleted[tableID] = true
		linked, err := w.deleteTableTx(tx, tableID, deletedAt)
		if err != nil {
			return nil, nil, err
		}

		linkedFieldIDs = append(linkedFieldIDs, linked...)
	}

	linkedFields, err := w.collectParentFieldsTx(tx, linkedFieldIDs, deleted)
	if err != nil {
		return nil, nil, err
	}

	folderArgs := append([]any{deletedAt, workspaceID}, toInterfaceSlice(ids)...)
	if _, err := tx.ExecContext(ctx, `UPDATE workspace_folders SET deleted_at = ?
		WHERE workspace_id = ? AND id IN (`+sqlPlaceholders(len(ids))+`)`, folderArgs...); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return tableIDs, linkedFields, nil
}

// MoveTable puts a table in folderID, or at the workspace root when folderID
// is empty. It returns sql.ErrNoRows when the table or the folder is not live
// in the workspace.
func (w *workspaceRepository) MoveTable(ctx context.Context, workspaceID, tableID, folderID string) error {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := lockFolderTreeTx(ctx, tx, workspaceID); err != nil {
		return err
	}

	if folderID != "" {
		if err := requireLiveFolderTx(ctx, tx, workspaceID, folderID); err != nil {
			return err
		}
	}

	var id string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM workspace_tables
		WHERE id = ? AND workspace_id = ? AND deleted_at = 0 AND linked = 0 AND single_select = 0`,
		tableID, workspaceID).Scan(&id); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE workspace_tables SET folder_id = ? WHERE id = ?`, folderID, tableID); err != nil {
		return err
	}

	return tx.Commit()
}

// MoveFolder puts a folder inside parentID, or at the workspace root when
// parentID is empty. It returns sql.ErrNoRows when either folder is not live
// in the workspace and model.ErrFolderCycle when parentID is the folder or
// lies inside it.
func (w *workspaceRepository) MoveFolder(ctx context.Context, workspaceID, folderID, parentID string) error {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := lockFolderTreeTx(ctx, tx, workspaceID); err != nil {
		return err
	}

	if err := requireLiveFolderTx(ctx, tx, workspaceID, folderID); err != nil {
		return err
	}

	if parentID != "" {
		if err := requireLiveFolderTx(ctx, tx, workspaceID, parentID); err != nil {
			return err
		}

		ids, err := folderSubtreeTx(ctx, tx, workspaceID, folderID)
		if err != nil {
			return err
		}

		if slices.Contains(ids, parentID) {
			return model.ErrFolderCycle
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE workspace_folders SET parent_folder_id = ?, updated_at = UNIX_TIMESTAMP() WHERE id = ?`,
		parentID, folderID); err != nil {
		return err
	}

	return tx.Commit()
}

// lockFolderTreeTx holds the workspace's row until tx ends, so changes to one
// workspace's folder tree run one at a time. Without it two moves checked
// at once could each pass and together put two folders inside each other.
func lockFolderTreeTx(ctx context.Context, tx *sql.Tx, workspaceID string) error {
	var id string
	return tx.QueryRowContext(ctx, `SELECT id FROM workspaces WHERE id = ? AND (deleted_at IS NULL OR deleted_at = 0) FOR UPDATE`,
		workspaceID).Scan(&id)
}

func requireLiveFolderTx(ctx context.Context, tx *sql.Tx, workspaceID, folderID string) error {
	var id string
	return tx.QueryRowContext(ctx, `SELECT id FROM workspace_folders
		WHERE id = ? AND workspace_id = ? AND (deleted_at IS NULL OR deleted_at = 0)`,
		folderID, workspaceID).Scan(&id)
}

func folderTablesTx(ctx context.Context, tx *sql.Tx, workspaceID string, folderIDs []string) ([]string, error) {
	args := append([]any{workspaceID}, toInterfaceSlice(folderIDs)...)
	rows, err := tx.QueryContext(ctx, `SELECT id FROM workspace_tables
		WHERE workspace_id = ? AND deleted_at = 0 AND linked = 0 AND single_select = 0
		AND folder_id IN (`+sqlPlaceholders(len(folderIDs))+`)`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tableIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		tableIDs = append(tableIDs, id)
	}

	return tableIDs, rows.Err()
}

// GetFolderTableIDs returns the tables in a folder and in the folders inside it.
func (w *workspaceRepository) GetFolderTableIDs(ctx context.Context, workspaceID, folderID string) ([]string, error) {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	ids, err := folderSubtreeTx(ctx, tx, workspaceID, folderID)
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	return folderTablesTx(ctx, tx, workspaceID, ids)
}

// GetLinkingTables returns the live tables, other than those in tableIDs,
// with a link field pointing at any of tableIDs.
func (w *workspaceRepository) GetLinkingTables(ctx context.Context, workspaceID string, tableIDs []string) ([]model.LinkingTable, error) {
	tables := []model.LinkingTable{}
	if len(tableIDs) == 0 {
		return tables, nil
	}

	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	targets := map[string]bool{}
	for _, id := range tableIDs {
		targets[id] = true
	}

	var linking []string
	seen := map[string]bool{}
	for _, id := range tableIDs {
		incoming, err := incomingLinkFieldsTx(tx, id)
		if err != nil {
			return nil, err
		}

		for _, field := range incoming {
			if !targets[field.tableID] && !seen[field.tableID] {
				seen[field.tableID] = true
				linking = append(linking, field.tableID)
			}
		}
	}

	if len(linking) == 0 {
		return tables, nil
	}

	args := append([]any{workspaceID}, toInterfaceSlice(linking)...)
	rows, err := tx.QueryContext(ctx, `SELECT id, COALESCE(NULLIF(display_name, ''), name) FROM workspace_tables
		WHERE workspace_id = ? AND deleted_at = 0 AND id IN (`+sqlPlaceholders(len(linking))+`)
		ORDER BY 2`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var table model.LinkingTable
		if err := rows.Scan(&table.ID, &table.Name); err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// folderSubtreeTx returns folderID and the ids of the live folders inside it.
func folderSubtreeTx(ctx context.Context, tx *sql.Tx, workspaceID, folderID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `WITH RECURSIVE subtree (id) AS (
			SELECT id FROM workspace_folders WHERE id = ? AND workspace_id = ?
			UNION
			SELECT f.id FROM workspace_folders f
			JOIN subtree s ON f.parent_folder_id = s.id
			WHERE f.workspace_id = ? AND (f.deleted_at IS NULL OR f.deleted_at = 0)
		)
		SELECT id FROM subtree`, folderID, workspaceID, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (w *workspaceRepository) DeleteMember(workspaceID string, memberID string) (bool, error) {
	result, err := w.Db.Exec(`
		DELETE FROM workspace_members
		WHERE user_id = ? AND workspace_id = ?
	`, memberID, workspaceID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

// GetTableHeaderMeta loads what a table's headers are built from in four
// queries, whatever the number of columns.
func (w *workspaceRepository) GetTableHeaderMeta(tableID string) (*model.TableHeaderMeta, error) {
	meta := &model.TableHeaderMeta{
		Fields:           map[string]model.WorkspaceFieldData{},
		Links:            map[string]string{},
		LinkedTables:     map[string]model.LinkedTableMeta{},
		LinkedFieldNames: map[string]string{},
	}

	rows, err := w.Db.Query(`
		SELECT id, COALESCE(workspace_id, ''), table_id, field_name, COALESCE(field_display_name, ''), COALESCE(field_type, ''),
			formula, COALESCE(created_at, 0), COALESCE(updated_at, 0), deleted_at
		FROM workspace_fields WHERE table_id = ? AND deleted_at = 0`, tableID)
	if err != nil {
		return nil, err
	}

	var fieldIDs []any
	for rows.Next() {
		var wf model.WorkspaceFieldData
		var rawFormula sql.NullString
		if err := rows.Scan(&wf.ID, &wf.WorkspaceID, &wf.TableID, &wf.Name, &wf.DisplayName, &wf.FieldType,
			&rawFormula, &wf.CreatedAt, &wf.UpdatedAt, &wf.DeletedAt); err != nil {
			rows.Close()
			return nil, err
		}

		if rawFormula.Valid && len(rawFormula.String) > 0 {
			var spec model.FormulaSpec
			if err := json.Unmarshal([]byte(rawFormula.String), &spec); err == nil {
				wf.Formula = &spec
			}
		}

		if _, seen := meta.Fields[wf.Name]; !seen {
			meta.Fields[wf.Name] = wf
			fieldIDs = append(fieldIDs, wf.ID)
		}
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = w.Db.Query(`SELECT table_name, linked_table_id FROM workspace_relationships WHERE table_id = ? ORDER BY created_at DESC`, tableID)
	if err != nil {
		return nil, err
	}

	var linkedIDs []any
	for rows.Next() {
		var name, linked string
		if err := rows.Scan(&name, &linked); err != nil {
			rows.Close()
			return nil, err
		}

		if _, seen := meta.Links[name]; !seen {
			meta.Links[name] = linked
			linkedIDs = append(linkedIDs, linked)
		}
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(linkedIDs) == 0 {
		return meta, nil
	}

	rows, err = w.Db.Query(`SELECT id, COALESCE(parent_table_id, ''), COALESCE(single_select, 0), COALESCE(both_direction_link, 0)
		FROM workspace_tables WHERE id IN (`+sqlPlaceholders(len(linkedIDs))+`)`, linkedIDs...)
	if err != nil {
		return nil, err
	}

	var parentIDs []any
	for rows.Next() {
		var id string
		var lt model.LinkedTableMeta
		if err := rows.Scan(&id, &lt.ParentTableID, &lt.SingleSelect, &lt.BothDirections); err != nil {
			rows.Close()
			return nil, err
		}

		meta.LinkedTables[id] = lt
		if lt.ParentTableID != "" {
			parentIDs = append(parentIDs, lt.ParentTableID)
		}
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(parentIDs) == 0 || len(fieldIDs) == 0 {
		return meta, nil
	}

	rows, err = w.Db.Query(`SELECT table_id, parent_field_id, field_name FROM workspace_fields
		WHERE deleted_at = 0 AND table_id IN (`+sqlPlaceholders(len(parentIDs))+`) AND parent_field_id IN (`+sqlPlaceholders(len(fieldIDs))+`)`,
		append(append([]any{}, parentIDs...), fieldIDs...)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var table, parentField, name string
		if err := rows.Scan(&table, &parentField, &name); err != nil {
			return nil, err
		}

		if key := table + "/" + parentField; meta.LinkedFieldNames[key] == "" {
			meta.LinkedFieldNames[key] = name
		}
	}

	return meta, rows.Err()
}

func (w *workspaceRepository) GetLinkedTaskData(parentTableID string, parentTaskID string) (*model.TaskOrderField, error) {
	linkedData, _ := w.GetSingleSelectValues(parentTableID, parentTaskID)
	return linkedData, nil
}

// GetSingleSelectOptions loads the options of a single-select table with the
// given ids, keyed by id. Ids with no option are left out.
func (w *workspaceRepository) GetSingleSelectOptions(ctx context.Context, linkedID string, ids []string) (map[string]*model.TaskOrderField, error) {
	options := map[string]*model.TaskOrderField{}
	if len(ids) == 0 {
		return options, nil
	}

	var tableName string
	err := w.Db.QueryRowContext(ctx, `SELECT name FROM workspace_tables WHERE id = ?`, linkedID).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return options, nil
	}

	if err != nil {
		return nil, err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	rows, err := w.Db.QueryContext(ctx, fmt.Sprintf("SELECT id, COALESCE(name, ''), COALESCE(color, '') FROM %s WHERE id IN (%s)",
		table, sqlPlaceholders(len(ids))), toInterfaceSlice(ids)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var o model.TaskOrderField
		if err := rows.Scan(&o.ID, &o.Name, &o.Color); err != nil {
			return nil, err
		}

		options[o.ID] = &o
	}

	return options, rows.Err()
}

func (w *workspaceRepository) GetSingleSelectValues(linkedID string, itemID string) (*model.TaskOrderField, error) {
	if itemID == "0" {
		return nil, nil
	}

	var parentTableName string
	err := w.Db.QueryRow(`
		SELECT name FROM workspace_tables
		WHERE id = ?
	`, linkedID).Scan(&parentTableName)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	parentTable, err := quoteIdent(parentTableName)
	if err != nil {
		return nil, err
	}

	var task model.TaskOrderField
	err = w.Db.QueryRow(fmt.Sprintf(`
		SELECT id, name, color FROM %s
		WHERE id = ?
	`, parentTable), itemID).Scan(&task.ID, &task.Name, &task.Color)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &task, nil
}

func (w *workspaceRepository) GetLinkedIDs(linkedID string, itemTableID string) ([]model.LinkedItem, error) {
	var linkTable, parentTable string
	err := w.Db.QueryRow(`
		SELECT wt.name, pt.name
		FROM workspace_tables wt
		JOIN workspace_tables pt ON pt.id = wt.parent_table_id
		WHERE wt.id = ?
	`, linkedID).Scan(&linkTable, &parentTable)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	parent, err := quoteIdent(parentTable)
	if err != nil {
		return nil, err
	}

	link, err := quoteIdent(linkTable)
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(`
		SELECT p.id, p.name
		FROM %s AS p
		JOIN %s AS l ON l.parent_table_item_id = p.id
		WHERE l.table_item_id = ?
	`, parent, link)

	rows, err := w.Db.Query(q, itemTableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]model.LinkedItem, 0, 8)
	for rows.Next() {
		var it model.LinkedItem
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			return nil, err
		}

		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items, nil
}

func (w *workspaceRepository) GetLinkedIDsBatch(ctx context.Context, linkedID string, taskIDs []string) (map[string][]model.LinkedItem, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}

	var linkTable, parentTable string
	err := w.Db.QueryRowContext(ctx, `
		SELECT wt.name, pt.name
		FROM workspace_tables wt
		JOIN workspace_tables pt ON pt.id = wt.parent_table_id
		WHERE wt.id = ?
	`, linkedID).Scan(&linkTable, &parentTable)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	link, err := quoteIdent(linkTable)
	if err != nil {
		return nil, err
	}

	parent, err := quoteIdent(parentTable)
	if err != nil {
		return nil, err
	}

	placeholders := make([]string, len(taskIDs))
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	q := fmt.Sprintf(`
		SELECT l.table_item_id, p.id, p.name
		FROM %s AS l
		JOIN %s AS p ON p.id = l.parent_table_item_id
		WHERE l.table_item_id IN (%s)
	`, link, parent, strings.Join(placeholders, ","))

	rows, err := w.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := map[string][]model.LinkedItem{}
	for rows.Next() {
		var taskID string
		var item model.LinkedItem
		if err := rows.Scan(&taskID, &item.ID, &item.Name); err != nil {
			return nil, err
		}

		result[taskID] = append(result[taskID], item)
	}

	return result, rows.Err()
}

type LinkedItem struct {
	TaskID       string
	LinkedItemID string
}

// ChangeTaskLinks adds and removes links in one of a task's link fields,
// leaving every other link as it is. The field is found by its name on the
// task's own table in the workspace, never taken from the caller. It returns
// sql.ErrNoRows when the table, field or task is not there, and
// model.ErrLinkTargetMissing when a row to link is not in the linked table.
func (w *workspaceRepository) ChangeTaskLinks(ctx context.Context, workspaceID, tableID, taskID, field string, add, remove []string) (*map[string]interface{}, error) {
	add = slices.Compact(slices.Sorted(slices.Values(add)))
	remove = slices.Compact(slices.Sorted(slices.Values(remove)))

	// Read committed, so that once the task is locked every read sees what
	// an earlier save to the same task wrote, rather than the snapshot taken
	// before it was locked.
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var taskTable string
	if err := tx.QueryRowContext(ctx, `SELECT name FROM workspace_tables
		WHERE id = ? AND workspace_id = ? AND (deleted_at IS NULL OR deleted_at = 0)`,
		tableID, workspaceID).Scan(&taskTable); err != nil {
		return nil, err
	}

	var linkTableID string
	if err := tx.QueryRowContext(ctx, `SELECT linked_table_id FROM workspace_relationships
		WHERE workspace_id = ? AND table_id = ? AND table_name = ? AND (deleted_at IS NULL OR deleted_at = 0)
		ORDER BY created_at DESC
		LIMIT 1`,
		workspaceID, tableID, field).Scan(&linkTableID); err != nil {
		return nil, err
	}

	var linkTable, targetTableID, reverseTableID string
	var bothDirections bool
	if err := tx.QueryRowContext(ctx, `SELECT name, parent_table_id, both_direction_link, second_table_id
		FROM workspace_tables WHERE id = ? AND workspace_id = ?`,
		linkTableID, workspaceID).Scan(&linkTable, &targetTableID, &bothDirections, &reverseTableID); err != nil {
		return nil, err
	}

	var targetTable string
	if err := tx.QueryRowContext(ctx, `SELECT name FROM workspace_tables WHERE id = ? AND workspace_id = ?`,
		targetTableID, workspaceID).Scan(&targetTable); err != nil {
		return nil, err
	}

	if err := checkIdents(taskTable, linkTable, targetTable); err != nil {
		return nil, err
	}

	// Locking the task makes two saves to its links wait for each other. A
	// link both ways also writes the other tasks' links, so they are locked
	// too, all in one order by id: two saves from opposite ends then queue
	// instead of deadlocking.
	type lockedRow struct{ table, id string }
	locks := []lockedRow{{taskTable, taskID}}
	if bothDirections {
		for _, id := range slices.Concat(add, remove) {
			locks = append(locks, lockedRow{targetTable, id})
		}
	}

	slices.SortFunc(locks, func(a, b lockedRow) int {
		return cmp.Or(strings.Compare(a.id, b.id), strings.Compare(a.table, b.table))
	})
	locks = slices.Compact(locks)

	for _, row := range locks {
		query := "SELECT id FROM `" + row.table + "` WHERE id = ? FOR UPDATE"
		own := row == lockedRow{taskTable, taskID}
		if own {
			query = "SELECT id FROM `" + row.table + "` WHERE id = ? AND (deleted_at IS NULL OR deleted_at = 0) FOR UPDATE"
		}

		var locked string
		err := tx.QueryRowContext(ctx, query, row.id).Scan(&locked)
		if err != nil && (own || !errors.Is(err, sql.ErrNoRows)) {
			return nil, err
		}
	}

	if err := w.requireLinkTargetsTx(ctx, tx, targetTable, add); err != nil {
		return nil, err
	}

	current := map[string]bool{}
	rows, err := tx.QueryContext(ctx, "SELECT parent_table_item_id FROM `"+linkTable+"` WHERE table_id = ? AND table_item_id = ? AND parent_table_id = ?",
		tableID, taskID, linkTableID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}

		current[id] = true
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var added []string
	for _, id := range add {
		if current[id] {
			continue
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO `"+linkTable+"` (id, table_id, table_item_id, parent_table_id, parent_table_item_id) VALUES (UUID(), ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id",
			tableID, taskID, linkTableID, id); err != nil {
			return nil, err
		}

		current[id] = true
		added = append(added, id)
	}

	var removed []string
	for _, id := range remove {
		if !current[id] {
			continue
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM `"+linkTable+"` WHERE table_id = ? AND table_item_id = ? AND parent_table_id = ? AND parent_table_item_id = ?",
			tableID, taskID, linkTableID, id); err != nil {
			return nil, err
		}

		delete(current, id)
		removed = append(removed, id)
	}

	reverseAdded := []map[string]string{}
	reverseRemoved := []map[string]string{}
	var reverseField, reverseTable string

	if bothDirections {
		if err := tx.QueryRowContext(ctx, `SELECT name FROM workspace_tables WHERE id = ? AND workspace_id = ?`,
			reverseTableID, workspaceID).Scan(&reverseTable); err != nil {
			return nil, err
		}

		if err := checkIdents(reverseTable); err != nil {
			return nil, err
		}

		if len(added) > 0 {
			items, fieldName, err := w.reconcileBidirectionalLinksTx(tx, workspaceID, tableID, taskID, targetTableID, reverseTableID, reverseTable, added)
			if err != nil {
				return nil, err
			}

			reverseAdded = items
			reverseField = fieldName
		}

		for _, id := range removed {
			if _, err := tx.ExecContext(ctx, "DELETE FROM `"+reverseTable+"` WHERE parent_table_item_id = ? AND table_item_id = ?",
				taskID, id); err != nil {
				return nil, err
			}

			reverseRemoved = append(reverseRemoved, map[string]string{
				"taskID":        taskID,
				"linkedItemID":  id,
				"secondTableID": targetTableID,
			})
		}
	}

	links, err := w.getCurrentLinkedItemsTx(tx, tableID, taskID, linkTableID, workspaceID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &map[string]interface{}{
		"newLinkedItems":            links,
		"secondLinkedItemsToAdd":    reverseAdded,
		"secondLinkedItemsToDelete": reverseRemoved,
		"bothDirection":             fmt.Sprintf("%t", bothDirections),
		"secondTableFieldName":      reverseField,
		"secondTableID":             targetTableID,
	}, nil
}

// requireLinkTargetsTx checks that every id is a row of the table, and not a
// deleted one where the table records deletions.
func (w *workspaceRepository) requireLinkTargetsTx(ctx context.Context, tx *sql.Tx, table string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var hasDeletedAt int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = 'deleted_at'`,
		table).Scan(&hasDeletedAt); err != nil {
		return err
	}

	where := "id IN (" + sqlPlaceholders(len(ids)) + ")"
	if hasDeletedAt > 0 {
		where += " AND (deleted_at IS NULL OR deleted_at = 0)"
	}

	var found int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+table+"` WHERE "+where,
		toInterfaceSlice(ids)...).Scan(&found); err != nil {
		return err
	}

	if found != len(ids) {
		return model.ErrLinkTargetMissing
	}

	return nil
}

// reconcileBidirectionalLinksTx ensures the reverse side of a bidirectional link is up to date.
// Returns the list of added item descriptors and the field name on the second table.
func (w *workspaceRepository) reconcileBidirectionalLinksTx(tx *sql.Tx, workspaceID, tableID, taskID, secondParentTableID, secondTableID, secondTableName string, linkedItemIDs []string) ([]map[string]string, string, error) {
	var added []map[string]string
	var fieldName string

	secondTable, err := quoteIdent(secondTableName)
	if err != nil {
		return nil, "", err
	}

	var secondParentTableName string
	if err := tx.QueryRow(`SELECT name FROM workspace_tables WHERE id = ?`, tableID).Scan(&secondParentTableName); err != nil {
		return nil, "", err
	}

	for _, linkedItemID := range linkedItemIDs {
		current, err := w.getCurrentLinkedItemsTx(tx, secondParentTableID, linkedItemID, secondTableID, workspaceID)
		if err != nil {
			return nil, "", err
		}

		fieldName = current.ItemFieldName

		alreadyLinked := false
		for _, item := range current.ItemNames {
			if item.ID == taskID {
				alreadyLinked = true
				break
			}
		}

		if alreadyLinked {
			continue
		}

		if _, err = tx.Exec(
			"INSERT INTO "+secondTable+" (id, table_id, table_item_id, parent_table_id, parent_table_item_id) VALUES (UUID(), ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id",
			secondParentTableID, linkedItemID, secondTableID, taskID,
		); err != nil {
			return nil, "", err
		}

		taskName, _ := w.getLinkedItemNameTx(tx, secondParentTableName, taskID)
		added = append(added, map[string]string{
			"taskID":        taskID,
			"linkedItemID":  linkedItemID,
			"tableName":     secondTableName,
			"secondTableID": secondParentTableID,
			"taskName":      taskName,
		})
	}

	return added, fieldName, nil
}

func (w *workspaceRepository) CreateTaskComment(workspaceID string, tableID string, taskID string, comment string, userID string, commentID string) (*model.TaskComment, error) {
	_, err := w.Db.Exec(`
		INSERT INTO workspace_comments (id, user_id, workspace_id, table_id, item_id, content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP())
	`, commentID, userID, workspaceID, tableID, taskID, comment)
	if err != nil {
		return nil, err
	}

	var taskComment model.TaskComment
	err = w.Db.QueryRow(`
		SELECT id, user_id, item_id, content, created_at
		FROM workspace_comments
		WHERE id = ?
	`, commentID).Scan(&taskComment.ID, &taskComment.UserID, &taskComment.ItemID, &taskComment.Content, &taskComment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &taskComment, nil
}

// CreateFolder adds a folder inside parentFolderID, or at the workspace root
// when parentFolderID is empty. It returns sql.ErrNoRows when the parent is
// not live in the workspace.
func (w *workspaceRepository) CreateFolder(ctx context.Context, workspaceID, name, parentFolderID, folderID string) (*model.WorkspaceFolder, error) {
	tx, err := w.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if err := lockFolderTreeTx(ctx, tx, workspaceID); err != nil {
		return nil, err
	}

	if parentFolderID != "" {
		if err := requireLiveFolderTx(ctx, tx, workspaceID, parentFolderID); err != nil {
			return nil, err
		}
	}

	_, err = tx.ExecContext(ctx, `
    INSERT INTO workspace_folders (id, workspace_id, parent_folder_id, name, description, created_at, updated_at, deleted_at)
    VALUES (?, ?, ?, ?, ?, UNIX_TIMESTAMP(), 0, 0)
`, folderID, workspaceID, parentFolderID, name, "des")

	if err != nil {
		return nil, err
	}

	var workspaceFolder model.WorkspaceFolder
	err = tx.QueryRowContext(ctx, `
		SELECT id, workspace_id, parent_folder_id, name, created_at
		FROM workspace_folders
		WHERE id = ?
	`, folderID).Scan(&workspaceFolder.ID, &workspaceFolder.WorkspaceID, &workspaceFolder.ParentFolderID, &workspaceFolder.Name, &workspaceFolder.CreatedAt)
	if err != nil {
		return nil, err
	}

	workspaceFolder.IsFolder = true

	return &workspaceFolder, tx.Commit()
}

func (w *workspaceRepository) GetComments(taskID string) ([]model.TaskComment, error) {
	rows, err := w.Db.Query(`
		SELECT id, item_id, user_id, content, created_at
		FROM workspace_comments
		WHERE item_id = ?
		ORDER BY created_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []model.TaskComment
	for rows.Next() {
		var c model.TaskComment
		if err := rows.Scan(&c.ID, &c.ItemID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}

		c.ActivityType = "Comment"
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (w *workspaceRepository) GetTaskActivities(taskID string) ([]model.Activity, error) {
	rows, err := w.Db.Query(`
		SELECT activity.id, activity.app, activity.type, activity.user_id, activity.affected_user,
		       activity.item_id, activity.parent_id, activity.parameters, activity.created_at,
		       users.name, users.lastname
		FROM activity
		JOIN users ON users.id = activity.user_id
		WHERE activity.item_id = ?
		ORDER BY activity.created_at DESC`, taskID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var activities []model.Activity
	for rows.Next() {
		var a model.Activity
		var parameters string
		var firstName, lastName sql.NullString
		if err := rows.Scan(&a.ID, &a.App, &a.Type, &a.UserID, &a.AffectedUser,
			&a.ItemID, &a.ParentID, &parameters, &a.CreatedAt,
			&firstName, &lastName); err != nil {
			return nil, err
		}

		a.UserName = fmt.Sprintf("%s %s", firstName.String, lastName.String)
		a.ParametersJSON = parameters
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

func (w *workspaceRepository) getLinkedItemNameTx(tx *sql.Tx, tableName, itemID string) (string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return "", err
	}

	var itemName string
	err = tx.QueryRow(`
		SELECT name FROM `+table+`
		WHERE id = ?
	`, itemID).Scan(&itemName)
	if err != nil {
		return "", err
	}

	return itemName, nil
}

func (w *workspaceRepository) getCurrentLinkedItemsTx(tx *sql.Tx, tableID string, taskID string, parentTableID string, workspaceID string) (*model.WorkspaceLinkedItems, error) {
	var linkedItems []string
	var linkedItemID string
	var tableName string

	err := tx.QueryRow(`
        SELECT name FROM workspace_tables
        WHERE id = ?
    `, parentTableID).Scan(&tableName)
	if err != nil {
		return nil, err
	}

	table, err := quoteIdent(tableName)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT parent_table_item_id
        FROM ` + table + `
        WHERE table_id = ? AND table_item_id = ? AND parent_table_id = ?
    `
	rows, err := tx.Query(query, tableID, taskID, parentTableID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&linkedItemID)
		if err != nil {
			return nil, err
		}

		linkedItems = append(linkedItems, linkedItemID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var warehouseParentTableID string
	err = tx.QueryRow(`
        SELECT parent_table_id
        FROM workspace_tables
        WHERE id = ?
    `, parentTableID).Scan(&warehouseParentTableID)
	if err != nil {
		return nil, err
	}

	var warehouseTableName string
	err = tx.QueryRow(`
        SELECT name
        FROM workspace_tables
        WHERE id = ?
    `, warehouseParentTableID).Scan(&warehouseTableName)
	if err != nil {
		return nil, err
	}

	warehouseTable, err := quoteIdent(warehouseTableName)
	if err != nil {
		return nil, err
	}

	var fieldName string
	err = tx.QueryRow(`
        SELECT table_name
        FROM workspace_relationships
        WHERE workspace_id = ? AND table_id = ? AND linked_table_id = ?
    `, workspaceID, tableID, parentTableID).Scan(&fieldName)
	if err != nil {
		return nil, err
	}

	var linkedItemObjects []model.LinkedItem
	for _, itemID := range linkedItems {
		var itemName string
		err = tx.QueryRow(`
        SELECT name
        FROM `+warehouseTable+`
        WHERE id = ?
    `, itemID).Scan(&itemName)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		linkedItemObjects = append(linkedItemObjects, model.LinkedItem{ID: itemID, Name: itemName})
	}

	workspaceLinkedItems := model.WorkspaceLinkedItems{
		TableID:       warehouseParentTableID,
		TableName:     removePrefix(warehouseTableName),
		LinkedItems:   linkedItems,
		ItemFieldName: fieldName,
		ItemNames:     linkedItemObjects,
	}

	return &workspaceLinkedItems, nil
}

func (w *workspaceRepository) GetAttachmentByID(fileID string) (*model.WorkspaceAttachment, error) {
	query := `SELECT id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, COALESCE(storage_id, ''), created_at, updated_at, deleted_at
              FROM workspace_attachments
              WHERE id = ? AND deleted_at = 0`

	row := w.Db.QueryRow(query, fileID)

	attachment := &model.WorkspaceAttachment{}
	err := row.Scan(
		&attachment.ID,
		&attachment.WorkspaceID,
		&attachment.TableID,
		&attachment.FieldID,
		&attachment.TaskID,
		&attachment.UserID,
		&attachment.Name,
		&attachment.Size,
		&attachment.MimeType,
		&attachment.Width,
		&attachment.Height,
		&attachment.StorageID,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

		return nil, err
	}

	return attachment, nil
}

func (w *workspaceRepository) UpdateTaskNameTx(tx *sql.Tx, tableName, taskID, name string) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("UPDATE %s SET name = ? WHERE id = ?", table), name, taskID)
	return err
}

func (w *workspaceRepository) UpdateAttachment(attachment *model.WorkspaceAttachment) error {
	query := `INSERT INTO workspace_attachments
              (id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, storage_id, created_at, updated_at, deleted_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := w.Db.Exec(query,
		attachment.ID,
		attachment.WorkspaceID,
		attachment.TableID,
		attachment.FieldID,
		attachment.TaskID,
		attachment.UserID,
		attachment.Name,
		attachment.Size,
		attachment.MimeType,
		attachment.Width,
		attachment.Height,
		attachment.StorageID,
		attachment.CreatedAt,
		attachment.UpdatedAt,
		attachment.DeletedAt,
	)

	return err
}

func (w *workspaceRepository) GetUserAttachment(fileID string) (*model.WorkspaceAttachment, error) {
	attachmentQuery := `SELECT id, workspace_id, table_id, field_id, task_id, user_id, name, size, mime_type, width, height, COALESCE(storage_id, ''), created_at, updated_at, deleted_at
                        FROM workspace_attachments
                        WHERE id = ? AND deleted_at = 0`

	attachment := &model.WorkspaceAttachment{}
	err := w.Db.QueryRow(attachmentQuery, fileID).Scan(
		&attachment.ID,
		&attachment.WorkspaceID,
		&attachment.TableID,
		&attachment.FieldID,
		&attachment.TaskID,
		&attachment.UserID,
		&attachment.Name,
		&attachment.Size,
		&attachment.MimeType,
		&attachment.Width,
		&attachment.Height,
		&attachment.StorageID,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
		&attachment.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return attachment, nil
}

func (w *workspaceRepository) IsMember(workspaceID, userID string) (bool, error) {
	var count int
	err := w.Db.QueryRow(`
        SELECT COUNT(1)
        FROM workspace_members
        WHERE workspace_id = ? AND user_id = ?
    `, workspaceID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// renameKanbanSectionName updates oldName → newName in a kanban item_order JSON blob.
// It patches both the top-level field "name" entries and the nested section sub-object "name" values
// found inside each order item. Returns the updated JSON, whether anything changed, and any error.
func renameKanbanSectionName(itemOrderJSON, oldName, newName string) (string, bool, error) {
	var itemOrder map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrderJSON), &itemOrder); err != nil {
		return "", false, err
	}

	fields, ok := itemOrder["fields"].([]interface{})
	if !ok {
		return itemOrderJSON, false, nil
	}

	updated := false

	for _, f := range fields {
		fieldData, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		if currentName, ok := fieldData["name"].(string); ok && currentName == oldName {
			fieldData["name"] = newName
			updated = true
		}
	}

	if section, ok := itemOrder["section"].(string); ok {
		for _, f := range fields {
			fieldData, ok := f.(map[string]interface{})
			if !ok {
				continue
			}

			if orders, ok := fieldData["order"].([]interface{}); ok {
				for _, o := range orders {
					orderData, ok := o.(map[string]interface{})
					if !ok {
						continue
					}

					if singleSelect, ok := orderData[section].(map[string]interface{}); ok {
						if currentName, ok := singleSelect["name"].(string); ok && currentName == oldName {
							singleSelect["name"] = newName
							updated = true
						}
					}
				}
			}
		}
	}

	if !updated {
		return itemOrderJSON, false, nil
	}

	b, err := json.Marshal(itemOrder)
	if err != nil {
		return "", false, err
	}

	return string(b), true, nil
}

func (w *workspaceRepository) UpdateKanbanSectionsFieldNameTx(tx *sql.Tx, workspaceID string, tableID string, taskID string, field string, newSingleSelect string, oldName string, LinkedTableID string) (bool, error) {
	rows, err := tx.Query(`
        SELECT id, item_order
        FROM workspace_view
        WHERE workspace_id = ? AND table_id = ? AND parent_table_id = ? AND view_type = 'kanban'
        FOR UPDATE
    `, workspaceID, tableID, LinkedTableID)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	type KanbanView struct {
		ViewID    string
		ItemOrder string
	}

	var kanbanViews []KanbanView

	for rows.Next() {
		var view KanbanView
		if err := rows.Scan(&view.ViewID, &view.ItemOrder); err != nil {
			return false, err
		}

		kanbanViews = append(kanbanViews, view)
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	for _, view := range kanbanViews {
		updatedJSON, modified, err := renameKanbanSectionName(view.ItemOrder, oldName, newSingleSelect)
		if err != nil {
			return false, err
		}

		if !modified {
			continue
		}

		_, err = tx.Exec(`
                UPDATE workspace_view
                SET item_order = ?
                WHERE id = ?
            `, updatedJSON, view.ViewID)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// moveTaskBetweenKanbanSections removes taskID from its current section and appends it to the section
// whose id matches targetID (use "0" for Unassigned) within a kanban item_order JSON blob.
// The view is only considered a match when its "section" key equals field.
// Returns the updated JSON, whether a modification was made, and any error.
func moveTaskBetweenKanbanSections(itemOrderJSON, taskID, field, targetID string) (string, bool, error) {
	if itemOrderJSON == "" || itemOrderJSON == "null" {
		return itemOrderJSON, false, nil
	}

	var itemOrderMap map[string]interface{}
	if err := json.Unmarshal([]byte(itemOrderJSON), &itemOrderMap); err != nil {
		return "", false, err
	}

	sectionValue, ok := itemOrderMap["section"]
	if !ok || sectionValue != field {
		return itemOrderJSON, false, nil
	}

	fields, ok := itemOrderMap["fields"]
	if !ok {
		return itemOrderJSON, false, nil
	}

	sections, ok := fields.([]interface{})
	if !ok {
		return itemOrderJSON, false, nil
	}

	var sourceSection, targetSection map[string]interface{}
	for _, section := range sections {
		sectionMap, ok := section.(map[string]interface{})
		if !ok {
			continue
		}

		orders, ok := sectionMap["order"].([]interface{})
		if !ok {
			sectionMap["order"] = []interface{}{}
			orders = sectionMap["order"].([]interface{})
		}

		for i, orderItem := range orders {
			orderMap, ok := orderItem.(map[string]interface{})
			if ok && orderMap["id"] == taskID {
				sourceSection = sectionMap
				sectionMap["order"] = append(orders[:i], orders[i+1:]...)
				break
			}
		}

		if sectionMap["id"] == targetID {
			targetSection = sectionMap
		}
	}

	if sourceSection == nil || targetSection == nil {
		return itemOrderJSON, false, nil
	}

	targetSection["order"] = append(targetSection["order"].([]interface{}), map[string]interface{}{
		"id":            taskID,
		"single_select": targetID,
	})

	updated, err := json.Marshal(itemOrderMap)
	if err != nil {
		return "", false, err
	}

	return string(updated), true, nil
}

func (w *workspaceRepository) UpdateKanbanSectionsTx(tx *sql.Tx, workspaceID string, tableID string, taskID string, field string, newSingleSelect string) (bool, error) {
	// if empty, put into Unassigned (id "0")
	targetID := newSingleSelect
	if targetID == "" {
		targetID = "0"
	}

	applyToViews := func(views []kanbanViewRow) error {
		for _, v := range views {
			updatedJSON, modified, err := moveTaskBetweenKanbanSections(v.itemOrder, taskID, field, targetID)
			if err != nil {
				return err
			}

			if !modified {
				continue
			}

			if _, err = tx.Exec(`UPDATE workspace_view SET item_order = ? WHERE id = ?`, updatedJSON, v.viewID); err != nil {
				return err
			}
		}

		return nil
	}

	rows, err := tx.Query(`
		SELECT id, item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND view_type = 'kanban'
		FOR UPDATE
	`, workspaceID, tableID)
	if err != nil {
		return false, err
	}

	defer rows.Close()
	kanbanViews, err := scanKanbanViewRowsTx(rows, false)
	if err != nil {
		return false, err
	}

	if err := applyToViews(kanbanViews); err != nil {
		return false, err
	}

	namedRows, err := tx.Query(`
		SELECT id, item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND view_type = 'kanban' AND name = ?
		FOR UPDATE
	`, workspaceID, tableID, field)
	if err != nil {
		return false, err
	}

	defer namedRows.Close()
	namedViews, err := scanKanbanViewRowsTx(namedRows, false)
	if err != nil {
		return false, err
	}

	if err := applyToViews(namedViews); err != nil {
		return false, err
	}

	return true, nil
}

func (w *workspaceRepository) UpdateTable(workspaceID, tableID, name string) (*string, error) {
	_, err := w.Db.Exec(`
		UPDATE workspace_tables
		SET display_name = ?
		WHERE workspace_id = ? AND id = ?
	`, name, workspaceID, tableID)

	if err != nil {
		return nil, err
	}

	return &name, nil
}

func (w *workspaceRepository) Update(workspaceID, name string, description *string) (*string, error) {
	_, err := w.Db.Exec(`
		UPDATE workspaces
		SET title = ?, description = COALESCE(?, description)
		WHERE id = ?
	`, name, description, workspaceID)

	if err != nil {
		return nil, err
	}

	return &name, nil
}

func (w *workspaceRepository) GetTableMetas(workspaceIDs []string) ([]model.WorkspaceTable, error) {
	if len(workspaceIDs) == 0 {
		rows, err := w.Db.Query(`
			SELECT DISTINCT wm.workspace_id
			FROM workspace_members wm
			JOIN workspaces ws ON wm.workspace_id = ws.id
			WHERE ws.deleted_at IS NULL OR ws.deleted_at = 0
		`)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}

			workspaceIDs = append(workspaceIDs, id)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		if len(workspaceIDs) == 0 {
			return nil, nil
		}
	}

	placeholders := strings.Repeat("?,", len(workspaceIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
		SELECT wt.id, wt.name, wt.workspace_id, wt.display_name, ws.title, ws.pre_fix
		FROM workspace_tables wt
		JOIN workspaces ws ON wt.workspace_id = ws.id
		WHERE wt.workspace_id IN (%s)
		  AND wt.linked = 0
		  AND wt.single_select = 0
		  AND wt.deleted_at = 0
		  AND (ws.deleted_at IS NULL OR ws.deleted_at = 0)
	`, placeholders)

	tableRows, err := w.Db.Query(query, toInterfaceSlice(workspaceIDs)...)
	if err != nil {
		return nil, err
	}

	defer tableRows.Close()

	var tables []model.WorkspaceTable
	for tableRows.Next() {
		var t model.WorkspaceTable
		if err := tableRows.Scan(&t.ID, &t.Name, &t.WorkspaceID, &t.DisplayName, &t.WorkspaceName, &t.Prefix); err != nil {
			return nil, err
		}

		tables = append(tables, t)
	}

	if err := tableRows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (w *workspaceRepository) GetTablesByIDs(tableIDs []string) ([]model.WorkspaceTable, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(tableIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`SELECT id, name, display_name FROM workspace_tables WHERE id IN (%s)`, placeholders)

	rows, err := w.Db.Query(query, toInterfaceSlice(tableIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tables []model.WorkspaceTable
	for rows.Next() {
		var t model.WorkspaceTable
		if err := rows.Scan(&t.ID, &t.Name, &t.DisplayName); err != nil {
			return nil, err
		}

		tables = append(tables, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (w *workspaceRepository) GetTableMetasForUser(userID string) ([]model.WorkspaceTable, error) {
	rows, err := w.Db.Query(`
		SELECT ws.id, ws.title, ws.pre_fix
		FROM workspace_members wm
		JOIN workspaces ws ON wm.workspace_id = ws.id
		WHERE wm.user_id = ? AND (ws.deleted_at IS NULL OR ws.deleted_at = 0)
	`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	type workspaceMeta struct {
		name   string
		prefix string
	}

	wsMap := make(map[string]workspaceMeta)
	var wsIDs []string
	for rows.Next() {
		var id, name, prefix string
		if err := rows.Scan(&id, &name, &prefix); err != nil {
			return nil, err
		}

		wsIDs = append(wsIDs, id)
		wsMap[id] = workspaceMeta{name: name, prefix: prefix}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(wsIDs) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(wsIDs))
	placeholders = placeholders[:len(placeholders)-1]

	tableRows, err := w.Db.Query(fmt.Sprintf(`
		SELECT wt.id, wt.name, wt.workspace_id, wt.display_name
		FROM workspace_tables wt
		WHERE wt.workspace_id IN (%s)
		  AND wt.linked = 0
		  AND wt.single_select = 0
		  AND wt.deleted_at = 0
	`, placeholders), toInterfaceSlice(wsIDs)...)
	if err != nil {
		return nil, err
	}

	defer tableRows.Close()

	var tables []model.WorkspaceTable
	for tableRows.Next() {
		var t model.WorkspaceTable
		if err := tableRows.Scan(&t.ID, &t.Name, &t.WorkspaceID, &t.DisplayName); err != nil {
			return nil, err
		}

		if m, ok := wsMap[t.WorkspaceID]; ok {
			t.WorkspaceName = m.name
			t.Prefix = m.prefix
		}

		tables = append(tables, t)
	}

	if err := tableRows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

const taskSourceWhere = "(main.deleted_at IS NULL OR main.deleted_at = 0)"

// TaskOrParentMatches reports whether a task, or any task above it, is not
// deleted and matches filter. The grid shows a matching task with every
// subtask under it, so this is whether the grid shows the task.
func (w *workspaceRepository) TaskOrParentMatches(ctx context.Context, tableName, taskID string, filter model.SQLFilter) (bool, error) {
	if err := checkIdents(tableName); err != nil {
		return false, err
	}

	matches := "1"
	if filter.SQL != "" {
		matches = "CASE WHEN " + filter.SQL + " THEN 1 ELSE 0 END"
	}

	query := fmt.Sprintf("SELECT COALESCE(main.parent_task_id, ''), %s FROM `%s` AS main WHERE main.id = ? AND %s", matches, tableName, taskSourceWhere)

	seen := map[string]bool{}
	for id := taskID; id != "" && !seen[id]; {
		seen[id] = true
		var parent string
		var matched int
		err := w.Db.QueryRowContext(ctx, query, append(append([]any{}, filter.Args...), id)...).Scan(&parent, &matched)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if matched == 1 {
			return true, nil
		}

		if parent == "0" {
			parent = ""
		}

		id = parent
	}

	return false, nil
}

// StreamTableRows calls fn with each of a table's rows that match filter, in
// the order orderSQL gives, reading them from the database as fn takes them
// so a table of any size is never held in memory.
func (w *workspaceRepository) StreamTableRows(ctx context.Context, tableName string, filter model.SQLFilter, orderSQL string, fn func(map[string]interface{}) error) error {
	if err := checkIdents(tableName); err != nil {
		return err
	}

	where := taskSourceWhere
	if filter.SQL != "" {
		where += " AND " + filter.SQL
	}

	rows, err := w.Db.QueryContext(ctx, fmt.Sprintf("SELECT main.* FROM `%s` AS main WHERE %s %s", tableName, where, orderSQL), filter.Args...)
	if err != nil {
		return err
	}

	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	ptrs := make([]interface{}, len(columns))
	for i := range values {
		ptrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}

		if err := fn(row); err != nil {
			return err
		}
	}

	return rows.Err()
}

// GetTableRowsPage returns the first limit of a table's rows that match
// filter. Rows listed in firstIDs come first, in that order, then the rest by
// when they were created.
func (w *workspaceRepository) GetTableRowsPage(ctx context.Context, tableName string, filter model.SQLFilter, firstIDs []string, limit int) ([]map[string]interface{}, error) {
	if err := checkIdents(tableName); err != nil {
		return nil, err
	}

	where := taskSourceWhere
	args := []any{}
	if filter.SQL != "" {
		where += " AND " + filter.SQL
		args = append(args, filter.Args...)
	}

	order := "main.created_at, main.id"
	if len(firstIDs) > 0 {
		in := sqlPlaceholders(len(firstIDs))
		order = "FIELD(main.id, " + in + ") = 0, FIELD(main.id, " + in + "), " + order
		ids := toInterfaceSlice(firstIDs)
		args = append(append(args, ids...), ids...)
	}

	args = append(args, limit)

	rows, err := w.Db.QueryContext(ctx, fmt.Sprintf("SELECT main.* FROM `%s` AS main WHERE %s ORDER BY %s LIMIT ?", tableName, where, order), args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	items, err := scanRowsToMaps(rows)
	if items == nil {
		items = []map[string]interface{}{}
	}

	return items, err
}

// GetTasksAcrossTables returns the first limit of the tasks in several task
// tables, ordered by due date and then id, each row carrying its table_id.
// Every table is limited first, so no table sends more rows than the page can
// use.
func (w *workspaceRepository) GetTasksAcrossTables(ctx context.Context, sources []model.TaskSource, limit int) ([]map[string]interface{}, error) {
	if err := checkSourceTables(sources); err != nil {
		return nil, err
	}

	if len(sources) == 0 {
		return []map[string]interface{}{}, nil
	}

	parts := make([]string, 0, len(sources))
	var args []any
	for _, s := range sources {
		where := taskSourceWhere
		if s.Filter.SQL != "" {
			where += " AND " + s.Filter.SQL
		}

		parts = append(parts, fmt.Sprintf(`(SELECT ? AS table_id, main.id, main.name, main.start_date, main.due_date,
			main.assignee, main.status, main.updated_at, main.parent_task_id
			FROM `+"`%s`"+` AS main WHERE %s ORDER BY main.due_date, main.id LIMIT ?)`, s.TableName, where))
		args = append(args, s.TableID)
		args = append(args, s.Filter.Args...)
		args = append(args, limit)
	}

	query := "SELECT * FROM (" + strings.Join(parts, " UNION ALL ") + ") tasks ORDER BY due_date, id LIMIT ?"
	args = append(args, limit)

	rows, err := w.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return scanRowsToMaps(rows)
}

// GetTaskPageKeys returns which tasks, among several task tables, make up one
// page when they are sorted by sort. With no sort, tasks with a due date come
// first by due date, then the rest by when they were created. Only the keys
// are read, so the rows of a deep page are never loaded; each table is
// limited to the page's end first.
func (w *workspaceRepository) GetTaskPageKeys(ctx context.Context, sources []model.TaskSource, sort []model.SortParam, limit, offset int) ([]model.TaskKey, error) {
	if err := checkSourceTables(sources); err != nil {
		return nil, err
	}

	if len(sources) == 0 {
		return []model.TaskKey{}, nil
	}

	type key struct{ field, dir string }
	var keys []key
	for _, s := range sort {
		if checkIdents(s.Field) != nil {
			continue
		}

		dir := "ASC"
		if strings.EqualFold(s.Direction, "desc") {
			dir = "DESC"
		}

		keys = append(keys, key{s.Field, dir})
	}

	// The default order reads dated tasks by due date and the rest by when
	// they were created as two parts, so each can follow its own index.
	if len(keys) == 0 {
		parts := make([]string, 0, 2*len(sources))
		var args []any
		for _, src := range sources {
			where := taskSourceWhere
			if src.Filter.SQL != "" {
				where += " AND " + src.Filter.SQL
			}

			for group, part := range []struct{ cond, col string }{
				{"main.due_date > 0", "main.due_date"},
				{"(main.due_date IS NULL OR main.due_date = 0)", "main.created_at"},
			} {
				parts = append(parts, fmt.Sprintf("(SELECT ? AS table_id, main.id AS id, %d AS g, %s AS s0 FROM `%s` AS main WHERE %s AND %s ORDER BY %s, main.id LIMIT ?)",
					group, part.col, src.TableName, where, part.cond, part.col))
				args = append(args, src.TableID)
				args = append(args, src.Filter.Args...)
				args = append(args, offset+limit)
			}
		}

		query := "SELECT table_id, id FROM (" + strings.Join(parts, " UNION ALL ") + ") page ORDER BY g, s0, table_id, id LIMIT ? OFFSET ?"
		return w.scanTaskKeys(ctx, query, append(args, limit, offset))
	}

	var outerOrder []string
	for i, k := range keys {
		outerOrder = append(outerOrder, fmt.Sprintf("s%d %s", i, k.dir))
	}

	parts := make([]string, 0, len(sources))
	var args []any
	for _, src := range sources {
		cols := []string{"? AS table_id", "main.id AS id"}
		for i, k := range keys {
			expr := "NULL"
			if src.Columns == nil || src.Columns[k.field] {
				expr = "main.`" + k.field + "`"
			}

			cols = append(cols, fmt.Sprintf("%s AS s%d", expr, i))
		}

		where := taskSourceWhere
		if src.Filter.SQL != "" {
			where += " AND " + src.Filter.SQL
		}

		parts = append(parts, fmt.Sprintf("(SELECT %s FROM `%s` AS main WHERE %s ORDER BY %s, main.id LIMIT ?)",
			strings.Join(cols, ", "), src.TableName, where, strings.Join(outerOrder, ", ")))
		args = append(args, src.TableID)
		args = append(args, src.Filter.Args...)
		args = append(args, offset+limit)
	}

	query := "SELECT table_id, id FROM (" + strings.Join(parts, " UNION ALL ") + ") page ORDER BY " +
		strings.Join(outerOrder, ", ") + ", table_id, id LIMIT ? OFFSET ?"
	return w.scanTaskKeys(ctx, query, append(args, limit, offset))
}

func (w *workspaceRepository) scanTaskKeys(ctx context.Context, query string, args []any) ([]model.TaskKey, error) {
	rows, err := w.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result := []model.TaskKey{}
	for rows.Next() {
		var k model.TaskKey
		if err := rows.Scan(&k.TableID, &k.ID); err != nil {
			return nil, err
		}

		result = append(result, k)
	}

	return result, rows.Err()
}

// GetAttachmentNames returns the names of the files attached to each of the
// given tasks in a file field, newest first.
func (w *workspaceRepository) GetAttachmentNames(ctx context.Context, tableID, fieldID string, taskIDs []string) (map[string][]string, error) {
	names := map[string][]string{}
	if len(taskIDs) == 0 {
		return names, nil
	}

	rows, err := w.Db.QueryContext(ctx, `SELECT task_id, name FROM workspace_attachments
		WHERE table_id = ? AND field_id = ? AND deleted_at = 0 AND task_id IN (`+sqlPlaceholders(len(taskIDs))+`)
		ORDER BY created_at DESC`, append([]any{tableID, fieldID}, toInterfaceSlice(taskIDs)...)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var task, name string
		if err := rows.Scan(&task, &name); err != nil {
			return nil, err
		}

		names[task] = append(names[task], name)
	}

	return names, rows.Err()
}

// GetAttachmentsForTasks returns the files attached to each of the tasks in
// one file field, newest first.
func (w *workspaceRepository) GetAttachmentsForTasks(ctx context.Context, tableID, fieldID string, taskIDs []string) (map[string][]map[string]interface{}, error) {
	files := map[string][]map[string]interface{}{}

	for chunk := range slices.Chunk(taskIDs, 1000) {
		rows, err := w.Db.QueryContext(ctx, `
			SELECT task_id, id, name, size, mime_type, width, height, created_at
			FROM workspace_attachments
			WHERE table_id = ? AND field_id = ? AND deleted_at = 0 AND task_id IN (`+sqlPlaceholders(len(chunk))+`)
			ORDER BY created_at DESC
		`, append([]any{tableID, fieldID}, toInterfaceSlice(chunk)...)...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var taskID, id, name, mimeType string
			var size, createdAt int64
			var width, height int
			if err := rows.Scan(&taskID, &id, &name, &size, &mimeType, &width, &height, &createdAt); err != nil {
				rows.Close()
				return nil, err
			}

			files[taskID] = append(files[taskID], map[string]interface{}{
				"id": id, "name": name, "size": size, "type": mimeType,
				"width": width, "height": height, "created_at": createdAt,
			})
		}

		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return files, nil
}

// CountTasksAcrossTables counts the tasks in several task tables that fall in
// each of groups, returning one count per group. A task is counted in the
// first group it matches.
func (w *workspaceRepository) CountTasksAcrossTables(ctx context.Context, sources []model.TaskSource, groups []model.SQLFilter) ([]int, error) {
	if err := checkSourceTables(sources); err != nil {
		return nil, err
	}

	counts := make([]int, len(groups))
	if len(sources) == 0 || len(groups) == 0 {
		return counts, nil
	}

	var cases strings.Builder
	var caseArgs []any
	for i, g := range groups {
		fmt.Fprintf(&cases, " WHEN %s THEN %d", g.SQL, i)
		caseArgs = append(caseArgs, g.Args...)
	}

	parts := make([]string, 0, len(sources))
	var args []any
	for _, s := range sources {
		where := taskSourceWhere
		if s.Filter.SQL != "" {
			where += " AND " + s.Filter.SQL
		}

		parts = append(parts, fmt.Sprintf("SELECT CASE%s ELSE -1 END AS task_group FROM `%s` AS main WHERE %s", cases.String(), s.TableName, where))
		args = append(args, caseArgs...)
		args = append(args, s.Filter.Args...)
	}

	query := "SELECT task_group, COUNT(*) FROM (" + strings.Join(parts, " UNION ALL ") + ") tasks GROUP BY task_group"

	rows, err := w.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var group, n int
		if err := rows.Scan(&group, &n); err != nil {
			return nil, err
		}

		if group >= 0 && group < len(counts) {
			counts[group] = n
		}
	}

	return counts, rows.Err()
}

func toInterfaceSlice(strs []string) []interface{} {
	interfaces := make([]interface{}, len(strs))
	for i, v := range strs {
		interfaces[i] = v
	}

	return interfaces
}

func (w *workspaceRepository) UpdateMemberRole(workspaceID string, memberID string, role string, user model.User) (bool, error) {
	return w.updateMemberRole(workspaceID, memberID, role)
}

func (w *workspaceRepository) updateMemberRole(workspaceID, memberID, role string) (bool, error) {
	result, err := w.Db.Exec(`
		UPDATE workspace_members
		SET role = ?
		WHERE user_id = ? AND workspace_id = ?
	`, role, memberID, workspaceID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (w *workspaceRepository) GetMainViewByTableID(workspaceID, tableID string) (string, error) {
	var viewID string

	err := w.Db.QueryRow(`
		SELECT id
		FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND main_view = 1
	`, workspaceID, tableID).Scan(&viewID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}

		return "", err
	}

	return viewID, nil
}

// GetMainViewIDs is GetMainViewByTableID for several tables, given as table
// id to workspace id. Tables without a main view are left out.
func (w *workspaceRepository) GetMainViewIDs(workspaceByTable map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(workspaceByTable))
	if len(workspaceByTable) == 0 {
		return out, nil
	}

	tableIDs := make([]string, 0, len(workspaceByTable))
	for id := range workspaceByTable {
		tableIDs = append(tableIDs, id)
	}

	rows, err := w.Db.Query(`
		SELECT table_id, workspace_id, id
		FROM workspace_view
		WHERE table_id IN (`+sqlPlaceholders(len(tableIDs))+`) AND main_view = 1
		ORDER BY table_id, view_type, id
	`, toInterfaceSlice(tableIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var tableID, workspaceID, viewID string
		if err := rows.Scan(&tableID, &workspaceID, &viewID); err != nil {
			return nil, err
		}

		if _, seen := out[tableID]; !seen && workspaceByTable[tableID] == workspaceID {
			out[tableID] = viewID
		}
	}

	return out, rows.Err()
}

func (w *workspaceRepository) GetStatusNameByID(tableName, id string) (string, error) {
	table, err := quoteIdent(tableName)
	if err != nil {
		return "", err
	}

	var name string
	err = w.Db.QueryRow(fmt.Sprintf(`SELECT name FROM %s WHERE id = ?`, table), id).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (w *workspaceRepository) UpdateKanbanOptionColorsTx(tx *sql.Tx, workspaceID, statusTableID, optionID, newColor string) error {
	relRows, err := tx.Query(`
		SELECT table_id, table_name FROM workspace_relationships
		WHERE workspace_id = ? AND linked_table_id = ? AND deleted_at = 0
	`, workspaceID, statusTableID)
	if err != nil {
		return err
	}

	type parentTable struct {
		tableID   string
		fieldName string
	}

	var parents []parentTable
	for relRows.Next() {
		var p parentTable
		if err := relRows.Scan(&p.tableID, &p.fieldName); err != nil {
			relRows.Close()
			return err
		}

		parents = append(parents, p)
	}

	if err := relRows.Err(); err != nil {
		return err
	}

	if err := relRows.Close(); err != nil {
		return err
	}

	if len(parents) == 0 {
		return nil
	}

	fieldByTable := make(map[string]string, len(parents))
	placeholders := make([]string, len(parents))
	queryArgs := make([]interface{}, len(parents)+1)
	queryArgs[0] = workspaceID
	for i, p := range parents {
		placeholders[i] = "?"
		queryArgs[i+1] = p.tableID
		fieldByTable[p.tableID] = p.fieldName
	}

	viewRows, err := tx.Query(fmt.Sprintf(`
		SELECT id, table_id, item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id IN (%s) AND view_type = 'kanban' AND deleted_at = 0
		FOR UPDATE
	`, strings.Join(placeholders, ",")), queryArgs...)
	if err != nil {
		return err
	}

	defer viewRows.Close()
	views, err := scanKanbanViewRowsTx(viewRows, true)
	if err != nil {
		return err
	}

	for _, v := range views {
		updatedJSON, changed, err := updateColorInKanbanOrder(v.itemOrder, fieldByTable[v.tableID], optionID, newColor)
		if err != nil {
			return err
		}

		if !changed {
			continue
		}

		if _, err := tx.Exec(`UPDATE workspace_view SET item_order = ? WHERE id = ?`, updatedJSON, v.viewID); err != nil {
			return err
		}
	}

	return nil
}

func (w *workspaceRepository) GetRolesByName(roleNames []string, workspaceID string) ([]model.ProjectWorkspaceRole, error) {
	if len(roleNames) == 0 {
		return []model.ProjectWorkspaceRole{}, nil
	}

	var roles []model.ProjectWorkspaceRole

	args := make([]interface{}, len(roleNames)+1)
	for i, name := range roleNames {
		args[i] = name
	}

	args[len(roleNames)] = workspaceID

	stmt := `SELECT id, name, displayname, description, permissions, per_table_mode, created_at, updated_at
             FROM workspace_roles
             WHERE name IN (?` + strings.Repeat(",?", len(roleNames)-1) + `) AND workspace_id = ?`

	rows, err := w.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var role model.ProjectWorkspaceRole
		var permissions string

		err = rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&permissions,
			&role.PerTableMode,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		var parsed []string
		if err = json.Unmarshal([]byte(permissions), &parsed); err != nil {
			parsed = strings.Split(permissions, " ")
		}

		role.Permissions = parsed
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (w *workspaceRepository) GetAllActiveIDs() ([]string, error) {
	rows, err := w.Db.Query("SELECT id FROM workspaces WHERE deleted_at = 0")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetViewVisibility reports whether a view is public and who created it.
func (w *workspaceRepository) GetViewVisibility(ctx context.Context, viewID string) (bool, string, error) {
	var isPublic bool
	var createdBy string
	err := w.Db.QueryRowContext(ctx, `SELECT COALESCE(is_public, 1), created_by FROM workspace_view WHERE id = ?`, viewID).Scan(&isPublic, &createdBy)
	return isPublic, createdBy, err
}

func (w *workspaceRepository) GetView(ctx context.Context, viewID string) (*model.WorkspaceView, error) {
	var v model.WorkspaceView
	err := w.Db.QueryRowContext(ctx, `
		SELECT id, workspace_id, table_id, view_type, parent_table_id, name, item_order, created_by, created_at, updated_at, deleted_at, main_view, COALESCE(is_public, 1)
		FROM workspace_view WHERE id = ?
	`, viewID).Scan(&v.ID, &v.WorkspaceID, &v.TableID, &v.ViewType, &v.ParentTableID, &v.Name, &v.TaskOrder,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt, &v.MainView, &v.IsPublic)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// GetMemberUserIDs returns those of userIDs who belong to the workspace,
// directly or through a group.
func (w *workspaceRepository) GetMemberUserIDs(ctx context.Context, workspaceID string, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	in := sqlPlaceholders(len(userIDs))
	args := append([]any{workspaceID}, toInterfaceSlice(userIDs)...)
	args = append(append(args, workspaceID), toInterfaceSlice(userIDs)...)
	rows, err := w.Db.QueryContext(ctx, `
		SELECT wm.user_id FROM workspace_members wm
		WHERE wm.workspace_id = ? AND wm.user_id IN (`+in+`)
		UNION
		SELECT gm.user_id FROM workspace_groups wg
		JOIN group_members gm ON gm.group_id = wg.group_id
		JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		WHERE wg.workspace_id = ? AND gm.user_id IN (`+in+`)
	`, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (w *workspaceRepository) UserHasAnyGroupAccess(workspaceID, userID string) (bool, error) {
	var count int
	err := w.Db.QueryRow(
		`SELECT COUNT(*) FROM workspace_groups wg
		 JOIN group_members gm ON gm.group_id = wg.group_id
		 JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		 WHERE wg.workspace_id = ? AND gm.user_id = ?`,
		workspaceID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (w *workspaceRepository) UserHasOtherGroupAccess(workspaceID, groupID, userID string) (bool, error) {
	var count int
	err := w.Db.QueryRow(
		`SELECT COUNT(*) FROM workspace_groups wg
		 JOIN group_members gm ON gm.group_id = wg.group_id
		 JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		 WHERE wg.workspace_id = ? AND wg.group_id != ? AND gm.user_id = ?`,
		workspaceID, groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (w *workspaceRepository) GetIDsForGroup(groupID string) ([]string, error) {
	rows, err := w.Db.Query(
		`SELECT workspace_id FROM workspace_groups WHERE group_id = ?`, groupID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetAllIDsForUser returns workspace IDs the user belongs to either directly or via a group.
func (w *workspaceRepository) GetAllIDsForUser(userID string) ([]string, error) {
	rows, err := w.Db.Query(`
		SELECT DISTINCT workspace_id FROM workspace_members WHERE user_id = ?
		UNION
		SELECT DISTINCT wg.workspace_id
		FROM workspace_groups wg
		JOIN group_members gm ON gm.group_id = wg.group_id
		WHERE gm.user_id = ?
	`, userID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// GetFieldType returns the kind a custom field was created as, such as url or
// bool, or "" for a column that is not a custom field.
func (w *workspaceRepository) GetFieldType(ctx context.Context, tableID, fieldName string) (string, error) {
	var fieldType string
	err := w.Db.QueryRowContext(ctx,
		`SELECT field_type FROM workspace_fields WHERE table_id = ? AND field_name = ? AND deleted_at = 0 LIMIT 1`,
		tableID, fieldName,
	).Scan(&fieldType)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return fieldType, err
}

// GetPersonFieldNamesForTable returns all column names that store a user reference (assignee, default_assignee, user types).
func (w *workspaceRepository) GetPersonFieldNamesForTable(tableID string) ([]string, error) {
	rows, err := w.Db.Query(
		`SELECT field_name FROM workspace_fields WHERE table_id = ? AND field_type IN ('assignee', 'default_assignee', 'user') AND deleted_at = 0`,
		tableID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		names = append(names, name)
	}

	return names, rows.Err()
}

func (w *workspaceRepository) ClearAssigneeByUserID(tableName, fieldName, userID string) error {
	table, err := quoteIdent(tableName)
	if err != nil {
		return err
	}

	field, err := quoteIdent(fieldName)
	if err != nil {
		return err
	}

	_, err = w.Db.Exec(
		fmt.Sprintf("UPDATE %s SET %s = '' WHERE %s = ?", table, field, field),
		userID,
	)
	return err
}

func (w *workspaceRepository) GetRoles(workspaceID string) ([]model.ProjectWorkspaceRole, error) {
	query := `
		SELECT id, name, displayname, description, permissions, per_table_mode, created_at, updated_at
		FROM workspace_roles
		WHERE workspace_id = ?
	`

	rows, err := w.Db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	roles := make([]model.ProjectWorkspaceRole, 0)

	for rows.Next() {
		var role model.ProjectWorkspaceRole
		var permissionsStr string

		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&permissionsStr,
			&role.PerTableMode,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if permissionsStr != "" {
			role.Permissions = strings.Split(permissionsStr, " ")
		} else {
			role.Permissions = []string{}
		}

		role.WorkspaceID = workspaceID
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (w *workspaceRepository) CreateRole(role model.ProjectWorkspaceRole) (*model.ProjectWorkspaceRole, error) {
	permissionsStr := strings.Join(role.Permissions, " ")

	perTableMode := 0
	if role.PerTableMode {
		perTableMode = 1
	}

	_, err := w.Db.Exec(`
		INSERT INTO workspace_roles
		(id, name, displayname, description, permissions, workspace_id, per_table_mode, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
	`, role.ID, role.Name, role.DisplayName, role.Description, permissionsStr, role.WorkspaceID, perTableMode)

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetRoleNamesForUsers returns each user's role names in a workspace: those
// of their membership first, then those their groups give them, each name
// once. Users with no role are left out.
func (w *workspaceRepository) GetRoleNamesForUsers(ctx context.Context, workspaceID string, userIDs []string) (map[string][]string, error) {
	names := map[string][]string{}
	if len(userIDs) == 0 {
		return names, nil
	}

	seen := map[string]map[string]bool{}
	add := func(userID, role string) {
		if seen[userID] == nil {
			seen[userID] = map[string]bool{}
		}

		if !seen[userID][role] {
			seen[userID][role] = true
			names[userID] = append(names[userID], role)
		}
	}

	in := sqlPlaceholders(len(userIDs))

	rows, err := w.Db.QueryContext(ctx, `SELECT user_id, role FROM workspace_members
		WHERE workspace_id = ? AND user_id IN (`+in+`)`, append([]any{workspaceID}, toInterfaceSlice(userIDs)...)...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var userID, role string
		if err := rows.Scan(&userID, &role); err != nil {
			rows.Close()
			return nil, err
		}

		for _, r := range strings.Fields(role) {
			add(userID, r)
		}
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = w.Db.QueryContext(ctx, `
		SELECT DISTINCT gm.user_id, wg.roles
		FROM workspace_groups wg
		JOIN group_members gm ON gm.group_id = wg.group_id
		JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		WHERE wg.workspace_id = ? AND gm.user_id IN (`+in+`) AND wg.roles IS NOT NULL AND wg.roles != ''
	`, append([]any{workspaceID}, toInterfaceSlice(userIDs)...)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var userID, roles string
		if err := rows.Scan(&userID, &roles); err != nil {
			return nil, err
		}

		for _, r := range strings.Fields(roles) {
			add(userID, r)
		}
	}

	return names, rows.Err()
}

// GetGroupRolesForUser returns all distinct roles a user has via group membership in a workspace.
func (w *workspaceRepository) GetGroupRolesForUser(userID, workspaceID string) ([]string, error) {
	rows, err := w.Db.Query(`
		SELECT DISTINCT wg.roles
		FROM workspace_groups wg
		JOIN group_members gm ON gm.group_id = wg.group_id
		JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		WHERE wg.workspace_id = ? AND gm.user_id = ? AND wg.roles IS NOT NULL AND wg.roles != ''
	`, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	seen := map[string]bool{}
	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}

		for _, part := range strings.Fields(r) {
			if !seen[part] {
				seen[part] = true
				roles = append(roles, part)
			}
		}
	}

	return roles, rows.Err()
}

func (w *workspaceRepository) GetUserByUserID(userID, workspaceId string) (*model.WorkspaceMember, error) {
	query := `
	SELECT id, user_id, workspace_id, role, date_joined
	FROM workspace_members
	WHERE user_id = ? AND workspace_id = ?
	`

	var member model.WorkspaceMember
	err := w.Db.QueryRow(query, userID, workspaceId).Scan(
		&member.ID,
		&member.UserID,
		&member.WorkspaceID,
		&member.Role,
		&member.DateJoined,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &member, nil
}

func (w *workspaceRepository) GetMembers(workspaceID string) ([]model.WorkspaceMember, error) {
	query := `
	SELECT wm.id, wm.user_id, wm.workspace_id, wm.role, wm.date_joined,
	       u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
	FROM workspace_members wm
	JOIN users u ON u.id = wm.user_id
	WHERE wm.workspace_id = ?
	`

	rows, err := w.Db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]model.WorkspaceMember, 0)
	for rows.Next() {
		var member model.WorkspaceMember
		if err := rows.Scan(
			&member.ID,
			&member.UserID,
			&member.WorkspaceID,
			&member.Role,
			&member.DateJoined,
			&member.UserInfo.Email,
			&member.UserInfo.Username,
			&member.UserInfo.Name,
			&member.UserInfo.LastName,
			&member.UserInfo.Photo,
		); err != nil {
			return nil, err
		}

		member.UserInfo.ID = member.UserID
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (w *workspaceRepository) GetSavedFilterOwner(savedFilterID, workspaceID string) (bool, string, error) {
	var isPrivate bool
	var createdBy string

	err := w.Db.QueryRow(`
		SELECT wf.is_private, wf.created_by
		FROM workspace_saved_filters wsf
		INNER JOIN workspace_filters wf ON wf.id = wsf.filter_id
		WHERE wsf.id = ? AND wsf.workspace_id = ?
	`, savedFilterID, workspaceID).Scan(&isPrivate, &createdBy)

	if err != nil {
		return false, "", err
	}

	return isPrivate, createdBy, nil
}

func (w *workspaceRepository) DeleteSavedFilter(savedFilterID, workspaceID string) error {
	_, err := w.Db.Exec(`
		DELETE wf, wsf FROM workspace_saved_filters wsf
		INNER JOIN workspace_filters wf ON wf.id = wsf.filter_id
		WHERE wsf.id = ? AND wsf.workspace_id = ?
	`, savedFilterID, workspaceID)
	return err
}

func (w *workspaceRepository) CreateSavedFilter(workspaceID, tableID, viewID, filterID, userID, id string) (string, error) {
	_, err := w.Db.Exec(`
    INSERT INTO workspace_saved_filters
    (id, workspace_id, table_id, view_id, filter_id, user_id, is_active, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, 0, NOW(), NOW())
    `, id, workspaceID, tableID, viewID, filterID, userID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (w *workspaceRepository) CreateFilter(workspaceID, tableID string, filters model.FilterPayload, name string, isPrivate bool, userID, id string) (string, error) {
	filterJSON, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}

	_, err = w.Db.Exec(`
    INSERT INTO workspace_filters
    (id, workspace_id, table_id, filter_settings, name, is_private, created_by, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
    `, id, workspaceID, tableID, string(filterJSON), name, isPrivate, userID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (w *workspaceRepository) GetViewItemOrder(workspaceID, tableID, viewID string) ([]byte, error) {
	var itemOrder []byte
	err := w.Db.QueryRow(`
		SELECT item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND id = ?
	`, workspaceID, tableID, viewID).Scan(&itemOrder)
	if err != nil {
		return nil, err
	}

	return itemOrder, nil
}

func (w *workspaceRepository) GetGridViewItemOrder(workspaceID, tableID, viewID string) ([]byte, error) {
	var itemOrder []byte
	err := w.Db.QueryRow(`
		SELECT item_order FROM workspace_view
		WHERE workspace_id = ? AND table_id = ? AND id = ? AND view_type = 'grid'
	`, workspaceID, tableID, viewID).Scan(&itemOrder)
	if err != nil {
		return nil, err
	}

	return itemOrder, nil
}

func (w *workspaceRepository) UpdateViewItemOrder(workspaceID, tableID, viewID string, itemOrder []byte) error {
	_, err := w.Db.Exec(`
		UPDATE workspace_view SET item_order = ?
		WHERE workspace_id = ? AND table_id = ? AND id = ?
	`, itemOrder, workspaceID, tableID, viewID)
	return err
}

func (w *workspaceRepository) AddMember(members []model.WorkspaceMember) ([]model.WorkspaceMember, error) {
	tx, err := w.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	if len(members) == 0 {
		return members, nil
	}

	keys := make([]any, 0, len(members)*2)
	for _, m := range members {
		_, err := tx.Exec(`
    INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined)
    VALUES (?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE id = id
    `, m.ID, m.UserID, m.WorkspaceID, m.Role, m.DateJoined)
		if err != nil {
			return nil, err
		}

		keys = append(keys, m.WorkspaceID, m.UserID)
	}

	rows, err := tx.Query(`SELECT id, user_id, workspace_id, role, date_joined FROM workspace_members
		WHERE (workspace_id, user_id) IN (`+rowPlaceholders(len(members), 2)+`)`, keys...)
	if err != nil {
		return nil, err
	}

	stored := make(map[string]model.WorkspaceMember, len(members))
	for rows.Next() {
		var m model.WorkspaceMember
		if err := rows.Scan(&m.ID, &m.UserID, &m.WorkspaceID, &m.Role, &m.DateJoined); err != nil {
			rows.Close()
			return nil, err
		}

		stored[m.WorkspaceID+"/"+m.UserID] = m
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	result := make([]model.WorkspaceMember, 0, len(stored))
	for _, m := range members {
		key := m.WorkspaceID + "/" + m.UserID
		if s, ok := stored[key]; ok {
			result = append(result, s)
			delete(stored, key)
		}
	}

	return result, nil
}

func (w *workspaceRepository) UpdateFilter(filterID, workspaceID, tableID string, filters model.FilterPayload, name string, isPrivate bool, userID string) error {
	filterJSON, err := json.Marshal(filters)
	if err != nil {
		return err
	}

	query := `
    UPDATE workspace_filters 
    SET filter_settings = ?, name = ?, is_private = ?, updated_at = NOW()
    WHERE id = ? AND workspace_id = ? AND table_id = ?
    `

	result, err := w.Db.Exec(query,
		string(filterJSON),
		name,
		isPrivate,
		filterID,
		workspaceID,
		tableID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (w *workspaceRepository) DeactivateFiltersForView(workspaceID, tableID, viewID string) error {
	_, err := w.Db.Exec(`
		UPDATE workspace_saved_filters
		SET is_active = 0, updated_at = NOW()
		WHERE workspace_id = ? AND table_id = ? AND view_id = ? AND is_active = 1
	`, workspaceID, tableID, viewID)
	return err
}

func (w *workspaceRepository) UpdateFilterActiveStatus(workspaceID, tableID, viewID, savedFilterID string, isActive bool) (bool, error) {
	result, err := w.Db.Exec(`
		UPDATE workspace_saved_filters
		SET is_active = ?, updated_at = NOW()
		WHERE id = ? AND workspace_id = ? AND table_id = ? AND view_id = ?
	`, isActive, savedFilterID, workspaceID, tableID, viewID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func (w *workspaceRepository) UpdateSavedFilter(savedFilterID, workspaceID, tableID, viewID string, isActive bool) (bool, error) {
	result, err := w.Db.Exec(`
		UPDATE workspace_saved_filters
		SET is_active = ?, updated_at = NOW()
		WHERE id = ? AND workspace_id = ? AND table_id = ? AND view_id = ?
	`, isActive, savedFilterID, workspaceID, tableID, viewID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

func (w *workspaceRepository) UpdateViewPublic(viewID, workspaceID, tableID string, isPublic bool) error {
	_, err := w.Db.Exec(`
		UPDATE workspace_view SET is_public = ?, updated_at = UNIX_TIMESTAMP()
		WHERE id = ? AND workspace_id = ? AND table_id = ?
	`, isPublic, viewID, workspaceID, tableID)
	return err
}

func (w *workspaceRepository) GetViewSharedUserIDs(viewID string) ([]string, error) {
	rows, err := w.Db.Query(`SELECT user_id FROM workspace_view_shares WHERE view_id = ?`, viewID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}

		ids = append(ids, uid)
	}

	return ids, rows.Err()
}

func (w *workspaceRepository) UpdateViewShares(viewID, workspaceID string, userIDs []string) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM workspace_view_shares WHERE view_id = ?`, viewID); err != nil {
		return err
	}

	for _, uid := range userIDs {
		if _, err := tx.Exec(
			`INSERT INTO workspace_view_shares (id, view_id, workspace_id, user_id, created_at) VALUES (?, ?, ?, ?, UNIX_TIMESTAMP())`,
			model.NewID(), viewID, workspaceID, uid,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (w *workspaceRepository) GetSharedViewIDsForUser(userID, workspaceID string) ([]string, error) {
	rows, err := w.Db.Query(`SELECT view_id FROM workspace_view_shares WHERE user_id = ? AND workspace_id = ?`, userID, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var vid string
		if err := rows.Scan(&vid); err != nil {
			return nil, err
		}

		ids = append(ids, vid)
	}

	return ids, rows.Err()
}

func (w *workspaceRepository) GetTablePermissionsForRole(roleID, workspaceID string) ([]model.TablePermission, error) {
	rows, err := w.Db.Query(`
		SELECT id, role_id, workspace_id, table_id, action, created_at, updated_at
		FROM workspace_table_permissions
		WHERE role_id = ? AND workspace_id = ?
	`, roleID, workspaceID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	perms := make([]model.TablePermission, 0)
	for rows.Next() {
		var p model.TablePermission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.WorkspaceID, &p.TableID, &p.Action, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		perms = append(perms, p)
	}

	return perms, rows.Err()
}

func (w *workspaceRepository) UpdateTablePermissionsForRole(roleID, workspaceID string, perms []model.TablePermission) error {
	tx, err := w.Db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM workspace_table_permissions WHERE role_id = ? AND workspace_id = ?`, roleID, workspaceID); err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, p := range perms {
		if _, err = tx.Exec(`
			INSERT INTO workspace_table_permissions (id, role_id, workspace_id, table_id, action, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, p.ID, roleID, workspaceID, p.TableID, p.Action, now, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (w *workspaceRepository) GetTablePermissionsForRoles(roleIDs []string, workspaceID string) ([]model.TablePermission, error) {
	if len(roleIDs) == 0 {
		return []model.TablePermission{}, nil
	}

	args := make([]interface{}, len(roleIDs)+1)
	for i, id := range roleIDs {
		args[i] = id
	}

	args[len(roleIDs)] = workspaceID

	stmt := `SELECT id, role_id, workspace_id, table_id, action, created_at, updated_at
	         FROM workspace_table_permissions
	         WHERE role_id IN (?` + strings.Repeat(",?", len(roleIDs)-1) + `) AND workspace_id = ?`

	rows, err := w.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	perms := make([]model.TablePermission, 0)
	for rows.Next() {
		var p model.TablePermission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.WorkspaceID, &p.TableID, &p.Action, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		perms = append(perms, p)
	}

	return perms, rows.Err()
}
