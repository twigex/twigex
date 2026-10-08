// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, ref, watch } from "vue";

const EXCLUDED_USAGE = new Set(["link", "master link", "status", "assignee", "default_assignee"]);
const USAGE_KEY = "calc_header_usage_v1";

const headerKey = (h) => h?.header_name || h?.name || "";
const displayName = (h) => h?.display_name || h?.display_name?.String || headerKey(h);
const usageOf = (h) => (h.header_usage || "").toLowerCase();

// useFormulaEditor holds what a calculation field's formula editor shows: the
// formula picked, its arguments, and whether a formula over many columns
// reads several of them or its own column. Fields a person picks often are
// offered first, remembered in this browser.
export function useFormulaEditor(tableHeaders) {
    const chosenFormula = ref(null);
    const formulaArgs = ref([]);
    const useMultipleFields = ref(false);

    const usageCounts = ref({});

    try {
        usageCounts.value = JSON.parse(localStorage.getItem(USAGE_KEY) || "{}");
    } catch {
        usageCounts.value = {};
    }

    const getUsage = (h) => usageCounts.value[headerKey(h)] || 0;
    const bumpUsage = (h) => {
        const k = headerKey(h);

        if (!k) return;
        usageCounts.value[k] = (usageCounts.value[k] || 0) + 1;
        try {
            localStorage.setItem(USAGE_KEY, JSON.stringify(usageCounts.value));
        } catch {
            // Remembering is only a convenience.
        }
    };
    const byUsageThenName = (a, b) =>
        getUsage(b) - getUsage(a) || (displayName(a) || "").localeCompare(displayName(b) || "");

    const allHeaders = computed(() =>
        (tableHeaders.value || []).filter((h) => !EXCLUDED_USAGE.has(usageOf(h))),
    );
    const fieldsOf = (usages) =>
        computed(() =>
            allHeaders.value.filter((h) => usages.includes(usageOf(h))).sort(byUsageThenName),
        );
    const numberFields = fieldsOf(["number", "decimal"]);
    const dateFields = fieldsOf(["default_date", "date"]);
    const boolFields = fieldsOf(["bool"]);
    const textFields = fieldsOf(["text", "url"]);

    const fieldOptionsForArg = (arg) => {
        const typ = (arg?.type || "").toLowerCase();

        if (typ === "number") return numberFields.value;
        if (typ === "date") return dateFields.value;
        if (typ === "boolean") return boolFields.value;
        if (typ === "text") return textFields.value;

        return [
            ...numberFields.value,
            ...dateFields.value,
            ...textFields.value,
            ...boolFields.value,
        ];
    };

    watch(
        () => formulaArgs.value.map((v) => v && headerKey(v)),
        (curr, prev) => {
            const specs = chosenFormula.value?.args || [];

            curr.forEach((key, i) => {
                if (specs[i]?.kind === "field" && key && key !== prev?.[i])
                    bumpUsage(formulaArgs.value[i]);
            });
        },
        { deep: true },
    );

    watch(chosenFormula, (formula) => {
        useMultipleFields.value = false;
        if (formula?.arity === "variadic") formulaArgs.value = [];
    });

    // payload is the formula as it is saved: its name, and its arguments as
    // fields by name or as typed values.
    const payload = () => {
        const f = chosenFormula.value;

        if (!f) return null;

        if (f.arity === "variadic") {
            const args =
                useMultipleFields.value && Array.isArray(formulaArgs.value)
                    ? formulaArgs.value.map((v) => ({
                          type: "field",
                          value: v ? headerKey(v) : "",
                      }))
                    : [];

            return { name: f.name, args };
        }

        const args = (f.args || []).map((spec, i) => {
            const v = formulaArgs.value[i];

            if (spec.kind === "field") return { type: "field", value: v ? headerKey(v) : "" };

            return { type: "const", value: String(v ?? spec.default ?? "") };
        });

        return { name: f.name, args };
    };

    const reset = () => {
        chosenFormula.value = null;
        formulaArgs.value = [];
        useMultipleFields.value = false;
    };

    return {
        chosenFormula,
        formulaArgs,
        useMultipleFields,
        allHeaders,
        fieldOptionsForArg,
        payload,
        reset,
    };
}
