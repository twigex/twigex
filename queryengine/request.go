// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "time"

type Request struct {
	Dimensions []string
	Measures   []string
	DateFields []DateField
	Filters    []Filter
	SortBy     []SortField
	RowLimit   int

	// RowsOnly returns one row per database row, without adding rows up.
	// Only dimensions can be asked for.
	RowsOnly bool

	// Now is the moment relative date ranges such as "this month" are worked
	// out from. Empty means the current time.
	Now time.Time
}

type DateStep string

const (
	DateStepSecond  DateStep = "second"
	DateStepMinute  DateStep = "minute"
	DateStepHour    DateStep = "hour"
	DateStepDay     DateStep = "day"
	DateStepWeek    DateStep = "week"
	DateStepMonth   DateStep = "month"
	DateStepQuarter DateStep = "quarter"
	DateStepYear    DateStep = "year"
)

// DateField groups rows by a date field, one row per step, and keeps only
// rows inside its range. Without a step it only keeps rows inside the range,
// and the field is not shown.
type DateField struct {
	Field string
	Step  DateStep
	Range DateRange
}

// DateRange is either a named range such as "this month", or a start and an
// end. A start or end without a time means the whole day. An empty DateRange
// keeps every row.
type DateRange struct {
	Named string
	Start string
	End   string
}

type FilterOperator string

const (
	FilterEquals         FilterOperator = "equals"
	FilterNotEquals      FilterOperator = "not_equals"
	FilterContains       FilterOperator = "contains"
	FilterNotContains    FilterOperator = "not_contains"
	FilterStartsWith     FilterOperator = "starts_with"
	FilterEndsWith       FilterOperator = "ends_with"
	FilterGreaterThan    FilterOperator = "greater_than"
	FilterGreaterOrEqual FilterOperator = "greater_or_equal"
	FilterLessThan       FilterOperator = "less_than"
	FilterLessOrEqual    FilterOperator = "less_or_equal"
	FilterIsSet          FilterOperator = "is_set"
	FilterIsNotSet       FilterOperator = "is_not_set"
	FilterInDateRange    FilterOperator = "in_date_range"
	FilterNotInDateRange FilterOperator = "not_in_date_range"
	FilterBeforeDate     FilterOperator = "before_date"
	FilterAfterDate      FilterOperator = "after_date"
)

// Filter is either one condition on a field, or a group: AllOf when every
// filter in it must match, AnyOf when at least one must.
type Filter struct {
	Field    string
	Operator FilterOperator
	Values   []string

	AllOf []Filter
	AnyOf []Filter
}

type SortField struct {
	Field      string
	Descending bool
}
