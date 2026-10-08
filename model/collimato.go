// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var (
	CollimatoWorkspaceStatusDraft    = "draft"
	CollimatoWorkspaceStatusFinished = "finished"
	CollimatoWorkspaceStatusInactive = "inactive"
	CollimatoWorkspaceStatusActive   = "active"
)

var CollimatoDataFolder = "data/collimato/data"

var (
	CollimatoFileTypeCube   = "cube"
	CollimatoFileTypeView   = "view"
	CollimatoFileTypeEnv    = "env"
	CollimatoFileTypeCubejs = "cubejs"
)

const CollimatoMaxQueryLimit = 50000

type CollimatoWorkspaceFile struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	FileType    string `json:"file_type"`
	Content     string `json:"content"`
	Encrypted   bool   `json:"-"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	// BuilderModel is the structured JSON the visual builder round-trips from.
	// NULL for hand-written/legacy files, which open in the raw code editor.
	BuilderModel *string `json:"builder_model,omitempty"`
}

type Connection struct {
	ID              string           `json:"id"`
	WorkspaceID     string           `json:"workspace_id"`
	Type            string           `json:"type"`
	DisplayName     string           `json:"displayname"`
	Host            string           `json:"host"`
	Port            string           `json:"port"`
	Username        string           `json:"username"`
	Password        string           `json:"-"`
	Database        string           `json:"database"`
	SSLMode         string           `json:"ssl_mode"`
	Config          ConnectionConfig `json:"config"`
	EncryptedConfig sql.NullString   `json:"-"`
	Default         bool             `json:"default"`
	CreatedBy       string           `json:"createdby"`
	CreatedAt       int64            `json:"createdat"`
	UpdatedAt       int64            `json:"updatedat"`
	DeletedAt       int64            `json:"deletedat"`
}

// SSL modes a connection may use, named as libpq names them. A verify mode may
// carry a CA certificate in the config; without one the host's roots are used.
const (
	SSLModeDisable    = "disable"
	SSLModeRequire    = "require"
	SSLModeVerifyCA   = "verify-ca"
	SSLModeVerifyFull = "verify-full"
)

func SSLModeSupported(mode string) bool {
	switch mode {
	case SSLModeDisable, SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull:
		return true
	}

	return false
}

// ConnectionConfig holds engine specific settings that are not secret. Secrets
// belong in the encrypted column instead.
type ConnectionConfig map[string]string

func (c ConnectionConfig) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}

	return json.Marshal(c)
}

func (c *ConnectionConfig) Scan(value any) error {
	*c = ConnectionConfig{}

	if value == nil {
		return nil
	}

	switch raw := value.(type) {
	case []byte:
		return json.Unmarshal(raw, c)
	case string:
		return json.Unmarshal([]byte(raw), c)
	}

	return fmt.Errorf("connection config: unsupported type %T", value)
}

type NewConnection struct {
	WorkspaceID     string           `json:"workspace_id"`
	Type            string           `json:"type"`
	DisplayName     string           `json:"displayname"`
	Host            string           `json:"host"`
	Port            string           `json:"port"`
	Username        string           `json:"username"`
	Password        string           `json:"password"`
	Database        string           `json:"database"`
	SSLMode         string           `json:"ssl_mode"`
	Config          ConnectionConfig `json:"config"`
	EncryptedConfig sql.NullString   `json:"-"`
	Default         bool             `json:"default"`
	CreatedBy       string           `json:"createdby"`
}

type ConnectionPatch struct {
	Type        *string           `json:"type"`
	DisplayName *string           `json:"displayname"`
	Host        *string           `json:"host"`
	Port        *string           `json:"port"`
	Username    *string           `json:"username"`
	Password    *string           `json:"password"`
	Database    *string           `json:"database"`
	SSLMode     *string           `json:"ssl_mode"`
	Config      *ConnectionConfig `json:"config"`
	Default     *bool             `json:"default"`
	CreatedBy   *string           `json:"createdby"`
}

// NeedsVerification reports whether the patch changes how the database is
// reached, so renaming a connection does not require the server to be up. The
// comparison is against current because forms tend to send every field.
func (p ConnectionPatch) NeedsVerification(current Connection) bool {
	if p.Password != nil && *p.Password != "" {
		return true
	}

	return differs(p.Type, current.Type) ||
		differs(p.Host, current.Host) ||
		differs(p.Port, current.Port) ||
		differs(p.Username, current.Username) ||
		differs(p.Database, current.Database) ||
		differs(p.SSLMode, current.SSLMode) ||
		configDiffers(p.Config, current.Config)
}

func differs(patched *string, current string) bool {
	return patched != nil && *patched != current
}

func configDiffers(patched *ConnectionConfig, current ConnectionConfig) bool {
	if patched == nil {
		return false
	}

	if len(*patched) != len(current) {
		return true
	}

	for key, value := range *patched {
		if current[key] != value {
			return true
		}
	}

	return false
}

type NewChart struct {
	Name          string                 `json:"name"`
	Query         DataQuery              `json:"query"`
	ChartType     string                 `json:"chart_type"`
	Configuration map[string]interface{} `json:"configuration"`
	OwnerID       string                 `json:"owner_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	Model         string                 `json:"model"`
}

