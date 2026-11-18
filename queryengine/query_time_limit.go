// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

// QueryReachedTimeLimit says whether the database stopped a query because it
// ran past ConnectorOptions.QueryTimeLimit.
func QueryReachedTimeLimit(err error) bool {
	for _, database := range supportedDatabases {
		if database.stoppedAtTimeLimit(err) {
			return true
		}
	}

	return false
}
