// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func onPostgres(models Models) Models {
	models.DatabaseTypes = map[string]DatabaseType{"connection-1": DatabasePostgres}

	return models
}

func buildPostgresQuery(t *testing.T, models Models, request Request) BuiltQuery {
	t.Helper()

	query, err := BuildQuery(onPostgres(models), request)
	if err != nil {
		t.Fatalf("BuildQuery: %v", err)
	}

	return query
}

func TestBuildQueryPostgresNamesAndNumberedValues(t *testing.T) {
	request := Request{
		Dimensions: []string{"sales.country"},
		Measures:   []string{"sales.count"},
		Filters: []Filter{
			{Field: "sales.channel", Operator: FilterEquals, Values: []string{"Shop", "Online"}},
			{Field: "sales.note", Operator: FilterContains, Values: []string{"50%"}},
			{Field: "sales.country", Operator: FilterStartsWith, Values: []string{"L"}},
			{Field: "sales.count", Operator: FilterGreaterThan, Values: []string{"1"}},
		},
		RowLimit: 10,
	}

	query := buildPostgresQuery(t, salesModels(), request)

	wantSQL := `SELECT "sales"."country" AS "sales__country", COUNT("sales"."id") AS "sales__count"` + "\n" +
		`FROM "sales" AS "sales"` + "\n" +
		`WHERE "sales"."channel" IN ($1, $2) AND (CAST("sales"."note" AS TEXT) ILIKE ('%' || $3 || '%') ESCAPE '!') AND (CAST("sales"."country" AS TEXT) ILIKE ($4 || '%') ESCAPE '!')` + "\n" +
		"GROUP BY 1\n" +
		`HAVING COUNT("sales"."id") > $5` + "\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Shop", "Online", "50!%", "L", "1"})
}

// A ? in SQL written by hand is a Postgres operator, not a value.
func TestBuildQueryPostgresLeavesQuestionMarkInWrittenSQL(t *testing.T) {
	models := salesModels()
	sales := models.Datasets[0]
	sales.Dimensions = append(sales.Dimensions, Dimension{Name: "has_key", Type: FieldTypeYesNo, SQL: "{DATASET}.data ? 'key'"})
	models.Datasets[0] = sales

	request := Request{
		Dimensions: []string{"sales.has_key"},
		Filters:    []Filter{{Field: "sales.country", Operator: FilterEquals, Values: []string{"Latvia"}}},
		RowLimit:   10,
	}

	query := buildPostgresQuery(t, models, request)

	wantSQL := `SELECT ("sales".data ? 'key') AS "sales__has_key"` + "\n" +
		`FROM "sales" AS "sales"` + "\n" +
		`WHERE "sales"."country" = $1` + "\n" +
		"GROUP BY 1\n" +
		"LIMIT 10"

	checkSQL(t, query, wantSQL, []any{"Latvia"})
}

func TestBuildQueryPostgresDates(t *testing.T) {
	request := Request{
		Measures:   []string{"events.count"},
		DateFields: []DateField{{Field: "events.happened_at", Step: DateStepWeek, Range: DateRange{Start: "2025-03-01", End: "2025-03-31"}}},
		Filters:    []Filter{{Field: "events.happened_at", Operator: FilterBeforeDate, Values: []string{"2025-03-20"}}},
		RowLimit:   10,
	}

	query := buildPostgresQuery(t, eventsModels(), request)

	wantSQL := `SELECT date_trunc('week', ("events"."happened_at"::timestamptz AT TIME ZONE 'UTC')) AS "events__happened_at", COUNT("events"."id") AS "events__count"` + "\n" +
		`FROM "events" AS "events"` + "\n" +
		`WHERE "events"."happened_at" < $1::timestamptz AND ("events"."happened_at" >= $2::timestamptz AND "events"."happened_at" < $3::timestamptz)` + "\n" +
		"GROUP BY 1\n" +
		"LIMIT 10"

	wantValues := []any{"2025-03-20T00:00:00.000Z", "2025-03-01T00:00:00.000Z", "2025-04-01T00:00:00.000Z"}

	checkSQL(t, query, wantSQL, wantValues)
}

func TestBuildQueryPostgresCombinedQueryMatchesEmptyValues(t *testing.T) {
	request := Request{
		Dimensions: []string{"shops.format"},
		Measures:   []string{"shop_sales.revenue", "shops.total_floor_area"},
		RowLimit:   10,
	}

	query := buildPostgresQuery(t, shopModels(), request)

	wantLines := []string{
		`FROM (SELECT "shops__format" FROM "part_1" UNION SELECT "shops__format" FROM "part_2") AS "all_groups"`,
		`LEFT JOIN "part_1" ON "all_groups"."shops__format" IS NOT DISTINCT FROM "part_1"."shops__format"`,
		`LEFT JOIN "part_2" ON "all_groups"."shops__format" IS NOT DISTINCT FROM "part_2"."shops__format"`,
	}

	gotLines := strings.Split(query.SQL, "\n")

	for _, line := range wantLines {
		if !slices.Contains(gotLines, line) {
			t.Errorf("want line %q in\n%s", line, query.SQL)
		}
	}
}

func TestBuildQueryUnknownDatabaseType(t *testing.T) {
	models := salesModels()
	models.DatabaseTypes = map[string]DatabaseType{"connection-1": "sqlite"}

	_, err := BuildQuery(models, Request{Measures: []string{"sales.count"}, RowLimit: 10})

	var requestError *RequestError
	if !errors.As(err, &requestError) || requestError.Code != ErrorUnknownDatabaseType {
		t.Fatalf("want ErrorUnknownDatabaseType, got %v", err)
	}

	if requestError.Field != "sales" {
		t.Errorf("want field sales, got %q", requestError.Field)
	}
}