func (c *NewChart) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.ChartType == "" {
		return fmt.Errorf("chart_type is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	return nil
}

type Chart struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	ChartType     string                 `json:"chart_type"`
	Configuration map[string]interface{} `json:"configuration"`
	Position      map[string]interface{} `json:"position,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	OwnerID       string                 `json:"owner_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	CreatedAt     int64                  `json:"created_at"`
	UpdatedAt     int64                  `json:"updated_at"`
	DeletedAt     int64                  `json:"deleted_at"`
}

func (c *Chart) Patch(patch ChartPatch) {
	if patch.Name != nil {
		c.Name = *patch.Name
	}

	if patch.ChartType != nil {
		c.ChartType = *patch.ChartType
	}

	if patch.Configuration != nil {
		c.Configuration = *patch.Configuration
	}

	if patch.Query != nil {
		c.Data["query"] = patch.Query
	}

	if patch.Model != nil {
		c.Data["model"] = patch.Model
	}
}

type ChartPatch struct {
	Name          *string                 `json:"name"`
	Query         *DataQuery              `json:"query"`
	ChartType     *string                 `json:"chart_type"`
	Configuration *map[string]interface{} `json:"configuration"`
	Model         *string                 `json:"model"`
}

type DataQuery struct {
	Measures       []string             `json:"measures"`
	Dimensions     []string             `json:"dimensions"`
	Filters        []QueryFilter        `json:"filters"`
	TimeDimensions []QueryTimeDimension `json:"timeDimensions"`
	Order          [][]string           `json:"order"`
	Limit          int                  `json:"limit"`
}

func (q *DataQuery) SetLimit() {
	if q.Limit <= 0 || q.Limit > CollimatoMaxQueryLimit {
		q.Limit = CollimatoMaxQueryLimit
	}
}

type QueryFilter struct {
	Member   string        `json:"member,omitempty"`
	Operator string        `json:"operator,omitempty"`
	Values   []string      `json:"values,omitempty"`
	And      []QueryFilter `json:"and,omitempty"`
	Or       []QueryFilter `json:"or,omitempty"`
}

type QueryTimeDimension struct {
	Dimension   string         `json:"dimension"`
	DateRange   QueryDateRange `json:"dateRange"`
	Granularity string         `json:"granularity"`
}

// DateRange can represent either a string or an array of strings
type QueryDateRange struct {
	StringValue string
	ArrayValue  []string
	IsArray     bool
}

func (dr *QueryDateRange) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		dr.StringValue = single
		dr.IsArray = false
		return nil
	}

	var array []string
	if err := json.Unmarshal(data, &array); err == nil {
		dr.ArrayValue = array
		dr.IsArray = true
		return nil
	}

	return fmt.Errorf("DateRange must be either a string or an array of strings")
}

