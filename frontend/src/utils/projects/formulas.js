// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const headerKey = (h) => h?.header_name || h?.name || "";

const stripQuotes = (s) =>
    typeof s === "string" &&
    s.length >= 2 &&
    ((s.startsWith('"') && s.endsWith('"')) || (s.startsWith("'") && s.endsWith("'")))
        ? s.slice(1, -1)
        : s;

export const toNumber = (v) => {
    if (typeof v === "number") return Number.isFinite(v) ? v : NaN;
    if (typeof v === "string") {
        const cleaned = v.replace(",", ".").replace(/[^\d.-]/g, "");
        const n = parseFloat(cleaned);

        return Number.isFinite(n) ? n : NaN;
    }

    return NaN;
};

const coerceDate = (v) => {
    if (v == null || v === "") return null;
    if (v instanceof Date && !Number.isNaN(v.getTime())) return v;

    if (typeof v === "string") v = stripQuotes(v.trim());

    if (typeof v === "number" || /^-?\d+(\.\d+)?$/.test(String(v))) {
        const n = Number(v);

        if (Number.isFinite(n)) {
            const ms = n >= 1e12 ? n : n * 1000;
            const d = new Date(ms);

            return Number.isNaN(d.getTime()) ? null : d;
        }
    }

    // A date field comes as its day at midnight UTC, which is that day
    // wherever the user is, not the moment it names.
    if (typeof v === "string") {
        const m = v.match(/^(\d{4})-(\d{2})-(\d{2})(?:T00:00:00(?:\.0+)?Z)?$/);

        if (m) {
            const y = parseInt(m[1], 10);
            const mo = parseInt(m[2], 10) - 1;
            const d = parseInt(m[3], 10);
            const dt = new Date(y, mo, d);

            return Number.isNaN(dt.getTime()) ? null : dt;
        }
    }

    const dt = new Date(v);

    return Number.isNaN(dt.getTime()) ? null : dt;
};

const datedif = (start, end, unit) => {
    const s = coerceDate(start);
    const e = coerceDate(end);

    if (!s || !e) return null;

    const u = String(unit || "D").toUpperCase();

    if (u === "D") return Math.floor((e - s) / 86400000);

    if (u === "M") {
        let months = (e.getFullYear() - s.getFullYear()) * 12 + (e.getMonth() - s.getMonth());

        if (e.getDate() < s.getDate()) months -= 1;

        return months;
    }

    if (u === "Y") {
        let years = e.getFullYear() - s.getFullYear();

        if (
            e.getMonth() < s.getMonth() ||
            (e.getMonth() === s.getMonth() && e.getDate() < s.getDate())
        )
            years -= 1;

        return years;
    }

    return null;
};

// Whether a formula reads any field. A field slot left empty when it was set
// up reads none.
export const readsFields = (spec) =>
    Array.isArray(spec?.args) &&
    spec.args.some((a) => a?.type === "field" && String(stripQuotes(a.value ?? "")).trim() !== "");

// formulaArgsFor turns a saved formula's arguments back into what the formula
// editor holds: the header for an argument that names a field, or the value
// typed. A variadic formula marks its field arguments itself; a fixed one
// says in its catalog entry which of its arguments are fields.
export const formulaArgsFor = (formula, spec, headers) =>
    (spec?.args || []).map((a, i) => {
        const isField =
            formula.arity === "variadic" ? a.type === "field" : formula.args?.[i]?.kind === "field";

        return isField
            ? headers.find((h) => (h.header_name || h.name) === a.value) || null
            : (a.value ?? "");
    });

export const normalizeFormulaSpec = (spec) => {
    if (!spec) return null;
    try {
        return JSON.parse(JSON.stringify(spec));
    } catch {}

    return typeof spec === "object" ? { ...spec } : null;
};

// Returns true when a calculations cell shows a computed value, so editing
// what it stores would change nothing. Only an aggregate over no fields
// shows the value stored in the cell; TODAY or ROUND of constants compute.
export const isComputedCalculations = (header) => {
    if ((header?.header_usage || "").toLowerCase() !== "calculations") return false;
    const spec = normalizeFormulaSpec(header?.formula);
    const aggregate = ["SUM", "AVERAGE", "MIN", "MAX", "PRODUCT"].includes(
        String(spec?.name || "").toUpperCase(),
    );
    const overFields = readsFields(spec);

    return !aggregate || overFields;
};

const resolveArg = (argSpec, row) => {
    if (!argSpec) return null;
    if (argSpec.type === "field") {
        const key = stripQuotes(argSpec.value);

        return row?.[key];
    }

    if (argSpec.type === "const") return stripQuotes(argSpec.value);

    return null;
};

