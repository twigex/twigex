// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

var connectionFailures = map[queryengine.ConnectionFailure]string{
	queryengine.ConnectionFailureCredentials:   "connection.verification.credentials",
	queryengine.ConnectionFailureDatabase:      "connection.verification.database",
	queryengine.ConnectionFailureCertificate:   "connection.verification.certificate",
	queryengine.ConnectionFailureHostname:      "connection.verification.hostname",
	queryengine.ConnectionFailureCACertificate: "connection.verification.ca_cert",
	queryengine.ConnectionFailureUnreachable:   "connection.verification.unreachable",
}

func verificationError(err error) *model.AppError {
	id, ok := connectionFailures[queryengine.ConnectionFailureOf(err)]
	if !ok {
		id = "connection.verification_failed"
	}

	return model.NewAppError(id, http.StatusBadRequest)
}
