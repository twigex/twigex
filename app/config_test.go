// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
)

func clientConfigApp(knowledgeGraph bool) *App {
	serverConfig := model.ServerConfig{}
	serverConfig.SetDefaults()
	serverConfig.FeatureFlags.KnowledgeGraph = model.NewBool(knowledgeGraph)

	a := &App{
		ConfigStore: config.ConfigStore{Config: &serverConfig},
	}

	return a
}

func TestClientConfigSendsFeatureFlagsToSignedInUsers(t *testing.T) {
	a := clientConfigApp(true)

	clientConfig := a.GetClientConfig(true)

	flags, ok := clientConfig["FeatureFlags"].(map[string]bool)
	if !ok {
		t.Fatalf("want FeatureFlags in the client config, got %v", clientConfig["FeatureFlags"])
	}

	if !flags["KnowledgeGraph"] {
		t.Errorf("want KnowledgeGraph on, got %v", flags)
	}
}

func TestClientConfigHidesFeatureFlagsFromAnonymousUsers(t *testing.T) {
	a := clientConfigApp(true)

	clientConfig := a.GetClientConfig(false)

	if _, ok := clientConfig["FeatureFlags"]; ok {
		t.Fatalf("want no FeatureFlags for anonymous users, got %v", clientConfig["FeatureFlags"])
	}
}