export const evaluateFormulaForRow = (header, row) => {
    const spec = header?.formula;

    if (!spec || !spec.name) return "";

    const rawArgs = Array.isArray(spec.args) ? spec.args.map((a) => resolveArg(a, row)) : [];
    const fn = String(spec.name).toUpperCase();

    const isAggregator = ["SUM", "AVERAGE", "MIN", "MAX", "PRODUCT"].includes(fn);
    const hasFieldArgs = readsFields(spec);

    if (isAggregator && hasFieldArgs) {
        const nums = rawArgs.map(toNumber).filter(Number.isFinite);

        switch (fn) {
            case "SUM":
                return nums.reduce((acc, n) => acc + n, 0);
            case "AVERAGE":
                return nums.length ? nums.reduce((a, b) => a + b, 0) / nums.length : 0;
            case "MIN":
                return nums.length ? Math.min(...nums) : "";
            case "MAX":
                return nums.length ? Math.max(...nums) : "";
            case "PRODUCT":
                return nums.length ? nums.reduce((a, b) => a * b, 1) : "";
        }
    }

    if (isAggregator && !hasFieldArgs) {
        const key = headerKey(header);
        const v = toNumber(row?.[key]);

        return Number.isFinite(v) ? v : "";
    }

    switch (fn) {
        case "SUM": {
            const nums = rawArgs.map(toNumber).filter(Number.isFinite);

            return nums.reduce((acc, n) => acc + n, 0);
        }

        case "AVERAGE": {
            const nums = rawArgs.map(toNumber).filter(Number.isFinite);

            return nums.length ? nums.reduce((a, b) => a + b, 0) / nums.length : 0;
        }

        case "PRODUCT": {
            const nums = rawArgs.map(toNumber).filter(Number.isFinite);

            return nums.length ? nums.reduce((a, b) => a * b, 1) : 0;
        }

        case "MIN": {
            const nums = rawArgs.map(toNumber).filter(Number.isFinite);

            return nums.length ? Math.min(...nums) : "";
        }

        case "MAX": {
            const nums = rawArgs.map(toNumber).filter(Number.isFinite);

            return nums.length ? Math.max(...nums) : "";
        }

        case "ROUND": {
            const n = toNumber(rawArgs[0]);
            const d = Math.max(0, parseInt(rawArgs[1] ?? 0, 10) || 0);

            if (!Number.isFinite(n)) return "";
            const p = Math.pow(10, d);

            return Math.round(n * p) / p;
        }

        case "ROUNDUP": {
            const n = toNumber(rawArgs[0]);
            const d = Math.max(0, parseInt(rawArgs[1] ?? 0, 10) || 0);

            if (!Number.isFinite(n)) return "";
            const p = Math.pow(10, d);

            return Math.ceil(n * p) / p;
        }

        case "ROUNDDOWN": {
            const n = toNumber(rawArgs[0]);
            const d = Math.max(0, parseInt(rawArgs[1] ?? 0, 10) || 0);

            if (!Number.isFinite(n)) return "";
            const p = Math.pow(10, d);

            return Math.floor(n * p) / p;
        }

        case "ABS": {
            const n = toNumber(rawArgs[0]);

            return Number.isFinite(n) ? Math.abs(n) : "";
        }

        case "POWER": {
            const base = toNumber(rawArgs[0]);
            const exp = toNumber(rawArgs[1]);

            return Number.isFinite(base) && Number.isFinite(exp) ? Math.pow(base, exp) : "";
        }

        case "IF": {
            const test = !!rawArgs[0];

            return test ? rawArgs[1] : rawArgs[2];
        }

        case "AND": {
            return rawArgs.every((v) => !!v);
        }

        case "OR": {
            return rawArgs.some((v) => !!v);
        }

        case "NOT": {
            return !rawArgs[0];
        }

        case "CONCATENATE": {
            return rawArgs.map((v) => (v == null ? "" : String(v))).join("");
        }

        case "LEFT": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);
            const n = parseInt(rawArgs[1] ?? 1, 10) || 1;

            return s.slice(0, Math.max(0, n));
        }

        case "RIGHT": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);
            const n = parseInt(rawArgs[1] ?? 1, 10) || 1;

            return s.slice(-Math.max(0, n));
        }

        case "MID": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);
            const start1 = parseInt(rawArgs[1] ?? 1, 10) || 1;
            const len = Math.max(0, parseInt(rawArgs[2] ?? 1, 10) || 1);
            const start0 = Math.max(0, start1 - 1);

            return s.substring(start0, start0 + len);
        }

        case "LEN": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);

            return s.length;
        }

        case "TRIM": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);

            return s.trim().replace(/\s+/g, " ");
        }

        case "UPPER": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);

            return s.toUpperCase();
        }

        case "LOWER": {
            const s = rawArgs[0] == null ? "" : String(rawArgs[0]);

            return s.toLowerCase();
        }

        case "TODAY": {
            const d = new Date();

            d.setHours(0, 0, 0, 0);

            return d;
        }

        case "NOW": {
            return new Date();
        }

        case "DATE": {
            const y = parseInt(rawArgs[0], 10) || 0;
            const m = (parseInt(rawArgs[1], 10) || 1) - 1;
            const d = parseInt(rawArgs[2], 10) || 1;
            const dt = new Date(y, m, d);

            return Number.isNaN(dt.getTime()) ? "" : dt;
        }

        case "YEAR": {
            const d = coerceDate(rawArgs[0]);

            return d ? d.getFullYear() : "";
        }

        case "MONTH": {
            const d = coerceDate(rawArgs[0]);

            return d ? d.getMonth() + 1 : "";
        }

        case "DAY": {
            const d = coerceDate(rawArgs[0]);

            return d ? d.getDate() : "";
        }

        case "DATEDIF": {
            const unit = stripQuotes(String(rawArgs[2] ?? "D"));
            const out = datedif(rawArgs[0], rawArgs[1], unit);

            return out == null ? "" : out;
        }

        case "WEEKDAY": {
            const d = coerceDate(rawArgs[0]);

            if (!d) return "";
            const type = parseInt(rawArgs[1] ?? 1, 10) || 1;
            const js = d.getDay(); // 0=Sun..6=Sat

            if (type === 2 || type === 3) return js === 0 ? 7 : js; // Mon=1..Sun=7

            return js + 1; // Sun=1..Sat=7
        }

        default:
            return "";
    }
};

