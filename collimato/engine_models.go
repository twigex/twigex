// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
)

type UnsupportedReason string

const (
	UnsupportedWrittenByHand UnsupportedReason = "written_by_hand"
	UnsupportedCannotRead    UnsupportedReason = "cannot_read"
)

type UnsupportedFile struct {
	Name   string
	Reason UnsupportedReason
	Detail string
}

var engineFieldTypes = map[string]queryengine.FieldType{
	"string":  queryengine.FieldTypeText,
	"number":  queryengine.FieldTypeNumber,
	"time":    queryengine.FieldTypeTime,
	"boolean": queryengine.FieldTypeYesNo,
}

// The approximate distinct count is counted exactly: MySQL and Postgres have
// no built-in approximate count, and an exact one is always correct.
var engineMeasureTypes = map[string]queryengine.MeasureType{
	"count":                 queryengine.MeasureCount,
	"count_distinct":        queryengine.MeasureCountDistinct,
	"countDistinct":         queryengine.MeasureCountDistinct,
	"count_distinct_approx": queryengine.MeasureCountDistinct,
	"sum":                   queryengine.MeasureSum,
	"avg":                   queryengine.MeasureAverage,
	"min":                   queryengine.MeasureMinimum,
	"max":                   queryengine.MeasureMaximum,
	"number":                queryengine.MeasureFormula,
}

var engineRelationships = map[string]queryengine.Relationship{
	"one_to_one":  queryengine.OneToOne,
	"one_to_many": queryengine.OneToMany,
	"many_to_one": queryengine.ManyToOne,
}

// EngineModels turns a workspace's saved model files into the query engine's
// models. Only files made in the builder are supported; the rest are returned
// as unsupported. Files of any other type are skipped.
func EngineModels(files []model.CollimatoWorkspaceFile) (queryengine.Models, []UnsupportedFile) {
	models := queryengine.Models{}
	var unsupported []UnsupportedFile

	for _, file := range files {
		isDatasetFile := file.FileType == model.CollimatoFileTypeDataset
		isViewFile := file.FileType == model.CollimatoFileTypeView

		if !isDatasetFile && !isViewFile {
			continue
		}

		if file.BuilderModel == nil {
			writtenByHand := UnsupportedFile{Name: file.Name, Reason: UnsupportedWrittenByHand}
			unsupported = append(unsupported, writtenByHand)
			continue
		}

		if isDatasetFile {
			dataset, err := datasetFromBuilderModel(*file.BuilderModel)
			if err != nil {
				cannotRead := UnsupportedFile{Name: file.Name, Reason: UnsupportedCannotRead, Detail: err.Error()}
				unsupported = append(unsupported, cannotRead)
				continue
			}

			models.Datasets = append(models.Datasets, dataset)
			continue
		}

		view, err := viewFromBuilderModel(*file.BuilderModel)
		if err != nil {
			cannotRead := UnsupportedFile{Name: file.Name, Reason: UnsupportedCannotRead, Detail: err.Error()}
			unsupported = append(unsupported, cannotRead)
			continue
		}

		models.Views = append(models.Views, view)
	}

	return models, unsupported
}

func datasetFromBuilderModel(builderModel string) (queryengine.Dataset, error) {
	var saved model.Dataset
	if err := json.Unmarshal([]byte(builderModel), &saved); err != nil {
		return queryengine.Dataset{}, err
	}

	dataset := queryengine.Dataset{
		Name:         saved.Name,
		Table:        saved.SqlTable,
		ConnectionID: saved.ConnectionID,
	}

	for _, savedDimension := range saved.Dimensions {
		dimension, err := dimensionFromBuilder(savedDimension)
		if err != nil {
			return queryengine.Dataset{}, err
		}

		dataset.Dimensions = append(dataset.Dimensions, dimension)
	}

	for _, savedMeasure := range saved.Measures {
		measure, err := measureFromBuilder(savedMeasure)
		if err != nil {
			return queryengine.Dataset{}, err
		}

		dataset.Measures = append(dataset.Measures, measure)
	}

	for _, savedJoin := range saved.Joins {
		join, err := joinFromBuilder(savedJoin)
		if err != nil {
			return queryengine.Dataset{}, err
		}

		dataset.Joins = append(dataset.Joins, join)
	}

	return dataset, nil
}

func dimensionFromBuilder(saved model.DatasetDimension) (queryengine.Dimension, error) {
	fieldType, known := engineFieldTypes[saved.Type]
	if !known {
		return queryengine.Dimension{}, fmt.Errorf("dimension %q has unknown type %q", saved.Name, saved.Type)
	}

	// The builder hides a primary key unless it is marked public, and shows
	// every other field.
	isHidden := saved.PrimaryKey && !saved.Public

	dimension := queryengine.Dimension{
		Name:         saved.Name,
		Title:        saved.Title,
		Type:         fieldType,
		SQL:          saved.SQL,
		IsPrimaryKey: saved.PrimaryKey,
		IsHidden:     isHidden,
	}

	return dimension, nil
}

func measureFromBuilder(saved model.DatasetMeasure) (queryengine.Measure, error) {
	measureType, known := engineMeasureTypes[saved.Type]
	if !known {
		return queryengine.Measure{}, fmt.Errorf("measure %q has unknown type %q", saved.Name, saved.Type)
	}

	measure := queryengine.Measure{
		Name:   saved.Name,
		Title:  saved.Title,
		Type:   measureType,
		SQL:    saved.SQL,
		Format: saved.Format,
	}

	return measure, nil
}

func joinFromBuilder(saved model.DatasetJoin) (queryengine.Join, error) {
	relationship, known := engineRelationships[saved.Relationship]
	if !known {
		return queryengine.Join{}, fmt.Errorf("join to %q has unknown relationship %q", saved.Name, saved.Relationship)
	}

	join := queryengine.Join{
		OtherDataset:    saved.Name,
		Relationship:    relationship,
		Column:          saved.ThisColumn,
		OtherColumn:     saved.OtherColumn,
		CustomCondition: saved.SQL,
	}

	return join, nil
}

func viewFromBuilderModel(builderModel string) (queryengine.View, error) {
	var saved model.View
	if err := json.Unmarshal([]byte(builderModel), &saved); err != nil {
		return queryengine.View{}, err
	}

	view := queryengine.View{Name: saved.Name}

	for _, savedPart := range saved.Parts {
		part := viewPartFromBuilder(savedPart)

		view.Parts = append(view.Parts, part)
	}

	return view, nil
}

func viewPartFromBuilder(saved model.ViewPart) queryengine.ViewPart {
	path := strings.Split(saved.JoinPath, ".")

	choseAllFields := len(saved.Includes) == 1 && saved.Includes[0] == "*"

	var chosenFields []string
	if !choseAllFields {
		chosenFields = saved.Includes
	}

	part := queryengine.ViewPart{
		Path:         path,
		ChosenFields: chosenFields,
		AddPrefix:    saved.Prefix,
	}

	return part
}
