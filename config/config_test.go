// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package config

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func featureFlagsFromEnv() model.FeatureFlags {
	config := model.ServerConfig{}
	config.SetDefaults()

	store := &ConfigStore{Config: &config}
	store.loadFeatureFlags()

	return config.FeatureFlags
}

func TestKnowledgeGraphFlagFromEnv(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty", "", false},
		{"true", "true", true},
		{"false", "false", false},
		{"anything else", "yes", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TWIGEX_FEATURE_KNOWLEDGE_GRAPH", tc.value)

			flags := featureFlagsFromEnv()

			if *flags.KnowledgeGraph != tc.want {
				t.Fatalf("want %v, got %v", tc.want, *flags.KnowledgeGraph)
			}
		})
	}
}
