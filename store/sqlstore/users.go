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

var userColumnNames = []string{
	"id", "email", "role", "password", "auth_service", "username", "name",
	"lastname", "storage_limit", "timezone", "mfa_active", "mfa_secret",
	"created_at", "updated_at", "deactivated_at", "auth_data",
}

const statusColumns = `user_id, status, last_activity, user_defined`

func userColumns(table string) string {
	if table == "" {
		table = "users"
	}

	qualified := make([]string, 0, len(userColumnNames)+1)
	for i, c := range userColumnNames {
		if i == 9 {
			qualified = append(qualified, photoColumn(table))
		}

		qualified = append(qualified, table+"."+c)
	}

	return strings.Join(qualified, ", ")
}

func photoColumn(table string) string {
	return "COALESCE((SELECT photo_id FROM user_photos WHERE user_id = " + table + ".id), '')"
}

type userRepository struct {
	Db *sql.DB
}

func NewUserRepository(Db *sql.DB) (*userRepository, error) {
	repo := &userRepository{}

	repo.Db = Db
	return repo, nil
}

func (m *userRepository) GetAll() ([]model.User, error) {
	results, err := m.Db.Query("SELECT " + userColumns("") + " FROM users")
	if err != nil {
		return nil, err
	}

	defer results.Close()

	Users := make([]model.User, 0)
	user := model.User{}
	for results.Next() {
		var timezone []byte
		err = results.Scan(&user.ID, &user.Email, &user.Role,
			&user.Password, &user.AuthService, &user.Username,
			&user.Name, &user.LastName, &user.StorageLimit, &user.Photo,
			&timezone, &user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
			return nil, err
		}

		Users = append(Users, user)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return Users, nil
}

// Shared with Count so a page and its total always cover the same set.
func userListFilter(query string, includeDeactivated bool) (string, []any) {
	conditions := []string{}
	args := []any{}

	if !includeDeactivated {
		conditions = append(conditions, "deactivated_at = 0")
	}

	if query != "" {
		like := "%" + escapeLike(query) + "%"
		conditions = append(conditions,
			"(name LIKE ? OR lastname LIKE ? OR email LIKE ? OR username LIKE ?)")
		args = append(args, like, like, like, like)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

var userSortColumns = map[string][]string{
	"name":          {"name", "lastname"},
	"email":         {"email"},
	"storage_limit": {"storage_limit"},
	"status":        {"deactivated_at"},
	"role":          {"role"},
}

func (m *userRepository) GetAllPaged(ctx context.Context, query string, sort model.Sort, limit, offset int, includeDeactivated bool) ([]model.User, error) {
	q := `SELECT id, email, username, name, lastname,
	             COALESCE((SELECT photo_id FROM user_photos WHERE user_id = users.id), ''), role, storage_limit, deactivated_at
	      FROM users`

	where, args := userListFilter(query, includeDeactivated)
	q += where + orderBy(sort, userSortColumns, "name", "id") + ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := m.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.Username,
			&u.Name,
			&u.LastName,
			&u.Photo,
			&u.Role,
			&u.StorageLimit,
			&u.DeactivatedAt,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m *userRepository) Count(ctx context.Context, query string, includeDeactivated bool) (int, error) {
	where, args := userListFilter(query, includeDeactivated)

	var total int
	if err := m.Db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`+where, args...).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

// Search returns up to `limit` active users whose name, lastname, email or
// username matches the (already-trimmed) query, ordered by name. An empty query
// returns the first `limit` users. Only the fields needed by user pickers are
// selected, never password/mfa columns. LIKE metacharacters in the query are
// escaped so user input cannot inject wildcards.
func (m *userRepository) Search(ctx context.Context, query string, limit int) ([]model.User, error) {
	q := `SELECT id, email, username, name, lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = users.id), '')
	      FROM users
	      WHERE deactivated_at = 0`
	args := []any{}

	if query != "" {
		like := "%" + escapeLike(query) + "%"
		q += ` AND (name LIKE ? OR lastname LIKE ? OR email LIKE ? OR username LIKE ?)`
		args = append(args, like, like, like, like)
	}

	q += ` ORDER BY name ASC, lastname ASC LIMIT ?`
	args = append(args, limit)

	rows, err := m.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// escapeLike escapes the LIKE metacharacters (\, %, _) so a search term is
// matched literally rather than as a pattern.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

func (m *userRepository) GetByEmail(email string) (*model.User, error) {
	row := m.Db.QueryRow("SELECT "+userColumns("")+" FROM users WHERE email=?", email)
	user := model.User{}

	var timezone []byte
	err := row.Scan(&user.ID, &user.Email, &user.Role, &user.Password,
		&user.AuthService, &user.Username,
		&user.Name, &user.LastName, &user.StorageLimit, &user.Photo,
		&timezone, &user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *userRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := m.Db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deactivated_at = 0").Scan(&count)

	return count, err
}

func (m *userRepository) Get(id string) (*model.User, error) {
	row := m.Db.QueryRow("SELECT "+userColumns("")+" FROM users WHERE id=?", id)
	user := model.User{}
	var timezone []byte
	err := row.Scan(&user.ID, &user.Email, &user.Role,
		&user.Password, &user.AuthService, &user.Username,
		&user.Name, &user.LastName, &user.StorageLimit, &user.Photo,
		&timezone, &user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *userRepository) GetByUsername(username string) (*model.User, error) {
	rows := m.Db.QueryRow("SELECT "+userColumns("")+" FROM users WHERE username=?", username)

	user := model.User{}
	var timezone []byte
	err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.Password,
		&user.AuthService, &user.Username, &user.Name,
		&user.LastName, &user.StorageLimit, &user.Photo, &timezone,
		&user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	err = json.Unmarshal(timezone, &user.Timezone)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *userRepository) GetByAuthData(providerID, ID string) (*model.User, error) {
	row := m.Db.QueryRow("SELECT "+userColumns("")+" FROM users WHERE auth_service = ? AND auth_data = ?", providerID, ID)
	user := model.User{}

	var timezone []byte
	err := row.Scan(&user.ID, &user.Email, &user.Role,
		&user.Password, &user.AuthService, &user.Username,
		&user.Name, &user.LastName, &user.StorageLimit, &user.Photo,
		&timezone, &user.MfaActive, &user.MfaSecret, &user.CreatedAt, &user.UpdatedAt, &user.DeactivatedAt, &user.AuthData)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *userRepository) GetByIDs(ids []string) ([]model.User, error) {
	if len(ids) == 0 {
		users := make([]model.User, 0)
		return users, nil
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	users := make([]model.User, 0)

	stmt := "SELECT " + userColumns("") + ` FROM users WHERE id IN (?` +
		strings.Repeat(",?", len(args)-1) +
		`)`

	results, err := m.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	for results.Next() {
		var user model.User
		var timezone []byte

		err = results.Scan(
			&user.ID,
			&user.Email,
			&user.Role,
			&user.Password,
			&user.AuthService,
			&user.Username,
			&user.Name,
			&user.LastName,
			&user.StorageLimit,
			&user.Photo,
			&timezone,
			&user.MfaActive,
			&user.MfaSecret,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeactivatedAt,
			&user.AuthData,
		)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m *userRepository) GetByUsernames(ctx context.Context, usernames []string) ([]model.User, error) {
	if len(usernames) == 0 {
		users := make([]model.User, 0)
		return users, nil
	}

	args := make([]any, len(usernames))
	for i, username := range usernames {
		args[i] = username
	}

	q := `SELECT id, email, username, name, lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = users.id), '')
	      FROM users
	      WHERE deactivated_at = 0
	        AND username IN (?` + strings.Repeat(",?", len(args)-1) + `)`

	rows, err := m.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, len(usernames))
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetActiveByIDs batch-loads active users by id in a single query, including
// role so callers can run permission checks without a per-user lookup.
func (m *userRepository) GetActiveByIDs(ctx context.Context, ids []string) ([]model.User, error) {
	if len(ids) == 0 {
		users := make([]model.User, 0)
		return users, nil
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	q := `SELECT id, email, role, name, lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = users.id), '')
	      FROM users
	      WHERE deactivated_at = 0
	        AND id IN (?` + strings.Repeat(",?", len(args)-1) + `)`

	rows, err := m.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, len(ids))
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetSharedUsers returns every user with access to fileID: direct user shares
// unioned with members of every group the file is shared with. Used for
// notification fan-out where the full effective audience matters.
//
// For the share-management UI, use GetSharedUsersForFile (users, inheritance
// aware) alongside GroupStore.GetSharesForFile (groups listed separately).
func (m *userRepository) GetSharedUsers(fileID string) ([]model.User, error) {
	results, err := m.Db.Query(`SELECT DISTINCT
				users.id,
				users.email,
				users.role,
				users.password,
				users.username,
				users.name,
				users.lastname,
				COALESCE((SELECT photo_id FROM user_photos WHERE user_id = users.id), '')
				FROM users
				WHERE users.id IN (
					SELECT fs.share_with FROM file_share fs WHERE fs.file_id = ? AND fs.share_type = ?
					UNION
					SELECT gm.user_id FROM file_share fs
						JOIN group_members gm ON gm.group_id = fs.share_with
						WHERE fs.file_id = ? AND fs.share_type = ?
				)`, fileID, model.SHARE_TYPE_USER, fileID, model.SHARE_TYPE_GROUP)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	arr := make([]model.User, 0)
	user := model.User{}
	for results.Next() {
		err := results.Scan(&user.ID, &user.Email, &user.Role,
			&user.Password, &user.Username,
			&user.Name, &user.LastName, &user.Photo)
		if err != nil {
			return nil, err
		}

		arr = append(arr, user)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return arr, nil
}

// GetSharedUsersForFile lists the users who can access file, resolving grants up
// the ancestor chain (nearest grant per user). A grant on an ancestor rather
// than the file itself is flagged Inherited with GrantedBy set to that folder's
// name, so the UI can escalate a revoke/edit to it.
func (m *userRepository) GetSharedUsersForFile(ctx context.Context, file model.File) ([]model.SharedUsers, error) {
	results, err := m.Db.QueryContext(ctx, `
		WITH RECURSIVE anc AS (
			SELECT id, parent, displayname, 0 AS depth FROM files WHERE id = ?
			UNION ALL
			SELECT f.id, f.parent, f.displayname, a.depth + 1
			FROM files f JOIN anc a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		)
		SELECT u.id, u.name, u.lastname, u.email, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), ''),
			fs.expiration, fs.access_level,
			anc.id, anc.displayname
		FROM anc
		JOIN file_share fs ON fs.file_id = anc.id AND fs.share_type = ?
		JOIN users u ON u.id = fs.share_with
		ORDER BY u.id, anc.depth`,
		file.ID, model.SHARE_TYPE_USER,
	)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	sharedUsers := make([]model.SharedUsers, 0)
	seen := make(map[string]bool)
	for results.Next() {
		var su model.SharedUsers
		var grantedByID, grantedByName string
		if err := results.Scan(&su.ID, &su.Name, &su.LastName, &su.Email, &su.Photo,
			&su.Expiration, &su.AccessLevel,
			&grantedByID, &grantedByName); err != nil {
			return nil, err
		}

		if seen[su.ID] {
			continue
		}

		seen[su.ID] = true

		su.Owner = file.Owner == su.ID
		if grantedByID != file.ID {
			su.Inherited = true
			su.GrantedBy = grantedByName
		}

		sharedUsers = append(sharedUsers, su)
	}

	return sharedUsers, results.Err()
}

func (m *userRepository) Create(user model.NewUser) (*model.User, error) {
	const queryTimeout = 5 * time.Second

	// Start a transaction
	tx, err := m.Db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	userID := model.NewID()
	now := time.Now().Unix()

	timezone := map[string]any{
		"useAutomaticTimezone": true,
		"automaticTimezone":    "",
		"manualTimezone":       "",
	}

	tz, err := json.Marshal(timezone)
	if err != nil {
		return nil, err
	}

	// Insert user into users table
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	_, err = tx.ExecContext(ctx, `INSERT INTO users (`+strings.Join(userColumnNames, ", ")+`)
	                              VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, user.Email, user.Role, user.Password, user.AuthService, user.Username, user.Name, user.LastName, user.StorageLimit, tz, false, "", now, now, 0, user.AuthData)
	if err != nil {
		return nil, err
	}

	// Add user Clock Display to preferences
	ctx, cancel = context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	_, err = tx.ExecContext(ctx, "INSERT INTO preferences(user_id, category, name, value) VALUES(?,?,?,?);", userID, "display_settings", "clock_display", user.ClockDisplay)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:            userID,
		Email:         user.Email,
		Role:          user.Role,
		Password:      user.Password,
		AuthService:   user.Username,
		Username:      user.Username,
		Name:          user.Name,
		LastName:      user.LastName,
		StorageLimit:  user.StorageLimit,
		Photo:         user.Photo,
		DeactivatedAt: user.DeactivatedAt,
	}, nil
}

