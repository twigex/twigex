// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "time"

// The date conditions are the same for every database; only where a date
// value goes and how it is written differ (sqlSyntax.datePlaceholder, dateForSQL).

func dateRangeCondition(syntax sqlSyntax, expression string, dateRange dateRangeMoments) filterSQL {
	placeholder := syntax.datePlaceholder()
	end := endBoundOf(dateRange.end)

	condition := "(" + expression + " >= " + placeholder + " AND " + expression + " " + end.upTo + " " + placeholder + ")"
	values := []any{syntax.dateForSQL(dateRange.start), syntax.dateForSQL(end.moment)}

	return filterSQL{condition: condition, values: values}
}

func outsideDateRangeCondition(syntax sqlSyntax, expression string, dateRange dateRangeMoments) filterSQL {
	placeholder := syntax.datePlaceholder()
	end := endBoundOf(dateRange.end)

	condition := "(" + expression + " < " + placeholder + " OR " + expression + " " + end.after + " " + placeholder + ")"
	values := []any{syntax.dateForSQL(dateRange.start), syntax.dateForSQL(end.moment)}

	return filterSQL{condition: condition, values: values}
}

// endBound is how a condition compares with the end of a range: up to and
// including the end, or after it.
type endBound struct {
	moment time.Time
	upTo   string
	after  string
}

// endBoundOf compares an end that is the last millisecond of a day with the
// start of the next day instead. Times can hold microseconds, and a row at
// 23:59:59.9995 belongs to the day though it is after 23:59:59.999.
func endBoundOf(end time.Time) endBound {
	nextMoment := end.Add(time.Millisecond)

	isEndOfDay := nextMoment.Hour() == 0 && nextMoment.Minute() == 0 && nextMoment.Second() == 0 && nextMoment.Nanosecond() == 0
	if isEndOfDay {
		return endBound{moment: nextMoment, upTo: "<", after: ">="}
	}

	return endBound{moment: end, upTo: "<=", after: ">"}
}

// writeDateCondition writes the date filters. "Before" a date means before
// the start of that day, and "after" means after its end.
func writeDateCondition(syntax sqlSyntax, expression string, filter Filter) (filterSQL, error) {
	switch filter.Operator {
	case FilterInDateRange, FilterNotInDateRange:
		return writeDateRangeFilter(syntax, expression, filter)
	case FilterBeforeDate:
		return writeDateComparison(syntax, expression, filter, " < ", rangeStart)
	case FilterAfterDate:
		return writeAfterDate(syntax, expression, filter)
	}

	return filterSQL{}, newRequestError(ErrorUnknownFilterType, filter.Field)
}

func writeDateRangeFilter(syntax sqlSyntax, expression string, filter Filter) (filterSQL, error) {
	if len(filter.Values) < 2 {
		return filterSQL{}, newRequestError(ErrorMissingFilterValue, filter.Field)
	}

	start, err := rangeStart(filter.Field, filter.Values[0])
	if err != nil {
		return filterSQL{}, err
	}

	end, err := rangeEnd(filter.Field, filter.Values[1])
	if err != nil {
		return filterSQL{}, err
	}

	dateRange := dateRangeMoments{start: start, end: end}

	if filter.Operator == FilterInDateRange {
		return dateRangeCondition(syntax, expression, dateRange), nil
	}

	return outsideDateRangeCondition(syntax, expression, dateRange), nil
}

// dateReader reads a date a filter gives as the moment it starts or ends.
type dateReader func(field string, text string) (time.Time, error)

// writeDateComparison compares with one date: its start or its end, as
// readDate gives it.
func writeDateComparison(syntax sqlSyntax, expression string, filter Filter, comparison string, readDate dateReader) (filterSQL, error) {
	if len(filter.Values) == 0 {
		return filterSQL{}, newRequestError(ErrorMissingFilterValue, filter.Field)
	}

	moment, err := readDate(filter.Field, filter.Values[0])
	if err != nil {
		return filterSQL{}, err
	}

	condition := expression + comparison + syntax.datePlaceholder()
	values := []any{syntax.dateForSQL(moment)}

	return filterSQL{condition: condition, values: values}, nil
}

// writeAfterDate compares with the end of the date, as dateRangeCondition
// does.
func writeAfterDate(syntax sqlSyntax, expression string, filter Filter) (filterSQL, error) {
	if len(filter.Values) == 0 {
		return filterSQL{}, newRequestError(ErrorMissingFilterValue, filter.Field)
	}

	moment, err := rangeEnd(filter.Field, filter.Values[0])
	if err != nil {
		return filterSQL{}, err
	}

	end := endBoundOf(moment)

	condition := expression + " " + end.after + " " + syntax.datePlaceholder()
	values := []any{syntax.dateForSQL(end.moment)}

	return filterSQL{condition: condition, values: values}, nil
}

func isDateOperator(operator FilterOperator) bool {
	switch operator {
	case FilterInDateRange, FilterNotInDateRange, FilterBeforeDate, FilterAfterDate:
		return true
	}

	return false
}

func isKnownDateStep(step DateStep) bool {
	switch step {
	case DateStepSecond, DateStepMinute, DateStepHour, DateStepDay, DateStepWeek, DateStepMonth, DateStepQuarter, DateStepYear:
		return true
	}

	return false
}
