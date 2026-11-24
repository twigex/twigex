// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"strconv"
	"strings"
)

// The lines of a SELECT after FROM, written the same for every database. Each
// gives no line when it has nothing to say.

func whereLine(conditions []string) []string {
	if len(conditions) == 0 {
		return nil
	}

	return []string{"WHERE " + strings.Join(conditions, " AND ")}
}

// groupByLine groups by the first columns of the SELECT, by position: the
// grouped fields always come first.
func groupByLine(groupedCount int) []string {
	if groupedCount == 0 {
		return nil
	}

	positions := make([]string, 0, groupedCount)

	for position := 1; position <= groupedCount; position++ {
		positions = append(positions, strconv.Itoa(position))
	}

	return []string{"GROUP BY " + strings.Join(positions, ", ")}
}

func havingLine(conditions []string) []string {
	if len(conditions) == 0 {
		return nil
	}

	return []string{"HAVING " + strings.Join(conditions, " AND ")}
}

func orderByLine(orderParts []string) []string {
	if len(orderParts) == 0 {
		return nil
	}

	return []string{"ORDER BY " + strings.Join(orderParts, ", ")}
}
