// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/twigex/twigex/model"
)

const (
	fileMetadataColumns = `id, type, title, fields, created_at, updated_at, created_by`

	fileMetadataEntryColumns = `id, file_id, metadata_id, value, created_at, updated_at`
)

var fileColumnNames = []string{
	"id", "owner", "parent", "storage", "name", "displayname", "size", "type",
	"is_folder", "shared", "created_at", "modified_at", "deleted_at",
	"lock_owner", "lock_expires_at", "version",
}

// A file inside a shared folder is deliberately not marked: only a share of
// its own counts, so this needs no ancestor walk.
func sharedExpr(idRef string) string {
	return `(EXISTS (SELECT 1 FROM file_share sh_fs
	                  WHERE sh_fs.file_id = ` + idRef + `
	                    AND (sh_fs.expiration = 0 OR sh_fs.expiration > UNIX_TIMESTAMP()))
	          OR EXISTS (SELECT 1 FROM external_share sh_es
	                  WHERE sh_es.file_id = ` + idRef + ` AND sh_es.active = 1)) AS shared`
}

func fileColumns(table string) string {
	prefix := ""
	if table != "" {
		prefix = table + "."
	}

	// Qualified deliberately: file_share has an id of its own, and a bare one
	// resolves to that instead.
	idRef := prefix + "id"
	if table == "" {
		idRef = "files.id"
	}

	out := make([]string, len(fileColumnNames))
	for i, c := range fileColumnNames {
		if c == "shared" {
			out[i] = sharedExpr(idRef)
			continue
		}

		out[i] = prefix + c
	}

	return strings.Join(out, ", ")
}

// For a CTE or derived table that already projected these, so shared is not
// computed a second time over the same rows.
func fileColumnsFrom(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	out := make([]string, len(fileColumnNames))
	for i, c := range fileColumnNames {
		out[i] = prefix + c
	}

	return strings.Join(out, ", ")
}

type fileRepository struct {
	Db *sql.DB
}

func NewFileRepository(Db *sql.DB) (*fileRepository, error) {
	repo := &fileRepository{}

	repo.Db = Db
	return repo, nil
}

func scanFiles(results *sql.Rows) ([]model.File, error) {
	dbFiles := make([]model.File, 0)
	file := model.File{}
	for results.Next() {
		err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt, &file.LockOwner,
			&file.LockExpiresAt, &file.Version)
		if err != nil {
			return nil, err
		}

		dbFiles = append(dbFiles, file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return dbFiles, nil
}

// mostPermissiveExpiration combines two grant expirations: 0 means "never" and
// always wins; otherwise the later timestamp (the least-restrictive) wins.
func mostPermissiveExpiration(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}

	if b > a {
		return b
	}

	return a
}

// sharedFileSet collects files while collapsing a file reached through several
// grants (direct plus one or more groups) into a single entry with the most-
// permissive access level and expiration.
type sharedFileSet struct {
	files []model.File
	pos   map[string]int
}

func newSharedFileSet() *sharedFileSet {
	// files is a non-nil empty slice so an empty result marshals as [] not null.
	return &sharedFileSet{files: make([]model.File, 0), pos: make(map[string]int)}
}

func (s *sharedFileSet) add(f model.File) {
	if i, ok := s.pos[f.ID]; ok {
		if f.AccessLevel > s.files[i].AccessLevel {
			s.files[i].AccessLevel = f.AccessLevel
		}

		s.files[i].Expiration = mostPermissiveExpiration(s.files[i].Expiration, f.Expiration)
		return
	}

	s.pos[f.ID] = len(s.files)
	s.files = append(s.files, f)
}

// reachableShareRootIDs returns the file IDs the user has a live (non-expired)
// share on, directly or via a group. These are the entry points from which
// inherited access flows down; callers expand them into subtrees. It is a plain
// query (not a CTE) because MariaDB rejects the effectiveShareSubquery derived
// table inside a recursive CTE's anchor member.
func (m *fileRepository) reachableShareRootIDs(ctx context.Context, userID string) ([]string, error) {
	shareJoin, shareArgs := effectiveShareSubquery(userID)
	args := append(shareArgs, time.Now().Unix())

	rows, err := m.Db.QueryContext(ctx, `
		SELECT DISTINCT files.id
		FROM files
		JOIN `+shareJoin+` ON efs.file_id = files.id
		WHERE files.deleted_at = 0 AND (efs.expiration = 0 OR efs.expiration > ?)`, args...)
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

	return ids, rows.Err()
}

func (m *fileRepository) ancestorIDs(ctx context.Context, fileID string) ([]string, error) {
	rows, err := m.Db.QueryContext(ctx, `
		WITH RECURSIVE anc AS (
			SELECT id, parent, 0 AS depth FROM files WHERE id = ?
			UNION ALL
			SELECT f.id, f.parent, a.depth + 1
			FROM files f JOIN anc a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		)
		SELECT id FROM anc`, fileID)
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

	return ids, rows.Err()
}

