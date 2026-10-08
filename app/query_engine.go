// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
)

// QueryEngine is everything Twigex needs from an analytics backend. The Cube.js
// implementation is cubeEngine; a future Go-native engine would implement the
// same interface, so callers never change.
type QueryEngine interface {
	Meta(ctx context.Context, workspaceID string) (*collimato.CubeCollection, error)
	SQL(ctx context.Context, workspaceID string, query model.DataQuery) (map[string]any, error)
	Query(ctx context.Context, workspaceID string, query model.DataQuery) (map[string]any, error)
}

// cubeEngine talks to the shared Cube.js server over HTTP. It owns the Cube
// specifics (URL, JWT auth) so the rest of the app stays engine-agnostic.
type cubeEngine struct {
	client    *http.Client
	apiURL    func() string                   // read lazily: config can change at runtime
	mintToken func(workspaceID string) string // = App.GenerateToken
}

func (e *cubeEngine) Meta(ctx context.Context, workspaceID string) (*collimato.CubeCollection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.apiURL()+"/cubejs-api/v1/meta", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.mintToken(workspaceID))

	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if !json.Valid(body) {
		return nil, fmt.Errorf("cube: invalid JSON in meta response")
	}

	var collection collimato.CubeCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

func (e *cubeEngine) SQL(ctx context.Context, workspaceID string, query model.DataQuery) (map[string]any, error) {
	return e.post(ctx, workspaceID, "/cubejs-api/v1/sql", query)
}

func (e *cubeEngine) Query(ctx context.Context, workspaceID string, query model.DataQuery) (map[string]any, error) {
	return e.post(ctx, workspaceID, "/cubejs-api/v1/load", query)
}

// post runs a query against a Cube endpoint. Cube reports query errors inside a
// 200 body, so the decoded body is returned (and logged by logCubeError) even on
// a Cube-level error, for the caller to surface.
func (e *cubeEngine) post(ctx context.Context, workspaceID, endpoint string, query model.DataQuery) (map[string]any, error) {
	payload, err := json.Marshal(struct {
		Query model.DataQuery `json:"query"`
	}{Query: query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.apiURL()+endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.mintToken(workspaceID))

	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, err
	}

	logCubeError(workspaceID, res.StatusCode, body)
	return body, nil
}
