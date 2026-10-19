// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"slices"
	"testing"
)

func TestFeatureFlagsAreOffByDefault(t *testing.T) {
	config := ServerConfig{}
	config.SetDefaults()

	for name, on := range config.FeatureFlags.ByName() {
		if on {
			t.Errorf("flag %s is on by default", name)
		}
	}

	enabled := config.FeatureFlags.Enabled()
	if len(enabled) != 0 {
		t.Errorf("want no enabled flags, got %v", enabled)
	}
}

func TestFeatureFlagsEnabledListsFlagsThatAreOn(t *testing.T) {
	flags := FeatureFlags{KnowledgeGraph: NewBool(true)}

	enabled := flags.Enabled()

	want := []string{"KnowledgeGraph"}
	if !slices.Equal(enabled, want) {
		t.Fatalf("want %v, got %v", want, enabled)
	}
}
