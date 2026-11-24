// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

// planJoins finds the datasets the request reads and how to join them all.
// Datasets only formulas refer to come last.
func (builder *queryBuilder) planJoins(parts []queryPart) (joinPlan, error) {
	fields := builder.fieldsInStartOrder()

	for _, part := range parts {
		fields = append(fields, part.measures...)
	}

	return builder.joinPlanForFields(fields)
}

// partPlan finds the joins one part of a combined query needs: those for the
// fields shown, sorted and filtered on, and for the part's own measures. A part
// does not join the datasets only other parts need.
func (builder *queryBuilder) partPlan(part queryPart) (joinPlan, error) {
	fields := builder.fieldsInStartOrder()

	var partFields []requestedField

	for _, field := range fields {
		if !field.isMeasure() {
			partFields = append(partFields, field)
		}
	}

	partFields = append(partFields, part.measures...)

	return builder.joinPlanForFields(partFields)
}

func (builder *queryBuilder) joinPlanForFields(fields []requestedField) (joinPlan, error) {
	var neededNames []string
	var viewPaths [][]string

	isNeeded := make(map[string]bool)
	_, isView := builder.viewsByName[builder.firstSourceName]

	for _, field := range fields {
		datasetName := field.dataset.Name

		// In a view, a formula can refer to a dataset the view does not
		// reach.
		if isView && field.viewPath == nil {
			return joinPlan{}, newRequestError(ErrorNoJoinPath, datasetName)
		}

		if field.viewPath != nil {
			viewPaths = append(viewPaths, field.viewPath)
		}

		if !isNeeded[datasetName] {
			isNeeded[datasetName] = true
			neededNames = append(neededNames, datasetName)
		}
	}

	plan, err := builder.findPlan(neededNames, viewPaths)
	if err != nil {
		return joinPlan{}, err
	}

	if err := builder.checkPlanDatasets(plan); err != nil {
		return joinPlan{}, err
	}

	return plan, nil
}

func (builder *queryBuilder) findPlan(neededNames []string, viewPaths [][]string) (joinPlan, error) {
	if len(viewPaths) > 0 {
		return viewJoinPlan(builder.firstSourceName, viewPaths, builder.datasetsByName)
	}

	return findJoinPlan(neededNames, builder.datasetsByName)
}

// checkPlanDatasets checks the datasets joined on the way too, not only those
// the request names.
func (builder *queryBuilder) checkPlanDatasets(plan joinPlan) error {
	for _, dataset := range plan.datasets() {
		if builder.brokenDatasets[dataset.Name] {
			return newRequestError(ErrorDatasetHasErrors, dataset.Name)
		}

		if dataset.ConnectionID != plan.start.ConnectionID {
			return newRequestError(ErrorDifferentConnections, dataset.Name)
		}
	}

	return nil
}

// fieldsInStartOrder lists every field the request uses, in the order in which
// their datasets are tried as the start of the joins: dimensions, sort
// fields, date fields, filters on fields, measures, then filters on measures.
// The order matters when two datasets join each other.
func (builder *queryBuilder) fieldsInStartOrder() []requestedField {
	var firstKeys []string
	firstKeys = append(firstKeys, builder.request.Dimensions...)

	for _, sortField := range builder.request.SortBy {
		firstKeys = append(firstKeys, sortField.Field)
	}

	for _, dateField := range builder.request.DateFields {
		firstKeys = append(firstKeys, dateField.Field)
	}

	firstFields := builder.fieldsByKeys(firstKeys)
	filterFields := builder.fieldsByKeys(filterFieldKeys(builder.request.Filters))
	measures := builder.fieldsByKeys(builder.request.Measures)

	ordered := firstFields

	for _, field := range filterFields {
		if !field.isMeasure() {
			ordered = append(ordered, field)
		}
	}

	ordered = append(ordered, measures...)

	for _, field := range filterFields {
		if field.isMeasure() {
			ordered = append(ordered, field)
		}
	}

	return ordered
}

func (builder *queryBuilder) fieldsByKeys(keys []string) []requestedField {
	fields := make([]requestedField, 0, len(keys))

	for _, key := range keys {
		fields = append(fields, builder.field(key))
	}

	return fields
}

func filterFieldKeys(filters []Filter) []string {
	var keys []string

	for _, filter := range filters {
		if filter.Field != "" {
			keys = append(keys, filter.Field)
		}

		keys = append(keys, filterFieldKeys(filter.AllOf)...)
		keys = append(keys, filterFieldKeys(filter.AnyOf)...)
	}

	return keys
}
