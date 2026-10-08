// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"time"

	"github.com/twigex/twigex/internal/ratelimit"
)

func (a *App) loginLimiter() ratelimit.Limiter {
	return ratelimit.New(a.Store.Auth, "login_attempt", 5, 5*time.Minute)
}

func (a *App) shareAuthLimiter() ratelimit.Limiter {
	return ratelimit.New(a.Store.Auth, "share_auth", 5, 5*time.Minute)
}
