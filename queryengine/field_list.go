// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"slices"
	"strings"
)

type FieldList struct {
	Sources []SourceFields
}

type SourceFields struct {
	Name   string
	Title  string
	IsView bool
	Fields []Field
	Errors []ModelError
}

type FieldKind string

const (
	FieldKindDimension FieldKind = "dimension"
	FieldKindMeasure   FieldKind = "measure"
)

type Field struct {
	Key   string
	Title string
	Kind  FieldKind
	Type  FieldType

	// MeasureType is how a measure adds rows up. Empty for a dimension.
	MeasureType MeasureType

	IsPrimaryKey bool
}

// ListFields lists the fields of every dataset and view. A dataset or view
// with errors lists its errors instead of its fields, and the others are
// not affected.
func ListFields(models Models) FieldList {
	checked := checkModels(models)
	list := FieldList{}

	sortedDatasets := sortDatasetsByName(models.Datasets)
	for _, dataset := range sortedDatasets {
		datasetErrors := checkDataset(dataset, checked.datasetsByName, checked.duplicateNames)
		datasetFields := listDatasetFields(dataset, datasetErrors)

		list.Sources = append(list.Sources, datasetFields)
	}

	sortedViews := sortViewsByName(models.Views)
	for _, view := range sortedViews {
		viewFields := listViewFields(view, checked)

		list.Sources = append(list.Sources, viewFields)
	}

	return list
}

func listDatasetFields(dataset Dataset, datasetErrors []ModelError) SourceFields {
	datasetFields := SourceFields{
		Name:   dataset.Name,
		Title:  titleOrMadeFromName(dataset.Title, dataset.Name),
		Errors: datasetErrors,
	}

	if len(datasetErrors) > 0 {
		return datasetFields
	}

	for _, dimension := range dataset.Dimensions {
		if dimension.IsHidden {
			continue
		}

		key := dataset.Name + "." + dimension.Name
		datasetFields.Fields = append(datasetFields.Fields, listedDimension(key, dimension.Name, dimension))
	}

	for _, measure := range dataset.Measures {
		key := dataset.Name + "." + measure.Name
		datasetFields.Fields = append(datasetFields.Fields, listedMeasure(key, measure.Name, measure))
	}

	return datasetFields
}

// listedDimension and listedMeasure describe a field for the field list,
// under the key and name it is listed with: its own, or a view's.
func listedDimension(key string, name string, dimension Dimension) Field {
	field := Field{
		Key:          key,
		Title:        titleOrMadeFromName(dimension.Title, name),
		Kind:         FieldKindDimension,
		Type:         dimension.Type,
		IsPrimaryKey: dimension.IsPrimaryKey,
	}

	return field
}

func listedMeasure(key string, name string, measure Measure) Field {
	field := Field{
		Key:         key,
		Title:       titleOrMadeFromName(measure.Title, name),
		Kind:        FieldKindMeasure,
		Type:        FieldTypeNumber,
		MeasureType: measure.Type,
	}

	return field
}

// checkedModels is the models looked up by name, with the datasets that have
// a problem. Listing fields and building a query both start from it.
type checkedModels struct {
	datasetsByName map[string]Dataset
	viewsByName    map[string]View
	duplicateNames map[string]bool
	brokenDatasets map[string]bool
}

func checkModels(models Models) checkedModels {
	checked := checkedModels{
		datasetsByName: mapDatasetsByName(models.Datasets),
		viewsByName:    make(map[string]View),
		duplicateNames: findDuplicateNames(models),
		brokenDatasets: make(map[string]bool),
	}

	for _, view := range models.Views {
		checked.viewsByName[view.Name] = view
	}

	for _, dataset := range models.Datasets {
		datasetErrors := checkDataset(dataset, checked.datasetsByName, checked.duplicateNames)
		if len(datasetErrors) > 0 {
			checked.brokenDatasets[dataset.Name] = true
		}
	}

	return checked
}

func checkDataset(dataset Dataset, datasetsByName map[string]Dataset, duplicateNames map[string]bool) []ModelError {
	var modelErrors []ModelError

	if duplicateNames[dataset.Name] {
		modelError := ModelError{Code: ErrorDuplicateName, Source: dataset.Name}
		modelErrors = append(modelErrors, modelError)
	}

	if isReservedName(dataset.Name) {
		modelError := ModelError{Code: ErrorReservedName, Source: dataset.Name}
		modelErrors = append(modelErrors, modelError)
	}

	if len(dataset.Name) > longestName {
		modelError := ModelError{Code: ErrorNameTooLong, Source: dataset.Name}
		modelErrors = append(modelErrors, modelError)
	}

	fieldNamesSeen := make(map[string]bool)

	for _, name := range dataset.fieldNames() {
		if fieldNamesSeen[name] {
			modelError := ModelError{Code: ErrorDuplicateName, Source: dataset.Name, Field: name}
			modelErrors = append(modelErrors, modelError)
		}

		fieldNamesSeen[name] = true
	}

	for _, join := range dataset.Joins {
		_, targetExists := datasetsByName[join.OtherDataset]
		if !targetExists {
			modelError := ModelError{Code: ErrorUnknownJoinTarget, Source: dataset.Name, Field: join.OtherDataset}
			modelErrors = append(modelErrors, modelError)
		}
	}

	return modelErrors
}

func findDuplicateNames(models Models) map[string]bool {
	timesUsed := make(map[string]int)

	for _, dataset := range models.Datasets {
		timesUsed[dataset.Name]++
	}

	for _, view := range models.Views {
		timesUsed[view.Name]++
	}

	duplicateNames := make(map[string]bool)
	for name, count := range timesUsed {
		if count > 1 {
			duplicateNames[name] = true
		}
	}

	return duplicateNames
}

func mapDatasetsByName(datasets []Dataset) map[string]Dataset {
	datasetsByName := make(map[string]Dataset, len(datasets))

	for _, dataset := range datasets {
		datasetsByName[dataset.Name] = dataset
	}

	return datasetsByName
}

func sortDatasetsByName(datasets []Dataset) []Dataset {
	sorted := slices.Clone(datasets)

	slices.SortStableFunc(sorted, func(first, second Dataset) int {
		return strings.Compare(first.Name, second.Name)
	})

	return sorted
}

func sortViewsByName(views []View) []View {
	sorted := slices.Clone(views)

	slices.SortStableFunc(sorted, func(first, second View) int {
		return strings.Compare(first.Name, second.Name)
	})

	return sorted
}

func titleOrMadeFromName(title string, name string) string {
	if title != "" {
		return title
	}

	return MakeTitle(name)
}
