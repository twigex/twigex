// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import GraphSearch from "@/components/Graph/GraphSearch.vue";

// Each search waits for its own answer, so a test decides when it arrives.
const searches = [];

vi.mock("@/services/graphService", () => ({
    default: {
        searchEntities: (type, text, signal) => {
            return new Promise((resolve, reject) => {
                searches.push({ text: text, signal: signal, answer: resolve, fail: reject });
            });
        },
    },
}));

const types = [{ id: "t-company", key: "company", name: "Company", source: "dataset" }];

function mountSearch() {
    return mount(GraphSearch, {
        props: { entityTypes: types, styles: { company: { color: "#0891b2" } } },
        attachTo: document.body,
    });
}

beforeEach(() => {
    searches.length = 0;
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

it("shows a spinner while searching, then the results, and keeps the typed text", async () => {
    const wrapper = mountSearch();
    const input = wrapper.get("input");

    await input.setValue("Tall");

    expect(input.element.value).toBe("Tall");
    expect(wrapper.text()).toContain("Searching…");

    await vi.advanceTimersByTimeAsync(400);

    searches.at(-1).answer({ data: [{ type: "company", id: "c-05", title: "Tallinn Tech" }] });

    await flushPromises();

    expect(wrapper.text()).toContain("Tallinn Tech");
    expect(wrapper.text()).not.toContain("Searching…");
    expect(input.element.value).toBe("Tall");

    const option = wrapper.findAll("li").find((item) => item.text() === "Tallinn Tech");

    // Headless UI picks an option on mousedown, before the input loses focus.
    await option.trigger("mousedown", { button: 0 });
    await flushPromises();

    expect(wrapper.emitted("pick")[0]).toEqual(["company", "c-05"]);

    wrapper.unmount();
});

it("sends one search for a word typed quickly", async () => {
    const wrapper = mountSearch();
    const input = wrapper.get("input");

    for (const text of ["T", "Ta", "Tal", "Tall"]) {
        await input.setValue(text);
        await vi.advanceTimersByTimeAsync(100);
    }

    await vi.advanceTimersByTimeAsync(400);

    expect(searches.map((search) => search.text)).toEqual(["Tall"]);

    wrapper.unmount();
});

it("cancels a search still running when typing goes on", async () => {
    const wrapper = mountSearch();
    const input = wrapper.get("input");

    await input.setValue("Tal");
    await vi.advanceTimersByTimeAsync(400);

    const first = searches.at(-1);

    await input.setValue("Tall");

    expect(first.signal.aborted).toBe(true);

    wrapper.unmount();
});

it("says the search failed instead of nothing found", async () => {
    const wrapper = mountSearch();
    const input = wrapper.get("input");

    await input.setValue("Tall");
    await vi.advanceTimersByTimeAsync(400);

    searches.at(-1).fail(new Error("server error"));

    await flushPromises();

    expect(wrapper.text()).toContain("The search failed");
    expect(wrapper.text()).not.toContain("Nothing found");

    wrapper.unmount();
});
