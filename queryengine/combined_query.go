// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strings"
)

// queryPart holds the measures of one dataset, worked out in one part of a
// combined query.
type queryPart struct {
	dataset  Dataset
	measures []requestedField
}

// queryMeasures is every measure the query works out, by part, and whether a
// formula is worked out from the parts' results.
type queryMeasures struct {
	parts                 []queryPart
	hasFormulaUsingOthers bool
}

// allGroupsName names the list of every group any part found, which the
// parts are joined to.
const allGroupsName = "all_groups"

// distinctRowsName names the list of each group's rows, by primary key, once.
const distinctRowsName = "distinct_rows"

const primaryKeyAlias = "distinct_rows__primary_key"

// findQueryMeasures lists the measures the query works out, grouped by dataset,
// in the order the request uses them. A formula using another dataset's
// measures is not listed: it is worked out from the parts' results, so the
// measures it refers to are listed instead.
func (builder *queryBuilder) findQueryMeasures() (queryMeasures, error) {
	fields := builder.fieldsInStartOrder()

	measureParts := queryMeasures{}
	partPositions := make(map[string]int)
	listed := make(map[string]bool)

	for _, field := range fields {
		if !field.isMeasure() {
			continue
		}

		partMeasures, usesOthers, err := builder.partMeasuresOf(field)
		if err != nil {
			return queryMeasures{}, err
		}

		if usesOthers {
			measureParts.hasFormulaUsingOthers = true
		}

		for _, measureField := range partMeasures {
			key := partMeasureKey(measureField)
			if listed[key] {
				continue
			}

			listed[key] = true

			position, exists := partPositions[measureField.dataset.Name]
			if !exists {
				measureParts.parts = append(measureParts.parts, queryPart{dataset: measureField.dataset})
				position = len(measureParts.parts) - 1
				partPositions[measureField.dataset.Name] = position
			}

			measureParts.parts[position].measures = append(measureParts.parts[position].measures, measureField)
		}
	}

	return measureParts, nil
}

// partMeasuresOf gives the measures the parts work out for one requested
// measure: the measure itself, or for a formula using other datasets, the
// measures it refers to. It also says whether the measure is such a formula.
func (builder *queryBuilder) partMeasuresOf(measureField requestedField) ([]requestedField, bool, error) {
	usesOthers, err := builder.usesOtherDatasets(measureField, make(map[string]bool))
	if err != nil {
		return nil, false, err
	}

	if !usesOthers {
		return []requestedField{measureField}, false, nil
	}

	// Worked out after the parts are put together, the formula can no longer
	// reach its dataset's columns, only measures.
	if strings.Contains(measureField.measure.SQL, "{"+ownDatasetReference+"}") {
		return nil, false, newRequestError(ErrorFormulaMixesDatasets, measureField.key)
	}

	references, err := builder.formulaReferences(measureField)
	if err != nil {
		return nil, false, err
	}

	var partMeasures []requestedField

	for _, reference := range references {
		referenceMeasures, _, err := builder.partMeasuresOf(reference)
		if err != nil {
			return nil, false, err
		}

		partMeasures = append(partMeasures, referenceMeasures...)
	}

	return partMeasures, true, nil
}

// needsCombinedQuery reports whether one SELECT cannot answer the request:
// measures come from more than one dataset, a formula uses another dataset's
// measures, or a dataset's rows repeat and its measures are more than counts.
func (builder *queryBuilder) needsCombinedQuery(measureParts queryMeasures) bool {
	if len(measureParts.parts) > 1 || measureParts.hasFormulaUsingOthers {
		return true
	}

	if len(measureParts.parts) == 0 {
		return false
	}

	return partNeedsDistinctRows(measureParts.parts[0], builder.plan)
}

// partNeedsDistinctRows reports whether the part's measures must be worked out
// from each row's primary key, because the joins repeat its rows. A count of
// rows can count distinct primary keys instead, and a distinct count is not
// changed by repeats.
func partNeedsDistinctRows(part queryPart, plan joinPlan) bool {
	if !plan.rowsRepeat(part.dataset.Name) {
		return false
	}

	for _, measureField := range part.measures {
		isDistinctCount := measureField.measure.Type == MeasureCountDistinct
		if !isRowCount(measureField.measure) && !isDistinctCount {
			return true
		}
	}

	return false
}

