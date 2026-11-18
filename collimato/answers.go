// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"sort"
	"strings"

	"github.com/twigex/twigex/queryengine"
)

var builderMeasureTypes = map[queryengine.MeasureType]string{
	queryengine.MeasureCount:         "count",
	queryengine.MeasureCountDistinct: "count_distinct",
	queryengine.MeasureSum:           "sum",
	queryengine.MeasureAverage:       "avg",
	queryengine.MeasureMinimum:       "min",
	queryengine.MeasureMaximum:       "max",
	queryengine.MeasureFormula:       "number",
}

// FieldListFor writes the engine's field list for the browser. A dataset or
// view with errors is left out; ModelErrors lists it.
func FieldListFor(models queryengine.Models, engineFieldList queryengine.FieldList) FieldList {
	joinGroups := joinGroupNumbers(models)

	fieldList := FieldList{Datasets: []DatasetFields{}}

	for _, dataset := range engineFieldList.Sources {
		if len(dataset.Errors) > 0 {
			continue
		}

		datasetFields := DatasetFields{
			Name:      dataset.Name,
			Title:     dataset.Title,
			IsView:    dataset.IsView,
			JoinGroup: joinGroups[dataset.Name],
			Fields:    []Field{},
		}

		for _, engineField := range dataset.Fields {
			field := Field{
				Key:          engineField.Key,
				Title:        engineField.Title,
				FullTitle:    dataset.Title + " " + engineField.Title,
				Kind:         string(engineField.Kind),
				Type:         builderFieldTypes[engineField.Type],
				MeasureType:  builderMeasureTypes[engineField.MeasureType],
				IsPrimaryKey: engineField.IsPrimaryKey,
			}

			datasetFields.Fields = append(datasetFields.Fields, field)
		}

		fieldList.Datasets = append(fieldList.Datasets, datasetFields)
	}

	return fieldList
}

// joinGroupNumbers numbers the groups of datasets joined to each other, in
// either direction, in name order. A dataset without joins has none (0).
func joinGroupNumbers(models queryengine.Models) map[string]int {
	neighbours := make(map[string][]string)

	knownDatasets := make(map[string]bool, len(models.Datasets))
	for _, dataset := range models.Datasets {
		knownDatasets[dataset.Name] = true
	}

	for _, dataset := range models.Datasets {
		for _, join := range dataset.Joins {
			if !knownDatasets[join.OtherDataset] {
				continue
			}

			neighbours[dataset.Name] = append(neighbours[dataset.Name], join.OtherDataset)
			neighbours[join.OtherDataset] = append(neighbours[join.OtherDataset], dataset.Name)
		}
	}

	joinedNames := make([]string, 0, len(neighbours))
	for name := range neighbours {
		joinedNames = append(joinedNames, name)
	}

	sort.Strings(joinedNames)

	groups := make(map[string]int)
	nextGroup := 1

	for _, name := range joinedNames {
		if groups[name] != 0 {
			continue
		}

		markGroup(name, nextGroup, neighbours, groups)
		nextGroup++
	}

	return groups
}

func markGroup(name string, group int, neighbours map[string][]string, groups map[string]int) {
	if groups[name] != 0 {
		return
	}

	groups[name] = group

	for _, neighbour := range neighbours[name] {
		markGroup(neighbour, group, neighbours, groups)
	}
}

// UnsupportedHasErrors is a model the engine read but cannot use, such as a
// join to a dataset that does not exist.
const UnsupportedHasErrors UnsupportedReason = "has_errors"

// ModelErrors lists the saved files the engine cannot read and the datasets
// and views it read but cannot use.
func ModelErrors(unsupported []UnsupportedFile, engineFieldList queryengine.FieldList) []ModelError {
	var modelErrors []ModelError

	for _, file := range unsupported {
		modelError := ModelError{Name: file.Name, Reason: file.Reason, Detail: file.Detail}
		modelErrors = append(modelErrors, modelError)
	}

	for _, dataset := range engineFieldList.Sources {
		if len(dataset.Errors) == 0 {
			continue
		}

		modelError := ModelError{
			Name:   dataset.Name,
			Reason: UnsupportedHasErrors,
			Detail: modelErrorsText(dataset.Errors),
		}

		modelErrors = append(modelErrors, modelError)
	}

	return modelErrors
}

func modelErrorsText(engineErrors []queryengine.ModelError) string {
	texts := make([]string, 0, len(engineErrors))

	for _, engineError := range engineErrors {
		text := string(engineError.Code)
		if engineError.Field != "" {
			text += ": " + engineError.Field
		}

		texts = append(texts, text)
	}

	return strings.Join(texts, "; ")
}

// AnswerFor writes the engine's result for the browser, with each column's
// titles and the builder's word for its type.
func AnswerFor(result queryengine.Result, models queryengine.Models) Answer {
	answer := Answer{
		Columns:    make([]AnswerColumn, 0, len(result.Columns)),
		Rows:       result.Rows,
		DateRanges: make([]AnswerDateRange, 0, len(result.DateRanges)),
	}

	for _, column := range result.Columns {
		answerColumn := AnswerColumn{
			Key:       column.Key,
			Title:     column.Title,
			FullTitle: datasetTitle(models, column.Key) + " " + column.Title,
			Kind:      string(column.Kind),
			Type:      builderFieldTypes[column.Type],
			DateStep:  string(column.DateStep),
		}

		answer.Columns = append(answer.Columns, answerColumn)
	}

	for _, dateRange := range result.DateRanges {
		answerRange := AnswerDateRange{Field: dateRange.Field, Start: dateRange.Start, End: dateRange.End}
		answer.DateRanges = append(answer.DateRanges, answerRange)
	}

	return answer
}

// datasetTitle is the title of the dataset or view a field key starts with.
func datasetTitle(models queryengine.Models, fieldKey string) string {
	name, _, _ := strings.Cut(fieldKey, ".")

	for _, dataset := range models.Datasets {
		if dataset.Name == name {
			return titleOrMadeFromName(dataset.Title, name)
		}
	}

	for _, view := range models.Views {
		if view.Name == name {
			return titleOrMadeFromName(view.Title, name)
		}
	}

	return queryengine.MakeTitle(name)
}

func titleOrMadeFromName(title string, name string) string {
	if title != "" {
		return title
	}

	return queryengine.MakeTitle(name)
}
