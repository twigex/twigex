// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrLinkTargetMissing is returned when a task is to be linked to a row that
// is not in the linked table.
var ErrLinkTargetMissing = errors.New("a linked row does not exist")

// NewTaskFields holds what a task may be created with besides its name. A
// zero value leaves that field empty.
type NewTaskFields struct {
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
	StartDate   int64  `json:"start_date"`
	DueDate     int64  `json:"due_date"`
	Description string `json:"description"`
}

// ErrFolderCycle is returned when a folder is to be moved into itself or into
// a folder inside it.
var ErrFolderCycle = errors.New("a folder cannot be moved inside itself")

// ErrFolderHasTables is returned when a folder to be deleted, or a folder
// inside it, still holds a table.
var ErrFolderHasTables = errors.New("the folder still holds tables")

// LinkingTable is a table with a link column to a table about to be deleted.
type LinkingTable struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var serverKeptColumns = map[string]bool{
	"id":             true,
	"created_at":     true,
	"created_by":     true,
	"updated_at":     true,
	"deleted_at":     true,
	"parent_task_id": true,
}

// IsServerKeptColumn reports whether a task table column is set only by the
// server, never from a request. MySQL matches column names in any case, so
// the check does too.
func IsServerKeptColumn(name string) bool {
	return serverKeptColumns[strings.ToLower(name)]
}

type Workspace struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	StartDate       int64             `json:"start_date"`
	EndDate         int64             `json:"end_date"`
	Prefix          string            `json:"pre_fix"`
	Tables          []WorkspaceTable  `json:"tables"`
	Members         []WorkspaceMember `json:"members"`
	CreatedAt       int64             `json:"created_at"`
	UpdatedAt       int64             `json:"updated_at"`
	DeletedAt       int64             `json:"deleted_at"`
	Folders         []WorkspaceFolder `json:"folders"`
	UserPermissions []string          `json:"user_permissions"`
	UserRoles       []string          `json:"user_roles,omitempty"`
	CanDelete       bool              `json:"can_delete,omitempty"`
	GroupCount      int               `json:"group_count,omitempty"`
	TableCount      int               `json:"table_count,omitempty"`
	MemberCount     int               `json:"member_count,omitempty"`
}

type ProjectWorkspacePatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectWorkspaceDetails struct {
	Workspace Workspace              `json:"workspace"`
	Members   []WorkspaceMember      `json:"members"`
	Groups    []WorkspaceGroup       `json:"groups"`
	Roles     []ProjectWorkspaceRole `json:"roles"`
}

type LinkedItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Restricted marks a linked task the user may not see, sent without its
	// name.
	Restricted bool `json:"restricted,omitempty"`
}

type WorkspaceLinkedItems struct {
	TableID       string       `json:"table_id"`
	TableName     string       `json:"table_name"`
	LinkedItems   []string     `json:"linked_items"`
	ItemFieldName string       `json:"item_field_name"`
	ItemNames     []LinkedItem `json:"item_names"`
}

type WorkspaceMember struct {
	ID               string            `json:"id"`
	UserID           string            `json:"user_id"`
	WorkspaceID      string            `json:"workspace_id"`
	Role             string            `json:"role"`
	DateJoined       int64             `json:"date_joined"`
	Permissions      []string          `json:"permissions,omitempty"`
	PerTableMode     bool              `json:"per_table_mode,omitempty"`
	TablePermissions []TablePermission `json:"table_permissions,omitempty"`
	UserInfo         User              `json:"user_info"`
}

type WorkspaceTable struct {
	ID                string                   `json:"id"`
	WorkspaceID       string                   `json:"workspace_id"`
	Name              string                   `json:"name"`
	DisplayName       sql.NullString           `json:"display_name"`
	ParentTableID     string                   `json:"parent_table_id"`
	Linked            bool                     `json:"linked"`
	SingleSelect      bool                     `json:"single_select"`
	DataBase          []map[string]interface{} `json:"data_base"`
	Headers           []WorkspaceHeaders       `json:"headers"`
	Views             []WorkspaceView          `json:"views"`
	Options           []WorkspaceView          `json:"options"`
	BothDirectionLink bool                     `json:"both_direction_link,omitempty"`
	SecondTableID     string                   `json:"second_table_id,omitempty"`
	FolderID          string                   `json:"folder_id,omitempty"`
	SubtaskProject    *WorkspaceTable          `json:"subtask_project,omitempty"`
	WorkspaceName     string                   `json:"workspace_name,omitempty"`
	Prefix            string                   `json:"prefix,omitempty"`
	Filters           []SavedFilterWithPayload `json:"filters,omitempty"`
	Total             int                      `json:"total,omitempty"`
	RootTotal         int                      `json:"root_total,omitempty"`
}

