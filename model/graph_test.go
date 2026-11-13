// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestGraphQueryFieldsListsEveryFieldTheQueryReads(t *testing.T) {
	query := DataQuery{
		Measures:   []string{"orders.count"},
		Dimensions: []string{"orders.status", "orders.created_at.day"},
		TimeDimensions: []QueryTimeDimension{
			{Dimension: "orders.shipped_at", Granularity: "month"},
		},
		Filters: []QueryFilter{
			{Member: "orders.country", Operator: "equals", Values: []string{"LV"}},
			{Or: []QueryFilter{
				{Member: "orders.channel", Operator: "equals", Values: []string{"web"}},
				{And: []QueryFilter{
					{Member: "orders.amount", Operator: "gt", Values: []string{"10"}},
				}},
			}},
		},
		Order: [][]string{{"orders.priority", "desc"}},
	}

	got := GraphQueryFields(query)

	want := []string{
		"orders.count",
		"orders.status",
		"orders.created_at",
		"orders.shipped_at",
		"orders.country",
		"orders.channel",
		"orders.amount",
		"orders.priority",
	}

	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestGraphQueryFieldsListsAFieldOnce(t *testing.T) {
	query := DataQuery{
		Dimensions: []string{"orders.status"},
		Filters: []QueryFilter{
			{Member: "orders.status", Operator: "equals", Values: []string{"paid"}},
		},
		Order: [][]string{{"orders.status", "asc"}},
	}

	got := GraphQueryFields(query)

	want := []string{"orders.status"}

	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestGraphQueryFieldsIgnoresAnEmptyOrder(t *testing.T) {
	query := DataQuery{
		Measures: []string{"orders.count"},
		Order:    [][]string{{}},
	}

	got := GraphQueryFields(query)

	want := []string{"orders.count"}

	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestGraphMapFieldsListsEveryFieldTheLayersRead(t *testing.T) {
	raw := `{
		"zoom": 3.6,
		"layers": [
			{
				"type": "point",
				"model": "stores",
				"filters": [{"member": "stores.country", "operator": "equals", "values": ["LV"]}],
				"config": {
					"fill_color_field": "stores.revenue",
					"radius_field": "stores.revenue",
					"coordinates": {"lat": "stores.latitude", "long": "stores.longitude"},
					"popover": {"fields": ["stores.name"]},
					"data": null
				}
			},
			{
				"type": "arc",
				"model": "routes",
				"filters": [],
				"config": {
					"source": {"lat": "routes.from_lat", "long": "routes.from_lng"},
					"target": {"lat": "routes.to_lat", "long": "routes.to_lng"},
					"source_color_field": null,
					"width_field": "routes.shipments",
					"popover": {"fields": []},
					"data": []
				}
			}
		]
	}`

	configuration := GraphMapConfiguration{}

	err := json.Unmarshal([]byte(raw), &configuration)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	got := GraphMapFields(configuration)

	want := []string{
		"stores.count",
		"routes.count",
		"stores.latitude",
		"stores.longitude",
		"stores.revenue",
		"stores.name",
		"routes.from_lat",
		"routes.from_lng",
		"routes.to_lat",
		"routes.to_lng",
		"routes.shipments",
		"stores.country",
	}

	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}