export const calculateRowValueForRow = (header, row) => {
    if (!row) return "";
    if ((header?.header_usage || "").toLowerCase() !== "calculations") return "";

    const spec = normalizeFormulaSpec(header?.formula);

    if (!spec || !spec.name) return "";

    return evaluateFormulaForRow({ ...header, formula: spec }, row);
};

export const formatCellValue = (v, header) => {
    if (v instanceof Date && !Number.isNaN(v.getTime())) {
        const dd = String(v.getDate()).padStart(2, "0");
        const mm = String(v.getMonth() + 1).padStart(2, "0");
        const yyyy = v.getFullYear();
        const dateStr = `${dd}.${mm}.${yyyy}`;

        const hasTime =
            v.getHours() !== 0 ||
            v.getMinutes() !== 0 ||
            v.getSeconds() !== 0 ||
            v.getMilliseconds() !== 0;

        if (hasTime) {
            const HH = String(v.getHours()).padStart(2, "0");
            const MI = String(v.getMinutes()).padStart(2, "0");

            return `${dateStr} ${HH}:${MI}`;
        }

        return dateStr;
    }

    if (typeof v === "number" && Number.isFinite(v)) {
        // A field that names its decimals shows that many; otherwise up to
        // two, so 1.25 + 1.25 is 2.5, not 3.
        if (header?.decimals != null && header.decimals !== "") {
            const d = Math.max(0, parseInt(header.decimals, 10) || 0);

            return v.toFixed(d);
        }

        return String(Math.round(v * 100) / 100);
    }

    if (v === null || v === undefined || v === "") return " ";

    return String(v);
};

// The footer of an aggregate column applies the column's own function to its
// rows' results: the average of the averages, not their sum.
const combineTotals = (name, nums) => {
    if (!nums.length) return "";

    switch (name) {
        case "SUM":
            return nums.reduce((a, b) => a + b, 0);
        case "AVERAGE":
            return nums.reduce((a, b) => a + b, 0) / nums.length;
        case "MIN":
            return Math.min(...nums);
        case "MAX":
            return Math.max(...nums);
        case "PRODUCT":
            return nums.reduce((a, b) => a * b, 1);
        default:
            return "";
    }
};

// The value a calculations column shows under its rows.
export const columnTotal = (header, rows) => {
    if (!Array.isArray(rows)) return "";
    if ((header?.header_usage || "").toLowerCase() !== "calculations") return "";

    const spec = normalizeFormulaSpec(header?.formula);

    if (!spec || !spec.name) return "";

    const name = String(spec.name).toUpperCase();
    const AGGREGATORS = new Set(["SUM", "AVERAGE", "MIN", "MAX", "PRODUCT"]);

    const hasFieldArgs = readsFields(spec);
    const isVariadicAggregator = AGGREGATORS.has(name) && hasFieldArgs;

    if (isVariadicAggregator) {
        const nums = rows
            .map((row) => toNumber(evaluateFormulaForRow({ ...header, formula: spec }, row)))
            .filter(Number.isFinite);

        return combineTotals(name, nums);
    }

    if (AGGREGATORS.has(name) && !hasFieldArgs) {
        const key = headerKey(header);
        const nums = rows.map((row) => toNumber(row?.[key])).filter(Number.isFinite);

        return combineTotals(name, nums);
    }

    const nums = rows
        .map((row) => toNumber(evaluateFormulaForRow({ ...header, formula: spec }, row)))
        .filter(Number.isFinite);

    if (!nums.length) return "";

    return nums.reduce((a, b) => a + b, 0);
};
