// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

// effectiveShareSubquery returns a SQL fragment usable in a FROM clause that
// resolves all file_share rows reachable by userID: direct user shares
// (share_type = SHARE_TYPE_USER, share_with = userID) plus shares to any
// group the user is a member of (share_type = SHARE_TYPE_GROUP, share_with
// = group_id IN user's groups).
//
// The returned alias is "efs" with columns matching file_share.
//
// Safety contract:
//   - userID flows into args, NEVER into the returned SQL string.
//   - The returned string is built only from compile-time constants and is
//     therefore safe to concat into a larger query via the `+` operator.
//   - The string contains exactly THREE `?` placeholders: one for the direct
//     user-share branch, and two for the group branch (group membership +
//     owner-suppression so a file owner doesn't see their own file via a group
//     they belong to).
//
// Arg-order contract (this is the easy thing to get wrong):
//   - The returned args MUST be passed to the driver BEFORE any args for
//     placeholders that appear later in the outer query, because the helper's
//     placeholders appear first in the executed SQL.
//   - Always build the final args slice as:
//     args := append(shareArgs, outerArgs...)
//     not:
//     args := append([]any{outerArg}, shareArgs...)   // WRONG, placeholders shift
//
// Usage:
//
//	join, joinArgs := effectiveShareSubquery(userID)
//	query := "SELECT " + fileColumns("files") + " FROM files JOIN " + join + " ON efs.file_id = files.id WHERE files.parent = ?"
//	args := append(joinArgs, parentID)        // joinArgs MUST come first
//	rows, err := s.Db.Query(query, args...)
func effectiveShareSubquery(userID string) (sql string, args []any) {
	const cols = `fs.id, fs.file_id, fs.initiator,
	              fs.share_type, fs.share_with,
	              fs.expiration, fs.access_level,
	              fs.time_shared, fs.created_at, fs.updated_at`
	return `(
		SELECT ` + cols + `
		  FROM file_share fs
		  WHERE fs.share_type = ` + fmtShareType(model.SHARE_TYPE_USER) + ` AND fs.share_with = ?
		UNION ALL
		SELECT ` + cols + `
		  FROM file_share fs
		  INNER JOIN files ON fs.file_id = files.id
		  JOIN group_members gm ON gm.group_id = fs.share_with
		  WHERE fs.share_type = ` + fmtShareType(model.SHARE_TYPE_GROUP) + ` AND gm.user_id = ?
		  AND files.owner != ?
	) AS efs`, []any{userID, userID, userID}
}

// fmtShareType formats a share-type constant as a literal for inline SQL.
// Safe because the inputs are compile-time constants from the model package.
func fmtShareType(t int) string {
	// strconv.Itoa avoids importing fmt just for this; share.go already imports fmt below.
	return fmt.Sprintf("%d", t)
}

