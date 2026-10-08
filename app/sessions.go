// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetSessionDetails(token string) (*model.SessionDetails, *model.AppError) {
	sd := model.SessionDetails{}
	if err := a.Store.Auth.Get(token, &sd); err != nil {
		return nil, model.NewAppError("session.invalid", http.StatusUnauthorized)
	}

	return &sd, nil
}

func (a *App) GetMySessions(user model.User) ([]model.SessionInfo, *model.AppError) {
	sessions, err := a.Store.Auth.GetSessions(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user sessions",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("session.invalid", http.StatusInternalServerError)
	}

	infos := make([]model.SessionInfo, 0, len(sessions))
	for _, v := range sessions {
		infos = append(infos, model.SessionInfo{
			ID:           sessionHandle(v.AccessToken),
			IP:           v.IP,
			DeviceID:     v.DeviceID,
			Browser:      v.Browser,
			Created:      v.Created,
			LastActivity: v.LastActivity,
			Expires:      v.Expires,
		})
	}

	return infos, nil
}

// sessionHandle names a session to the client without handing over the access
// token, which is the value of the session cookie.
func sessionHandle(accessToken string) string {
	sum := sha256.Sum256([]byte(accessToken))
	return hex.EncodeToString(sum[:])
}

func (a *App) LogOutSession(user model.User, id string) *model.AppError {
	sessions, err := a.Store.Auth.GetSessions(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user sessions",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("session.invalid", http.StatusInternalServerError)
	}

	for _, v := range sessions {
		if sessionHandle(v.AccessToken) != id {
			continue
		}

		if err = a.Store.Auth.Delete(v.AccessToken); err != nil {
			tlog.Errorw("Failed to delete session",
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("session.delete_failed", http.StatusInternalServerError)
		}
	}

	return nil
}

func (a *App) DeleteUserSessions(user model.User) *model.AppError {
	return a.deleteSessions(user, "")
}

// DeleteOtherUserSessions leaves keep alive, so changing a password signs out
// every other device without signing out the one doing the changing.
func (a *App) DeleteOtherUserSessions(user model.User, keep string) *model.AppError {
	return a.deleteSessions(user, keep)
}

func (a *App) deleteSessions(user model.User, keep string) *model.AppError {
	sessions, err := a.Store.Auth.GetSessions(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user sessions",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("session.invalid", http.StatusInternalServerError)
	}

	for _, v := range sessions {
		if keep != "" && v.AccessToken == keep {
			continue
		}

		if err = a.Store.Auth.Delete(v.AccessToken); err != nil {
			tlog.Errorw("Failed to delete session",
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("session.delete_failed", http.StatusInternalServerError)
		}
	}

	return nil
}
