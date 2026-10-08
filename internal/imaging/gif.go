// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package imaging

import (
	"net/url"
	"regexp"

	"github.com/twigex/twigex/model"
)

var (
	giphyRe = regexp.MustCompile(`https?://(?:media[0-9]*\.)?giphy\.com/media/([^/]+)/`)
	klipyRe = regexp.MustCompile(`https?://(?:media[0-9]*\.)?klipy\.com/(?:m/)?([^/]+)/([^/?#]+)`)
	imgurRe = regexp.MustCompile(`https?://i\.imgur\.com/([A-Za-z0-9]+)\.`)
)

func NormalizeGIF(raw string) model.GifNormalized {
	clean := stripURLParams(raw)

	if m := giphyRe.FindStringSubmatch(clean); len(m) >= 2 {
		id := m[1]
		return model.GifNormalized{
			Provider: "giphy",
			ID:       id,
			RawURL:   raw,
		}
	}

	if m := klipyRe.FindStringSubmatch(clean); len(m) >= 3 {
		id := m[1]

		return model.GifNormalized{
			Provider: "klipy",
			ID:       id,
			RawURL:   raw,
		}
	}

	if m := imgurRe.FindStringSubmatch(clean); len(m) >= 2 {
		id := m[1]
		return model.GifNormalized{
			Provider: "imgur",
			ID:       id,
			RawURL:   raw,
		}
	}

	return model.GifNormalized{
		Provider: "unknown",
		ID:       "",
		RawURL:   raw,
		GIFURL:   "",
	}
}

func stripURLParams(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
