// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestFilterReadsAListValueAsValues(t *testing.T) {
	for _, tc := range []struct {
		name       string
		json       string
		wantValue  string
		wantValues []string
	}{
		{"text", `{"field":"name","operator":"is","value":"Alpha"}`, "Alpha", nil},
		{"list", `{"field":"status","operator":"is","value":["s1","s2"]}`, "", []string{"s1", "s2"}},
		{"list next to values", `{"field":"status","operator":"is","value":["s1"],"values":["s9"]}`, "", []string{"s9"}},
		{"number", `{"field":"points","operator":"is","value":5}`, "5", nil},
		{"null", `{"field":"name","operator":"is_set","value":null}`, "", nil},
		{"missing", `{"field":"name","operator":"is_set"}`, "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var f Filter
			if err := json.Unmarshal([]byte(tc.json), &f); err != nil {
				t.Fatal(err)
			}

			if f.Value != tc.wantValue || !slices.Equal(f.Values, tc.wantValues) || f.Operator == "" {
				t.Errorf("filter = %+v, want value %q and values %v", f, tc.wantValue, tc.wantValues)
			}
		})
	}

	var payload FilterPayload
	if err := json.Unmarshal([]byte(`{"groups":[{"filters":[{"field":"status","operator":"is_not","value":["s1"]}]}]}`), &payload); err != nil {
		t.Fatal(err)
	}

	if got := payload.Groups[0].Filters[0].Values; !slices.Equal(got, []string{"s1"}) {
		t.Errorf("grouped filter values = %v, want [s1]", got)
	}
}
