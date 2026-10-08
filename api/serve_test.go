// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http/httptest"
	"testing"
)

func TestUploadMediaType(t *testing.T) {
	cases := []struct {
		name       string
		storedType string
		fileName   string
		want       string
	}{
		{"html collapses", "text/html", "page.html", "application/octet-stream"},
		{"html with charset collapses", "text/html; charset=utf-8", "page.html", "application/octet-stream"},
		{"xhtml collapses", "application/xhtml+xml", "page.xhtml", "application/octet-stream"},
		{"plain text collapses", "text/plain", "notes.txt", "application/octet-stream"},
		{"svg keeps its type", "image/svg+xml", "logo.svg", "image/svg+xml"},
		{"type is matched case insensitively", "IMAGE/PNG", "shot.png", "image/png"},
		{"png keeps its type", "image/png", "shot.png", "image/png"},
		{"pdf keeps its type", "application/pdf", "report.pdf", "application/pdf"},
		{"video keeps its type", "video/mp4", "clip.mp4", "video/mp4"},
		{"audio keeps its type", "audio/mpeg", "song.mp3", "audio/mpeg"},
		{"missing type falls back to the extension", "", "shot.png", "image/png"},
		{"missing type on an html name collapses", "", "page.html", "application/octet-stream"},
		{"unknown type and extension collapses", "", "data.unknown", "application/octet-stream"},
		{"unparsable type collapses", "not a media type", "page.html", "application/octet-stream"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := uploadMediaType(c.storedType, c.fileName); got != c.want {
				t.Errorf("want %s, got %s", c.want, got)
			}
		})
	}
}

func TestSetUploadHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	setUploadHeaders(w, "text/html", `evil".html`)

	if got := w.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("want a byte stream, got %s", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="evil.html"; filename*=UTF-8''evil.html` {
		t.Errorf("unexpected disposition: %s", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("want nosniff, got %s", got)
	}

	w = httptest.NewRecorder()
	setUploadHeaders(w, "image/png", "shot.png")

	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="shot.png"; filename*=UTF-8''shot.png` {
		t.Errorf("unexpected disposition: %s", got)
	}
}