// GetGrantingRoot returns the nearest ancestor-or-self of fileID that holds a
// share for (shareType, shareWith). Inherited descendants carry no share row of
// their own, so a revoke or permission edit targets this node (Drive escalates
// the change to the parent folder). Empty string if none grants that principal.
func (m *fileRepository) GetGrantingRoot(ctx context.Context, fileID string, shareType int, shareWith string) (string, error) {
	row := m.Db.QueryRowContext(ctx, `
		WITH RECURSIVE anc AS (
			SELECT id, parent, 0 AS depth FROM files WHERE id = ?
			UNION ALL
			SELECT f.id, f.parent, a.depth + 1
			FROM files f JOIN anc a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		)
		SELECT anc.id
		FROM anc
		JOIN file_share fs ON fs.file_id = anc.id AND fs.share_type = ? AND fs.share_with = ?
		ORDER BY anc.depth
		LIMIT 1`, fileID, shareType, shareWith)

	var rootID string
	err := row.Scan(&rootID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return rootID, nil
}

// InheritedAccessLevel returns the highest access level the principal already
// has from a strict ancestor of fileID (0 if none). Used to reject sharing a
// node below what the principal inherits from above, which would be a no-op
// under most-permissive resolution.
func (m *fileRepository) InheritedAccessLevel(ctx context.Context, fileID string, shareType int, shareWith string) (model.AccessLevel, error) {
	row := m.Db.QueryRowContext(ctx, `
		WITH RECURSIVE anc AS (
			SELECT id, parent, 0 AS depth FROM files WHERE id = ?
			UNION ALL
			SELECT f.id, f.parent, a.depth + 1
			FROM files f JOIN anc a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		)
		SELECT COALESCE(MAX(fs.access_level), 0)
		FROM anc
		JOIN file_share fs ON fs.file_id = anc.id
		 AND fs.share_type = ? AND fs.share_with = ?
		WHERE anc.depth > 0
		 AND (fs.expiration = 0 OR fs.expiration > ?)`,
		fileID, shareType, shareWith, time.Now().Unix())

	var lvl int
	if err := row.Scan(&lvl); err != nil {
		return 0, err
	}

	return model.AccessLevel(lvl), nil
}

// MoveAllowedWithinShare reports whether a non-owner may move fileID under
// targetID: true when a share root they can reach is a common ancestor (or
// self) of both, so the file stays inside the shared subtree it draws access from.
func (m *fileRepository) MoveAllowedWithinShare(ctx context.Context, userID string, fileID string, targetID string) (bool, error) {
	roots, err := m.reachableShareRootIDs(ctx, userID)
	if err != nil {
		return false, err
	}

	if len(roots) == 0 {
		return false, nil
	}

	rootSet := make(map[string]bool, len(roots))
	for _, r := range roots {
		rootSet[r] = true
	}

	targetAnc, err := m.ancestorIDs(ctx, targetID)
	if err != nil {
		return false, err
	}

	targetSet := make(map[string]bool, len(targetAnc))
	for _, t := range targetAnc {
		targetSet[t] = true
	}

	fileAnc, err := m.ancestorIDs(ctx, fileID)
	if err != nil {
		return false, err
	}

	for _, a := range fileAnc {
		if rootSet[a] && targetSet[a] {
			return true, nil
		}
	}

	return false, nil
}

// GetRecents returns the 50 most recently modified files the user can access:
// their owned files, plus each share root itself and any file added to a shared
// subtree at or after it was shared. Pre-existing contents of a shared folder do
// not flood recents; only the root and newly added files surface.
func (m *fileRepository) GetRecents(ctx context.Context, userID string) ([]model.File, error) {
	shareJoin, shareArgs := effectiveShareSubquery(userID)
	args := append(shareArgs, time.Now().Unix(), userID)

	results, err := m.Db.QueryContext(ctx, `
		WITH RECURSIVE subtree (id, shared_at, depth) AS (
			SELECT files.id, efs.time_shared, 0
			FROM files
			JOIN `+shareJoin+` ON efs.file_id = files.id
			WHERE files.deleted_at = 0
			  AND (efs.expiration = 0 OR efs.expiration > ?)

			UNION ALL

			SELECT f.id, s.shared_at, s.depth + 1
			FROM files f
			JOIN subtree s ON f.parent = s.id
			WHERE f.deleted_at = 0 AND s.depth < 100
		)
		SELECT `+fileColumnsFrom("")+` FROM (
			SELECT `+fileColumns("files")+`
			FROM files
			WHERE files.type <> 'cloud#drive' AND files.deleted_at = 0
			  AND files.owner = ?

			UNION

			SELECT `+fileColumns("files")+`
			FROM files
			JOIN subtree ON subtree.id = files.id
			WHERE files.type <> 'cloud#drive' AND files.deleted_at = 0
			  AND (subtree.depth = 0 OR files.created_at >= subtree.shared_at)
		) recents
		ORDER BY modified_at DESC
		LIMIT 50`, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	dbFiles, err := scanFiles(results)
	if err != nil {
		return nil, err
	}

	return dbFiles, nil
}

// GetBreadcrumbs returns the accessible ancestor chain for fileID, ordered from
// the file up to its access origin (exclusive of the storage root, which the
// caller appends). It walks the physical ancestor chain (stopping at a deleted
// ancestor) and keeps the contiguous prefix the user can actually see: a node
// is kept if the user owns it, or it sits at or below the highest reachable
// share root on the chain (inherited access). Ancestors above the share root
// (folders the user was never shared) are excluded so their names don't leak.
func (m *fileRepository) GetBreadcrumbs(ctx context.Context, userID string, fileID string) ([]model.File, error) {
	rows, err := m.Db.QueryContext(ctx, `
		WITH RECURSIVE chain AS (
			SELECT `+fileColumns("files")+`, 0 AS depth
			FROM files
			WHERE id = ?

			UNION ALL

			SELECT `+fileColumns("f")+`, c.depth + 1
			FROM files f
			JOIN chain c ON f.id = c.parent
			WHERE f.deleted_at = 0 AND c.depth < 100
		)
		SELECT `+fileColumnsFrom("")+`
		FROM chain
		ORDER BY depth`, fileID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	chain, err := scanFiles(rows)
	if err != nil {
		return nil, err
	}

	roots, err := m.reachableShareRootIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	rootSet := make(map[string]bool, len(roots))
	for _, id := range roots {
		rootSet[id] = true
	}

	// Highest position in the chain (furthest from the file) carrying a
	// reachable share; everything at or below it is inherited-accessible.
	maxRootIdx := -1
	for i := range chain {
		if rootSet[chain[i].ID] {
			maxRootIdx = i
		}
	}

	crumbs := make([]model.File, 0, len(chain))
	for i := range chain {
		if chain[i].Owner == userID || (maxRootIdx >= 0 && i <= maxRootIdx) {
			crumbs = append(crumbs, chain[i])
			continue
		}

		break
	}

	return crumbs, nil
}

func (m *fileRepository) GetByIDs(user model.User, IDS []string) ([]model.File, error) {
	if len(IDS) == 0 {
		return make([]model.File, 0), nil
	}

	args := make([]any, len(IDS))
	for i, id := range IDS {
		args[i] = id
	}

	file := model.File{}
	now := time.Now().Unix()
	set := newSharedFileSet()

	stmt := fmt.Sprintf(`SELECT `+fileColumns("files")+`
						FROM files
						WHERE files.id IN (?`+strings.Repeat(",?", len(args)-1)+`)
						AND files.owner="%s"
						AND files.type<>"cloud#drive"
						AND files.deleted_at=%d`, user.ID, 0)
	results, err := m.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()
	for results.Next() {
		err = results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt,
			&file.LockOwner, &file.LockExpiresAt, &file.Version)
		if err != nil {
			return nil, err
		}

		set.add(file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	shareJoin, shareArgs := effectiveShareSubquery(user.ID)
	stmt = `SELECT ` + fileColumns("files") + `, efs.expiration, efs.access_level
						FROM files
						JOIN ` + shareJoin + ` ON efs.file_id = files.id
						WHERE files.id IN (?` + strings.Repeat(",?", len(args)-1) + `)
						AND files.deleted_at = 0`

	queryArgs := append(shareArgs, args...)
	results, err = m.Db.Query(stmt, queryArgs...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	for results.Next() {
		err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt, &file.LockOwner, &file.LockExpiresAt, &file.Version,
			&file.Expiration, &file.AccessLevel)
		if err != nil {
			return nil, err
		}

		if file.Expiration != 0 && now >= file.Expiration {
			continue
		}

		set.add(file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return set.files, nil
}

func (m *fileRepository) Get(id string) (*model.File, error) { //*files.File
	results, err := m.Db.Query("SELECT "+fileColumns("")+" FROM files WHERE id=?", id)
	if err != nil {
		return &model.File{}, err
	}

	defer results.Close()

	file := model.File{}
	for results.Next() {
		err = results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt,
			&file.LockOwner, &file.LockExpiresAt, &file.Version)
		if err != nil {
			return &model.File{}, err
		}
	}

	if err := results.Err(); err != nil {
		return &model.File{}, err
	}

	if file.ID == "" {
		return nil, nil
	}

	return &file, nil
}

// ResolveEffectiveShare computes a user's effective permission for one file
// under inherited sharing: it walks the file and its ancestors and aggregates
// the most-permissive grant reachable via any share on that chain, returning 0
// or 1 rows to match GetFilePermissions' shape. It replaces the per-file lookup
// once shares live only on the share root instead of being fanned out.
//
// The base row is included regardless of deleted_at (parity with
// GetFilePermissions for a file's own grant), but the recursion only ascends
// through live ancestors: a deleted intermediate folder blocks inheritance, so
// a descendant can never resolve access through a trashed parent.
func (m *fileRepository) ResolveEffectiveShare(ctx context.Context, userID string, fileID string) ([]model.File, error) {
	shareJoin, shareArgs := effectiveShareSubquery(userID)
	args := append([]any{fileID}, shareArgs...)

	results, err := m.Db.QueryContext(ctx, `
		WITH RECURSIVE ancestors (id, parent, depth) AS (
			SELECT id, parent, 0
			FROM files
			WHERE id = ? AND deleted_at = 0

			UNION ALL

			SELECT f.id, f.parent, a.depth + 1
			FROM files f
			JOIN ancestors a ON f.id = a.parent
			WHERE f.deleted_at = 0 AND a.depth < 100
		)
		SELECT
			CASE WHEN MIN(efs.expiration) = 0 THEN 0 ELSE MAX(efs.expiration) END,
			MAX(efs.access_level)
		FROM ancestors
		JOIN `+shareJoin+` ON efs.file_id = ancestors.id
		HAVING COUNT(efs.file_id) > 0`, args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	now := time.Now().Unix()
	permissions := make([]model.File, 0)
	file := model.File{}
	for results.Next() {
		if err := results.Scan(&file.Expiration, &file.AccessLevel); err != nil {
			return nil, err
		}

		if now > file.Expiration && file.Expiration != 0 {
			continue
		}

		permissions = append(permissions, file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m *fileRepository) GetByParent(parent model.File, userID string) ([]model.File, error) {
	results, err := m.Db.Query(`SELECT `+fileColumns("files")+`
								FROM files
								WHERE files.parent=?
								AND files.owner=?
								AND files.type<>?
								AND files.deleted_at=?`, parent.ID, userID, "cloud#drive", 0)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	file := model.File{}
	set := newSharedFileSet()
	now := time.Now().Unix()

	for results.Next() {
		err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt,
			&file.LockOwner, &file.LockExpiresAt, &file.Version)
		if err != nil {
			return nil, err
		}

		set.add(file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	if parent.Type != "cloud#drive" {
		shareJoin, shareArgs := effectiveShareSubquery(userID)
		queryArgs := append(shareArgs, parent.ID)
		results, err = m.Db.Query(`SELECT `+fileColumns("files")+`, efs.expiration, efs.access_level
		FROM files
		JOIN `+shareJoin+` ON efs.file_id = files.id
		WHERE files.parent = ?
		AND files.deleted_at = 0`, queryArgs...)
		if err != nil {
			return nil, err
		}

		for results.Next() {
			err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
				&file.Name, &file.DisplayName, &file.Size,
				&file.Type, &file.IsFolder, &file.Shared,
				&file.Created, &file.Modified, &file.DeletedAt, &file.LockOwner, &file.LockExpiresAt, &file.Version,
				&file.Expiration, &file.AccessLevel)
			if err != nil {
				return nil, err
			}

			if file.Expiration != 0 && now >= file.Expiration {
				continue
			}

			set.add(file)
		}

		if err := results.Err(); err != nil {
			return nil, err
		}
	}

	return set.files, nil
}

// GetChildren returns every live, non-drive file directly under folderID.
// Access is decided at the folder: once a user can see a folder, they can see
// all of its children.
func (m *fileRepository) GetChildren(ctx context.Context, folderID string) ([]model.File, error) {
	results, err := m.Db.QueryContext(ctx, `SELECT `+fileColumns("files")+`
		FROM files
		WHERE files.parent = ?
		  AND files.type <> ?
		  AND files.deleted_at = 0`, folderID, "cloud#drive")
	if err != nil {
		return nil, err
	}

	defer results.Close()

	dbFiles, err := scanFiles(results)
	if err != nil {
		return nil, err
	}

	return dbFiles, nil
}

func (m *fileRepository) GetSubFilesByFolder(folderID string, storage string) ([]model.File, error) {
	query := `
		WITH RECURSIVE tree AS (
			SELECT ` + fileColumns("") + `
			FROM files
			WHERE id = ? AND storage = ?

			UNION ALL

			SELECT ` + fileColumns("f") + `
			FROM files f
			INNER JOIN tree t ON f.parent = t.id
			WHERE f.storage = ?
		)
		SELECT ` + fileColumnsFrom("") + ` FROM tree
	`

	rows, err := m.Db.Query(query, folderID, storage, storage)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files, err := scanFiles(rows)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (m *fileRepository) CanMoveFolder(id string, targetID string) (bool, error) {
	query := `
        WITH RECURSIVE folder_hierarchy AS (
            SELECT id
            FROM files
            WHERE id = ?
            UNION ALL
            SELECT f.id
            FROM files f
            INNER JOIN folder_hierarchy fh ON f.parent = fh.id
        )
        SELECT 1
        FROM folder_hierarchy
        WHERE id = ?;
    `

	var result int
	err := m.Db.QueryRow(query, id, targetID).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			// No match found, move is allowed
			return true, nil
		}

		return false, err
	}

	return false, nil
}

func (m *fileRepository) Create(file model.File) (*string, error) { // owner string, Parent string, storage string, path string, size int64, name string, displayname string, doctype string, shared bool
	created := time.Now().Unix()

	_, err := m.Db.Exec("INSERT INTO files(id, owner, parent, storage, name, displayname, size, type, is_folder, created_at, modified_at) VALUES(?,?,?,?,?,?,?,?,'0',?,?);", file.ID, file.Owner, file.Parent, file.Storage, file.Name, file.DisplayName, file.Size, file.Type, created, created)
	if err != nil {
		return nil, err
	}

	return &file.ID, nil
}

func (m *fileRepository) CreateFolder(file model.File) (*string, error) {
	created := time.Now().Unix()

	_, err := m.Db.Exec("INSERT INTO files(id, owner, parent, storage, name, displayname, size, type, is_folder, created_at, modified_at) VALUES(?,?,?,?,?,?,?,?,'1',?,?);", file.ID, file.Owner, file.Parent, file.Storage, file.Name, file.DisplayName, file.Size, "inode/directory", created, created)
	if err != nil {
		return nil, err
	}

	return &file.ID, nil
}

func (m *fileRepository) AddFavourite(userID string, fileID string) error {
	id := model.NewID()

	_, err := m.Db.Exec("INSERT INTO favorites(id, user_id, app, item_id) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE id = id", id, userID, "files", fileID)
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) RemoveFavourite(userID string, fileID string) error {
	_, err := m.Db.Exec("DELETE FROM favorites WHERE user_id=? AND item_id=? AND app=?", userID, fileID, "files")
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) GetFavourite(userID string, fileID string) (string, error) {
	var id string
	err := m.Db.QueryRow("SELECT id FROM favorites WHERE user_id=? AND item_id=? AND app=?",
		userID, fileID, "files").Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return id, nil
}

// GetFavourites returns the user's favourited files that they can access: owned
// files plus any favourited file within a share-root subtree the user can reach.
func (m *fileRepository) GetFavourites(ctx context.Context, userID string) ([]model.File, error) {
	rootIDs, err := m.reachableShareRootIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results *sql.Rows
	if len(rootIDs) == 0 {
		results, err = m.Db.QueryContext(ctx, `
			SELECT `+fileColumns("files")+`
			FROM files
			JOIN favorites ON favorites.item_id = files.id
			WHERE favorites.user_id = ? AND favorites.app = 'files'
			  AND files.type <> 'cloud#drive' AND files.deleted_at = 0
			  AND files.owner = ?`, userID, userID)
	} else {
		inClause := "?" + strings.Repeat(",?", len(rootIDs)-1)
		args := make([]any, 0, len(rootIDs)+2)
		for _, id := range rootIDs {
			args = append(args, id)
		}

		args = append(args, userID, userID)

		results, err = m.Db.QueryContext(ctx, `
			WITH RECURSIVE subtree AS (
				SELECT id FROM files WHERE id IN (`+inClause+`)
				UNION ALL
				SELECT f.id FROM files f JOIN subtree s ON f.parent = s.id WHERE f.deleted_at = 0
			)
			SELECT `+fileColumns("files")+`
			FROM files
			JOIN favorites ON favorites.item_id = files.id
			WHERE favorites.user_id = ?
			  AND favorites.app = 'files'
			  AND files.type <> 'cloud#drive'
			  AND files.deleted_at = 0
			  AND (files.owner = ? OR files.id IN (SELECT id FROM subtree))`, args...)
	}

	if err != nil {
		return nil, err
	}

	defer results.Close()

	dbFiles, err := scanFiles(results)
	if err != nil {
		return nil, err
	}

	return dbFiles, nil
}

func (m *fileRepository) UpdateParentID(
	rootIDS []string,
	subfileIDS []string,
	targetParentID string,
	storage string,
) error {
	if len(rootIDS) == 0 && len(subfileIDS) == 0 {
		return nil
	}

	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	now := time.Now().Unix()

	if len(rootIDS) > 0 {
		ph := make([]string, len(rootIDS))
		args := make([]any, 0, 2+len(rootIDS))

		args = append(args, targetParentID, storage)

		for i, id := range rootIDS {
			ph[i] = "?"
			args = append(args, id)
		}

		stmt := fmt.Sprintf(
			`UPDATE files
			 SET parent = ?
			 WHERE storage = ?
			   AND id IN (%s)
			   AND (
			        lock_owner IS NULL
			        OR lock_expires_at < ?
			   )`,
			strings.Join(ph, ","),
		)

		args = append(args, now)

		res, err := tx.Exec(stmt, args...)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows != int64(len(rootIDS)) {
			return errors.New("one or more files are locked")
		}
	}

	return tx.Commit()
}

func (m *fileRepository) Delete(id string, storage string) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		return false, err
	}

	defer tx.Rollback()

	now := time.Now().Unix()

	res, err := tx.Exec(
		`DELETE FROM files
		 WHERE id = ?
		   AND storage = ?
		   AND (
		        lock_owner IS NULL
		        OR lock_expires_at < ?
		   )`,
		id,
		storage,
		now,
	)
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rows == 0 {
		return false, errors.New("file is locked or not found")
	}

	if _, err := tx.Exec(
		`DELETE FROM file_trash WHERE file_id = ?`,
		id,
	); err != nil {
		return false, err
	}

	if _, err := tx.Exec(
		`DELETE FROM file_share WHERE file_id = ?`,
		id,
	); err != nil {
		return false, err
	}

	if _, err := tx.Exec(
		`DELETE FROM favorites WHERE item_id = ? AND app = ?`,
		id,
		"files",
	); err != nil {
		return false, err
	}

	if _, err := tx.Exec(
		`DELETE FROM external_share WHERE file_id = ?`,
		id,
	); err != nil {
		return false, err
	}

	if _, err := tx.Exec(
		`DELETE FROM file_metadata_entries WHERE file_id = ?`,
		id,
	); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (m *fileRepository) GetDeleted(userID string) ([]model.File, error) {
	results, err := m.Db.Query(`SELECT `+fileColumns("files")+`
								FROM files, file_trash
								WHERE files.id = file_trash.file_id
								AND files.owner=?
								AND files.type<>?`, userID, "cloud#drive")
	if err != nil {
		return nil, err
	}

	defer results.Close()

	dbFiles := make([]model.File, 0)
	file := model.File{}
	for results.Next() {
		err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt,
			&file.LockOwner, &file.LockExpiresAt, &file.Version)
		if err != nil {
			return nil, err
		}

		dbFiles = append(dbFiles, file)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return dbFiles, nil
}

func (m *fileRepository) Restore(userID string, rootIDs []string) error {
	if len(rootIDs) == 0 {
		return nil
	}

	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	inRoots := "(?" + strings.Repeat(",?", len(rootIDs)-1) + ")"
	args := make([]any, 0, len(rootIDs)+1)
	for _, id := range rootIDs {
		args = append(args, id)
	}

	args = append(args, userID)

	rows, err := tx.Query(
		`SELECT id, is_folder
		 FROM files
		 WHERE id IN `+inRoots+`
		   AND owner = ?
		   AND deleted_at != 0`,
		args...,
	)
	if err != nil {
		return err
	}

	defer rows.Close()

	fileIDs := make([]string, 0)
	folderIDs := make([]string, 0)

	for rows.Next() {
		var id string
		var isFolder bool
		if err := rows.Scan(&id, &isFolder); err != nil {
			return err
		}

		if isFolder {
			folderIDs = append(folderIDs, id)
		} else {
			fileIDs = append(fileIDs, id)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Search subfiles in folder tree
	treeIDs := make([]string, 0)

	if len(folderIDs) > 0 {
		inFolders := "(?" + strings.Repeat(",?", len(folderIDs)-1) + ")"
		folderArgs := make([]any, 0, len(folderIDs)+1)
		for _, id := range folderIDs {
			folderArgs = append(folderArgs, id)
		}

		folderArgs = append(folderArgs, userID)

		query := `
			WITH RECURSIVE tree (id, parent, depth) AS (
				SELECT id, parent, 0
				FROM files
				WHERE id IN ` + inFolders + `
				  AND owner = ?
				  AND deleted_at != 0

				UNION ALL

				SELECT f.id, f.parent, t.depth + 1
				FROM files f
				JOIN tree t ON f.parent = t.id
				WHERE f.owner = ?
				  AND f.deleted_at != 0
				  AND t.depth < 100
			)
			SELECT id FROM tree
		`

		rows, err := tx.Query(query, append(folderArgs, userID)...)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}

			treeIDs = append(treeIDs, id)
		}

		if err := rows.Err(); err != nil {
			return err
		}
	}

	// Merge and remove duplicates
	seen := make(map[string]struct{})
	allIDs := make([]string, 0)

	for _, id := range fileIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			allIDs = append(allIDs, id)
		}
	}

	for _, id := range treeIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			allIDs = append(allIDs, id)
		}
	}

	if len(allIDs) == 0 {
		return nil
	}

	// Restore
	inAll := "(?" + strings.Repeat(",?", len(allIDs)-1) + ")"
	allArgs := make([]any, len(allIDs))
	for i, id := range allIDs {
		allArgs[i] = id
	}

	if _, err := tx.Exec(
		`UPDATE files SET deleted_at = 0 WHERE id IN `+inAll,
		allArgs...,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`DELETE FROM file_trash WHERE file_id IN `+inAll+` AND owner = ?`,
		append(allArgs, userID)...,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *fileRepository) Trash(userID string, files, subFiles []model.File) error {
	now := time.Now().Unix()

	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Trash root files
	for _, file := range files {
		_, err = tx.Exec(
			`INSERT INTO file_trash (id, file_id, owner)
			 VALUES (?, ?, ?)`,
			model.NewID(),
			file.ID,
			userID,
		)
		if err != nil {
			return err
		}

		res, err := tx.Exec(
			`UPDATE files
			 SET deleted_at = ?
			 WHERE id = ?
			   AND owner = ?
			   AND (
			        lock_owner IS NULL
			        OR lock_expires_at < ?
			   )`,
			now,
			file.ID,
			userID,
			now,
		)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return errors.New("file is locked or not found")
		}
	}

	if len(subFiles) > 0 {
		for _, file := range subFiles {
			res, err := tx.Exec(
				`UPDATE files
			 SET deleted_at = ?
			 WHERE id = ?
			   AND (
			        lock_owner IS NULL
			        OR lock_expires_at < ?
			   )`,
				now,
				file.ID,
				now,
			)
			if err != nil {
				return err
			}

			rows, err := res.RowsAffected()
			if err != nil {
				return err
			}

			if rows == 0 {
				return errors.New("subfile is locked")
			}
		}
	}

	return tx.Commit()
}

func (m *fileRepository) Rename(
	file model.File,
	name string,
	newDisplayName string,
) error {
	now := time.Now().Unix()

	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	res, err := tx.Exec(
		`UPDATE files
		 SET name = ?, displayname = ?
		 WHERE id = ?
		   AND (
		        lock_owner IS NULL
		        OR lock_expires_at < ?
		   )`,
		name,
		newDisplayName,
		file.ID,
		now,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("file is locked or not found")
	}

	return tx.Commit()
}

func (m *fileRepository) MoveToStorage(
	lockOwner string,
	rootIDS []string,
	subIDS []string,
	targetStorageID,
	targetParentID string,
) error {
	tx, err := m.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	now := time.Now().Unix()

	if len(rootIDS) > 0 {
		args := make([]any, 0, 4+len(rootIDS))
		args = append(args, targetStorageID, targetParentID)

		ph := make([]string, len(rootIDS))
		for i, id := range rootIDS {
			ph[i] = "?"
			args = append(args, id)
		}

		args = append(args, lockOwner, now)

		stmt := fmt.Sprintf(
			`UPDATE files
			 SET storage = ?, parent = ?
			 WHERE id IN (%s)
			   AND (
			        lock_owner = ?
			        OR lock_expires_at < ?
			   )`,
			strings.Join(ph, ","),
		)

		res, err := tx.Exec(stmt, args...)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows != int64(len(rootIDS)) {
			return errors.New("one or more root files are locked by another process")
		}
	}

	if len(subIDS) > 0 {
		args := make([]any, 0, 3+len(subIDS))
		args = append(args, targetStorageID)

		ph := make([]string, len(subIDS))
		for i, id := range subIDS {
			ph[i] = "?"
			args = append(args, id)
		}

		args = append(args, lockOwner, now)

		stmt := fmt.Sprintf(
			`UPDATE files
			 SET storage = ?
			 WHERE id IN (%s)
			   AND (
			        lock_owner = ?
			        OR lock_expires_at < ?
			   )`,
			strings.Join(ph, ","),
		)

		res, err := tx.Exec(stmt, args...)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rows != int64(len(subIDS)) {
			return errors.New("one or more subfiles are locked by another process")
		}
	}

	return tx.Commit()
}

func (m *fileRepository) Update(file model.File) error {
	_, err := m.Db.Exec(`UPDATE files
							SET owner=?, parent=?,
								storage=?,
								name=?, displayname=?,
								size=?, type=?,
								is_folder=?,
								created_at=?, modified_at=?, deleted_at=?
						  WHERE id=?`, file.Owner, file.Parent,
		file.Storage,
		file.Name, file.DisplayName,
		file.Size, file.Type,
		file.IsFolder,
		file.Created, file.Modified, file.DeletedAt, file.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) UpdateSize(size int64, fileID string) error {
	now := time.Now().Unix()

	_, err := m.Db.Exec(`UPDATE files SET size=?, modified_at=? WHERE id=?`, size, now, fileID)
	if err != nil {
		return err
	}

	return nil
}

// GetUserRoot returns the user's cloud#drive root on the given storage, or nil
// if it does not exist yet. The deterministic pick guards against a rare
// duplicate from concurrent lazy creation.
func (m *fileRepository) GetUserRoot(ctx context.Context, userID string, storage string) (*model.File, error) {
	row := m.Db.QueryRowContext(ctx, `SELECT `+fileColumns("")+`
		FROM files
		WHERE type = 'cloud#drive' AND owner = ? AND storage = ?
		ORDER BY created_at, id
		LIMIT 1`, userID, storage)

	file := model.File{}
	err := row.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
		&file.Name, &file.DisplayName, &file.Size,
		&file.Type, &file.IsFolder, &file.Shared,
		&file.Created, &file.Modified, &file.DeletedAt,
		&file.LockOwner, &file.LockExpiresAt, &file.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &file, nil
}

// RenameStorageDrives renames every drive on a storage (the anchor and all
// per-user roots) in one statement, so a storage rename can't leave some drives
// with the old name.
func (m *fileRepository) RenameStorageDrives(ctx context.Context, storageID string, name string) error {
	_, err := m.Db.ExecContext(ctx, `UPDATE files SET name = ?, displayname = ? WHERE type = 'cloud#drive' AND storage = ?`, name, name, storageID)
	return err
}

func (m *fileRepository) DeleteDrive(id string) error {
	_, err := m.Db.Exec("DELETE FROM file_trash WHERE file_id in (SELECT DISTINCT id FROM files WHERE storage=?)", id)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("DELETE FROM file_share WHERE file_id in (SELECT DISTINCT id FROM files WHERE storage=?)", id)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("DELETE FROM files WHERE storage=?", id)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("DELETE FROM file_storage WHERE id=?", id)
	if err != nil {
		return err
	}

	return nil
}

// GetStorageUsed sums every file the user owns at any depth, trashed ones
// included since they occupy storage until the trash is emptied.
func (m *fileRepository) GetStorageUsed(ownerID string) (int64, error) {
	var size int64
	if err := m.Db.QueryRow(`SELECT COALESCE(SUM(size), 0) FROM files WHERE owner = ? AND type <> 'cloud#drive'`, ownerID).Scan(&size); err != nil {
		return 0, err
	}

	return size, nil
}

func (m *fileRepository) GetTotalStorage() (int64, error) {
	var size int64
	if err := m.Db.QueryRow(`SELECT COALESCE(SUM(size), 0) FROM files`).Scan(&size); err != nil {
		return 0, err
	}

	return size, nil
}

func (m *fileRepository) CreateMetadata(name string, metaType string, values any) (*model.FileMetadata, error) {
	id := uuid.New()
	t := time.Now().Unix()

	s, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	_, errInsert := m.Db.Exec(`INSERT INTO file_metadata (`+fileMetadataColumns+`) VALUES(?,?,?,?,?,?,?)`, id, metaType, name, s, t, t, "system")
	if errInsert != nil {
		return nil, errInsert
	}

	return &model.FileMetadata{
		ID:        id.String(),
		Type:      metaType,
		Title:     name,
		Fields:    values,
		CreatedAt: t,
		UpdatedAt: t,
		CreatedBy: "system",
	}, nil
}

func (m *fileRepository) GetAllMetadata() ([]model.FileMetadata, error) {
	results, err := m.Db.Query(`SELECT ` + fileMetadataColumns + ` FROM file_metadata`)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	metaArr := make([]model.FileMetadata, 0)
	meta := model.FileMetadata{}
	valueString := make([]byte, 0)
	for results.Next() {
		err := results.Scan(&meta.ID, &meta.Type, &meta.Title, &valueString, &meta.CreatedAt, &meta.UpdatedAt, &meta.CreatedBy)
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, 0)

		err = json.Unmarshal(valueString, &values)
		if err != nil {
			return nil, err
		}

		meta.Fields = values
		metaArr = append(metaArr, meta)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return metaArr, nil
}

func (m *fileRepository) CreateMetadataEntry(id string, value string, metadata model.FileMetadata) (*model.FileMetadataEntry, error) {
	t := time.Now().Unix()

	new := model.FileMetadataEntry{
		ID:         uuid.New().String(),
		FileID:     id,
		MetadataID: metadata.ID,
		Value:      value,
		Type:       metadata.Type,
		Title:      metadata.Title,
		CreatedAt:  t,
		UpdatedAt:  t,
	}

	_, err := m.Db.Exec(`INSERT INTO file_metadata_entries (`+fileMetadataEntryColumns+`) VALUES(?,?,?,?,?,?)`, new.ID, new.FileID, new.MetadataID, new.Value, new.CreatedAt, new.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &new, nil
}

func (m *fileRepository) GetMetadataByID(id string) (*model.FileMetadata, error) {
	results, err := m.Db.Query(`SELECT `+fileMetadataColumns+` FROM file_metadata WHERE id=?`, id)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	meta := model.FileMetadata{}
	val := make([]uint8, 0)
	for results.Next() {
		err := results.Scan(&meta.ID, &meta.Type, &meta.Title, &val, &meta.CreatedAt, &meta.UpdatedAt, &meta.CreatedBy)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(val, &meta.Fields)
		if err != nil {
			return nil, err
		}
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (m *fileRepository) DeleteMetadataEntry(fileID string, id string) error {
	_, err := m.Db.Exec("DELETE FROM file_metadata_entries WHERE id=? AND file_id=?", id, fileID)
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) DeleteMetadata(id string) error {
	_, err := m.Db.Exec("DELETE FROM file_metadata WHERE id=?", id)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("DELETE FROM file_metadata_entries WHERE metadata_id=?", id)
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) GetMetadataIDForFile(id string) (*string, error) {
	results, err := m.Db.Query(`SELECT file_id FROM file_metadata_entries WHERE id=?`, id)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	var fileID string
	for results.Next() {
		err := results.Scan(&fileID)
		if err != nil {
			return nil, err
		}
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return &fileID, nil
}

func (m *fileRepository) UpdateMetadata(id string, name string, metaType string, values interface{}) error {
	_, err := m.Db.Exec("UPDATE file_metadata SET title=?,type=?,fields=? WHERE id=?", name, metaType, values, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *fileRepository) UpdateMetadataEntry(id string, value any) error {
	s, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = m.Db.Exec("UPDATE file_metadata_entries SET value=? WHERE id=?", s, id)
	if err != nil {
		return err
	}

	return nil
}

// Scoped to fileIDs so a listing reads a row per file on screen, not a row per
// favourite the user has.
func (m *fileRepository) GetFavouriteIDs(userID string, fileIDs []string) ([]string, error) {
	if len(fileIDs) == 0 {
		return make([]string, 0), nil
	}

	args := make([]any, 0, len(fileIDs)+2)
	args = append(args, userID, "files")
	for _, id := range fileIDs {
		args = append(args, id)
	}

	results, err := m.Db.Query(
		`SELECT item_id FROM favorites
		 WHERE user_id=? AND app=? AND item_id IN (?`+strings.Repeat(",?", len(fileIDs)-1)+`)`,
		args...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	ids := make([]string, 0)
	var id string
	for results.Next() {
		if err = results.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (m *fileRepository) GetMetadataEntriesForFiles(ids []string) ([]model.FileMetadataEntry, error) {
	if len(ids) == 0 {
		return []model.FileMetadataEntry{}, nil
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	repeat := strings.Repeat(",?", len(ids)-1)
	stmt := `SELECT file_metadata_entries.*, file_metadata.fields, file_metadata.title, file_metadata.type
				FROM file_metadata_entries, file_metadata
				WHERE file_id IN (?` + repeat + `)
				AND file_metadata_entries.metadata_id = file_metadata.id`

	rows, err := m.Db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	entries := make([]model.FileMetadataEntry, 0)

	for rows.Next() {
		// Declared per row: json.Unmarshal leaves its target untouched when it
		// fails, so a reused one hands the previous row's value to this file.
		fm := model.FileMetadataEntry{}
		valueString := make([]byte, 0)
		fieldsString := make([]byte, 0)

		if err = rows.Scan(&fm.ID, &fm.FileID, &fm.MetadataID, &valueString, &fm.CreatedAt, &fm.UpdatedAt, &fieldsString, &fm.Title, &fm.Type); err != nil {
			return nil, err
		}

		fm.Value = ""
		fm.Fields = ""
		if len(valueString) > 0 {
			if err = json.Unmarshal(valueString, &fm.Value); err != nil {
				fm.Value = ""
			}
		}

		if len(fieldsString) > 0 {
			if err = json.Unmarshal(fieldsString, &fm.Fields); err != nil {
				fm.Fields = ""
			}
		}

		entries = append(entries, fm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (m *fileRepository) Lock(
	lockOwner string,
	fileIDs []string,
	ttl time.Duration,
) error {
	if len(fileIDs) == 0 {
		return nil
	}

	expiresAt := time.Now().Add(ttl).Unix()

	placeholders := make([]string, len(fileIDs))
	args := make([]any, 0, 2+len(fileIDs))

	for i := range fileIDs {
		placeholders[i] = "?"
	}

	args = append(args, lockOwner, expiresAt)
	for _, id := range fileIDs {
		args = append(args, id)
	}

	stmt := fmt.Sprintf(`
		UPDATE files
		SET lock_owner = ?, lock_expires_at = ?
		WHERE id IN (%s)
		  AND (
		        lock_owner IS NULL
		        OR lock_expires_at < ?
		      )
	`, strings.Join(placeholders, ","))

	// add "now" for expiration comparison
	args = append(args, time.Now().Unix())

	res, err := m.Db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows != int64(len(fileIDs)) {
		return errors.New("one or more files are already locked")
	}

	return nil
}

func (m *fileRepository) Unlock(lockOwner string) error {
	_, err := m.Db.Exec(`
		UPDATE files
		SET lock_owner = NULL,
		    lock_expires_at = NULL
		WHERE lock_owner = ?
	`, lockOwner)

	return err
}

func (m *fileRepository) RenewLock(
	lockOwner string,
	ttl time.Duration,
) error {
	expiresAt := time.Now().Add(ttl).Unix()

	_, err := m.Db.Exec(`
		UPDATE files
		SET lock_expires_at = ?
		WHERE lock_owner = ?
	`, expiresAt, lockOwner)

	return err
}
