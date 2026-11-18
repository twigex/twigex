// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

type filterSQL struct {
	condition    string
	values       []any
	isOnMeasures bool
}

type requestFilters struct {
	whereConditions  []string
	whereValues      []any
	havingConditions []string
	havingValues     []any
}

// filterConditions turns the request's filters into SQL conditions. Filters
// on fields go in WHERE; filters on measures go in HAVING, because they apply
// after rows are added up.
func (builder *queryBuilder) filterConditions() (requestFilters, error) {
	filters := requestFilters{}

	for _, filter := range builder.request.Filters {
		written, err := builder.writeFilter(filter)
		if err != nil {
			return requestFilters{}, err
		}

		if written.isOnMeasures {
			filters.havingConditions = append(filters.havingConditions, written.condition)
			filters.havingValues = append(filters.havingValues, written.values...)
			continue
		}

		filters.whereConditions = append(filters.whereConditions, written.condition)
		filters.whereValues = append(filters.whereValues, written.values...)
	}

	return filters, nil
}

func (builder *queryBuilder) writeFilter(filter Filter) (filterSQL, error) {
	isAllOfGroup := len(filter.AllOf) > 0
	isAnyOfGroup := len(filter.AnyOf) > 0

	if isAllOfGroup && isAnyOfGroup {
		return filterSQL{}, newRequestError(ErrorBadFilterGroup, "")
	}

	if isAllOfGroup {
		return builder.writeFilterGroup(filter.AllOf, " AND ")
	}

	if isAnyOfGroup {
		return builder.writeFilterGroup(filter.AnyOf, " OR ")
	}

	if filter.Field == "" {
		return filterSQL{}, newRequestError(ErrorBadFilterGroup, "")
	}

	return builder.writeFieldFilter(filter)
}

// writeFilterGroup joins the filters of a group. A group cannot mix fields and
// measures, because one part would belong in WHERE and the other in HAVING.
func (builder *queryBuilder) writeFilterGroup(filters []Filter, joinWord string) (filterSQL, error) {
	var conditions []string
	group := filterSQL{}

	for position, filter := range filters {
		written, err := builder.writeFilter(filter)
		if err != nil {
			return filterSQL{}, err
		}

		if position == 0 {
			group.isOnMeasures = written.isOnMeasures
		}

		if written.isOnMeasures != group.isOnMeasures {
			return filterSQL{}, newRequestError(ErrorMixedFilterGroup, filter.Field)
		}

		conditions = append(conditions, written.condition)
		group.values = append(group.values, written.values...)
	}

	group.condition = "(" + strings.Join(conditions, joinWord) + ")"

	return group, nil
}

func (builder *queryBuilder) writeFieldFilter(filter Filter) (filterSQL, error) {
	field := builder.field(filter.Field)

	if builder.request.RowsOnly && field.isMeasure() {
		return filterSQL{}, newRequestError(ErrorMeasuresInRowsOnly, filter.Field)
	}

	if isDateOperator(filter.Operator) {
		return builder.writeDateFieldFilter(field, filter)
	}

	expression, err := builder.fieldExpression(field)
	if err != nil {
		return filterSQL{}, err
	}

	isYesNoField := !field.isMeasure() && field.dimension.Type == FieldTypeYesNo
	if isYesNoField {
		filter.Values = yesNoFilterValues(filter.Values)
	}

	written, err := writeCondition(builder.syntax, expression, filter)
	if err != nil {
		return filterSQL{}, err
	}

	written.isOnMeasures = field.isMeasure()

	return written, nil
}

// writeDateFieldFilter compares the value as stored, so the database can use
// an index on the column.
func (builder *queryBuilder) writeDateFieldFilter(field requestedField, filter Filter) (filterSQL, error) {
	expression, err := builder.storedValueExpression(field)
	if err != nil {
		return filterSQL{}, err
	}

	written, err := writeDateCondition(builder.syntax, expression, filter)
	if err != nil {
		return filterSQL{}, err
	}

	written.isOnMeasures = field.isMeasure()

	return written, nil
}

var knownOperators = map[FilterOperator]bool{
	FilterEquals:         true,
	FilterNotEquals:      true,
	FilterContains:       true,
	FilterNotContains:    true,
	FilterStartsWith:     true,
	FilterEndsWith:       true,
	FilterGreaterThan:    true,
	FilterGreaterOrEqual: true,
	FilterLessThan:       true,
	FilterLessOrEqual:    true,
	FilterIsSet:          true,
	FilterIsNotSet:       true,
}

