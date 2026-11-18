// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strings"
)

type BuiltQuery struct {
	SQL string

	// Values are the filter values, sent to the database separately from the
	// SQL, in the order of the ? marks in it.
	Values []any

	Columns      []Column
	ConnectionID string

	// DateRanges has the actual start and end of every date field's range.
	DateRanges []ResolvedDateRange
}

type Column struct {
	Key   string
	Title string
	Kind  FieldKind
	Type  FieldType

	// DateStep is set for a date field grouped by day, month and so on. Each
	// value is then the start of its period.
	DateStep DateStep

	// SentAsText is set by RunQuery for number fields whose values can be too
	// long to be exact in a browser.
	SentAsText bool
}

type requestedField struct {
	key string

	// shownName is the field's name where it was asked for: in its dataset,
	// or in the view that shows it.
	shownName string

	// dataset is the dataset the value is read from, also for view fields.
	dataset Dataset

	// viewPath is set for view fields: the datasets the view goes through to
	// reach the field.
	viewPath []string

	dimension *Dimension
	measure   *Measure
	dateStep  DateStep
}

func (field requestedField) isMeasure() bool {
	return field.measure != nil
}

func (field requestedField) alias() string {
	return strings.ReplaceAll(field.key, ".", "__")
}

// queryBuilder writes the SQL for one request. build sets its fields in two
// stages, and nothing changes them after: reading the request's fields finds
// them and the database, then planJoins finds the joins. Writing one part
// of a combined query, or its outer SELECT, uses a copy with its own plan or
// part columns (withPlan, forOuterQuery).
type queryBuilder struct {
	models  Models
	request Request

	// The models looked up by name and checked, as for the field list.
	checkedModels

	checkedViews map[string]bool

	// Set while the request's fields are read (readFieldOnce).
	fieldsByKey     map[string]requestedField
	firstSourceName string

	// syntax is how the request's database writes SQL, set from the first
	// dataset the request uses. Every dataset of a query must use the same
	// connection, which checkPlanDatasets makes sure of.
	syntax sqlSyntax

	// plan is the joins of the query, or of the part of a combined query,
	// being written.
	plan joinPlan

	// partColumnByMeasure is set only for the outer SELECT of a combined
	// query: for each measure worked out in a part, the part's column that
	// holds it.
	partColumnByMeasure map[string]string
}

// BuildQuery turns a request into the SQL that answers it.
func BuildQuery(models Models, request Request) (BuiltQuery, error) {
	if !asksForAnyField(request) {
		return BuiltQuery{}, newRequestError(ErrorNoFields, "")
	}

	if request.RowLimit <= 0 {
		return BuiltQuery{}, newRequestError(ErrorMissingRowLimit, "")
	}

	if request.RowsOnly && len(request.Measures) > 0 {
		return BuiltQuery{}, newRequestError(ErrorMeasuresInRowsOnly, request.Measures[0])
	}

	builder := newQueryBuilder(models, request)

	query, err := builder.build()
	if err != nil {
		return BuiltQuery{}, err
	}

	query.SQL = builder.syntax.replaceValuePlaceholders(query.SQL)

	return query, nil
}

func asksForAnyField(request Request) bool {
	if len(request.Dimensions) > 0 || len(request.Measures) > 0 {
		return true
	}

	for _, dateField := range request.DateFields {
		if dateField.Step != "" {
			return true
		}
	}

	return false
}

func newQueryBuilder(models Models, request Request) *queryBuilder {
	builder := &queryBuilder{
		models:        models,
		request:       request,
		checkedModels: checkModels(models),
		checkedViews:  make(map[string]bool),
		fieldsByKey:   make(map[string]requestedField),
	}

	return builder
}

