// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

type groupRepository struct {
	Db *sql.DB
}

func NewGroupRepository(db *sql.DB) (*groupRepository, error) {
	return &groupRepository{Db: db}, nil
}

const groupColumns = `id, name, description, owner_id, created_at, updated_at, deleted_at, roles`

// Group roles are stored as a single space-separated string in the `roles`
// column, mirroring how user.Role and roles.permissions are stored. The role
// name regex forbids spaces, so the separator is unambiguous.
const groupRolesSeparator = " "

// memberCount counts one group's members through the unique_group_member
// prefix. A derived GROUP BY table would count every group in the install
// before the join kept the few that were asked for.
func memberCount(groupIDColumn string) string {
	return `(SELECT COUNT(*) FROM group_members gmc WHERE gmc.group_id = ` + groupIDColumn + `)`
}

// scanGroupWithCount scans a row that includes the standard group columns
// followed by member_count (BIGINT from COUNT(*)).
func scanGroupWithCount(scan func(dest ...any) error) (model.Group, error) {
	var group model.Group
	var description sql.NullString
	var updatedAt sql.NullInt64
	var roles sql.NullString
	var memberCount sql.NullInt64
	if err := scan(&group.ID, &group.Name, &description, &group.OwnerID,
		&group.CreatedAt, &updatedAt, &group.DeletedAt, &roles, &memberCount); err != nil {
		return group, err
	}

	group.Description = description.String
	group.UpdatedAt = updatedAt.Int64
	group.Roles = splitGroupRoles(roles.String)
	group.MemberCount = int(memberCount.Int64)
	return group, nil
}

// splitGroupRoles turns the stored space-separated role string into a slice,
// dropping empty entries so an empty/blank column yields an empty slice rather
// than [""]. strings.Fields handles any incidental extra whitespace.
func splitGroupRoles(s string) []string {
	return strings.Fields(s)
}

