// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

const defaultMediaType = "application/octet-stream"

// setUploadHeaders keeps an uploaded HTML or SVG file from running as a
// document in this origin.
func setUploadHeaders(w http.ResponseWriter, storedType, name string) {
	safe := sanitizeHeaderFilename(name)

	w.Header().Set("Content-Type", uploadMediaType(storedType, name))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.PathEscape(safe)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func uploadMediaType(storedType, name string) string {
	mediaType, _, err := mime.ParseMediaType(storedType)
	if err != nil {
		mediaType, _, err = mime.ParseMediaType(mime.TypeByExtension(filepath.Ext(name)))
		if err != nil {
			return defaultMediaType
		}
	}

	mediaType = strings.ToLower(mediaType)

	switch {
	case mediaType == "application/pdf",
		strings.HasPrefix(mediaType, "image/"),
		strings.HasPrefix(mediaType, "video/"),
		strings.HasPrefix(mediaType, "audio/"):
		return mediaType
	}

	return defaultMediaType
}