// buildCombinedQuery works out each part's measures in a query of its own, then
// puts them side by side, matched on the fields shown (as MySQL writes it):
//
//	WITH `part_1` AS (…), `part_2` AS (…)
//	SELECT `all_groups`.`shops__format`, `part_1`.`shop_sales__revenue`, …
//	FROM (SELECT `shops__format` FROM `part_1` UNION SELECT `shops__format` FROM `part_2`) AS `all_groups`
//	LEFT JOIN `part_1` ON `all_groups`.`shops__format` <=> `part_1`.`shops__format`
//	LEFT JOIN `part_2` ON …
//
// Every combination of shown values that any part finds is listed. A part
// that did not find it gives empty measures there, not 0.
func (builder *queryBuilder) buildCombinedQuery(groupedFields []requestedField, outputFields []requestedField, dates requestedDates, parts []queryPart) (BuiltQuery, error) {
	if err := builder.checkSortFieldsShown(outputFields); err != nil {
		return BuiltQuery{}, err
	}

	for _, part := range parts {
		if err := checkColumnNamesDiffer(part.measures, partMeasureAlias); err != nil {
			return BuiltQuery{}, err
		}
	}

	outer := builder.forOuterQuery(parts)

	filters, err := outer.filterConditions()
	if err != nil {
		return BuiltQuery{}, err
	}

	outerLines, err := outer.combinedOuterLines(groupedFields, outputFields, parts, filters)
	if err != nil {
		return BuiltQuery{}, err
	}

	var partTexts []string
	var values []any

	for position, part := range parts {
		partLines, err := builder.partLines(part, groupedFields, filters, dates)
		if err != nil {
			return BuiltQuery{}, err
		}

		partText := builder.quote(partName(position)) + " AS (\n" + strings.Join(partLines, "\n") + "\n)"
		partTexts = append(partTexts, partText)

		values = append(values, filters.whereValues...)
		values = append(values, dates.values...)
	}

	values = append(values, filters.havingValues...)

	var lines []string
	lines = append(lines, "WITH "+strings.Join(partTexts, ",\n"))
	lines = append(lines, outerLines...)

	query := BuiltQuery{
		SQL:          strings.Join(lines, "\n"),
		Values:       values,
		Columns:      builder.columns(outputFields),
		ConnectionID: builder.plan.start.ConnectionID,
		DateRanges:   dates.ranges,
	}

	return query, nil
}

// combinedOuterLines writes the SELECT that puts the parts together. Filters
// on measures apply to the measures put together, so they go in its WHERE.
func (builder *queryBuilder) combinedOuterLines(groupedFields []requestedField, outputFields []requestedField, parts []queryPart, filters requestFilters) ([]string, error) {
	groupedSource := partName(0)
	hasAllGroups := len(parts) > 1 && len(groupedFields) > 0
	if hasAllGroups {
		groupedSource = allGroupsName
	}

	selectLines := make([]string, 0, len(outputFields))

	for _, field := range outputFields {
		expression := builder.quote(groupedSource) + "." + builder.quote(field.alias())

		if field.isMeasure() {
			measureSQL, err := builder.measureExpression(field)
			if err != nil {
				return nil, err
			}

			expression = measureSQL
		}

		selectLines = append(selectLines, expression+" AS "+builder.quote(field.alias()))
	}

	orderParts, err := builder.orderParts(outputFields)
	if err != nil {
		return nil, err
	}

	var lines []string
	lines = append(lines, "SELECT "+strings.Join(selectLines, ", "))
	lines = append(lines, builder.combinedFromLines(groupedFields, len(parts), hasAllGroups)...)

	lines = append(lines, whereLine(filters.havingConditions)...)
	lines = append(lines, orderByLine(orderParts)...)
	lines = append(lines, builder.syntax.rowLimit(builder.request.RowLimit))

	return lines, nil
}

