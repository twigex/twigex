// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

// viewFieldSource is where a view field comes from: the path through the
// datasets, the dataset at its end and the field's name there.
type viewFieldSource struct {
	path      []string
	dataset   Dataset
	fieldName string
}

// findViewFieldSource finds the field a view shows under a name, using the
// same naming rule as the field list. The view must have no model errors.
func findViewFieldSource(view View, shownName string, datasetsByName map[string]Dataset) (viewFieldSource, bool) {
	for _, part := range view.Parts {
		lastName := part.Path[len(part.Path)-1]
		dataset := datasetsByName[lastName]

		for _, fieldName := range viewPartFieldNames(part, dataset) {
			nameInView := viewFieldName(dataset, fieldName, part.AddPrefix)
			if nameInView != shownName {
				continue
			}

			source := viewFieldSource{path: part.Path, dataset: dataset, fieldName: fieldName}

			return source, true
		}
	}

	return viewFieldSource{}, false
}

func viewPartFieldNames(part ViewPart, dataset Dataset) []string {
	if len(part.ChosenFields) > 0 {
		return part.ChosenFields
	}

	var names []string

	for _, dimension := range dataset.Dimensions {
		names = append(names, dimension.Name)
	}

	for _, measure := range dataset.Measures {
		names = append(names, measure.Name)
	}

	return names
}
