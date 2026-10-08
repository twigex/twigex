// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type GifNormalized struct {
	Provider string
	ID       string
	Poster   string
	RawURL   string
	GIFURL   string
	Width    int
	Height   int
}