func (s *fileRepository) CreateShare(file model.File, shareType int, user string, targetUser string, expiration int64, level model.AccessLevel) (*model.SharedFile, error) {
	shared := time.Now().Unix()

	_, err := s.Db.Exec(`INSERT INTO file_share(id, file_id, initiator, share_type, share_with, expiration, access_level, time_shared, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE access_level = VALUES(access_level), expiration = VALUES(expiration), updated_at = VALUES(updated_at)`,
		model.NewID(), file.ID, user, shareType, targetUser, expiration, level, shared, shared, shared)
	if err != nil {
		return nil, err
	}

	share := &model.SharedFile{}
	err = s.Db.QueryRow(`SELECT id, file_id, initiator, share_type, share_with, expiration, access_level, time_shared, created_at, updated_at
		FROM file_share WHERE file_id = ? AND share_type = ? AND share_with = ?`, file.ID, shareType, targetUser).Scan(
		&share.ID, &share.FileID, &share.Initiator, &share.ShareType, &share.ShareWith, &share.Expiration,
		&share.AccessLevel, &share.TimeShared, &share.CreatedAt, &share.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return share, nil
}

func (s *fileRepository) GetShared(userID string) ([]model.File, error) {
	now := time.Now().Unix()

	// Shared-with-me is a pure ACL view: every file the user has a share row
	// on, directly or via a group. Shares live only on the share root, so these
	// rows are exactly the top-level shared items.
	shareJoin, shareArgs := effectiveShareSubquery(userID)
	stmt := `SELECT ` + fileColumns("files") + `, efs.expiration, efs.access_level
	FROM files
	JOIN ` + shareJoin + ` ON efs.file_id = files.id
	WHERE files.deleted_at = 0`

	results, err := s.Db.Query(stmt, shareArgs...)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	set := newSharedFileSet()
	for results.Next() {
		var file model.File
		err := results.Scan(&file.ID, &file.Owner, &file.Parent, &file.Storage,
			&file.Name, &file.DisplayName, &file.Size,
			&file.Type, &file.IsFolder, &file.Shared,
			&file.Created, &file.Modified, &file.DeletedAt,
			&file.LockOwner, &file.LockExpiresAt, &file.Version,
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

	dbFiles := set.files
	ids := make([]string, 0, len(dbFiles))
	for _, f := range dbFiles {
		ids = append(ids, f.ID)
	}

	metaArr := make([]model.FileMetadataEntry, 0)
	fm := model.FileMetadataEntry{}

	if len(ids) > 0 {
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}

		repeat := strings.Repeat(",?", len(ids)-1)
		stmt := `SELECT file_metadata_entries.*, file_metadata.fields, file_metadata.title, file_metadata.type
					FROM file_metadata_entries, file_metadata
					WHERE file_id IN (?` + repeat + `)
					AND file_metadata_entries.metadata_id = file_metadata.id`
		m, err := s.Db.Query(stmt, args...)
		if err != nil {
			return nil, err
		}

		defer m.Close()
		valueString := make([]byte, 0)
		fieldsString := make([]byte, 0)

		for m.Next() {
			if err := m.Scan(&fm.ID, &fm.FileID, &fm.MetadataID, &valueString, &fm.CreatedAt, &fm.UpdatedAt, &fieldsString, &fm.Title, &fm.Type); err != nil {
				return nil, err
			}

			if len(valueString) > 0 {
				if err := json.Unmarshal(valueString, &fm.Value); err != nil {
					return nil, err
				}
			} else {
				fm.Value = ""
			}

			if len(fieldsString) > 0 {
				if err := json.Unmarshal(fieldsString, &fm.Fields); err != nil {
					return nil, err
				}
			} else {
				fm.Fields = ""
			}

			metaArr = append(metaArr, fm)
		}

		if err := m.Err(); err != nil {
			return nil, err
		}
	}

	favID := ""
	favourites := make([]string, 0)
	exists, err := s.Db.Query(`SELECT item_id
				FROM favorites
				WHERE user_id=? AND app=?`, userID, "files")
	if err != nil {
		return nil, err
	}

	defer exists.Close()

	for exists.Next() {
		if err := exists.Scan(&favID); err != nil {
			return nil, err
		}

		favourites = append(favourites, favID)
	}

	if err := exists.Err(); err != nil {
		return nil, err
	}

	for i := 0; i < len(dbFiles); i++ {
		for l := 0; l < len(favourites); l++ {
			if dbFiles[i].ID == favourites[l] {
				dbFiles[i].Favourite = true
			}
		}

		for _, v := range metaArr {
			if v.FileID == dbFiles[i].ID {
				dbFiles[i].Metadata = append(dbFiles[i].Metadata, v)
			}
		}

	}

	return dbFiles, nil
}

func (s *fileRepository) UpdateUserPermission(p model.FileSharePatch) error {
	t := time.Now().Unix()
	_, errDelete := s.Db.Exec(`UPDATE file_share SET expiration=?, access_level=?, updated_at=? WHERE share_with=? AND file_id=? AND share_type=?`,
		p.Expiration, p.AccessLevel, t, p.ID, p.FileID, p.ShareType)
	if errDelete != nil {
		return errDelete
	}

	return nil
}

func (s *fileRepository) RemoveUserFromShare(fileID string, userID string) error {
	_, err := s.Db.Exec("DELETE FROM file_share WHERE file_id=? AND share_with=? AND share_type=?", fileID, userID, model.SHARE_TYPE_USER)
	if err != nil {
		return err
	}

	return nil
}

func (s *fileRepository) RemoveGroupFromShare(ctx context.Context, fileID, groupID string) error {
	_, err := s.Db.ExecContext(ctx,
		`DELETE FROM file_share WHERE file_id = ? AND share_with = ? AND share_type = ?`,
		fileID, groupID, model.SHARE_TYPE_GROUP,
	)
	return err
}

func (s *fileRepository) RemoveUsersFromSharedFile(fileID string, users []string) error {
	if len(users) == 0 {
		return nil
	}

	args := make([]any, 0, len(users)+2)
	args = append(args, fileID, model.SHARE_TYPE_USER)
	for _, u := range users {
		args = append(args, u)
	}

	_, err := s.Db.Exec(
		`DELETE FROM file_share
		 WHERE file_id = ? AND share_type = ? AND share_with IN (`+sqlPlaceholders(len(users))+`)`,
		args...)

	return err
}

func (s *fileRepository) DeleteShares(fileID string) error {
	_, errDelete := s.Db.Exec("DELETE FROM file_share WHERE file_id=?", fileID)
	if errDelete != nil {
		return errDelete
	}

	return nil
}
