// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strings"
	"time"
)

// The grid's one-dataset shape is the sales table (test_tables_test.go) with
// fields written by hand. Its answers are worked out here in Go from the same
// five rows, so the grid checks the numbers, not only that the databases
// agree.

// salesGridRow is one row of the sales table. country and isPaid are nil
// where the table has no value.
type salesGridRow struct {
	id          int
	country     any
	channel     string
	amountCents int
	amountText  string
	quantity    int
	isPaid      any
}

var salesGridRows = []salesGridRow{
	{id: 1, country: "Latvia", channel: "Online", amountCents: 10050, amountText: "100.50", quantity: 2, isPaid: true},
	{id: 2, country: "Latvia", channel: "Shop", amountCents: 20000, amountText: "200.00", quantity: 1, isPaid: false},
	{id: 3, country: "Poland", channel: "Shop", amountCents: 5025, amountText: "50.25", quantity: 3, isPaid: true},
	{id: 4, country: nil, channel: "Online", amountCents: 1000, amountText: "10.00", quantity: 1, isPaid: nil},
	{id: 5, country: "Estonia", channel: "Online", amountCents: 30000, amountText: "300.00", quantity: 5, isPaid: true},
}

// The fields written by hand, as the SQL in salesGridShape works them out.

func (row salesGridRow) size() string {
	if row.amountCents >= 10000 {
		return "big"
	}

	return "small"
}

// paidOrOnline follows SQL: empty OR true is true, empty OR false is empty.
func (row salesGridRow) paidOrOnline() any {
	if row.channel == "Online" {
		return true
	}

	return row.isPaid
}

func (row salesGridRow) doubled() int {
	return row.quantity*2 + 1
}

// soldAt is 30 days apart for each id, and empty for the sale with no
// country.
func (row salesGridRow) soldAt() (time.Time, bool) {
	if row.country == nil {
		return time.Time{}, false
	}

	firstDay := time.Date(2025, time.January, 27, 0, 0, 0, 0, time.UTC)

	return firstDay.AddDate(0, 0, row.id*30), true
}

func salesGridTimeSQL(databaseType DatabaseType) string {
	if databaseType == DatabasePostgres {
		return "CASE WHEN {DATASET}.country IS NULL THEN NULL ELSE TIMESTAMP '2025-01-27' + ({DATASET}.id * 30) * INTERVAL '1 day' END"
	}

	return "CASE WHEN {DATASET}.country IS NULL THEN NULL ELSE TIMESTAMP('2025-01-27') + INTERVAL ({DATASET}.id * 30) DAY END"
}

