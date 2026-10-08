// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package interfaces

import (
	"context"

	"github.com/twigex/twigex/app/jobs"
	"github.com/twigex/twigex/model"
)

type LDAP interface {
	Authorize(c model.Credentials, u *model.User) (*model.User, error)
	TestConnection(in model.AdminLDAPSettings) *model.AppError
	SyncJob() jobs.Job
}

// OIDC covers both legs of the flow so the protocol libraries stay behind the
// seam; the API layer only ever handles a URL and a resulting user.
type OIDC interface {
	AuthCodeURL(ctx context.Context, providerID, state, nonce string) (string, *model.AppError)
	ExchangeToken(ctx context.Context, providerID, code, nonce string) (*model.User, *model.AppError)
}
