// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

// fromLines gives the FROM line and one LEFT JOIN line per join. LEFT JOIN
// keeps rows that match nothing in the other dataset.
func (plan joinPlan) fromLines(syntax sqlSyntax) ([]string, error) {
	startTable, err := quotedTableName(syntax, plan.start.Table)
	if err != nil {
		return nil, err
	}

	lines := []string{"FROM " + startTable + " AS " + syntax.quoteName(plan.start.Name)}

	for _, step := range plan.steps {
		table, err := quotedTableName(syntax, step.to.Table)
		if err != nil {
			return nil, err
		}

		condition, err := joinCondition(syntax, step)
		if err != nil {
			return nil, err
		}

		line := "LEFT JOIN " + table + " AS " + syntax.quoteName(step.to.Name) + " ON " + condition
		lines = append(lines, line)
	}

	return lines, nil
}

// joinCondition is "this column = other column", or the join's own condition
// with {DATASET} standing for the dataset that has the join and {name} for the
// dataset it joins to.
func joinCondition(syntax sqlSyntax, step joinStep) (string, error) {
	fromName := syntax.quoteName(step.from.Name)
	toName := syntax.quoteName(step.to.Name)
	join := step.join

	if join.CustomCondition != "" {
		condition := strings.ReplaceAll(join.CustomCondition, "{"+ownDatasetReference+"}", fromName)
		condition = strings.ReplaceAll(condition, "{"+step.to.Name+"}", toName)

		// A join condition may name the two datasets, but no field.
		if strings.Contains(condition, "{") {
			return "", newRequestError(ErrorFieldReferenceInSQL, step.from.Name)
		}

		return condition, nil
	}

	if !isPlainName(join.Column) || !isPlainName(join.OtherColumn) {
		return "", newRequestError(ErrorBadJoin, step.from.Name+" → "+step.to.Name)
	}

	thisColumn := fromName + "." + syntax.quoteName(join.Column)
	otherColumn := toName + "." + syntax.quoteName(join.OtherColumn)

	return thisColumn + " = " + otherColumn, nil
}