func salesGridShape() gridShape {
	models := func(databaseType DatabaseType) Models {
		dataset := salesDataset()
		dataset.Dimensions = append(dataset.Dimensions,
			Dimension{Name: "size", Type: FieldTypeText, SQL: "CASE WHEN {DATASET}.amount >= 100 THEN 'big' ELSE 'small' END"},
			Dimension{Name: "paid_or_online", Type: FieldTypeYesNo, SQL: "{DATASET}.is_paid OR {DATASET}.channel = 'Online'"},
			Dimension{Name: "doubled", Type: FieldTypeNumber, SQL: "{DATASET}.quantity * 2 + 1"},
			Dimension{Name: "sold_at", Type: FieldTypeTime, SQL: salesGridTimeSQL(databaseType)},
		)

		return Models{Datasets: []Dataset{dataset}}
	}

	filters := []gridFilter{
		{
			name: "no filter",
		},
		{
			name:    "written text equals",
			filters: []Filter{{Field: "sales.size", Operator: FilterEquals, Values: []string{"big"}}},
			keeps:   func(row salesGridRow) bool { return row.size() == "big" },
		},
		{
			name:    "text not equals",
			filters: []Filter{{Field: "sales.country", Operator: FilterNotEquals, Values: []string{"Latvia"}}},
			keeps:   func(row salesGridRow) bool { return row.country != "Latvia" },
		},
		{
			name:    "text contains",
			filters: []Filter{{Field: "sales.country", Operator: FilterContains, Values: []string{"A"}}},
			keeps: func(row salesGridRow) bool {
				country, hasCountry := row.country.(string)
				return hasCountry && strings.Contains(strings.ToLower(country), "a")
			},
		},
		{
			name:    "written number greater",
			filters: []Filter{{Field: "sales.doubled", Operator: FilterGreaterThan, Values: []string{"4"}}},
			keeps:   func(row salesGridRow) bool { return row.doubled() > 4 },
		},
		{
			name:    "number contains",
			filters: []Filter{{Field: "sales.price", Operator: FilterContains, Values: []string{"0.5"}}},
			keeps:   func(row salesGridRow) bool { return strings.Contains(row.amountText, "0.5") },
		},
		{
			name:    "written yes/no with OR",
			filters: []Filter{{Field: "sales.paid_or_online", Operator: FilterEquals, Values: []string{"true"}}},
			keeps:   func(row salesGridRow) bool { return row.paidOrOnline() == true },
		},
		{
			name:    "date range",
			filters: []Filter{{Field: "sales.sold_at", Operator: FilterInDateRange, Values: []string{"2025-03-01", "2025-04-27"}}},
			keeps: func(row salesGridRow) bool {
				soldAt, isSet := row.soldAt()
				start := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
				dayAfterEnd := time.Date(2025, time.April, 28, 0, 0, 0, 0, time.UTC)

				return isSet && !soldAt.Before(start) && soldAt.Before(dayAfterEnd)
			},
		},
		{
			name:    "after date",
			filters: []Filter{{Field: "sales.sold_at", Operator: FilterAfterDate, Values: []string{"2025-04-27"}}},
			keeps: func(row salesGridRow) bool {
				soldAt, isSet := row.soldAt()
				dayAfter := time.Date(2025, time.April, 28, 0, 0, 0, 0, time.UTC)

				return isSet && !soldAt.Before(dayAfter)
			},
		},
		{
			name:    "not set",
			filters: []Filter{{Field: "sales.country", Operator: FilterIsNotSet}},
			keeps:   func(row salesGridRow) bool { return row.country == nil },
		},
		{
			name: "two filters",
			filters: []Filter{
				{Field: "sales.country", Operator: FilterNotEquals, Values: []string{"Estonia"}},
				{Field: "sales.paid_or_online", Operator: FilterEquals, Values: []string{"true"}},
			},
			keeps: func(row salesGridRow) bool { return row.country != "Estonia" && row.paidOrOnline() == true },
		},
		{
			name:      "measure greater",
			filters:   []Filter{{Field: "sales.count", Operator: FilterGreaterThan, Values: []string{"1"}}},
			onMeasure: true,
		},
	}

	return gridShape{
		name:       "one dataset",
		models:     models,
		dimensions: []string{"sales.country", "sales.size"},
		measures:   []string{"sales.count", "sales.revenue"},
		timeField:  "sales.sold_at",
		notShown:   "sales.quantity",
		filters:    filters,
		answer:     salesGridAnswer,
	}
}

// salesGridAnswer works out the rows a combination of the one-dataset shape
// must give.
func salesGridAnswer(combination gridCombination) []map[string]any {
	var kept []salesGridRow

	for _, row := range salesGridRows {
		keeps := combination.filter.keeps
		if keeps == nil || keeps(row) {
			kept = append(kept, row)
		}
	}

	if combination.rowsOnly {
		var rows []map[string]any

		for _, row := range kept {
			rows = append(rows, map[string]any{"sales.country": row.country, "sales.size": row.size()})
		}

		return rows
	}

	return salesGroups(kept, combination)
}

type salesGroup struct {
	values       map[string]any
	count        int
	revenueCents int
}

func salesGroups(kept []salesGridRow, combination gridCombination) []map[string]any {
	groupsByKey := make(map[string]*salesGroup)
	var keys []string

	for _, row := range kept {
		values := map[string]any{"sales.country": row.country, "sales.size": row.size()}

		if combination.step != "" {
			values["sales.sold_at"] = startOfGridStep(row, combination.step)
		}

		key := fmt.Sprint(values)

		group, exists := groupsByKey[key]
		if !exists {
			group = &salesGroup{values: values}
			groupsByKey[key] = group
			keys = append(keys, key)
		}

		group.count++
		group.revenueCents += row.amountCents
	}

	var rows []map[string]any

	for _, key := range keys {
		group := groupsByKey[key]

		// The one filter on a measure is "more than one sale".
		if combination.filter.onMeasure && group.count <= 1 {
			continue
		}

		row := group.values
		row["sales.count"] = float64(group.count)
		row["sales.revenue"] = float64(group.revenueCents) / 100

		rows = append(rows, row)
	}

	return rows
}

// startOfGridStep is the start of the day, week (Monday) or quarter a sale
// is in, as results write it, or nil when the sale has no time.
func startOfGridStep(row salesGridRow, step DateStep) any {
	soldAt, isSet := row.soldAt()
	if !isSet {
		return nil
	}

	var start time.Time

	switch step {
	case DateStepDay:
		start = soldAt
	case DateStepWeek:
		daysSinceMonday := (int(soldAt.Weekday()) + 6) % 7
		start = soldAt.AddDate(0, 0, -daysSinceMonday)
	case DateStepQuarter:
		firstMonth := time.Month((int(soldAt.Month())-1)/3*3 + 1)
		start = time.Date(soldAt.Year(), firstMonth, 1, 0, 0, 0, 0, time.UTC)
	}

	return dateForResult(start)
}
