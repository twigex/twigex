// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"errors"
	"reflect"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

func TestEngineRequestConvertsFieldsSortAndLimit(t *testing.T) {
	query := model.DataQuery{
		Dimensions: []string{"orders.country"},
		Measures:   []string{"orders.count"},
		Order:      [][]string{{"orders.count", "desc"}, {"orders.country", "asc"}},
		Limit:      500,
		Ungrouped:  false,
	}

	request, err := EngineRequest(query)
	if err != nil {
		t.Fatalf("EngineRequest: %v", err)
	}

	want := queryengine.Request{
		Dimensions: []string{"orders.country"},
		Measures:   []string{"orders.count"},
		DateFields: []queryengine.DateField{},
		Filters:    []queryengine.Filter{},
		SortBy: []queryengine.SortField{
			{Field: "orders.count", Descending: true},
			{Field: "orders.country", Descending: false},
		},
		RowLimit: 500,
	}

	if !reflect.DeepEqual(request, want) {
		t.Errorf("want\n%+v\ngot\n%+v", want, request)
	}
}

func TestEngineRequestUngroupedBecomesRowsOnly(t *testing.T) {
	query := model.DataQuery{
		Dimensions: []string{"orders.country"},
		Limit:      10,
		Ungrouped:  true,
	}

	request, err := EngineRequest(query)
	if err != nil {
		t.Fatalf("EngineRequest: %v", err)
	}

	if !request.RowsOnly {
		t.Error("ungrouped did not become RowsOnly")
	}
}

func TestEngineRequestConvertsEveryFilterType(t *testing.T) {
	cases := []struct {
		savedOperator string
		want          queryengine.FilterOperator
	}{
		{"equals", queryengine.FilterEquals},
		{"notEquals", queryengine.FilterNotEquals},
		{"contains", queryengine.FilterContains},
		{"notContains", queryengine.FilterNotContains},
		{"startsWith", queryengine.FilterStartsWith},
		{"endsWith", queryengine.FilterEndsWith},
		{"gt", queryengine.FilterGreaterThan},
		{"gte", queryengine.FilterGreaterOrEqual},
		{"lt", queryengine.FilterLessThan},
		{"lte", queryengine.FilterLessOrEqual},
		{"set", queryengine.FilterIsSet},
		{"notSet", queryengine.FilterIsNotSet},
		{"inDateRange", queryengine.FilterInDateRange},
		{"notInDateRange", queryengine.FilterNotInDateRange},
		{"beforeDate", queryengine.FilterBeforeDate},
		{"afterDate", queryengine.FilterAfterDate},
	}

	for _, testCase := range cases {
		t.Run(testCase.savedOperator, func(t *testing.T) {
			query := model.DataQuery{
				Measures: []string{"orders.count"},
				Filters: []model.QueryFilter{
					{Member: "orders.country", Operator: testCase.savedOperator, Values: []string{"Latvia"}},
				},
				Limit: 10,
			}

			request, err := EngineRequest(query)
			if err != nil {
				t.Fatalf("EngineRequest: %v", err)
			}

			want := []queryengine.Filter{
				{Field: "orders.country", Operator: testCase.want, Values: []string{"Latvia"}},
			}

			if !reflect.DeepEqual(request.Filters, want) {
				t.Errorf("want %+v, got %+v", want, request.Filters)
			}
		})
	}
}

func TestEngineRequestConvertsFilterGroups(t *testing.T) {
	query := model.DataQuery{
		Measures: []string{"orders.count"},
		Filters: []model.QueryFilter{
			{
				And: []model.QueryFilter{
					{Member: "orders.channel", Operator: "equals", Values: []string{"Shop"}},
					{
						Or: []model.QueryFilter{
							{Member: "orders.country", Operator: "notSet"},
							{Member: "orders.country", Operator: "equals", Values: []string{"Latvia"}},
						},
					},
				},
			},
		},
		Limit: 10,
	}

	request, err := EngineRequest(query)
	if err != nil {
		t.Fatalf("EngineRequest: %v", err)
	}

	want := []queryengine.Filter{
		{
			AllOf: []queryengine.Filter{
				{Field: "orders.channel", Operator: queryengine.FilterEquals, Values: []string{"Shop"}},
				{
					AllOf: []queryengine.Filter{},
					AnyOf: []queryengine.Filter{
						{Field: "orders.country", Operator: queryengine.FilterIsNotSet},
						{Field: "orders.country", Operator: queryengine.FilterEquals, Values: []string{"Latvia"}},
					},
				},
			},
			AnyOf: []queryengine.Filter{},
		},
	}

	if !reflect.DeepEqual(request.Filters, want) {
		t.Errorf("want\n%+v\ngot\n%+v", want, request.Filters)
	}
}

