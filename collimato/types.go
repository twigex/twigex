// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package collimato

type DrillMembersGrouped struct {
	Dimensions []interface{} `json:"dimensions"`
	Measures   []interface{} `json:"measures"`
}

type Measure struct {
	AggType             string              `json:"aggType"`
	Cumulative          bool                `json:"cumulative"`
	CumulativeTotal     bool                `json:"cumulativeTotal"`
	DrillMembers        []interface{}       `json:"drillMembers"`
	DrillMembersGrouped DrillMembersGrouped `json:"drillMembersGrouped"`
	IsVisible           bool                `json:"isVisible"`
	Name                string              `json:"name"`
	Public              bool                `json:"public"`
	ShortTitle          string              `json:"shortTitle"`
	Title               string              `json:"title"`
	Type                string              `json:"type"`
}

type Dimension struct {
	IsVisible           bool   `json:"isVisible"`
	Name                string `json:"name"`
	PrimaryKey          bool   `json:"primaryKey"`
	Public              bool   `json:"public"`
	ShortTitle          string `json:"shortTitle"`
	SuggestFilterValues bool   `json:"suggestFilterValues"`
	Title               string `json:"title"`
	Type                string `json:"type"`
}

type Cube struct {
	Dimensions         []Dimension            `json:"dimensions"`
	Hierarchies        []interface{}          `json:"hierarchies"`
	IsVisible          bool                   `json:"isVisible"`
	Measures           []Measure              `json:"measures"`
	Name               string                 `json:"name"`
	Public             bool                   `json:"public"`
	Segments           []interface{}          `json:"segments"`
	Title              string                 `json:"title"`
	Type               string                 `json:"type"`
	ConnectedComponent int32                  `json:"connectedComponent"`
	Meta               map[string]interface{} `json:"meta"`
}

type CubeCollection struct {
	Cubes []Cube `json:"cubes"`
	Error string `json:"error"`
}