func (builder *queryBuilder) build() (BuiltQuery, error) {
	dimensions, err := builder.readFieldsOfKind(builder.request.Dimensions, FieldKindDimension)
	if err != nil {
		return BuiltQuery{}, err
	}

	dates, err := builder.findDateFields()
	if err != nil {
		return BuiltQuery{}, err
	}

	measures, err := builder.readFieldsOfKind(builder.request.Measures, FieldKindMeasure)
	if err != nil {
		return BuiltQuery{}, err
	}

	var outputFields []requestedField
	outputFields = append(outputFields, dimensions...)
	outputFields = append(outputFields, dates.groupedFields...)
	outputFields = append(outputFields, measures...)

	if err := checkNoFieldShownTwice(outputFields); err != nil {
		return BuiltQuery{}, err
	}

	if err := checkColumnNamesDiffer(outputFields, requestedField.alias); err != nil {
		return BuiltQuery{}, err
	}

	if err := builder.readSortAndFilterFields(); err != nil {
		return BuiltQuery{}, err
	}

	measureParts, err := builder.findQueryMeasures()
	if err != nil {
		return BuiltQuery{}, err
	}

	plan, err := builder.planJoins(measureParts.parts)
	if err != nil {
		return BuiltQuery{}, err
	}

	builder.plan = plan

	var groupedFields []requestedField
	groupedFields = append(groupedFields, dimensions...)
	groupedFields = append(groupedFields, dates.groupedFields...)

	if builder.needsCombinedQuery(measureParts) {
		return builder.buildCombinedQuery(groupedFields, outputFields, dates, measureParts.parts)
	}

	return builder.buildSingleQuery(groupedFields, outputFields, dates)
}

// buildSingleQuery writes one SELECT, which is enough when every measure comes
// from one dataset whose rows the joins do not repeat.
func (builder *queryBuilder) buildSingleQuery(groupedFields []requestedField, outputFields []requestedField, dates requestedDates) (BuiltQuery, error) {
	selectLines, err := builder.selectLines(outputFields)
	if err != nil {
		return BuiltQuery{}, err
	}

	filters, err := builder.filterConditions()
	if err != nil {
		return BuiltQuery{}, err
	}

	orderParts, err := builder.orderParts(outputFields)
	if err != nil {
		return BuiltQuery{}, err
	}

	fromLines, err := builder.plan.fromLines(builder.syntax)
	if err != nil {
		return BuiltQuery{}, err
	}

	var lines []string

	lines = append(lines, "SELECT "+strings.Join(selectLines, ", "))
	lines = append(lines, fromLines...)

	var whereConditions []string
	whereConditions = append(whereConditions, filters.whereConditions...)
	whereConditions = append(whereConditions, dates.conditions...)

	lines = append(lines, whereLine(whereConditions)...)

	if !builder.request.RowsOnly {
		lines = append(lines, groupByLine(len(groupedFields))...)
	}

	lines = append(lines, havingLine(filters.havingConditions)...)
	lines = append(lines, orderByLine(orderParts)...)
	lines = append(lines, builder.syntax.rowLimit(builder.request.RowLimit))

	var values []any
	values = append(values, filters.whereValues...)
	values = append(values, dates.values...)
	values = append(values, filters.havingValues...)

	query := BuiltQuery{
		SQL:          strings.Join(lines, "\n"),
		Values:       values,
		Columns:      builder.columns(outputFields),
		ConnectionID: builder.plan.start.ConnectionID,
		DateRanges:   dates.ranges,
	}

	return query, nil
}

// The request's fields are read in this order, each group checked before the
// next is read, so that a request with several mistakes reports the first:
// dimensions, date fields, measures, then sort fields and filters. The first
// field read decides the database. Everything after reads them with field,
// which finds them without changing anything.