func writeCondition(syntax sqlSyntax, expression string, filter Filter) (filterSQL, error) {
	if !knownOperators[filter.Operator] {
		return filterSQL{}, newRequestError(ErrorUnknownFilterType, filter.Field)
	}

	switch filter.Operator {
	case FilterIsSet:
		return filterSQL{condition: expression + " IS NOT NULL"}, nil
	case FilterIsNotSet:
		return filterSQL{condition: expression + " IS NULL"}, nil
	}

	if len(filter.Values) == 0 {
		return filterSQL{}, newRequestError(ErrorMissingFilterValue, filter.Field)
	}

	values := valuesForSQL(filter.Values)

	switch filter.Operator {
	case FilterEquals:
		return writeEquals(expression, values), nil
	case FilterNotEquals:
		return writeNotEquals(expression, values), nil
	case FilterContains:
		return writeTextMatch(syntax, expression, values, matchAnywhere, false), nil
	case FilterNotContains:
		return writeTextMatch(syntax, expression, values, matchAnywhere, true), nil
	case FilterStartsWith:
		return writeTextMatch(syntax, expression, values, matchAtStart, false), nil
	case FilterEndsWith:
		return writeTextMatch(syntax, expression, values, matchAtEnd, false), nil
	case FilterGreaterThan:
		return writeComparison(expression, ">", values), nil
	case FilterGreaterOrEqual:
		return writeComparison(expression, ">=", values), nil
	case FilterLessThan:
		return writeComparison(expression, "<", values), nil
	case FilterLessOrEqual:
		return writeComparison(expression, "<=", values), nil
	}

	return filterSQL{}, newRequestError(ErrorUnknownFilterType, filter.Field)
}

func writeEquals(expression string, values []any) filterSQL {
	if len(values) == 1 {
		return filterSQL{condition: expression + " = " + valuePlaceholder, values: values}
	}

	condition := expression + " IN (" + valuePlaceholders(len(values)) + ")"

	return filterSQL{condition: condition, values: values}
}

// writeNotEquals also matches empty values: "not Latvia" includes rows with
// no country.
func writeNotEquals(expression string, values []any) filterSQL {
	if len(values) == 1 {
		condition := "(" + expression + " <> " + valuePlaceholder + " OR " + expression + " IS NULL)"
		return filterSQL{condition: condition, values: values}
	}

	condition := "(" + expression + " NOT IN (" + valuePlaceholders(len(values)) + ") OR " + expression + " IS NULL)"

	return filterSQL{condition: condition, values: values}
}

// writeTextMatch writes contains, starts with and ends with. Upper and lower
// case are ignored, and % and _ in the searched text are matched as
// themselves, not as "anything".
func writeTextMatch(syntax sqlSyntax, expression string, values []any, kind textMatchKind, isNegated bool) filterSQL {
	joinWord := " OR "
	if isNegated {
		joinWord = " AND "
	}

	var matches []string
	var escapedValues []any

	for _, value := range values {
		text, _ := value.(string)
		escapedValues = append(escapedValues, escapeTextMatch(text))

		match := syntax.textMatch(expression, kind, isNegated)
		matches = append(matches, match)
	}

	condition := "(" + strings.Join(matches, joinWord) + ")"

	// "Does not contain" also matches empty values, as "not equals" does.
	if isNegated {
		condition = "(" + condition + " OR " + expression + " IS NULL)"
	}

	return filterSQL{condition: condition, values: escapedValues}
}

// writeComparison uses only the first value.
func writeComparison(expression string, comparison string, values []any) filterSQL {
	condition := expression + " " + comparison + " " + valuePlaceholder

	return filterSQL{condition: condition, values: values[:1]}
}

func escapeTextMatch(text string) string {
	escaped := strings.ReplaceAll(text, "!", "!!")
	escaped = strings.ReplaceAll(escaped, "%", "!%")
	escaped = strings.ReplaceAll(escaped, "_", "!_")

	return escaped
}

// yesNoFilterValues writes yes and no as 1 and 0, which both databases read
// as yes and no. Charts send them as true and false, as results give them,
// and MySQL would compare a 0 or 1 column with the text "true" as 0.
func yesNoFilterValues(texts []string) []string {
	values := make([]string, 0, len(texts))

	for _, text := range texts {
		isYes, isYesOrNo := readYesNo(text)
		if !isYesOrNo {
			values = append(values, text)
			continue
		}

		if isYes {
			values = append(values, "1")
			continue
		}

		values = append(values, "0")
	}

	return values
}

func valuesForSQL(texts []string) []any {
	values := make([]any, 0, len(texts))

	for _, text := range texts {
		values = append(values, text)
	}

	return values
}