func (dr *QueryDateRange) MarshalJSON() ([]byte, error) {
	if dr.IsArray {
		return json.Marshal(dr.ArrayValue)
	}

	return json.Marshal(dr.StringValue)
}

type CollimatoWorkspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ServerID    string `json:"server_id"`
	CreatedBy   string `json:"created_by"`
	Secret      string `json:"-"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	DeletedAt   int64  `json:"deleted_at"`
	// Only the workspace list fills these: a few member IDs for the card's
	// avatars and the full count.
	MemberIDs   []string `json:"member_ids,omitempty"`
	MemberCount int      `json:"member_count,omitempty"`
	CanDelete   bool     `json:"can_delete,omitempty"`
}

type CollimatoWorkspaceUser struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	UserInfo    User     `json:"user_info"`
	Permissions []string `json:"permissions,omitempty"`
	ViaGroup    bool     `json:"via_group,omitempty"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	DeletedAt   int64    `json:"deleted_at"`
}

// A workspace member's roles are the
// union of their direct roles and the roles from every attached group they
// belong to. Resolved live
type CollimatoWorkspaceGroup struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	GroupID     string   `json:"group_id"`
	Roles       []string `json:"roles"`
	AddedBy     string   `json:"added_by"`
	AddedAt     int64    `json:"added_at"`
	UpdatedAt   int64    `json:"updated_at,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	MemberCount int      `json:"member_count,omitempty"`
}

type AddWorkspaceGroupsRequest struct {
	GroupIDs []string `json:"group_ids"`
	Roles    []string `json:"roles"`
}

type UpdateWorkspaceGroupRolesRequest struct {
	Roles []string `json:"roles"`
}

type CollimatoWorkspacePatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CollimatoWorkspaceDetails struct {
	Workspace CollimatoWorkspace        `json:"workspace"`
	Users     []CollimatoWorkspaceUser  `json:"users"`
	Groups    []CollimatoWorkspaceGroup `json:"groups"`
	Roles     []CollimatoRole           `json:"roles"`
}

type Dashboard struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Charts      []Chart           `json:"charts"`
	OwnerID     string            `json:"owner_id"`
	WorkspaceID string            `json:"workspace_id"`
	Filters     []DashboardFilter `json:"filters"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	DeletedAt   int64             `json:"deleted_at"`
}

func (d *Dashboard) Patch(patch DashboardPatch) {
	if patch.Title != nil {
		d.Title = *patch.Title
	}

	if patch.Description != nil {
		d.Description = *patch.Description
	}
}

type DashboardFilter struct {
	ID          string   `json:"id"`
	DashboardID string   `json:"dashboard_id"`
	Name        string   `json:"name"`
	Table       string   `json:"table"`
	Column      string   `json:"column"`
	Operator    string   `json:"operator"`
	Values      []string `json:"values"`
	ApplyTo     []string `json:"apply_to"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	DeletedAt   int64    `json:"deleted_at"`
}

type DashboardFilterPatch struct {
	Name     *string   `json:"name"`
	Table    *string   `json:"table"`
	Column   *string   `json:"column"`
	Operator *string   `json:"operator"`
	Values   *[]string `json:"values"`
	ApplyTo  *[]string `json:"apply_to"`
}

func (f *DashboardFilter) Patch(patch DashboardFilterPatch) {
	if patch.Name != nil {
		f.Name = *patch.Name
	}

	if patch.Table != nil {
		f.Table = *patch.Table
	}

	if patch.Column != nil {
		f.Column = *patch.Column
	}

	if patch.Operator != nil {
		f.Operator = *patch.Operator
	}

	if patch.Values != nil {
		f.Values = *patch.Values
	}

	if patch.ApplyTo != nil {
		f.ApplyTo = *patch.ApplyTo
	}
}

type DashboardPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type DashboardChartPatch struct {
	ID       string                 `json:"id"`
	Position DashboardChartPosition `json:"position"`
}

type DashboardChartPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type DSN struct {
	Username string
	Password string
	Protocol string
	Host     string
	Port     string
	Database string
}

type DatabaseTable struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Database string      `json:"database"`
	Tables   []TableName `json:"tables"`
}

type TableName struct {
	Name    string         `json:"name"`
	Columns []ColumnDetail `json:"columns"`
}

type ColumnDetail struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type GenerateModelsResult struct {
	Created        []string            `json:"created"`
	Skipped        []string            `json:"skipped"`
	SkippedColumns map[string][]string `json:"skipped_columns"`
}

type CubeFile struct {
	Name         string  `json:"name"`
	Content      string  `json:"content"`
	Path         string  `json:"path"`
	Type         string  `json:"type"`
	BuilderModel *string `json:"builder_model,omitempty"`
}

type Dataset struct {
	Name         string          `json:"name"`
	SqlTable     string          `json:"sql_table"`
	ConnectionID string          `json:"connection_id"`
	Dimensions   []CubeDimension `json:"dimensions"`
	Measures     []CubeMeasure   `json:"measures"`
	Joins        []CubeJoin      `json:"joins,omitempty"`
}

type CubeDimension struct {
	Name       string `json:"name"`
	SQL        string `json:"sql"`
	Type       string `json:"type"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
	Title      string `json:"title,omitempty"`
	// Custom marks a raw-SQL dimension
	Custom bool `json:"custom,omitempty"`
}

