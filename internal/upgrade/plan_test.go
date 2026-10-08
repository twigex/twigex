// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func useStops(t *testing.T, stops []Stop) {
	t.Helper()

	original := Stops
	Stops = stops
	t.Cleanup(func() { Stops = original })
}

func fixture() []Stop {
	return []Stop{
		{"0.15.0", "0.14.0", "collapses per-descendant share rows"},
		{"0.19.0", "0.15.0", "moves chat attachments into per-storage paths"},
	}
}

func TestPlan(t *testing.T) {
	useStops(t, fixture())

	tests := []struct {
		name string
		cur  string
		bin  string
		want []Step
	}{
		{
			name: "walks back through every floor",
			cur:  "0.13.0",
			bin:  "0.20.0",
			want: []Step{
				{"0.14.0", "required by 0.15.0"},
				{"0.15.0", "collapses per-descendant share rows"},
			},
		},
		{
			name: "stops as soon as the floor is satisfied",
			cur:  "0.14.0",
			bin:  "0.20.0",
			want: []Step{{"0.15.0", "collapses per-descendant share rows"}},
		},
		{
			name: "direct when the recorded version clears the floor",
			cur:  "0.15.0",
			bin:  "0.20.0",
		},
		{
			name: "direct when no stop sits between the two",
			cur:  "0.14.0",
			bin:  "0.16.0",
		},
		{
			name: "direct when already ahead of every stop",
			cur:  "0.19.0",
			bin:  "0.20.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Plan(tt.cur, tt.bin)
			if len(got) != len(tt.want) {
				t.Fatalf("Plan(%q, %q) = %v, want %v", tt.cur, tt.bin, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("step %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestPlanEmptyTable(t *testing.T) {
	useStops(t, nil)

	if got := Plan("0.10.0", "0.20.0"); len(got) != 0 {
		t.Errorf("Plan with no stops = %v, want empty", got)
	}
}

func TestDecide(t *testing.T) {
	useStops(t, fixture())

	tests := []struct {
		name    string
		cur     string
		bin     string
		floor   string
		assumed bool
		wantErr string
	}{
		{name: "same version", cur: "0.15.0", bin: "0.15.0", floor: "0.15.0"},
		{name: "clears the floor", cur: "0.15.0", bin: "0.20.0", floor: "0.15.0"},
		{name: "default floor blocks nothing", cur: "0.13.0", bin: "0.20.0", floor: "0.0.0"},
		{name: "unset floor blocks nothing", cur: "0.13.0", bin: "0.20.0", floor: ""},
		{
			name:    "assumed version one release ahead of a release candidate",
			cur:     "0.14.0",
			bin:     "0.14.0-rc4",
			floor:   "0.14.0",
			assumed: true,
		},
		{
			name:    "downgrade",
			cur:     "0.16.0",
			bin:     "0.15.0",
			floor:   "0.14.0",
			wantErr: "newer than this 0.15.0 binary",
		},
		{
			name:    "skips a release",
			cur:     "0.13.0",
			bin:     "0.20.0",
			floor:   "0.15.0",
			wantErr: "Cannot upgrade directly from 0.13.0 to 0.20.0",
		},
		{
			name:    "assumed version still meets the floor",
			cur:     "0.14.0",
			bin:     "0.20.0",
			floor:   "0.15.0",
			assumed: true,
			wantErr: "Cannot upgrade directly from 0.14.0 to 0.20.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decide(tt.cur, tt.bin, tt.floor, tt.assumed)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("decide(%q, %q, %q) = %v, want nil", tt.cur, tt.bin, tt.floor, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("decide(%q, %q, %q) = nil, want error", tt.cur, tt.bin, tt.floor)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestBlockedMessageListsThePath(t *testing.T) {
	useStops(t, fixture())

	got := blockedMessage("0.13.0", "0.20.0", "0.15.0")

	for _, want := range []string{
		"1. 0.14.0  required by 0.15.0",
		"2. 0.15.0  collapses per-descendant share rows",
		"3. 0.20.0  this release",
		"Install 0.14.0, start it once",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("message missing %q:\n%s", want, got)
		}
	}
}

func TestBlockedMessageWithoutAPath(t *testing.T) {
	useStops(t, nil)

	got := blockedMessage("0.10.0", "0.20.0", "0.15.0")

	if !strings.Contains(got, "No upgrade path is available") {
		t.Errorf("message should admit there is no path:\n%s", got)
	}
}

func TestStopsInvariants(t *testing.T) {
	for i, s := range Stops {
		if Compare(s.MinFrom, s.Version) >= 0 {
			t.Errorf("stop %s: MinFrom %s is not below it", s.Version, s.MinFrom)
		}
		if Compare(s.Version, model.LastUngatedVersion) <= 0 {
			t.Errorf("stop %s is not above LastUngatedVersion %s, so installations "+
				"assumed to be on that version would skip it", s.Version, model.LastUngatedVersion)
		}
		if s.Reason == "" {
			t.Errorf("stop %s has no reason", s.Version)
		}
		if i == 0 {
			continue
		}

		prev := Stops[i-1]
		if Compare(prev.Version, s.Version) >= 0 {
			t.Errorf("stops are out of order: %s precedes %s", prev.Version, s.Version)
		}
		if Compare(prev.MinFrom, s.MinFrom) > 0 {
			t.Errorf("floor went backwards: %s demands %s but %s demands %s",
				prev.Version, prev.MinFrom, s.Version, s.MinFrom)
		}
	}

	// Only a release build carries the ldflags this compares.
	if !IsReal(model.Version) {
		return
	}

	switch floor := floorFor(model.Version); {
	case floor == "":
		if Compare(model.MinUpgradableVersion, model.LastUngatedVersion) > 0 {
			t.Errorf("MIN_UPGRADABLE_VERSION is %s with no stop to match, so an install "+
				"assumed to be on %s would be refused with no path to offer",
				model.MinUpgradableVersion, model.LastUngatedVersion)
		}
	case Compare(model.MinUpgradableVersion, floor) != 0:
		t.Errorf("MIN_UPGRADABLE_VERSION is %s but the stop table says %s for %s",
			model.MinUpgradableVersion, floor, model.Version)
	}
}
