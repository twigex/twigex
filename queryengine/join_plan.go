// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

type joinStep struct {
	from Dataset
	to   Dataset
	join Join
}

// joinPlan is the datasets a query reads, starting from one of them, and the
// joins that bring in the others.
type joinPlan struct {
	start Dataset
	steps []joinStep
}

func (plan joinPlan) datasets() []Dataset {
	datasets := []Dataset{plan.start}

	for _, step := range plan.steps {
		datasets = append(datasets, step.to)
	}

	return datasets
}

// findJoinPlan finds how to join the needed datasets. Each needed dataset is
// tried as the start, in the order given; joins are followed only in the
// direction they were made. The start needing the fewest joins wins, and on a
// tie the one given first.
func findJoinPlan(neededNames []string, datasetsByName map[string]Dataset) (joinPlan, error) {
	var best joinPlan
	found := false

	for _, startName := range neededNames {
		plan, reachesAll := planFromStart(startName, neededNames, datasetsByName)
		if !reachesAll {
			continue
		}

		if !found || len(plan.steps) < len(best.steps) {
			best = plan
			found = true
		}
	}

	if !found {
		return joinPlan{}, newRequestError(ErrorNoJoinPath, strings.Join(neededNames, ", "))
	}

	return best, nil
}

func planFromStart(startName string, neededNames []string, datasetsByName map[string]Dataset) (joinPlan, bool) {
	stepReaching := shortestPathsFrom(startName, datasetsByName)

	plan := joinPlan{start: datasetsByName[startName]}
	joined := map[string]bool{startName: true}

	for _, name := range neededNames {
		path, reached := pathTo(name, startName, stepReaching)
		if !reached {
			return joinPlan{}, false
		}

		for _, step := range path {
			if joined[step.to.Name] {
				continue
			}

			joined[step.to.Name] = true
			plan.steps = append(plan.steps, step)
		}
	}

	return plan, true
}

// shortestPathsFrom follows joins outward from a dataset, nearest datasets
// first, and gives for every dataset it reaches the join step that reached it.
func shortestPathsFrom(startName string, datasetsByName map[string]Dataset) map[string]joinStep {
	stepReaching := make(map[string]joinStep)
	reached := map[string]bool{startName: true}
	waiting := []string{startName}

	for len(waiting) > 0 {
		current := datasetsByName[waiting[0]]
		waiting = waiting[1:]

		for _, join := range current.Joins {
			other, exists := datasetsByName[join.OtherDataset]
			if !exists || reached[other.Name] {
				continue
			}

			reached[other.Name] = true
			stepReaching[other.Name] = joinStep{from: current, to: other, join: join}
			waiting = append(waiting, other.Name)
		}
	}

	return stepReaching
}

func pathTo(name string, startName string, stepReaching map[string]joinStep) ([]joinStep, bool) {
	var backwards []joinStep

	current := name
	for current != startName {
		step, reached := stepReaching[current]
		if !reached {
			return nil, false
		}

		backwards = append(backwards, step)
		current = step.from.Name
	}

	path := make([]joinStep, 0, len(backwards))
	for position := len(backwards) - 1; position >= 0; position-- {
		path = append(path, backwards[position])
	}

	return path, true
}

// viewJoinPlan joins the datasets along the paths of the view fields a
// request uses, rather than searching for joins.
func viewJoinPlan(viewName string, paths [][]string, datasetsByName map[string]Dataset) (joinPlan, error) {
	startName := paths[0][0]
	plan := joinPlan{start: datasetsByName[startName]}
	joined := map[string]bool{startName: true}

	for _, path := range paths {
		if path[0] != startName {
			return joinPlan{}, newRequestError(ErrorViewPathsStartApart, viewName)
		}

		for position := 1; position < len(path); position++ {
			from := datasetsByName[path[position-1]]
			to := datasetsByName[path[position]]

			if joined[to.Name] {
				continue
			}

			join, exists := from.joinTo(to.Name)
			if !exists {
				return joinPlan{}, newRequestError(ErrorDatasetHasErrors, viewName)
			}

			joined[to.Name] = true
			plan.steps = append(plan.steps, joinStep{from: from, to: to, join: join})
		}
	}

	return plan, nil
}

// rowsRepeat reports whether the plan's joins can make a row of the dataset
// appear more than once, so that adding its rows up would count some twice.
//
// The rule: walk the joins outward from the dataset, and the rows
// repeat if any dataset along the way joins further out to one where its rows
// can match many. A one-to-many join repeats the rows of the dataset it starts
// from; a many-to-one join repeats the rows of the dataset it points to.
func (plan joinPlan) rowsRepeat(datasetName string) bool {
	visited := make(map[string]bool)

	return plan.rowsRepeatFrom(datasetName, visited)
}

func (plan joinPlan) rowsRepeatFrom(current string, visited map[string]bool) bool {
	visited[current] = true

	for _, step := range plan.steps {
		neighbour, touches := step.otherEnd(current)
		if !touches || visited[neighbour] {
			continue
		}

		if step.repeatsRowsOf(current) {
			return true
		}
	}

	for _, step := range plan.steps {
		neighbour, touches := step.otherEnd(current)
		if !touches || visited[neighbour] {
			continue
		}

		if plan.rowsRepeatFrom(neighbour, visited) {
			return true
		}
	}

	return false
}

func (step joinStep) otherEnd(datasetName string) (string, bool) {
	if step.from.Name == datasetName {
		return step.to.Name, true
	}

	if step.to.Name == datasetName {
		return step.from.Name, true
	}

	return "", false
}

func (step joinStep) repeatsRowsOf(datasetName string) bool {
	startsHereToMany := step.from.Name == datasetName && step.join.Relationship == OneToMany
	manyPointHere := step.to.Name == datasetName && step.join.Relationship == ManyToOne

	return startsHereToMany || manyPointHere
}
