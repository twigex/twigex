// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
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
