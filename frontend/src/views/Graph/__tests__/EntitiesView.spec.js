// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import GraphSectionView from "@/views/Graph/GraphSectionView.vue";
import EntitiesView from "@/views/Graph/EntitiesView.vue";

const searches = [];

// 60 companies, read 50 at a time in key order like the server does.
const companies = Array.from({ length: 60 }, (_, index) => {
    const id = `c-${String(index).padStart(3, "0")}`;

    return { type: "company", id: id, title: `Company ${index}` };
});

vi.mock("@/services/graphService", () => ({
    default: {
        getTypes: async () => ({
            data: {
                entity_types: [
                    {
                        id: "t-chart",
                        key: "chart",
                        name: "Chart",
                        source: "built_in",
                        built_in: true,
                    },
                    {
                        id: "t-company",
                        key: "company",
                        name: "Company",
                        plural_name: "Companies",
                        source: "dataset",
                    },
                ],
                relation_types: [],
            },
        }),
        searchEntities: async (type, search, signal, after) => {
            searches.push({ type: type, search: search, after: after });

            const matching = companies.filter((company) => {
                return company.title.toLowerCase().includes(search.toLowerCase());
            });
            const start = after ? matching.findIndex((company) => company.id === after) + 1 : 0;
            const items = matching.slice(start, start + 50);
            const next = matching.length > start + 50 ? items.at(-1).id : "";

            return { data: { items: items, next: next } };
        },
    },
}));

globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
};

function entitiesRouter() {
    return createRouter({
        history: createMemoryHistory(),
        routes: [
            {
                path: "/graph",
                component: GraphSectionView,
                children: [
                    { path: "entities/:key?", name: "graph-entities", component: EntitiesView },
                    { path: "explore", name: "graph-explore", component: EntitiesView },
                    { path: "ontology", name: "graph-ontology", component: EntitiesView },
                    {
                        path: "ontology/types/:key",
                        name: "graph-ontology-type",
                        component: EntitiesView,
                    },
                ],
            },
        ],
    });
}

async function mountEntities(router) {
    await router.push({ name: "graph-entities" });

    const wrapper = mount(GraphSectionView, { global: { plugins: [router] } });

    await flushPromises();

    return wrapper;
}

beforeEach(() => {
    setActivePinia(createPinia());
    searches.length = 0;
});

it("opens the first type read from a dataset and lists its entities", async () => {
    const router = entitiesRouter();
    const wrapper = await mountEntities(router);

    expect(router.currentRoute.value.params.key).toBe("company");
    expect(wrapper.text()).toContain("Companies");
    expect(wrapper.text()).not.toContain("Chart");
    expect(wrapper.findAll("tbody tr")).toHaveLength(50);
    expect(wrapper.text()).toContain("Shown: 50");

    const relation = wrapper.find("tbody tr a");

    expect(relation.attributes("href")).toBe("/graph/explore?type=company&id=c-000");
});

it("loads the next page after the last key shown", async () => {
    const wrapper = await mountEntities(entitiesRouter());

    const loadMore = wrapper.findAll("button").find((button) => button.text() === "Load more");

    await loadMore.trigger("click");
    await flushPromises();

    expect(searches.at(-1).after).toBe("c-049");
    expect(wrapper.findAll("tbody tr")).toHaveLength(60);
    expect(wrapper.text()).toContain("Shown: 60");
    expect(wrapper.text()).not.toContain("Load more");
});

it("searches by name after a pause in typing", async () => {
    vi.useFakeTimers();

    const wrapper = await mountEntities(entitiesRouter());

    await wrapper.find('input[type="search"]').setValue("company 5");
    await vi.advanceTimersByTimeAsync(450);
    await flushPromises();

    vi.useRealTimers();

    expect(searches.at(-1)).toEqual({ type: "company", search: "company 5", after: "" });
    expect(wrapper.findAll("tbody tr")).toHaveLength(11);
});