// AllWorkspaceTasksPage is one page of the task report. Total and RootTotal
// are left out when the caller asked not to count, as it does when only the
// page changes.
type AllWorkspaceTasksPage struct {
	DataBase  []map[string]interface{} `json:"data_base"`
	Total     *int                     `json:"total,omitempty"`
	RootTotal *int                     `json:"root_total,omitempty"`
	Headers   []WorkspaceHeaders       `json:"headers"`
	Filters   []SavedFilterWithPayload `json:"filters"`
}

// TaskSource is one task table to read from, with the conditions its rows
// must meet.
type TaskSource struct {
	TableID   string
	TableName string
	Filter    SQLFilter
	// Columns, when set, are the table's columns, so a sort by a column the
	// table lacks treats its rows as empty there.
	Columns map[string]bool
}

// TaskKey identifies a task among several tables.
type TaskKey struct {
	TableID string
	ID      string
}

// AssignedSection is one due-date group of the user's assigned tasks: how
// many there are and one page of them. Next is where the following page
// starts, and is empty on the last page.
type AssignedSection struct {
	Key   string                   `json:"key"`
	Total int                      `json:"total"`
	Tasks []map[string]interface{} `json:"tasks"`
	Next  string                   `json:"next,omitempty"`
}

// AssignedTable is what the assigned tasks page shows about a task's table.
type AssignedTable struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	TableName     string `json:"table_name"`
	ViewID        string `json:"view_id"`
}

// KanbanColumnPage is one column of a Kanban board: how many cards it holds
// and one page of them. Next is where the following page starts, and is
// empty on the last page.
type KanbanColumnPage struct {
	ID    string                   `json:"id"`
	Total int                      `json:"total"`
	Tasks []map[string]interface{} `json:"tasks"`
	Next  string                   `json:"next,omitempty"`
}

// KanbanPage is a page of a Kanban board, with the subtasks of the cards on
// it.
type KanbanPage struct {
	Columns  []KanbanColumnPage       `json:"columns"`
	Subtasks []map[string]interface{} `json:"subtasks"`
}

// TaskCompletion says what a task leaves open below it, and whether its
// parent is still open with nothing open below it but the task.
type TaskCompletion struct {
	OpenSubtasks int                   `json:"open_subtasks"`
	Parent       *TaskCompletionParent `json:"parent"`
}

type TaskCompletionParent struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Open         bool   `json:"open"`
	OpenSubtasks int    `json:"open_subtasks"`
}

type AssignedToMePage struct {
	Sections []AssignedSection        `json:"sections"`
	Tables   map[string]AssignedTable `json:"tables"`
}

type StatusIndex struct {
	NameToIDs map[string][]string `json:"name_to_ids"`
	FlatList  []StatusField       `json:"flat_list"`
}

type StatusField struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StatusType string `json:"status_type,omitempty"`
}

// MaxTaskCommentRunes is the largest rune count that always fits the
// workspace_comments.content TEXT column: 65535 bytes / 4 bytes per utf8mb4
// char (worst case) = 16383.
const MaxTaskCommentRunes = 16383

type TaskComment struct {
	ID           string `json:"id"`
	ItemID       string `json:"item_id"`
	UserID       string `json:"user_id"`
	AffectedUser string `json:"affected_user"`
	Content      string `json:"content"`
	CreatedAt    int64  `json:"created_at"`
	ActivityType string `json:"activity_type"`
}

