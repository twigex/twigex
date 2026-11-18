// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

// FieldList is what a workspace's charts can use: its datasets and views,
// each with its fields, and the saved models that cannot be used.
type FieldList struct {
	Datasets    []DatasetFields `json:"datasets"`
	ModelErrors []ModelError    `json:"modelErrors,omitempty"`
}

type DatasetFields struct {
	Name   string `json:"name"`
	Title  string `json:"title"`
	IsView bool   `json:"isView"`

	// JoinGroup is shared by datasets joined to each other, in either
	// direction. 0 for a dataset without joins, and for a view.
	JoinGroup int `json:"joinGroup"`

	Fields []Field `json:"fields"`
}

const (
	FieldKindDimension = "dimension"
	FieldKindMeasure   = "measure"
)

type Field struct {
	Key   string `json:"key"`
	Title string `json:"title"`

	// FullTitle is the dataset's title followed by the field's, as charts
	// and tables label a field: "Orders Revenue".
	FullTitle string `json:"fullTitle"`

	Kind string `json:"kind"`

	// Type is the builder's word: string, number, time or boolean.
	Type string `json:"type"`

	// MeasureType is how a measure adds rows up, in the builder's words: count,
	// count_distinct, sum, avg, min, max or number. Empty for a dimension.
	MeasureType string `json:"measureType,omitempty"`

	IsPrimaryKey bool `json:"isPrimaryKey,omitempty"`
}

// HasModelsThatMayComeBack says whether a model is missing from Datasets only
// until someone fixes it, so a missing dataset must not be read as deleted.
// A file written by hand stays unsupported.
func (fieldList FieldList) HasModelsThatMayComeBack() bool {
	for _, modelError := range fieldList.ModelErrors {
		if modelError.Reason != UnsupportedWrittenByHand {
			return true
		}
	}

	return false
}

// ModelError is a saved model the engine cannot use.
type ModelError struct {
	Name   string            `json:"name"`
	Reason UnsupportedReason `json:"reason"`

	// Detail is what is wrong, in the engine's words, for a model that
	// cannot be read or has errors.
	Detail string `json:"detail,omitempty"`

	// Message says the same in the reader's language, filled in by the API.
	Message string `json:"message,omitempty"`
}

// Answer is a chart's data: its columns, in the order asked for, and rows
// keyed by field key.
type Answer struct {
	Columns []AnswerColumn   `json:"columns"`
	Rows    []map[string]any `json:"rows"`

	// DateRanges has the actual start and end of every date field's range,
	// such as the days "this month" stands for.
	DateRanges []AnswerDateRange `json:"dateRanges"`
}

type AnswerColumn struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	FullTitle string `json:"fullTitle"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`

	// DateStep is set for a date field grouped by day, month and so on. Each
	// value is then the start of its period.
	DateStep string `json:"dateStep,omitempty"`
}

type AnswerDateRange struct {
	Field string `json:"field"`
	Start string `json:"start"`
	End   string `json:"end"`
}
