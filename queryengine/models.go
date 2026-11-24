// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

type Models struct {
	Datasets []Dataset
	Views    []View

	// DatabaseTypes gives the kind of database of each connection, by
	// connection ID. A query on a connection not listed cannot be built.
	DatabaseTypes map[string]DatabaseType
}

type Dataset struct {
	Name         string
	Title        string
	Table        string
	ConnectionID string
	Dimensions   []Dimension
	Measures     []Measure
	Joins        []Join
}

func (dataset Dataset) dimensionNamed(name string) (*Dimension, bool) {
	for position := range dataset.Dimensions {
		if dataset.Dimensions[position].Name == name {
			return &dataset.Dimensions[position], true
		}
	}

	return nil, false
}

func (dataset Dataset) measureNamed(name string) (*Measure, bool) {
	for position := range dataset.Measures {
		if dataset.Measures[position].Name == name {
			return &dataset.Measures[position], true
		}
	}

	return nil, false
}

// fieldNames lists the names of the dataset's dimensions, then its measures.
func (dataset Dataset) fieldNames() []string {
	names := make([]string, 0, len(dataset.Dimensions)+len(dataset.Measures))

	for _, dimension := range dataset.Dimensions {
		names = append(names, dimension.Name)
	}

	for _, measure := range dataset.Measures {
		names = append(names, measure.Name)
	}

	return names
}

func (dataset Dataset) joinTo(otherName string) (Join, bool) {
	for _, join := range dataset.Joins {
		if join.OtherDataset == otherName {
			return join, true
		}
	}

	return Join{}, false
}

type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeNumber FieldType = "number"
	FieldTypeTime   FieldType = "time"
	FieldTypeYesNo  FieldType = "yes_no"
)

type Dimension struct {
	Name         string
	Title        string
	Type         FieldType
	SQL          string
	IsPrimaryKey bool
	IsHidden     bool
}

type MeasureType string

const (
	MeasureCount         MeasureType = "count"
	MeasureCountDistinct MeasureType = "count_distinct"
	MeasureSum           MeasureType = "sum"
	MeasureAverage       MeasureType = "average"
	MeasureMinimum       MeasureType = "minimum"
	MeasureMaximum       MeasureType = "maximum"
	MeasureFormula       MeasureType = "formula"
)

type Measure struct {
	Name   string
	Title  string
	Type   MeasureType
	SQL    string
	Format string
}

type Relationship string

const (
	OneToOne  Relationship = "one_to_one"
	OneToMany Relationship = "one_to_many"
	ManyToOne Relationship = "many_to_one"
)

type Join struct {
	OtherDataset    string
	Relationship    Relationship
	Column          string
	OtherColumn     string
	CustomCondition string
}

type View struct {
	Name  string
	Title string
	Parts []ViewPart
}

type ViewPart struct {
	Path []string

	// Empty means every field of the last dataset on the path.
	ChosenFields []string

	AddPrefix bool
}