type WorkspaceFolder struct {
	ID             string            `json:"id"`
	WorkspaceID    string            `json:"workspace_id"`
	ParentFolderID string            `json:"parent_folder_id,omitempty"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	IsFolder       bool              `json:"is_folder"`
	CreatedAt      int64             `json:"created_at"`
	UpdatedAt      int64             `json:"updated_at,omitempty"`
	DeletedAt      int64             `json:"deleted_at,omitempty"`
	Tables         []WorkspaceTable  `json:"tables,omitempty"`
	Children       []WorkspaceFolder `json:"children"`
}

type WorkspaceHeaders struct {
	ID                 string          `json:"id,omitempty"`
	Name               string          `json:"name"`
	HeaderType         string          `json:"header_type"`
	LinkedID           string          `json:"linked_id"`
	Value              interface{}     `json:"value,omitempty"`
	SingleSelect       bool            `json:"single_select"`
	ParentTableID      string          `json:"parent_table_id,omitempty"`
	ParentLinkTableID  string          `json:"parent_link_table_id,omitempty"`
	HeaderUsage        string          `json:"header_usage,omitempty"`
	LinkBothDirections bool            `json:"link_both_directions,omitempty"`
	LinkedHeaderID     string          `json:"linked_header_id,omitempty"`
	NewValues          interface{}     `json:"new_value,omitempty"`
	ChildTable         bool            `json:"child_table,omitempty"`
	LinkedName         string          `json:"linked_name"`
	DisplayName        string          `json:"display_name"`
	TableData          *WorkspaceTable `json:"table_data,omitempty"`
	Formula            *FormulaSpec    `json:"formula,omitempty"`
	StatusType         string          `json:"status_type,omitempty"`
}

type WorkspaceView struct {
	ID            string                   `json:"id"`
	WorkspaceID   string                   `json:"workspace_id"`
	TableID       string                   `json:"table_id"`
	ViewType      string                   `json:"view_type"`
	ParentTableID string                   `json:"parent_table_id,omitempty"`
	Name          string                   `json:"name"`
	TaskOrder     string                   `json:"order"`
	SelectField   string                   `json:"select_field,omitempty"`
	CreatedBy     string                   `json:"created_by"`
	CreatedAt     int64                    `json:"created_at"`
	UpdatedAt     int64                    `json:"updated_at"`
	DeletedAt     int64                    `json:"deleted_at"`
	MainView      bool                     `json:"main_view,omitempty"`
	IsPublic      bool                     `json:"is_public"`
	Filter        *FilterPayload           `json:"filter"`
	Sort          []map[string]interface{} `json:"sort,omitempty"`
}

// WorkspaceTableViewMetadata is what a table's views need besides its rows.
// Every view carries the user's active filter and its sort.
type WorkspaceTableViewMetadata struct {
	Views        []WorkspaceView
	GridSettings []WorkspaceView
	SavedFilters []SavedFilterWithPayload
}

type WorkspaceFieldData struct {
	ID            string       `json:"id"`
	WorkspaceID   string       `json:"workspace_id"`
	TableID       string       `json:"table_id"`
	Name          string       `json:"name"`
	DisplayName   string       `json:"display_name"`
	FieldType     string       `json:"field_type"`
	ParentFieldID string       `json:"parent_field_id"`
	Formula       *FormulaSpec `json:"formula,omitempty"`
	StatusType    string       `json:"status_type,omitempty"`
	CreatedAt     int64        `json:"created_at"`
	UpdatedAt     int64        `json:"updated_at"`
	DeletedAt     int64        `json:"deleted_at"`
}

// LinkedTableMeta is what a table's headers need to know about a table one
// of its columns links to.
type LinkedTableMeta struct {
	ParentTableID  string
	SingleSelect   bool
	BothDirections bool
}

// TableHeaderMeta is everything a table's column headers are built from,
// loaded at once rather than column by column.
type TableHeaderMeta struct {
	Fields       map[string]WorkspaceFieldData // by field name
	Links        map[string]string             // column name to linked table id
	LinkedTables map[string]LinkedTableMeta    // by linked table id
	// LinkedFieldNames maps a table id and a parent field id, joined by "/",
	// to the name of the field in that table that links back to it.
	LinkedFieldNames map[string]string
}

// systemFields maps built-in column names to their field type and display name.
var systemFields = map[string][2]string{
	"id":          {"text", "ID"},
	"name":        {"text", "Name"},
	"description": {"text", "Description"},
	"start_date":  {"default_date", "Start date"},
	"due_date":    {"default_date", "Due date"},
	"updated_at":  {"default_date", "Updated at"},
	"created_at":  {"default_date", "Created at"},
	"assignee":    {"default_assignee", "Assignee"},
	"created_by":  {"default_assignee", "Created by"},
	"status":      {"status", "Status"},
}

func SystemFieldData(tableID, fieldName string) *WorkspaceFieldData {
	meta, ok := systemFields[fieldName]
	if !ok {
		return nil
	}

	return &WorkspaceFieldData{TableID: tableID, Name: fieldName, FieldType: meta[0], DisplayName: meta[1]}
}

type TaskOrderField struct {
	Color       string                   `json:"color,omitempty"`
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name"`
	DisplayName string                   `json:"display_name"`
	Width       string                   `json:"width"`
	Order       []map[string]interface{} `json:"order,omitempty"` // Order is an optional array of tasks
	Visible     bool                     `json:"visible"`
}