func (g *groupRepository) Create(ctx context.Context, ownerID, name, description string, roles []string) (*model.Group, error) {
	now := time.Now().Unix()
	group := &model.Group{
		ID:          model.NewID(),
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
		Roles:       roles,
	}

	_, err := g.Db.ExecContext(ctx,
		`INSERT INTO user_groups (id, name, description, owner_id, created_at, updated_at, deleted_at, roles)
		 VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
		group.ID, group.Name, group.Description, group.OwnerID, group.CreatedAt, group.UpdatedAt,
		strings.Join(roles, groupRolesSeparator),
	)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (g *groupRepository) Get(ctx context.Context, id string) (*model.Group, error) {
	group, err := scanGroupWithCount(g.Db.QueryRowContext(ctx,
		`SELECT `+groupColumnsPrefixed("ug")+`, `+memberCount("ug.id")+`
		 FROM user_groups ug
		 WHERE ug.id = ? AND ug.deleted_at = 0`,
		id,
	).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &group, nil
}

// GetByIDs returns the non-deleted groups matching the given IDs,
// enriched with member counts. Groups that don't exist (or are soft-deleted)
// are simply absent from the result; the caller compares the result length
// against the input to detect missing IDs. Duplicate input IDs collapse to a
// single row, so callers should dedupe before comparing lengths.
func (g *groupRepository) GetByIDs(ctx context.Context, ids []string) ([]model.Group, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := g.Db.QueryContext(ctx,
		`SELECT `+groupColumnsPrefixed("ug")+`, `+memberCount("ug.id")+`
		 FROM user_groups ug
		 WHERE ug.id IN (`+strings.Join(placeholders, ",")+`) AND ug.deleted_at = 0
		 ORDER BY ug.name ASC`,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	groups := make([]model.Group, 0, len(ids))
	for rows.Next() {
		group, err := scanGroupWithCount(rows.Scan)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, rows.Err()
}

var groupSortColumns = map[string][]string{
	"name":         {"ug.name"},
	"member_count": {memberCount("ug.id")},
	"created_at":   {"ug.created_at"},
}

func (g *groupRepository) GetAllPaged(ctx context.Context, query string, sort model.Sort, limit, offset int) ([]model.Group, error) {
	q := `SELECT ` + groupColumnsPrefixed("ug") + `, ` + memberCount("ug.id") + `
	      FROM user_groups ug
	      WHERE ug.deleted_at = 0`
	args := []any{}

	filter, filterArgs := groupSearchFilter(query)
	q += filter
	args = append(args, filterArgs...)

	q += orderBy(sort, groupSortColumns, "name", "ug.id") + ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := g.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	groups := make([]model.Group, 0, limit)
	for rows.Next() {
		group, err := scanGroupWithCount(rows.Scan)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (g *groupRepository) CountAll(ctx context.Context, query string) (int, error) {
	q := `SELECT COUNT(*) FROM user_groups ug WHERE ug.deleted_at = 0`

	filter, args := groupSearchFilter(query)
	q += filter

	var count int
	err := g.Db.QueryRowContext(ctx, q, args...).Scan(&count)

	return count, err
}

func (g *groupRepository) SearchForUser(ctx context.Context, userID, query string, limit int) ([]model.Group, error) {
	q := `SELECT ` + groupColumnsPrefixed("ug") + `, ` + memberCount("ug.id") + `
	      FROM user_groups ug
	      WHERE ug.deleted_at = 0
	        AND (ug.owner_id = ?
	             OR EXISTS (SELECT 1 FROM group_members gm WHERE gm.group_id = ug.id AND gm.user_id = ?))`
	args := []any{userID, userID}

	filter, filterArgs := groupSearchFilter(query)
	q += filter
	args = append(args, filterArgs...)

	q += ` ORDER BY ug.name ASC, ug.id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := g.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	groups := make([]model.Group, 0, limit)
	for rows.Next() {
		group, err := scanGroupWithCount(rows.Scan)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func groupSearchFilter(query string) (string, []any) {
	if query == "" {
		return "", nil
	}

	like := "%" + escapeLike(query) + "%"

	return ` AND (ug.name LIKE ? OR ug.description LIKE ?)`, []any{like, like}
}

func (g *groupRepository) Update(ctx context.Context, id, name, description string, roles []string) error {
	res, err := g.Db.ExecContext(ctx,
		`UPDATE user_groups SET name = ?, description = ?, roles = ?, updated_at = ?
		 WHERE id = ? AND deleted_at = 0`,
		name, description, strings.Join(roles, groupRolesSeparator), time.Now().Unix(), id,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SoftDelete flips deleted_at on user_groups and hard-deletes the matching
// group_members rows in the same transaction. Membership doesn't survive group
// deletion, and leaving the rows around would leak past-membership through any
// join that forgets to filter on user_groups.deleted_at = 0.
func (g *groupRepository) SoftDelete(ctx context.Context, id string) error {
	tx, err := g.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE user_groups SET deleted_at = ? WHERE id = ? AND deleted_at = 0`,
		time.Now().Unix(), id,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = ?`, id,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (g *groupRepository) AddMembers(ctx context.Context, groupID string, userIDs []string, role string) error {
	if len(userIDs) == 0 {
		return nil
	}

	if role == "" {
		role = model.GroupRoleMember
	}

	now := time.Now().Unix()

	args := make([]any, 0, len(userIDs)*5)
	for _, userID := range userIDs {
		args = append(args, model.NewID(), groupID, userID, role, now)
	}

	_, err := g.Db.ExecContext(ctx,
		`INSERT INTO group_members (id, group_id, user_id, role, joined_at)
		 VALUES `+rowPlaceholders(len(userIDs), 5)+`
		 ON DUPLICATE KEY UPDATE role = VALUES(role)`, args...)

	return err
}

func (g *groupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	_, err := g.Db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID,
	)
	return err
}

func (g *groupRepository) RemoveAllMembershipsForUser(ctx context.Context, userID string) error {
	_, err := g.Db.ExecContext(ctx,
		`DELETE FROM group_members WHERE user_id = ?`, userID,
	)
	return err
}

func (g *groupRepository) GetMembers(ctx context.Context, groupID string) ([]model.GroupMember, error) {
	rows, err := g.Db.QueryContext(ctx,
		`SELECT id, group_id, user_id, role, joined_at
		 FROM group_members WHERE group_id = ? ORDER BY joined_at ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]model.GroupMember, 0)
	for rows.Next() {
		var m model.GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}

		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

var groupMemberSortColumns = map[string][]string{
	"name":      {"u.name", "u.lastname"},
	"email":     {"u.email"},
	"joined_at": {"gm.joined_at"},
}

// GetMembersPaged returns a page of a group's members hydrated with user
// info. When query is non-empty, results are filtered by name/lastname/email/
// username (LIKE, escaped). Ties in the requested sort fall back to the member
// id so pages stay stable.
func (g *groupRepository) GetMembersPaged(ctx context.Context, groupID, query string, sort model.Sort, limit, offset int) ([]model.GroupMember, error) {
	q := `SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.joined_at,
	             u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
	      FROM group_members gm
	      JOIN users u ON u.id = gm.user_id
	      WHERE gm.group_id = ?`
	args := []any{groupID}

	filter, filterArgs := memberSearchFilter(query)
	q += filter
	args = append(args, filterArgs...)

	q += orderBy(sort, groupMemberSortColumns, "name", "gm.id") + ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := g.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]model.GroupMember, 0, limit)
	for rows.Next() {
		var m model.GroupMember
		var u model.User
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.JoinedAt,
			&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		m.UserInfo = &u
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (g *groupRepository) CountMembers(ctx context.Context, groupID, query string) (int, error) {
	q := `SELECT COUNT(*)
	      FROM group_members gm
	      JOIN users u ON u.id = gm.user_id
	      WHERE gm.group_id = ?`
	args := []any{groupID}

	filter, filterArgs := memberSearchFilter(query)
	q += filter
	args = append(args, filterArgs...)

	var count int
	err := g.Db.QueryRowContext(ctx, q, args...).Scan(&count)

	return count, err
}

func memberSearchFilter(query string) (string, []any) {
	if query == "" {
		return "", nil
	}

	like := "%" + escapeLike(query) + "%"

	return ` AND (u.name LIKE ? OR u.lastname LIKE ? OR u.email LIKE ? OR u.username LIKE ?)`,
		[]any{like, like, like, like}
}

// GetSharesForFile returns the groups a file is shared with, each
// hydrated with member_count and the permissions from the file_share row.
func (g *groupRepository) GetSharesForFile(ctx context.Context, fileID string) ([]model.SharedGroup, error) {
	rows, err := g.Db.QueryContext(ctx,
		`WITH RECURSIVE anc AS (
			SELECT id, parent, displayname, 0 AS depth FROM files WHERE id = ?
			UNION ALL
			SELECT f.id, f.parent, f.displayname, a.depth + 1
			FROM files f JOIN anc a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		 )
		 SELECT
			ug.id, ug.name, ug.description,
			`+memberCount("ug.id")+` AS member_count,
			fs.expiration, fs.access_level,
			anc.id, anc.displayname
		 FROM anc
		 JOIN file_share fs ON fs.file_id = anc.id AND fs.share_type = ?
		 JOIN user_groups ug ON ug.id = fs.share_with AND ug.deleted_at = 0
		 ORDER BY ug.id, anc.depth`,
		fileID, model.SHARE_TYPE_GROUP,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	shares := make([]model.SharedGroup, 0)
	seen := make(map[string]bool)
	for rows.Next() {
		var sg model.SharedGroup
		var description sql.NullString
		var grantedByID, grantedByName string
		if err := rows.Scan(&sg.ID, &sg.Name, &description, &sg.MemberCount,
			&sg.Expiration, &sg.AccessLevel,
			&grantedByID, &grantedByName); err != nil {
			return nil, err
		}

		if seen[sg.ID] {
			continue
		}

		seen[sg.ID] = true

		sg.Description = description.String
		if grantedByID != fileID {
			sg.Inherited = true
			sg.GrantedBy = grantedByName
		}

		shares = append(shares, sg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shares, nil
}

func (g *groupRepository) GetIDsForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := g.Db.QueryContext(ctx,
		`SELECT gm.group_id
		 FROM group_members gm
		 JOIN user_groups ug ON ug.id = gm.group_id
		 WHERE gm.user_id = ? AND ug.deleted_at = 0`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// GetRolesForUser returns the distinct system role names granted to the
// user through every non-deleted group they are a member of. The owner of a
// group is not granted its roles implicitly; only actual group_members rows
// confer membership, matching GetIDsForUser. Returns an empty slice when the
// user belongs to no role-granting groups.
func (g *groupRepository) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := g.Db.QueryContext(ctx,
		`SELECT ug.roles
		 FROM group_members gm
		 JOIN user_groups ug ON ug.id = gm.group_id
		 WHERE gm.user_id = ? AND ug.deleted_at = 0
		   AND ug.roles IS NOT NULL AND ug.roles != ''`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	seen := make(map[string]struct{})
	roles := make([]string, 0)
	for rows.Next() {
		var rolesStr string
		if err := rows.Scan(&rolesStr); err != nil {
			return nil, err
		}

		for _, name := range splitGroupRoles(rolesStr) {
			if _, ok := seen[name]; ok {
				continue
			}

			seen[name] = struct{}{}
			roles = append(roles, name)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func groupColumnsPrefixed(prefix string) string {
	parts := strings.Split(groupColumns, ", ")
	for i, p := range parts {
		parts[i] = prefix + "." + p
	}

	return strings.Join(parts, ", ")
}
