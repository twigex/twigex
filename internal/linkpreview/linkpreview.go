// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Package linkpreview extracts link-preview metadata (title, description,
// image, favicon, site name) from an HTML document. It is a pure parser and
// performs no network access: callers fetch the page themselves and pass the
// body to Parse, keeping timeout, redirect, and SSRF controls in caller code.
package linkpreview

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type Options struct {
	Title       bool
	Description bool
	Image       bool
	Favicon     bool
	SiteName    bool
}

// Parse returns metadata as JSON. pageURL is the final, post-redirect URL: it
// is echoed as the "url" field and resolves relative image/favicon references,
// and may be empty.
func Parse(r io.Reader, pageURL string, opts Options) ([]byte, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	// base is nil when pageURL is empty or unparseable, leaving relative
	// references untouched rather than failing the whole parse.
	base, _ := url.Parse(pageURL)

	data := make(map[string]any)
	data["type"] = "opengraph"
	data["url"] = pageURL

	if opts.Title {
		title := extractMetaContent(doc, "property", "og:title")
		if title == "" {
			title = doc.Find("title").Text()
			if title != "" {
				data["type"] = "html"
			}
		}

		data["title"] = title
	}

	if opts.Description {
		description := extractMetaContent(doc, "property", "og:description")
		if description == "" {
			description = extractMetaContent(doc, "name", "description")
		}

		data["description"] = description
	}

	if opts.Image {
		imageURL := extractMetaContent(doc, "property", "og:image")
		width := extractMetaContent(doc, "property", "og:image:width")
		height := extractMetaContent(doc, "property", "og:image:height")

		widthInt, err := strconv.Atoi(width)
		if err != nil {
			widthInt = 0
		}

		heightInt, err := strconv.Atoi(height)
		if err != nil {
			heightInt = 0
		}

		data["image"] = map[string]any{
			"url":    resolveURL(base, imageURL),
			"width":  widthInt,
			"height": heightInt,
		}
	}

	if opts.Favicon {
		favicon := doc.Find("link[rel='icon']").AttrOr("href", "")
		if favicon == "" {
			favicon = doc.Find("link[rel='shortcut icon']").AttrOr("href", "")
			if favicon == "" {
				// apple-touch-icon is often used as favicon
				favicon = doc.Find("link[rel='apple-touch-icon']").AttrOr("href", "")
			}
		}

		data["favicon"] = resolveURL(base, favicon)
	}

	if opts.SiteName {
		siteName := extractMetaContent(doc, "property", "og:site_name")
		data["site_name"] = siteName
	}

	return json.Marshal(data)
}

func extractMetaContent(doc *goquery.Document, key, value string) string {
	return doc.Find("meta["+key+"='"+value+"']").AttrOr("content", "")
}

func resolveURL(base *url.URL, ref string) string {
	if ref == "" || base == nil {
		return ref
	}

	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	return base.ResolveReference(u).String()
}