// readFieldsOfKind reads fields and checks that each is a dimension, or each
// a measure.
func (builder *queryBuilder) readFieldsOfKind(keys []string, wantedKind FieldKind) ([]requestedField, error) {
	fields := make([]requestedField, 0, len(keys))

	for _, key := range keys {
		field, err := builder.readFieldOnce(key)
		if err != nil {
			return nil, err
		}

		isWantedKind := field.isMeasure() == (wantedKind == FieldKindMeasure)
		if !isWantedKind {
			return nil, newRequestError(ErrorWrongFieldKind, key)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

func (builder *queryBuilder) readSortAndFilterFields() error {
	var keys []string

	for _, sortField := range builder.request.SortBy {
		keys = append(keys, sortField.Field)
	}

	keys = append(keys, filterFieldKeys(builder.request.Filters)...)

	for _, key := range keys {
		if _, err := builder.readFieldOnce(key); err != nil {
			return err
		}
	}

	return nil
}

// readFieldOnce reads a field the first time the request names it, and gives
// the same field after.
func (builder *queryBuilder) readFieldOnce(key string) (requestedField, error) {
	field, isRead := builder.fieldsByKey[key]
	if isRead {
		return field, nil
	}

	field, err := builder.readField(key)
	if err != nil {
		return requestedField{}, err
	}

	builder.fieldsByKey[key] = field

	return field, nil
}

// field gives a field the request names, as it was read.
func (builder *queryBuilder) field(key string) requestedField {
	return builder.fieldsByKey[key]
}

// readField finds a field by its key, such as "orders.country" or
// "sales_view.store_name". The first field read decides the database.
func (builder *queryBuilder) readField(key string) (requestedField, error) {
	sourceName, fieldName, hasDot := strings.Cut(key, ".")
	if !hasDot {
		return requestedField{}, newRequestError(ErrorUnknownField, key)
	}

	view, isView := builder.viewsByName[sourceName]
	dataset, isDataset := builder.datasetsByName[sourceName]

	if !isView && !isDataset {
		return requestedField{}, newRequestError(ErrorUnknownField, key)
	}

	if err := builder.checkSourceNotMixed(sourceName); err != nil {
		return requestedField{}, err
	}

	if isView {
		return builder.findViewField(key, view, fieldName)
	}

	if builder.brokenDatasets[dataset.Name] {
		return requestedField{}, newRequestError(ErrorDatasetHasErrors, dataset.Name)
	}

	field, found := findDatasetField(dataset, fieldName)
	if !found {
		return requestedField{}, newRequestError(ErrorUnknownField, key)
	}

	if err := builder.chooseDatabase(dataset); err != nil {
		return requestedField{}, err
	}

	field.key = key

	return field, nil
}

// chooseDatabase sets how SQL is written from the first dataset the request
// uses.
func (builder *queryBuilder) chooseDatabase(dataset Dataset) error {
	if builder.syntax != nil {
		return nil
	}

	databaseType := builder.models.DatabaseTypes[dataset.ConnectionID]

	syntax, known := databaseFor(databaseType)
	if !known {
		return newRequestError(ErrorUnknownDatabaseType, dataset.Name)
	}

	builder.syntax = syntax

	return nil
}

func (builder *queryBuilder) quote(name string) string {
	return builder.syntax.quoteName(name)
}

func (builder *queryBuilder) findViewField(key string, view View, shownName string) (requestedField, error) {
	if !builder.checkedViews[view.Name] {
		viewFields := listViewFields(view, builder.checkedModels)
		if len(viewFields.Errors) > 0 {
			return requestedField{}, newRequestError(ErrorDatasetHasErrors, view.Name)
		}

		builder.checkedViews[view.Name] = true
	}

	source, found := findViewFieldSource(view, shownName, builder.datasetsByName)
	if !found {
		return requestedField{}, newRequestError(ErrorUnknownField, key)
	}

	field, found := findDatasetField(source.dataset, source.fieldName)
	if !found {
		return requestedField{}, newRequestError(ErrorUnknownField, key)
	}

	if err := builder.chooseDatabase(source.dataset); err != nil {
		return requestedField{}, err
	}

	field.key = key
	field.shownName = shownName
	field.viewPath = source.path

	return field, nil
}

func findDatasetField(dataset Dataset, fieldName string) (requestedField, bool) {
	dimension, isDimension := dataset.dimensionNamed(fieldName)
	if isDimension {
		field := requestedField{shownName: fieldName, dataset: dataset, dimension: dimension}
		return field, true
	}

	measure, isMeasure := dataset.measureNamed(fieldName)
	if isMeasure {
		field := requestedField{shownName: fieldName, dataset: dataset, measure: measure}
		return field, true
	}

	return requestedField{}, false
}

// checkSourceNotMixed makes sure a view's fields are not asked for together
// with fields of datasets or other views. A view already brings together the
// datasets it needs.
func (builder *queryBuilder) checkSourceNotMixed(sourceName string) error {
	if builder.firstSourceName == "" {
		builder.firstSourceName = sourceName
		return nil
	}

	if sourceName == builder.firstSourceName {
		return nil
	}

	_, firstIsView := builder.viewsByName[builder.firstSourceName]
	_, isView := builder.viewsByName[sourceName]

	if firstIsView || isView {
		return newRequestError(ErrorViewMixedWithOthers, sourceName)
	}

	return nil
}

func (builder *queryBuilder) selectLines(outputFields []requestedField) ([]string, error) {
	lines := make([]string, 0, len(outputFields))

	for _, field := range outputFields {
		expression, err := builder.fieldExpression(field)
		if err != nil {
			return nil, err
		}

		line := expression + " AS " + builder.quote(field.alias())
		lines = append(lines, line)
	}

	return lines, nil
}

// fieldExpression gives the SQL for a field's shown value. Date fields are
// shown in UTC, cut to the start of their step when they have one.
func (builder *queryBuilder) fieldExpression(field requestedField) (string, error) {
	if field.isMeasure() {
		return builder.measureExpression(field)
	}

	expression, err := builder.columnExpression(field, field.dimension.SQL)
	if err != nil {
		return "", err
	}

	if field.dimension.Type != FieldTypeTime {
		return expression, nil
	}

	utcTime := builder.syntax.timeInUTC(expression)

	if field.dateStep == "" {
		return utcTime, nil
	}

	if !isKnownDateStep(field.dateStep) {
		return "", newRequestError(ErrorUnknownDateStep, field.key)
	}

	return builder.syntax.startOfStep(utcTime, field.dateStep), nil
}

// storedValueExpression gives the SQL for a field's value as stored, which
// date filters compare with.
func (builder *queryBuilder) storedValueExpression(field requestedField) (string, error) {
	if field.isMeasure() {
		return builder.measureExpression(field)
	}

	return builder.columnExpression(field, field.dimension.SQL)
}

// columnExpression gives the SQL for a field's value. A plain column name gets
// the dataset in front of it; SQL written by hand is used as written, with
// {DATASET} standing for the dataset.
func (builder *queryBuilder) columnExpression(field requestedField, fieldSQL string) (string, error) {
	datasetName := builder.quote(field.dataset.Name)

	if isPlainName(fieldSQL) {
		return datasetName + "." + builder.quote(fieldSQL), nil
	}

	expression := strings.ReplaceAll(fieldSQL, "{"+ownDatasetReference+"}", datasetName)

	// Only a formula may refer to another field as {name}.
	if strings.Contains(expression, "{") {
		return "", newRequestError(ErrorFieldReferenceInSQL, field.key)
	}

	// In brackets, so that an OR or a + in it cannot mix with the SQL around
	// it, such as a filter's AND.
	return "(" + expression + ")", nil
}

// measureExpression gives the SQL for a measure: worked out from the rows, or
// in the outer SELECT of a combined query, read from the part that worked it
// out.
func (builder *queryBuilder) measureExpression(field requestedField) (string, error) {
	if builder.partColumnByMeasure != nil {
		return builder.combinedMeasureExpression(field)
	}

	return builder.rowsMeasureExpression(field)
}

// rowsMeasureExpression works a measure out from the rows of the query.
func (builder *queryBuilder) rowsMeasureExpression(field requestedField) (string, error) {
	measure := field.measure

	// A measure over repeated rows is either counted distinctly here, or,
	// when that is not enough, worked out from each row's primary key in a
	// part of a combined query (combined_query.go), where its rows no longer
	// repeat.
	rowsRepeat := builder.plan.rowsRepeat(field.dataset.Name)
	if rowsRepeat && isRowCount(measure) {
		return builder.countDistinctRowsExpression(field)
	}

	// needsCombinedQuery sends every other such measure to a part of its
	// own, so this is only reached by a mistake in the engine. It is checked
	// so that the mistake gives an error, not a wrong answer.
	if rowsRepeat && measure.Type != MeasureCountDistinct {
		return "", fmt.Errorf("query engine: %s would add up repeated rows", field.key)
	}

	if isRowCount(measure) {
		return builder.countRowsExpression(field)
	}

	if measure.SQL == "" {
		return "", newRequestError(ErrorFieldHasNoSQL, field.key)
	}

	if measure.Type == MeasureFormula {
		return builder.fillInFormula(field, builder.measureExpression)
	}

	columnSQL, err := builder.columnExpression(field, measure.SQL)
	if err != nil {
		return "", err
	}

	return aggregateSQL(field, columnSQL)
}

// aggregateSQL adds up a column's values with the measure's type.
func aggregateSQL(field requestedField, columnSQL string) (string, error) {
	switch field.measure.Type {
	case MeasureCount:
		return "COUNT(" + columnSQL + ")", nil
	case MeasureCountDistinct:
		return "COUNT(DISTINCT " + columnSQL + ")", nil
	case MeasureSum:
		return "SUM(" + columnSQL + ")", nil
	case MeasureAverage:
		return "AVG(" + columnSQL + ")", nil
	case MeasureMinimum:
		return "MIN(" + columnSQL + ")", nil
	case MeasureMaximum:
		return "MAX(" + columnSQL + ")", nil
	}

	return "", newRequestError(ErrorUnknownMeasureType, field.key)
}

func isRowCount(measure *Measure) bool {
	return measure.Type == MeasureCount && measure.SQL == ""
}

// countRowsExpression counts the dataset's primary key, or every row when the
// dataset has none.
func (builder *queryBuilder) countRowsExpression(field requestedField) (string, error) {
	primaryKey, hasPrimaryKey, err := builder.primaryKeyExpression(field)
	if err != nil {
		return "", err
	}

	if !hasPrimaryKey {
		return "COUNT(*)", nil
	}

	return "COUNT(" + primaryKey + ")", nil
}

// countDistinctRowsExpression counts a dataset's rows when joins repeat them:
// each primary key once.
func (builder *queryBuilder) countDistinctRowsExpression(field requestedField) (string, error) {
	primaryKey, hasPrimaryKey, err := builder.primaryKeyExpression(field)
	if err != nil {
		return "", err
	}

	if !hasPrimaryKey {
		return "", newRequestError(ErrorNeedsPrimaryKey, field.dataset.Name)
	}

	return "COUNT(DISTINCT " + primaryKey + ")", nil
}

func (builder *queryBuilder) primaryKeyExpression(field requestedField) (string, bool, error) {
	for _, dimension := range field.dataset.Dimensions {
		if !dimension.IsPrimaryKey {
			continue
		}

		primaryKey, err := builder.columnExpression(field, dimension.SQL)
		if err != nil {
			return "", false, err
		}

		return primaryKey, true, nil
	}

	return "", false, nil
}

func (builder *queryBuilder) orderParts(outputFields []requestedField) ([]string, error) {
	outputKeys := fieldKeys(outputFields)

	parts := make([]string, 0, len(builder.request.SortBy))

	for _, sortField := range builder.request.SortBy {
		field := builder.field(sortField.Field)

		if builder.request.RowsOnly && field.isMeasure() {
			return nil, newRequestError(ErrorMeasuresInRowsOnly, sortField.Field)
		}

		direction := "ASC"
		if sortField.Descending {
			direction = "DESC"
		}

		if outputKeys[field.key] {
			part := builder.quote(field.alias()) + " " + direction
			parts = append(parts, part)
			continue
		}

		// Grouped rows have no single value of a field they are not grouped
		// by: Postgres and MySQL 8 refuse to sort by one, and MariaDB picks
		// any row's value, so the order changes between runs.
		if !builder.request.RowsOnly && !field.isMeasure() {
			return nil, newRequestError(ErrorSortFieldNotShown, sortField.Field)
		}

		expression, err := builder.fieldExpression(field)
		if err != nil {
			return nil, err
		}

		part := expression + " " + direction
		parts = append(parts, part)
	}

	return parts, nil
}

func (builder *queryBuilder) columns(outputFields []requestedField) []Column {
	columns := make([]Column, 0, len(outputFields))

	for _, field := range outputFields {
		column := Column{Key: field.key}

		if field.isMeasure() {
			column.Title = titleOrMadeFromName(field.measure.Title, field.shownName)
			column.Kind = FieldKindMeasure
			column.Type = FieldTypeNumber
		} else {
			column.Title = titleOrMadeFromName(field.dimension.Title, field.shownName)
			column.Kind = FieldKindDimension
			column.Type = field.dimension.Type
			column.DateStep = field.dateStep
		}

		columns = append(columns, column)
	}

	return columns
}

func checkNoFieldShownTwice(outputFields []requestedField) error {
	keysSeen := make(map[string]bool)

	for _, field := range outputFields {
		if keysSeen[field.key] {
			return newRequestError(ErrorFieldShownTwice, field.key)
		}

		keysSeen[field.key] = true
	}

	return nil
}

// checkColumnNamesDiffer refuses two fields that would get the same column
// name in the SQL, as the database keeps it.
func checkColumnNamesDiffer(fields []requestedField, columnName func(requestedField) string) error {
	keyByName := make(map[string]string)

	for _, field := range fields {
		name := nameAsKept(columnName(field))

		otherKey, taken := keyByName[name]
		if taken && otherKey != field.key {
			return newRequestError(ErrorColumnNamesClash, otherKey+", "+field.key)
		}

		keyByName[name] = field.key
	}

	return nil
}

func fieldKeys(fields []requestedField) map[string]bool {
	keys := make(map[string]bool, len(fields))

	for _, field := range fields {
		keys[field.key] = true
	}

	return keys
}
