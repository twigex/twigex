// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"strings"

	"github.com/twigex/twigex/internal/catalog"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

var connectionFailures = map[string]string{
	catalog.FailureCredentials: "connection.verification.credentials",
	catalog.FailureDatabase:    "connection.verification.database",
	catalog.FailureCertificate: "connection.verification.certificate",
	catalog.FailureHostname:    "connection.verification.hostname",
	catalog.FailureCACert:      "connection.verification.ca_cert",
	catalog.FailureUnreachable: "connection.verification.unreachable",
}

func verificationError(err error) *model.AppError {
	id, ok := connectionFailures[catalog.Failure(err)]
	if !ok {
		id = "connection.verification_failed"
	}

	return model.NewAppError(id, http.StatusBadRequest)
}

var cubeErrorReasons = []struct {
	id       string
	fragment []string
}{
	// Checked first: a name mismatch is not the authority's fault, and Node
	// words it as an ordinary certificate error.
	{"connection.verification.hostname", []string{
		"altnames", "ip sans", "hostname/ip does not match",
		"does not match certificate", "cannot validate certificate for",
	}},
	{"connection.verification.certificate", []string{
		"self-signed certificate", "self signed certificate",
		"unable to verify", "unable to get local issuer",
		"cert_", "certificate has expired",
	}},
	{"connection.verification.unreachable", []string{
		"etimedout", "econnrefused", "enotfound", "ehostunreach", "econnreset",
	}},
	{"connection.verification.credentials", []string{"access denied"}},
	{"connection.verification.database", []string{"unknown database", "does not exist"}},
}

// CubeErrorID names the setting behind a driver error Cube reported, so the
// chart shows the same wording as saving the connection would. Empty when the
// message is not recognised, leaving Cube's own text in place.
func CubeErrorID(cubeError string) string {
	if cubeError == "" || cubeError == "Continue wait" {
		return ""
	}

	lower := strings.ToLower(cubeError)
	for _, reason := range cubeErrorReasons {
		for _, fragment := range reason.fragment {
			if strings.Contains(lower, fragment) {
				return reason.id
			}
		}
	}

	return ""
}

// logCubeError records a Cube.js error response for observability. Cube reports
// query errors in a 200 body via an "error" field (and sometimes a non-2xx
// status); the body is still returned so the client can show the specific
// message. "Continue wait" is the long-poll signal, not an error.
func logCubeError(workspaceID string, statusCode int, body map[string]any) {
	errVal, ok := body["error"].(string)
	if (ok && errVal != "Continue wait") || statusCode >= 300 {
		tlog.Warnw("CubeJS returned an error",
			"workspace_id", workspaceID,
			"status", statusCode,
			"cube_error", body["error"],
		)
	}
}
