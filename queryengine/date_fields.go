// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "time"

type requestedDates struct {
	groupedFields []requestedField
	conditions    []string
	values        []any
	ranges        []ResolvedDateRange
}

// findDateFields reads the request's date fields: those with a step are
// shown and grouped by, and every range becomes a condition on the rows.
func (builder *queryBuilder) findDateFields() (requestedDates, error) {
	dates := requestedDates{}

	now := builder.request.Now
	if now.IsZero() {
		now = time.Now()
	}

	for _, dateField := range builder.request.DateFields {
		field, err := builder.readFieldOnce(dateField.Field)
		if err != nil {
			return requestedDates{}, err
		}

		if err := checkDateField(dateField, field); err != nil {
			return requestedDates{}, err
		}

		if dateField.Step != "" {
			groupedField := field
			groupedField.dateStep = dateField.Step

			dates.groupedFields = append(dates.groupedFields, groupedField)
		}

		dateRange, hasRange, err := resolveDateRange(dateField.Field, dateField.Range, now)
		if err != nil {
			return requestedDates{}, err
		}

		if !hasRange {
			continue
		}

		if err := builder.addDateRange(&dates, field, dateRange); err != nil {
			return requestedDates{}, err
		}
	}

	return dates, nil
}

func checkDateField(dateField DateField, field requestedField) error {
	if field.isMeasure() {
		return newRequestError(ErrorWrongFieldKind, dateField.Field)
	}

	if field.dimension.Type != FieldTypeTime {
		return newRequestError(ErrorNotADateField, dateField.Field)
	}

	if dateField.Step != "" && !isKnownDateStep(dateField.Step) {
		return newRequestError(ErrorUnknownDateStep, dateField.Field)
	}

	return nil
}

// addDateRange keeps only the rows in a date field's range, and records the
// range's actual start and end for the result.
func (builder *queryBuilder) addDateRange(dates *requestedDates, field requestedField, dateRange dateRangeMoments) error {
	storedValue, err := builder.storedValueExpression(field)
	if err != nil {
		return err
	}

	condition := dateRangeCondition(builder.syntax, storedValue, dateRange)
	dates.conditions = append(dates.conditions, condition.condition)
	dates.values = append(dates.values, condition.values...)

	resolved := ResolvedDateRange{
		Field: field.key,
		Start: dateForResult(dateRange.start),
		End:   dateForResult(dateRange.end),
	}

	dates.ranges = append(dates.ranges, resolved)

	return nil
}
