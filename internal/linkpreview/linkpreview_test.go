// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package linkpreview

import (
	"encoding/json"
	"strings"
	"testing"
)

const ogPage = `<!doctype html><html><head>
<meta property="og:title" content="Example Title">
<meta property="og:description" content="Example description.">
<meta property="og:site_name" content="Example">
<meta property="og:image" content="/static/og.png">
<meta property="og:image:width" content="1200">
<meta property="og:image:height" content="630">
<link rel="icon" href="/favicon.ico">
<title>Fallback Title</title>
</head><body></body></html>`

func parse(t *testing.T, html, pageURL string) map[string]any {
	t.Helper()
	out, err := Parse(strings.NewReader(html), pageURL, Options{
		Title: true, Description: true, Image: true, Favicon: true, SiteName: true,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestParseOpenGraph(t *testing.T) {
	m := parse(t, ogPage, "https://example.com/page")

	if m["title"] != "Example Title" {
		t.Errorf("title = %v", m["title"])
	}
	if m["type"] != "opengraph" {
		t.Errorf("type = %v, want opengraph", m["type"])
	}

	img, ok := m["image"].(map[string]any)
	if !ok {
		t.Fatalf("image is %T, want object", m["image"])
	}
	// Relative og:image resolves against the page URL.
	if img["url"] != "https://example.com/static/og.png" {
		t.Errorf("image.url = %v", img["url"])
	}
	if img["width"].(float64) != 1200 || img["height"].(float64) != 630 {
		t.Errorf("image dims = %v x %v", img["width"], img["height"])
	}
	if m["favicon"] != "https://example.com/favicon.ico" {
		t.Errorf("favicon = %v", m["favicon"])
	}
}

func TestParseTitleFallbackMarksHTML(t *testing.T) {
	html := `<html><head><title>Only Title</title></head></html>`
	m := parse(t, html, "https://example.com")

	if m["title"] != "Only Title" {
		t.Errorf("title = %v", m["title"])
	}
	if m["type"] != "html" {
		t.Errorf("type = %v, want html (title fallback)", m["type"])
	}
}

func TestParseEmptyPageURLLeavesRelativeRefs(t *testing.T) {
	m := parse(t, ogPage, "")
	img := m["image"].(map[string]any)
	if img["url"] != "/static/og.png" {
		t.Errorf("image.url = %v, want unchanged relative ref", img["url"])
	}
}
