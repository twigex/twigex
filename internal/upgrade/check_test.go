// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func setBuild(t *testing.T, version, floor string) {
	t.Helper()

	originalVersion, originalFloor := model.Version, model.MinUpgradableVersion
	model.Version, model.MinUpgradableVersion = version, floor
	t.Cleanup(func() {
		model.Version, model.MinUpgradableVersion = originalVersion, originalFloor
	})
}

func TestCheck(t *testing.T) {
	useStops(t, fixture())

	installed := Schema{Installed: true, Version: 94}

	tests := []struct {
		name     string
		version  string
		floor    string
		recorded string
		schema   Schema
		wantErr  string
	}{
		{
			name:    "database install has not touched",
			version: "0.20.0",
			floor:   "0.19.0",
			schema:  Schema{},
		},
		{
			name:     "development build",
			version:  "dev",
			floor:    "0.19.0",
			recorded: "0.10.0",
			schema:   installed,
		},
		{
			name:     "recorded version clears the floor",
			version:  "0.20.0",
			floor:    "0.19.0",
			recorded: "0.19.0",
			schema:   installed,
		},
		{
			name:    "legacy install within the floor",
			version: "0.14.0",
			floor:   model.LastUngatedVersion,
			schema:  installed,
		},
		{
			name:    "legacy install meeting the release that squashes migrations",
			version: "0.15.0",
			floor:   "0.14.0",
			schema:  installed,
			wantErr: "Cannot upgrade directly from " + model.LastUngatedVersion + " to 0.15.0",
		},
		{
			name:    "release candidate over a legacy install",
			version: model.LastUngatedVersion + "-rc4",
			floor:   model.LastUngatedVersion,
			schema:  installed,
		},
		{
			name:    "legacy install below the floor",
			version: "0.20.0",
			floor:   "0.15.0",
			schema:  installed,
			wantErr: "Cannot upgrade directly from " + model.LastUngatedVersion,
		},
		{
			name:     "downgrade",
			version:  "0.15.0",
			floor:    "0.14.0",
			recorded: "0.16.0",
			schema:   installed,
			wantErr:  "newer than this 0.15.0 binary",
		},
		{
			name:    "half applied schema",
			version: "0.15.0",
			floor:   "0.14.0",
			schema:  Schema{Installed: true, Version: 88, Dirty: true},
			wantErr: "Migration 88 did not finish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setBuild(t, tt.version, tt.floor)

			err := Check(tt.recorded, tt.schema)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Check(%q, %+v) = %v, want nil", tt.recorded, tt.schema, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Check(%q, %+v) = nil, want error", tt.recorded, tt.schema)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
