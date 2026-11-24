// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"fmt"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

var engineFilterOperators = map[string]queryengine.FilterOperator{
	"equals":         queryengine.FilterEquals,
	"notEquals":      queryengine.FilterNotEquals,
	"contains":       queryengine.FilterContains,
	"notContains":    queryengine.FilterNotContains,
	"startsWith":     queryengine.FilterStartsWith,
	"endsWith":       queryengine.FilterEndsWith,
	"gt":             queryengine.FilterGreaterThan,
	"gte":            queryengine.FilterGreaterOrEqual,
	"lt":             queryengine.FilterLessThan,
	"lte":            queryengine.FilterLessOrEqual,
	"set":            queryengine.FilterIsSet,
	"notSet":         queryengine.FilterIsNotSet,
	"inDateRange":    queryengine.FilterInDateRange,
	"notInDateRange": queryengine.FilterNotInDateRange,
	"beforeDate":     queryengine.FilterBeforeDate,
	"afterDate":      queryengine.FilterAfterDate,
}

// EngineRequest turns a chart's request into a request for the query engine.
func EngineRequest(query model.DataQuery) (queryengine.Request, error) {
	dateFields, err := engineDateFields(query.TimeDimensions)
	if err != nil {
		return queryengine.Request{}, err
	}

	filters, err := engineFilters(query.Filters)
	if err != nil {
		return queryengine.Request{}, err
	}

	sortBy, err := engineSortFields(query.Order)
	if err != nil {
		return queryengine.Request{}, err
	}

	request := queryengine.Request{
		Dimensions: query.Dimensions,
		Measures:   query.Measures,
		DateFields: dateFields,
		Filters:    filters,
		SortBy:     sortBy,
		RowLimit:   query.Limit,
		RowsOnly:   query.Ungrouped,
	}

	return request, nil
}

func engineFilters(savedFilters []model.QueryFilter) ([]queryengine.Filter, error) {
	filters := make([]queryengine.Filter, 0, len(savedFilters))

	for _, savedFilter := range savedFilters {
		filter, err := engineFilter(savedFilter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, filter)
	}

	return filters, nil
}

func engineFilter(savedFilter model.QueryFilter) (queryengine.Filter, error) {
	isGroup := len(savedFilter.And) > 0 || len(savedFilter.Or) > 0
	if isGroup {
		return engineFilterGroup(savedFilter)
	}

	operator, known := engineFilterOperators[savedFilter.Operator]
	if !known {
		return queryengine.Filter{}, &queryengine.RequestError{Code: queryengine.ErrorUnknownFilterType, Field: savedFilter.Member}
	}

	filter := queryengine.Filter{
		Field:    savedFilter.Member,
		Operator: operator,
		Values:   savedFilter.Values,
	}

	return filter, nil
}

func engineFilterGroup(savedFilter model.QueryFilter) (queryengine.Filter, error) {
	allOf, err := engineFilters(savedFilter.And)
	if err != nil {
		return queryengine.Filter{}, err
	}

	anyOf, err := engineFilters(savedFilter.Or)
	if err != nil {
		return queryengine.Filter{}, err
	}

	group := queryengine.Filter{
		AllOf: allOf,
		AnyOf: anyOf,
	}

	return group, nil
}

// engineSortFields reads the sort order, which charts send as pairs such as
// ["orders.count", "desc"].
func engineSortFields(order [][]string) ([]queryengine.SortField, error) {
	sortFields := make([]queryengine.SortField, 0, len(order))

	for _, pair := range order {
		if len(pair) != 2 {
			return nil, fmt.Errorf("sort entry %v is not a field and a direction", pair)
		}

		field := pair[0]
		direction := pair[1]

		if direction != "asc" && direction != "desc" {
			return nil, fmt.Errorf("sort direction %q for %s is not asc or desc", direction, field)
		}

		sortField := queryengine.SortField{
			Field:      field,
			Descending: direction == "desc",
		}

		sortFields = append(sortFields, sortField)
	}

	return sortFields, nil
}

func engineDateFields(timeDimensions []model.QueryTimeDimension) ([]queryengine.DateField, error) {
	dateFields := make([]queryengine.DateField, 0, len(timeDimensions))

	for _, timeDimension := range timeDimensions {
		dateRange, err := engineDateRange(timeDimension)
		if err != nil {
			return nil, err
		}

		dateField := queryengine.DateField{
			Field: timeDimension.Dimension,
			Step:  queryengine.DateStep(timeDimension.Granularity),
			Range: dateRange,
		}

		dateFields = append(dateFields, dateField)
	}

	return dateFields, nil
}

// engineDateRange reads a chart's date range, which is either a name such as
// "this month" or a start and an end date.
func engineDateRange(timeDimension model.QueryTimeDimension) (queryengine.DateRange, error) {
	savedRange := timeDimension.DateRange

	if !savedRange.IsArray {
		return queryengine.DateRange{Named: savedRange.StringValue}, nil
	}

	if len(savedRange.ArrayValue) != 2 {
		return queryengine.DateRange{}, &queryengine.RequestError{Code: queryengine.ErrorBadDate, Field: timeDimension.Dimension}
	}

	dateRange := queryengine.DateRange{
		Start: savedRange.ArrayValue[0],
		End:   savedRange.ArrayValue[1],
	}

	return dateRange, nil
}