type TaskOrder struct {
	Fields        []TaskOrderField `json:"fields"`
	Section       string           `json:"section"`
	FieldsVisible []string         `json:"fields_visible"`
}

type Filter struct {
	Field           string   `json:"field"`
	Operator        string   `json:"operator"`
	Value           string   `json:"value"`
	OperatorBetween string   `json:"operatorBetween"`
	Date            string   `json:"date,omitempty"`   //for Exact/Before/After
	Date2           string   `json:"date2,omitempty"`  //for Between
	Values          []string `json:"values,omitempty"` //for Select/MultipleSelect
}

// UnmarshalJSON reads a value sent as a list, as filters for a multiple
// choice can be, into Values. Any other value is read as text.
func (f *Filter) UnmarshalJSON(data []byte) error {
	type plain Filter
	var raw struct {
		plain
		Value json.RawMessage `json:"value"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*f = Filter(raw.plain)

	value := bytes.TrimSpace(raw.Value)
	switch {
	case len(value) == 0 || bytes.Equal(value, []byte("null")):
		f.Value = ""

	case value[0] == '[':
		var list []any
		if err := json.Unmarshal(value, &list); err != nil {
			return err
		}

		f.Value = ""
		if len(f.Values) == 0 {
			for _, v := range list {
				f.Values = append(f.Values, fmt.Sprint(v))
			}
		}

	case value[0] == '"':
		return json.Unmarshal(value, &f.Value)

	default:
		f.Value = string(value)
	}

	return nil
}

type FilterGroup struct {
	Filters        []Filter `json:"filters"`
	RelationToNext *string  `json:"relationToNext"` // null-safe
}

type SortParam struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type FilterPayload struct {
	Groups      []FilterGroup `json:"groups"`
	FlatFilters []Filter      `json:"flatFilters"` // optional - if no groups
	Page        int           `json:"page,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Sort        []SortParam   `json:"sort,omitempty"`
	Timezone    string        `json:"timezone,omitempty"`
}

type FormulaArg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type FormulaSpec struct {
	Name string       `json:"name"`
	Args []FormulaArg `json:"args,omitempty"`
}

type ProjectFileMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type WorkspaceAttachment struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	TableID     string `json:"table_id"`
	FieldID     string `json:"field_id"`
	TaskID      string `json:"task_id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	MimeType    string `json:"mime_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	StorageID   string `json:"storage_id,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	DeletedAt   int64  `json:"deleted_at"`
}

type SavedFilterInfo struct {
	ID             string `json:"id"`
	FilterID       string `json:"filter_id"`
	Name           string `json:"name"`
	IsPrivate      bool   `json:"is_private"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      int64  `json:"created_at"`
	UserID         string `json:"user_id,omitempty"`
	ViewID         string `json:"view_id,omitempty"`
	TableID        string `json:"table_id,omitempty"`
	FilterSettings string `json:"-"`
}

type UserTimezone struct {
	AutomaticTimezone    string `json:"automaticTimezone"`
	ManualTimezone       string `json:"manualTimezone"`
	UseAutomaticTimezone bool   `json:"useAutomaticTimezone"`
}

type SavedFilterWithPayload struct {
	SavedFilterInfo
	Filters *FilterPayload `json:"filters,omitempty"`
}

type CreateFieldParams struct {
	WorkspaceID                   string
	TableID                       string
	FieldName                     string
	FieldType                     string
	SQLFieldType                  string
	LinkedTableID                 string
	UserID                        string
	SelectedType                  string
	LinkBothDirections            bool
	FieldNameInSecondTable        string
	FieldNameDisplay              string
	FieldNameInSecondTableDisplay string
	HeaderUsage                   string
	Formula                       *FormulaSpec
	// Pre-generated IDs (set by app layer before calling store)
	MainFieldID                 string
	SecondFieldID               string
	LinkTableID                 string
	SecondLinkID                string
	LinkTablePhysicalName       string
	SecondLinkTablePhysicalName string
	SingleSelectTableID         string
	SingleSelectPhysicalName    string
	LinkViewID                  string
	SecondLinkViewID            string
}

// DeleteCascade describes one cleanup operation that must run alongside a task delete.
type DeleteCascade struct {
	TableName      string // physical table name to update
	FieldName      string // if set: NULL out this column (single-select)
	JunctionColumn string // junction column to filter: "parent_table_item_id" or "table_item_id"
}

type KanbanStatusOption struct {
	ID         string
	Name       string
	Color      string
	StatusType string
}
