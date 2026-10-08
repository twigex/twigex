// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package parse

import (
	"strconv"
	"strings"
)

// Int converts a string value to int with a fallback default.
func Int(val string, defaultVal int) int {
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return n
}

// Bool converts a string value to bool with a fallback default.
func Bool(val string, defaultVal bool) bool {
	if val == "" {
		return defaultVal
	}

	return strings.ToLower(val) == "true"
}
