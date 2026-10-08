// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import "net/http"

type localeKey struct{}

func getLocale(r *http.Request) string {
	locale, ok := r.Context().Value(localeKey{}).(string)
	if !ok {
		return "en"
	}

	return locale
}
