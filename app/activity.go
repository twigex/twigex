// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"encoding/json"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetActivity(user model.User, fileID string) ([]model.Activity, *model.AppError) {
	file, appErr := a.HasPermission(fileID, user)
	if appErr != nil {
		return nil, appErr
	}

	activity, err := a.Store.Activity.GetForFile(*file)
	if err != nil || activity == nil {
		tlog.Errorw("File activity not found", "error", err)
		return nil, model.NewAppError("activity.get_file_activity.not_found", http.StatusNotFound)
	}

	return activity, nil
}

func (a *App) RecordActivity(userID, context, action, parentID, itemID string, data map[string]any) {
	j, err := json.Marshal(data)
	if err != nil {
		tlog.Errorw("Failed to marshal activity data",
			"user_id", userID,
			"action", action,
			"error", err,
		)
		return
	}

	if err = a.Store.Activity.Create(userID, context, action, parentID, itemID, string(j)); err != nil {
		tlog.Errorw("Failed to record activity",
			"user_id", userID,
			"context", context,
			"action", action,
			"item_id", itemID,
			"error", err,
		)
	}
}
