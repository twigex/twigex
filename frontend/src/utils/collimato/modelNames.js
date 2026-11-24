// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// A plain name is a column name as written, not SQL such as a CASE.
const PLAIN_NAME = /^[A-Za-z_][A-Za-z0-9_]*$/;

function findByName(items, name) {
    const exact = items.find((item) => item.name === name);

    if (exact) {
        return { exact: true, found: exact.name };
    }

    const lowered = name.toLowerCase();
    const sameButCase = items.find((item) => item.name.toLowerCase() === lowered);

    return { exact: false, found: sameButCase ? sameButCase.name : null };
}

// columnProblem checks one plain column name. MySQL ignores case in column
// names, so there only a missing column is a problem; Postgres reads a name
// exactly as the engine writes it.
function columnProblem(columns, savedName, databaseType) {
    if (!PLAIN_NAME.test(savedName ?? "")) {
        return null;
    }

    const column = findByName(columns, savedName);

    if (column.exact) {
        return null;
    }

    if (column.found && databaseType === "mysql") {
        return null;
    }

    return { saved: savedName, found: column.found };
}

// modelNameProblems compares a dataset made in the builder with the tables of
// its connection, as the table list gives them. A name the database has only
// in another case comes with the database's name to use instead; one it does
// not have at all comes without. Table names are compared exactly on both
// databases: MySQL on Linux reads table names exactly too.
export function modelNameProblems(model, connectionTables, databaseType) {
    const problems = [];

    const table = findByName(connectionTables, model.sql_table);

    if (!table.exact) {
        problems.push({ kind: "table", saved: model.sql_table, found: table.found });
    }

    const tableName = table.exact ? model.sql_table : table.found;

    if (!tableName) {
        return problems;
    }

    const columns = connectionTables.find((t) => t.name === tableName)?.columns ?? [];

    (model.dimensions ?? []).forEach((dimension) => {
        const problem = columnProblem(columns, dimension.sql, databaseType);

        if (problem) {
            problems.push({ kind: "column", field: dimension.name, ...problem });
        }
    });

    (model.measures ?? []).forEach((measure) => {
        const problem = columnProblem(columns, measure.sql, databaseType);

        if (problem) {
            problems.push({ kind: "total", field: measure.name, ...problem });
        }
    });

    return problems;
}

// withDatabaseNames gives the model with each name the database has in
// another case replaced by the database's. A field keeps its own name, so the
// charts using it keep working; only the column it reads changes.
export function withDatabaseNames(model, problems) {
    const fixed = JSON.parse(JSON.stringify(model));

    problems.forEach((problem) => {
        if (!problem.found) {
            return;
        }

        if (problem.kind === "table") {
            fixed.sql_table = problem.found;

            return;
        }

        if (problem.kind === "total") {
            const measure = fixed.measures.find((m) => m.name === problem.field);

            measure.sql = problem.found;

            return;
        }

        const dimension = fixed.dimensions.find((d) => d.name === problem.field);

        dimension.sql = problem.found;

        if (dimension.name !== dimension.sql) {
            dimension.custom = true;
        }
    });

    return fixed;
}