func (m *userRepository) Update(user model.UserPatch) error {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET email=?, username=?, role=?, name=?, lastname=?, storage_limit=?, updated_at=? WHERE id=?",
		user.Email, user.Username, user.Role, user.Name, user.LastName, user.StorageLimit, t, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *userRepository) Deactivate(id string, deactivatedAt int64) (int64, error) {
	now := time.Now().Unix()
	if deactivatedAt != 0 {
		now = 0
	}

	_, err := m.Db.Exec("UPDATE users SET deactivated_at=? WHERE id=?", now, id)
	if err != nil {
		return 0, err
	}

	return now, nil
}

func (m *userRepository) UpdateName(id string, name string, lastname string) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET name=?, lastname=?, updated_at=? WHERE id=?", name, lastname, t, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateTimezone(id string, timezone []byte) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET timezone=?, updated_at=? WHERE id=?", timezone, t, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateEmail(id string, email string) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET email=?, updated_at=? WHERE id=?", email, t, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateUsername(id string, username string) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET username=?, updated_at=? WHERE id=?", username, t, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateProfile(id string, email string, name string, lastname string) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec(
		"UPDATE users SET email=?, name=?, lastname=?, updated_at=? WHERE id=?",
		email, name, lastname, t, id,
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) GetByAuthService(ctx context.Context, authService string) ([]model.User, error) {
	q := `SELECT id, email, username, name, lastname, auth_service, auth_data, deactivated_at
	      FROM users
	      WHERE auth_service = ?`

	rows, err := m.Db.QueryContext(ctx, q, authService)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		u := model.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName,
			&u.AuthService, &u.AuthData, &u.DeactivatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (m *userRepository) UpdatePassword(id string, password string) (bool, error) {
	_, err := m.Db.Exec("UPDATE users SET password=? WHERE id=?", password, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateMfaSecret(user model.User, secret string) error {
	_, err := m.Db.Exec("UPDATE users SET mfa_secret=? WHERE id=?", secret, user.ID)
	if err != nil {
		return err
	}

	return err
}

func (m *userRepository) UpdateMfaActive(user model.User, status bool) error {
	_, err := m.Db.Exec("UPDATE users SET mfa_active=? WHERE id=?", status, user.ID)
	if err != nil {
		return err
	}

	return err
}

func (m *userRepository) UpdateAuthData(id string, authData string) (bool, error) {
	t := time.Now().Unix()
	_, err := m.Db.Exec("UPDATE users SET auth_data=?, updated_at=? WHERE id=?", authData, t, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *userRepository) UpdateStatus(userID string, status string, userDefined bool) (*model.UserStatus, error) {
	t := time.Now().Unix()

	query := `
		INSERT INTO status (user_id, status, last_activity, user_defined)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			last_activity = VALUES(last_activity),
			user_defined = VALUES(user_defined)
	`

	_, err := m.Db.Exec(query, userID, status, t, userDefined)
	if err != nil {
		return nil, err
	}

	return &model.UserStatus{
		UserID:       userID,
		Status:       status,
		LastActivity: t,
		UserDefined:  userDefined,
	}, nil
}

func (m *userRepository) GetStatus(userID string) (*model.UserStatus, error) {
	row := m.Db.QueryRow("SELECT "+statusColumns+" FROM status WHERE user_id=?", userID)

	status := model.UserStatus{}

	err := row.Scan(&status.UserID, &status.Status, &status.LastActivity, &status.UserDefined)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	if status.Status == "" {
		return nil, nil
	}

	return &status, nil
}

func (m *userRepository) GetStatuses() ([]model.UserStatus, error) {
	results, err := m.Db.Query("SELECT " + statusColumns + " FROM status")
	if err != nil {
		return nil, err
	}

	defer results.Close()

	statuses := make([]model.UserStatus, 0)
	for results.Next() {
		var status model.UserStatus

		err = results.Scan(&status.UserID, &status.Status, &status.LastActivity, &status.UserDefined)
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, status)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return statuses, nil
}