type CubeMeasure struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	SQL    string `json:"sql,omitempty"`
	Title  string `json:"title,omitempty"`
	Format string `json:"format,omitempty"`
	// Custom marks a raw-SQL measure
	Custom bool `json:"custom,omitempty"`
}

type CubeJoin struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	// ThisColumn/OtherColumn are the structured condition the builder uses; the
	// renderer builds the join SQL from them. SQL is an optional raw override and
	// round-trips so an advanced hand-written condition survives re-open.
	ThisColumn  string `json:"this_column,omitempty"`
	OtherColumn string `json:"other_column,omitempty"`
	SQL         string `json:"sql,omitempty"`
}

type View struct {
	Name  string        `json:"name"`
	Cubes []ViewCubeRef `json:"cubes"`
}

type ViewCubeRef struct {
	// JoinPath is the base cube name, or a dotted path following joins
	// ("files.channels"). Includes lists member names, or a single "*" for all.
	JoinPath string   `json:"join_path"`
	Includes []string `json:"includes"`
	Prefix   bool     `json:"prefix,omitempty"`
}

type CubeSchemaFile struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

// CubeConnection is the database connection Cube.js resolves at query time via
// its driverFactory for a given workspace and dataSource.
type CubeConnection struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	// SSLMode and SSLCA are passed through rather than translated here: the
	// shape a driver wants for TLS belongs with the engine that talks to it.
	SSLMode string `json:"ssl_mode"`
	SSLCA   string `json:"ssl_ca,omitempty"`
}

func (c *Connection) UsesCACertificate() bool {
	return sslModeUsesCACertificate(c.SSLMode)
}

func (c *NewConnection) UsesCACertificate() bool {
	return sslModeUsesCACertificate(c.SSLMode)
}

func sslModeUsesCACertificate(mode string) bool {
	return mode == SSLModeVerifyCA || mode == SSLModeVerifyFull
}

func (c *Connection) Patch(new ConnectionPatch) error {
	if new.Database != nil {
		c.Database = *new.Database
	}

	if new.Default != nil {
		c.Default = *new.Default
	}

	if new.DisplayName != nil {
		c.DisplayName = *new.DisplayName
	}

	if new.Host != nil {
		c.Host = *new.Host
	}

	if new.Password != nil && *new.Password != "" {
		c.Password = *new.Password
	}

	if new.Port != nil {
		c.Port = *new.Port
	}

	if new.SSLMode != nil {
		c.SSLMode = *new.SSLMode
	}

	if new.Config != nil {
		c.Config = *new.Config
	}

	if new.Type != nil {
		c.Type = *new.Type
	}

	if new.Username != nil {
		c.Username = *new.Username
	}

	return nil
}
