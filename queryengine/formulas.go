// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"regexp"
	"strings"
)

// A formula refers to another measure as {name} for one of its own dataset, or
// {dataset.name} for one of any dataset. {DATASET} means the formula's own
// dataset.
var measureReference = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\}`)

const ownDatasetReference = "DATASET"

// formulaReferences finds the measures a formula refers to, in the order they
// appear, each once.
func (builder *queryBuilder) formulaReferences(formula requestedField) ([]requestedField, error) {
	var references []requestedField
	seen := make(map[string]bool)

	for _, match := range measureReference.FindAllStringSubmatch(formula.measure.SQL, -1) {
		name := match[1]
		if name == ownDatasetReference || seen[name] {
			continue
		}

		seen[name] = true

		reference, err := builder.findReferencedMeasure(formula, name)
		if err != nil {
			return nil, err
		}

		references = append(references, reference)
	}

	return references, nil
}

func (builder *queryBuilder) findReferencedMeasure(formula requestedField, name string) (requestedField, error) {
	dataset := formula.dataset
	measureName := name

	datasetName, nameInDataset, hasDot := strings.Cut(name, ".")
	if hasDot {
		otherDataset, exists := builder.datasetsByName[datasetName]
		if !exists {
			return requestedField{}, newRequestError(ErrorUnknownField, formula.key+": "+name)
		}

		dataset = otherDataset
		measureName = nameInDataset
	}

	if builder.brokenDatasets[dataset.Name] {
		return requestedField{}, newRequestError(ErrorDatasetHasErrors, dataset.Name)
	}

	reference, found := findDatasetField(dataset, measureName)
	if !found || !reference.isMeasure() {
		return requestedField{}, newRequestError(ErrorUnknownField, formula.key+": "+name)
	}

	reference.key = dataset.Name + "." + measureName
	reference.viewPath = builder.viewPathTo(formula, dataset.Name)

	return reference, nil
}

// viewPathTo gives the path a view takes to a dataset that one of its
// formulas refers to: the formula's own path when it is the same dataset,
// otherwise the path of a view part that ends at that dataset.
func (builder *queryBuilder) viewPathTo(formula requestedField, datasetName string) []string {
	if formula.viewPath == nil {
		return nil
	}

	if datasetName == formula.dataset.Name {
		return formula.viewPath
	}

	view := builder.viewsByName[builder.firstSourceName]
	for _, part := range view.Parts {
		lastName := part.Path[len(part.Path)-1]
		if lastName == datasetName {
			return part.Path
		}
	}

	return nil
}

// usesOtherDatasets reports whether a formula needs measures of another
// dataset, directly or through the measures it refers to. Such a formula is
// worked out after each dataset's measures are put together.
func (builder *queryBuilder) usesOtherDatasets(measureField requestedField, visiting map[string]bool) (bool, error) {
	if measureField.measure.Type != MeasureFormula {
		return false, nil
	}

	if visiting[measureField.key] {
		return false, newRequestError(ErrorFormulaLoop, measureField.key)
	}

	visiting[measureField.key] = true
	defer delete(visiting, measureField.key)

	references, err := builder.formulaReferences(measureField)
	if err != nil {
		return false, err
	}

	usesOthers := false

	for _, reference := range references {
		if reference.dataset.Name != measureField.dataset.Name {
			usesOthers = true
		}

		referenceUsesOthers, err := builder.usesOtherDatasets(reference, visiting)
		if err != nil {
			return false, err
		}

		if referenceUsesOthers {
			usesOthers = true
		}
	}

	return usesOthers, nil
}

// fillInFormula writes a formula's SQL with {DATASET} and every reference
// replaced. A referenced formula is put in brackets, so that "{a} * 2" stays
// right when a is itself "x - y".
func (builder *queryBuilder) fillInFormula(formula requestedField, referenceSQL func(reference requestedField) (string, error)) (string, error) {
	references, err := builder.formulaReferences(formula)
	if err != nil {
		return "", err
	}

	sqlByName := make(map[string]string)

	for _, reference := range references {
		expression, err := referenceSQL(reference)
		if err != nil {
			return "", err
		}

		sqlByName[reference.key] = expression
	}

	ownDataset := builder.quote(formula.dataset.Name)

	filledIn := measureReference.ReplaceAllStringFunc(formula.measure.SQL, func(match string) string {
		name := strings.Trim(match, "{}")

		if name == ownDatasetReference {
			return ownDataset
		}

		if !strings.Contains(name, ".") {
			name = formula.dataset.Name + "." + name
		}

		return sqlByName[name]
	})

	// In brackets for the same reason as columnExpression's.
	return "(" + filledIn + ")", nil
}
