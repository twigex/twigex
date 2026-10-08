// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"time"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) ResetPassword(token, password string) *model.AppError {
	hashedToken := crypto.HashSHA256(token)

	p, err := a.Store.PasswordReset.Get(hashedToken)
	if err != nil {
		return model.NewAppError("auth.invalid_credentials", http.StatusNotFound)
	}

	if time.Now().Unix() > int64(p.ExpiresAt) {
		_ = a.Store.PasswordReset.Delete(hashedToken)
		return model.NewAppError("auth.invalid_credentials", http.StatusNotFound)
	}

	if _, appErr := a.isValidPassword(password); appErr != nil {
		return appErr
	}

	u, err := a.Store.User.Get(p.UserID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for password reset",
			"user_id", p.UserID,
			"error", err,
		)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	hashedPW, err := crypto.HashPassword(password)
	if err != nil {
		tlog.Errorw("Failed to hash password",
			"user_id", p.UserID,
			"error", err,
		)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	if _, err = a.Store.User.UpdatePassword(u.ID, hashedPW); err != nil {
		tlog.Errorw("Failed to update password",
			"user_id", u.ID,
			"error", err,
		)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	_ = a.Store.PasswordReset.Delete(hashedToken)

	if appErr := a.DeleteUserSessions(*u); appErr != nil {
		return appErr
	}

	return nil
}

func (a *App) SendResetPasswordEmail(email string) *model.AppError {
	user, err := a.Store.User.GetByEmail(email)
	if err != nil {
		tlog.Errorw("Failed to retrieve user by email",
			"error", err,
		)
		return model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	if user == nil {
		return nil
	}

	token := crypto.RandomText()

	if err = a.Store.PasswordReset.Create(*user, crypto.HashSHA256(token)); err != nil {
		tlog.Errorw("Failed to create password reset token",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	go a.SendPasswordResetEmail(*user, token)

	return nil
}
