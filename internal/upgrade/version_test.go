// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"0.14.0", "0.14.0", 0},
		{"0.14.0", "0.15.0", -1},
		{"0.15.0", "0.14.0", 1},
		{"0.9.0", "0.14.0", -1},
		{"1.0.0", "0.99.99", 1},
		{"0.14", "0.14.0", 0},
		{"0.14.1", "0.14", 1},
		{"0.14.0-rc4", "0.14.0", -1},
		{"0.14.0", "0.14.0-rc4", 1},
		{"0.14.0-rc2", "0.14.0-rc4", -1},
		{"0.14.0-rc4", "0.14.0-rc4", 0},
		{"0.14.0-rc4", "0.15.0-rc1", -1},
		{"", "0.0.0", 0},
	}

	for _, tt := range tests {
		if got := Compare(tt.a, tt.b); got != tt.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsReal(t *testing.T) {
	tests := []struct {
		v    string
		want bool
	}{
		{"0.14.0", true},
		{"0.14.0-rc4", true},
		{"1.0.0", true},
		{"dev", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsReal(tt.v); got != tt.want {
			t.Errorf("IsReal(%q) = %t, want %t", tt.v, got, tt.want)
		}
	}
}
