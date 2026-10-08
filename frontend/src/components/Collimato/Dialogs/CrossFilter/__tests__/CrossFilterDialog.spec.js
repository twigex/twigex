// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import CrossFilterDialog from "@/components/Collimato/Dialogs/CrossFilterDialog.vue";

vi.mock("@/services/collimatoService", () => ({
    default: {
        loadData: vi.fn().mockResolvedValue({ data: { data: [] } }),
    },
}));

vi.mock("vue-router", () => ({
    useRoute: () => ({ params: { workspaceId: "ws1", id: "d1" } }),
}));

const dataModels = [
    {
        name: "orders",
        dimensions: [
            { id: 1, name: "orders.status", type: "string" },
            { id: 2, name: "orders.created_at", type: "time" },
        ],
    },
];

const charts = [{ id: "chart-a", name: "A" }];

let wrappers = [];

function dialog(props = {}) {
    const w = mount(CrossFilterDialog, {
        props: {
            modelValue: true,
            dataModels,
            charts,
            editFilter: null,
            ...props,
        },
        attachTo: document.body,
    });

    wrappers.push(w);

    return w;
}

// The dialog renders through a HeadlessUI portal, so its markup lives on
// document.body rather than under the wrapper.
const find = (selector) => document.body.querySelector(selector);
const findAll = (selector) => [...document.body.querySelectorAll(selector)];

async function settled() {
    await vi.runAllTimersAsync();
}

async function click(label) {
    findAll("button")
        .find((b) => b.textContent.trim() === label)
        .click();
    await settled();
}

async function choose(picker, option) {
    await click(picker);
    findAll("li")
        .find((li) => li.textContent.trim() === option)
        .click();
    await settled();
}

async function fill(selector, value) {
    const el = find(selector);

    el.value = value;
    el.dispatchEvent(new Event("input"));
    await settled();
}

async function closeAndReopen(w) {
    await w.setProps({ modelValue: false });
    await settled();
    await w.setProps({ modelValue: true });
    await settled();
}

const pickerLabels = () => findAll("button").map((b) => b.textContent.trim());

beforeEach(() => {
    vi.stubGlobal(
        "ResizeObserver",
        class {
            observe() {}
            unobserve() {}
            disconnect() {}
        },
    );
    vi.useFakeTimers();
});
afterEach(() => {
    wrappers.forEach((w) => w.unmount());
    wrappers = [];
    document.body.innerHTML = "";
    vi.useRealTimers();
    vi.unstubAllGlobals();
});

const dateFilter = {
    table: "orders",
    column: "orders.created_at",
    operator: "inDateRange",
    values: ["2026-01-01", "2026-02-01"],
    name: "my filter",
    apply_to: ["chart-a"],
};

async function openEditing(filter) {
    const w = dialog();

    await settled();
    await w.setProps({ editFilter: filter });
    await settled();

    return w;
}

const dateInputs = () => findAll('input[readonly][type="text"]').map((i) => i.value);

describe("CrossFilterDialog resets", () => {
    it("clears what was typed once it closes", async () => {
        const w = await openEditing(dateFilter);

        expect(find("#name").value).toBe("my filter");

        await closeAndReopen(w);

        expect(find("#name").value).toBe("");
    });

    it("clears the dates", async () => {
        const w = await openEditing(dateFilter);

        expect(dateInputs()).toEqual(["2026-01-01", "2026-02-01"]);

        await closeAndReopen(w);

        // Re-open the same date operator so the fields render again: an empty
        // list here would only prove the control is gone, not that it cleared.
        await choose("Select data model", "orders");
        await choose("Select column", "orders.created_at");
        await choose("Select operator", "date range");

        expect(dateInputs()).toEqual(["", ""]);
    });

    it("resets the panel scope to all panels", async () => {
        const w = await openEditing(dateFilter);

        expect(find("#specific").checked).toBe(true);

        await closeAndReopen(w);

        expect(find("#all").checked).toBe(true);
        expect(find("#specific").checked).toBe(false);
    });

    it("clears the chosen model, column and operator", async () => {
        const w = await openEditing(dateFilter);

        expect(pickerLabels()).not.toContain("Select data model");

        await closeAndReopen(w);

        expect(pickerLabels()).toEqual(
            expect.arrayContaining(["Select data model", "Select column", "Select operator"]),
        );
    });

    it("does not reset while it is still open", async () => {
        dialog();
        await settled();

        await fill("#name", "my filter");
        await settled();

        expect(find("#name").value).toBe("my filter");
    });
});
