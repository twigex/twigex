// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/twigex/twigex/model"
)

type activityRepository struct {
	Db *sql.DB
}

const activityColumns = `id, app, type, user_id, affected_user, item_id, parent_id, parameters, created_at`

func NewActivityRepository(Db *sql.DB) (*activityRepository, error) {
	repo := &activityRepository{}

	repo.Db = Db
	return repo, nil
}

func (a *activityRepository) Create(userID, app, activityType, parentID, itemID, parameters string) error {
	created := time.Now().Unix()
	id := model.NewID()

	_, err := a.Db.Exec(`INSERT INTO activity (`+activityColumns+`) VALUES(?,?,?,?,?,?,?,?,?)`, id, app, activityType, userID, userID, itemID, parentID, parameters, created)
	if err != nil {
		return err
	}

	return nil
}

func (a *activityRepository) GetForFile(file model.File) ([]model.Activity, error) {
	arr := make([]model.Activity, 0)
	activity := model.Activity{}

	if file.IsFolder {
		results, err := a.Db.Query(`SELECT activity.*, users.name, users.lastname
		FROM activity, users
		WHERE activity.parent_id=?
		AND users.id = activity.affected_user
		ORDER BY activity.created_at DESC`, file.ID)
		if err != nil {
			return nil, err
		}

		defer results.Close()

		for results.Next() {
			var lastname string
			var parameters string
			m := make(map[string]interface{})

			err = results.Scan(&activity.ID, &activity.App, &activity.Type, &activity.UserID, &activity.AffectedUser,
				&activity.ItemID, &activity.ParentID, &parameters, &activity.CreatedAt, &activity.UserName, &lastname)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal([]byte(parameters), &m); err != nil {
				return nil, err
			}

			activity.Parameters = m
			activity.UserName = activity.UserName + " " + lastname

			arr = append(arr, activity)
		}

		if err := results.Err(); err != nil {
			return nil, err
		}
	}

	results, err := a.Db.Query(`SELECT activity.*, users.name, users.lastname
	FROM activity, users
	WHERE activity.item_id=?
	AND users.id = activity.affected_user
	ORDER BY activity.created_at DESC`, file.ID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	for results.Next() {
		var lastname string
		var parameters string
		m := make(map[string]interface{})

		err = results.Scan(&activity.ID, &activity.App, &activity.Type, &activity.UserID, &activity.AffectedUser,
			&activity.ItemID, &activity.ParentID, &parameters, &activity.CreatedAt, &activity.UserName, &lastname)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(parameters), &m); err != nil {
			return nil, err
		}

		activity.Parameters = m
		activity.UserName = activity.UserName + " " + lastname

		arr = append(arr, activity)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return arr, nil
}