// combinedFromLines lists every combination of shown values any part found,
// and joins each part to it. Empty values match each other
// (sameOrBothEmpty), so a group with an empty value is not lost. With nothing
// shown, each part is one row, and the rows are simply put side by side.
func (builder *queryBuilder) combinedFromLines(groupedFields []requestedField, partCount int, hasAllGroups bool) []string {
	if !hasAllGroups {
		lines := []string{"FROM " + builder.quote(partName(0))}

		for position := 1; position < partCount; position++ {
			lines = append(lines, "CROSS JOIN "+builder.quote(partName(position)))
		}

		return lines
	}

	groupedColumns := make([]string, 0, len(groupedFields))
	for _, field := range groupedFields {
		groupedColumns = append(groupedColumns, builder.quote(field.alias()))
	}

	groupSelects := make([]string, 0, partCount)
	for position := range partCount {
		groupSelect := "SELECT " + strings.Join(groupedColumns, ", ") + " FROM " + builder.quote(partName(position))
		groupSelects = append(groupSelects, groupSelect)
	}

	allGroups := builder.quote(allGroupsName)
	allGroupsList := strings.Join(groupSelects, " UNION ")

	lines := []string{"FROM (" + allGroupsList + ") AS " + allGroups}

	for position := range partCount {
		part := builder.quote(partName(position))

		matches := make([]string, 0, len(groupedFields))
		for _, column := range groupedColumns {
			match := builder.syntax.sameOrBothEmpty(allGroups+"."+column, part+"."+column)
			matches = append(matches, match)
		}

		lines = append(lines, "LEFT JOIN "+part+" ON "+strings.Join(matches, " AND "))
	}

	return lines
}

// partLines writes one part: its dataset's measures, grouped by the fields
// shown, over the rows the filters keep.
func (builder *queryBuilder) partLines(part queryPart, groupedFields []requestedField, filters requestFilters, dates requestedDates) ([]string, error) {
	plan, err := builder.partPlan(part)
	if err != nil {
		return nil, err
	}

	partBuilder := builder.withPlan(plan)

	var whereConditions []string
	whereConditions = append(whereConditions, filters.whereConditions...)
	whereConditions = append(whereConditions, dates.conditions...)

	if partNeedsDistinctRows(part, plan) {
		return partBuilder.distinctRowsPartLines(part, groupedFields, whereConditions)
	}

	groupedLines, err := partBuilder.selectLines(groupedFields)
	if err != nil {
		return nil, err
	}

	measureLines, err := partBuilder.partMeasureLines(part)
	if err != nil {
		return nil, err
	}

	fromLines, err := plan.fromLines(builder.syntax)
	if err != nil {
		return nil, err
	}

	var selectLines []string
	selectLines = append(selectLines, groupedLines...)
	selectLines = append(selectLines, measureLines...)

	var lines []string
	lines = append(lines, "SELECT "+strings.Join(selectLines, ", "))
	lines = append(lines, fromLines...)
	lines = append(lines, whereLine(whereConditions)...)
	lines = append(lines, groupByLine(len(groupedFields))...)

	return lines, nil
}

// distinctRowsPartLines works out measures over rows the joins repeat:
// first each row's primary key once per group, then the measures over the
// dataset's rows joined to those primary keys, where nothing repeats.
//
//	SELECT `distinct_rows`.`shop_sales__channel`, SUM(`shops`.`floor_area`)
//	FROM (SELECT DISTINCT <channel>, <store id> FROM … WHERE …) AS `distinct_rows`
//	LEFT JOIN `stores` AS `shops` ON `distinct_rows`.`distinct_rows__primary_key` = `shops`.`id`
//	GROUP BY 1
func (builder *queryBuilder) distinctRowsPartLines(part queryPart, groupedFields []requestedField, whereConditions []string) ([]string, error) {
	datasetField := requestedField{key: part.dataset.Name, dataset: part.dataset}

	primaryKey, hasPrimaryKey, err := builder.primaryKeyExpression(datasetField)
	if err != nil {
		return nil, err
	}

	if !hasPrimaryKey {
		return nil, newRequestError(ErrorNeedsPrimaryKey, part.dataset.Name)
	}

	distinctRowsSQL, err := builder.distinctRowsQuery(groupedFields, primaryKey, whereConditions)
	if err != nil {
		return nil, err
	}

	// The measures read the dataset's own table only, joined to the primary
	// keys.
	measuresBuilder := builder.withPlan(joinPlan{start: part.dataset})

	measureLines, err := measuresBuilder.partMeasureLines(part)
	if err != nil {
		return nil, err
	}

	table, err := quotedTableName(builder.syntax, part.dataset.Table)
	if err != nil {
		return nil, err
	}

	distinctRows := builder.quote(distinctRowsName)

	var selectLines []string
	for _, field := range groupedFields {
		alias := builder.quote(field.alias())
		selectLines = append(selectLines, distinctRows+"."+alias+" AS "+alias)
	}

	selectLines = append(selectLines, measureLines...)

	distinctRowsKey := distinctRows + "." + builder.quote(primaryKeyAlias)
	datasetName := builder.quote(part.dataset.Name)
	distinctRowsJoin := "LEFT JOIN " + table + " AS " + datasetName + " ON " + distinctRowsKey + " = " + primaryKey

	var lines []string
	lines = append(lines, "SELECT "+strings.Join(selectLines, ", "))
	lines = append(lines, "FROM ("+distinctRowsSQL+") AS "+distinctRows)
	lines = append(lines, distinctRowsJoin)
	lines = append(lines, groupByLine(len(groupedFields))...)

	return lines, nil
}

