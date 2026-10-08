// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package klipy

import (
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	APIKey   string `json:"api_key"`
	Provider string `json:"provider"`
}

type PostsResponse struct {
	Results []Post `json:"results"`
}

// A single GIF post
type Post struct {
	ID                       string           `json:"id"`
	Title                    string           `json:"title"`
	MediaFormats             map[string]Media `json:"media_formats"`
	Created                  float64          `json:"created"`
	ContentDescription       string           `json:"content_description"`
	ItemURL                  string           `json:"itemurl"`
	URL                      string           `json:"url"`
	Tags                     []string         `json:"tags"`
	Flags                    []string         `json:"flags"`
	HasAudio                 bool             `json:"hasaudio"`
	ContentDescriptionSource string           `json:"content_description_source"`
}

// A single media version (gif, mp4, tinygif, webm, webp, etc.)
type Media struct {
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
	Preview  string  `json:"preview"`
	Dims     []int   `json:"dims"` // [width, height]
	Size     int64   `json:"size"`
}

func New(apiKey string) *Client {
	return &Client{
		APIKey:   apiKey,
		Provider: "klipy",
	}
}

func (k *Client) FindByID(id string) (*PostsResponse, error) {
	resp, err := http.Get("https://api.klipy.com/v2/posts?ids=" + id + "&key=" + k.APIKey)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := PostsResponse{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
