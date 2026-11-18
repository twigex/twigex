// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

type viewFieldsBuilder struct {
	view    View
	checked checkedModels

	dimensions  []Field
	measures    []Field
	namesUsed   map[string]bool
	modelErrors []ModelError

	// pathTo is the path the view takes to each dataset it reaches.
	pathTo map[string]string
}

// listViewFields lists a view's fields: all dimensions before all measures.
func listViewFields(view View, checked checkedModels) SourceFields {
	viewFields := SourceFields{
		Name:   view.Name,
		Title:  titleOrMadeFromName(view.Title, view.Name),
		IsView: true,
	}

	builder := viewFieldsBuilder{
		view:      view,
		checked:   checked,
		namesUsed: make(map[string]bool),
		pathTo:    make(map[string]string),
	}

	if checked.duplicateNames[view.Name] {
		builder.addError(ErrorDuplicateName, "")
	}

	for _, part := range view.Parts {
		builder.addPart(part)
	}

	if len(builder.modelErrors) > 0 {
		viewFields.Errors = builder.modelErrors
		return viewFields
	}

	viewFields.Fields = append(viewFields.Fields, builder.dimensions...)
	viewFields.Fields = append(viewFields.Fields, builder.measures...)

	return viewFields
}

func (builder *viewFieldsBuilder) addPart(part ViewPart) {
	lastDataset, pathIsValid := builder.followPath(part.Path)
	if !pathIsValid {
		return
	}

	reachesEachDatasetOneWay := builder.checkOnePathToEachDataset(part.Path)
	if !reachesEachDatasetOneWay {
		return
	}

	chosenAllFields := len(part.ChosenFields) == 0
	if chosenAllFields {
		builder.addAllFields(lastDataset, part.AddPrefix)
		return
	}

	for _, fieldName := range part.ChosenFields {
		builder.addChosenField(lastDataset, fieldName, part.AddPrefix)
	}
}

// followPath checks that each step of the path is a dataset joined from the
// step before it, and returns the last dataset on the path.
func (builder *viewFieldsBuilder) followPath(path []string) (Dataset, bool) {
	if len(path) == 0 {
		builder.addError(ErrorBrokenViewPath, "")
		return Dataset{}, false
	}

	firstName := path[0]
	current, exists := builder.checked.datasetsByName[firstName]
	if !exists {
		builder.addError(ErrorBrokenViewPath, firstName)
		return Dataset{}, false
	}

	for _, nextName := range path[1:] {
		next, exists := builder.checked.datasetsByName[nextName]
		_, isJoined := current.joinTo(nextName)

		if !exists || !isJoined {
			builder.addError(ErrorBrokenViewPath, nextName)
			return Dataset{}, false
		}

		current = next
	}

	pathUsesBrokenDataset := false
	for _, name := range path {
		if builder.checked.brokenDatasets[name] {
			builder.addError(ErrorUsesBrokenDataset, name)
			pathUsesBrokenDataset = true
		}
	}

	if pathUsesBrokenDataset {
		return Dataset{}, false
	}

	return current, true
}

// checkOnePathToEachDataset refuses a path that reaches a dataset through
// other datasets than an earlier path did: a sale's region through its shop
// and through its customer. A query joins each dataset once, so fields of
// both would come from whichever path the query happened to join first.
func (builder *viewFieldsBuilder) checkOnePathToEachDataset(path []string) bool {
	reachesEachDatasetOneWay := true

	for position, name := range path {
		pathHere := strings.Join(path[:position+1], " → ")

		earlierPath, reached := builder.pathTo[name]
		if !reached {
			builder.pathTo[name] = pathHere
			continue
		}

		if earlierPath != pathHere {
			builder.addError(ErrorViewReachesDatasetTwice, name)
			reachesEachDatasetOneWay = false
		}
	}

	return reachesEachDatasetOneWay
}

// addAllFields adds every field of the dataset, hidden ones too: a dataset's
// hidden primary key is shown in any view that takes all its fields.
func (builder *viewFieldsBuilder) addAllFields(dataset Dataset, addPrefix bool) {
	for _, dimension := range dataset.Dimensions {
		builder.addDimension(dataset, dimension, addPrefix)
	}

	for _, measure := range dataset.Measures {
		builder.addMeasure(dataset, measure, addPrefix)
	}
}

func (builder *viewFieldsBuilder) addChosenField(dataset Dataset, fieldName string, addPrefix bool) {
	dimension, isDimension := dataset.dimensionNamed(fieldName)
	if isDimension {
		builder.addDimension(dataset, *dimension, addPrefix)
		return
	}

	measure, isMeasure := dataset.measureNamed(fieldName)
	if isMeasure {
		builder.addMeasure(dataset, *measure, addPrefix)
		return
	}

	builder.addError(ErrorUnknownChosenField, fieldName)
}

func (builder *viewFieldsBuilder) addDimension(dataset Dataset, dimension Dimension, addPrefix bool) {
	name := viewFieldName(dataset, dimension.Name, addPrefix)
	if !builder.claimName(name) {
		return
	}

	key := builder.view.Name + "." + name
	builder.dimensions = append(builder.dimensions, listedDimension(key, name, dimension))
}

func (builder *viewFieldsBuilder) addMeasure(dataset Dataset, measure Measure, addPrefix bool) {
	name := viewFieldName(dataset, measure.Name, addPrefix)
	if !builder.claimName(name) {
		return
	}

	key := builder.view.Name + "." + name
	builder.measures = append(builder.measures, listedMeasure(key, name, measure))
}

func (builder *viewFieldsBuilder) claimName(name string) bool {
	if builder.namesUsed[name] {
		builder.addError(ErrorDuplicateFieldInView, name)
		return false
	}

	builder.namesUsed[name] = true

	return true
}

func (builder *viewFieldsBuilder) addError(code ErrorCode, field string) {
	modelError := ModelError{Code: code, Source: builder.view.Name, Field: field}

	builder.modelErrors = append(builder.modelErrors, modelError)
}

func viewFieldName(dataset Dataset, fieldName string, addPrefix bool) string {
	if addPrefix {
		return dataset.Name + "_" + fieldName
	}

	return fieldName
}