// distinctRowsQuery lists each shown group with the primary key of every
// row in it, once.
func (builder *queryBuilder) distinctRowsQuery(groupedFields []requestedField, primaryKey string, whereConditions []string) (string, error) {
	groupedLines, err := builder.selectLines(groupedFields)
	if err != nil {
		return "", err
	}

	fromLines, err := builder.plan.fromLines(builder.syntax)
	if err != nil {
		return "", err
	}

	var selectLines []string
	selectLines = append(selectLines, groupedLines...)
	selectLines = append(selectLines, primaryKey+" AS "+builder.quote(primaryKeyAlias))

	var lines []string
	lines = append(lines, "SELECT DISTINCT "+strings.Join(selectLines, ", "))
	lines = append(lines, fromLines...)
	lines = append(lines, whereLine(whereConditions)...)

	return strings.Join(lines, "\n"), nil
}

func (builder *queryBuilder) partMeasureLines(part queryPart) ([]string, error) {
	lines := make([]string, 0, len(part.measures))

	for _, measureField := range part.measures {
		expression, err := builder.measureExpression(measureField)
		if err != nil {
			return nil, err
		}

		lines = append(lines, expression+" AS "+builder.quote(partMeasureAlias(measureField)))
	}

	return lines, nil
}

// combinedMeasureExpression gives a measure in the SELECT that puts the parts
// together: the part's column, or a formula using other datasets filled in
// with the parts' columns.
func (builder *queryBuilder) combinedMeasureExpression(measure requestedField) (string, error) {
	column, isInPart := builder.partColumnByMeasure[partMeasureKey(measure)]
	if isInPart {
		return column, nil
	}

	return builder.fillInFormula(measure, builder.combinedMeasureExpression)
}

// withPlan gives a copy of the builder that writes SQL for another join plan:
// one part of a combined query.
func (builder *queryBuilder) withPlan(plan joinPlan) *queryBuilder {
	copied := *builder
	copied.plan = plan

	return &copied
}

// forOuterQuery gives a copy of the builder that writes the outer SELECT of a
// combined query, where measures are read from the parts' columns.
func (builder *queryBuilder) forOuterQuery(parts []queryPart) *queryBuilder {
	copied := *builder
	copied.partColumnByMeasure = builder.partColumnNames(parts)

	return &copied
}

// checkSortFieldsShown refuses sorting by a field that is not shown: after the
// parts are put together, only shown fields and measures are left to sort by.
func (builder *queryBuilder) checkSortFieldsShown(outputFields []requestedField) error {
	outputKeys := fieldKeys(outputFields)

	for _, sortField := range builder.request.SortBy {
		field := builder.field(sortField.Field)

		if !field.isMeasure() && !outputKeys[field.key] {
			return newRequestError(ErrorSortFieldNotShown, sortField.Field)
		}
	}

	return nil
}

func (builder *queryBuilder) partColumnNames(parts []queryPart) map[string]string {
	columns := make(map[string]string)

	for position, part := range parts {
		for _, measureField := range part.measures {
			column := builder.quote(partName(position)) + "." + builder.quote(partMeasureAlias(measureField))
			columns[partMeasureKey(measureField)] = column
		}
	}

	return columns
}

func partName(position int) string {
	return fmt.Sprintf("part_%d", position+1)
}

// partMeasureKey names a measure by its dataset, also when the request asked
// for it through a view, so that the same measure is worked out once.
func partMeasureKey(measureField requestedField) string {
	return measureField.dataset.Name + "." + measureField.measure.Name
}

func partMeasureAlias(measureField requestedField) string {
	return measureField.dataset.Name + "__" + measureField.measure.Name
}
