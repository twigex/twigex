// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package parse

import "testing"

func TestInt(t *testing.T) {
	tests := []struct {
		val        string
		defaultVal int
		want       int
	}{
		{"12", 50, 12},
		{"-5", 50, -5},
		{"", 50, 50},
		{"12abc", 50, 50},
		{"abc", 50, 50},
		{" 12", 50, 50},
		{"12.9", 50, 50},
	}

	for _, tt := range tests {
		if got := Int(tt.val, tt.defaultVal); got != tt.want {
			t.Errorf("Int(%q, %d) = %d, want %d", tt.val, tt.defaultVal, got, tt.want)
		}
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		val        string
		defaultVal bool
		want       bool
	}{
		{"true", false, true},
		{"TRUE", false, true},
		{"", true, true},
		{"", false, false},
		{"false", true, false},
		{"1", false, false},
		{"yes", false, false},
	}

	for _, tt := range tests {
		if got := Bool(tt.val, tt.defaultVal); got != tt.want {
			t.Errorf("Bool(%q, %t) = %t, want %t", tt.val, tt.defaultVal, got, tt.want)
		}
	}
}