func TestEngineRequestErrors(t *testing.T) {
	threeDates := model.QueryDateRange{IsArray: true, ArrayValue: []string{"2025-01-01", "2025-02-01", "2025-03-01"}}
	badDateField := []model.QueryTimeDimension{{Dimension: "orders.ordered_at", DateRange: threeDates}}

	cases := []struct {
		name      string
		query     model.DataQuery
		wantCode  queryengine.ErrorCode
		wantField string
	}{
		{
			"date range that is not two dates",
			model.DataQuery{TimeDimensions: badDateField, Limit: 10},
			queryengine.ErrorBadDate, "orders.ordered_at",
		},
		{
			"unknown filter type",
			model.DataQuery{Limit: 10, Filters: []model.QueryFilter{{Member: "orders.country", Operator: "between"}}},
			queryengine.ErrorUnknownFilterType, "orders.country",
		},
		{
			"unknown filter type inside a group",
			model.DataQuery{Limit: 10, Filters: []model.QueryFilter{{Or: []model.QueryFilter{{Member: "orders.country", Operator: "between"}}}}},
			queryengine.ErrorUnknownFilterType, "orders.country",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := EngineRequest(testCase.query)

			var requestError *queryengine.RequestError
			if !errors.As(err, &requestError) {
				t.Fatalf("want a RequestError, got %v", err)
			}

			if requestError.Code != testCase.wantCode {
				t.Errorf("want code %q, got %q", testCase.wantCode, requestError.Code)
			}

			if requestError.Field != testCase.wantField {
				t.Errorf("want field %q, got %q", testCase.wantField, requestError.Field)
			}
		})
	}
}

func TestEngineRequestRejectsBadSortEntries(t *testing.T) {
	cases := []struct {
		name  string
		order [][]string
	}{
		{"missing direction", [][]string{{"orders.count"}}},
		{"unknown direction", [][]string{{"orders.count", "up"}}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			query := model.DataQuery{Measures: []string{"orders.count"}, Order: testCase.order, Limit: 10}

			_, err := EngineRequest(query)
			if err == nil {
				t.Error("want an error, got none")
			}
		})
	}
}

func TestEngineRequestConvertsDateFields(t *testing.T) {
	query := model.DataQuery{
		Measures: []string{"orders.count"},
		TimeDimensions: []model.QueryTimeDimension{
			{
				Dimension:   "orders.ordered_at",
				Granularity: "month",
				DateRange:   model.QueryDateRange{StringValue: "this year"},
			},
			{
				Dimension: "orders.shipped_at",
				DateRange: model.QueryDateRange{IsArray: true, ArrayValue: []string{"2025-01-01", "2025-03-31"}},
			},
			{
				Dimension:   "orders.paid_at",
				Granularity: "week",
			},
		},
		Limit: 10,
	}

	request, err := EngineRequest(query)
	if err != nil {
		t.Fatalf("EngineRequest: %v", err)
	}

	want := []queryengine.DateField{
		{
			Field: "orders.ordered_at",
			Step:  queryengine.DateStepMonth,
			Range: queryengine.DateRange{Named: "this year"},
		},
		{
			Field: "orders.shipped_at",
			Range: queryengine.DateRange{Start: "2025-01-01", End: "2025-03-31"},
		},
		{
			Field: "orders.paid_at",
			Step:  queryengine.DateStepWeek,
		},
	}

	if !reflect.DeepEqual(request.DateFields, want) {
		t.Errorf("want\n%+v\ngot\n%+v", want, request.DateFields)
	}
}
