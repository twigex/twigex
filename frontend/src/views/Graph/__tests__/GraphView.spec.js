// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The graph page sets up and mounts without an error, and starts from the
// entity in its address. Sigma needs WebGL, which the test browser lacks, so
// it is replaced by a stand-in.

import { it, expect, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import GraphSectionView from "@/views/Graph/GraphSectionView.vue";
import GraphView from "@/views/Graph/GraphView.vue";
import OntologyView from "@/views/Graph/OntologyView.vue";
import OntologyTypeView from "@/views/Graph/OntologyTypeView.vue";

vi.mock("sigma", () => ({
    default: class {
        on() {}
        refresh() {}
        resize() {}
        kill() {}
        getCamera() {
            return {};
        }
        graphToViewport() {
            return { x: 0, y: 0 };
        }
    },
}));

vi.mock("@sigma/edge-curve", () => ({ EdgeCurvedArrowProgram: class {} }));

vi.mock("@/services/graphService", () => ({
    default: {
        getTypes: async () => ({
            data: {
                entity_types: [
                    {
                        id: "t-company",
                        key: "company",
                        name: "Company",
                        plural_name: "Companies",
                        icon: "building-office",
                        color: "cyan",
                        status: "active",
                        description: "Customers and suppliers from the ERP.",
                        source: "dataset",
                        source_name: "companies",
                        key_property_name: "companies.key",
                        title_property_name: "companies.name",
                    },
                    { id: "t-order", key: "order", name: "Order", source: "dataset" },
                    {
                        id: "t-chart",
                        key: "chart",
                        name: "Chart",
                        source: "builtin",
                        built_in: true,
                    },
                ],
                relation_types: [
                    {
                        key: "company_places_order",
                        name: "places",
                        reverse_name: "placed by",
                        from_type_id: "t-company",
                        to_type_id: "t-order",
                        kind: "property",
                        cardinality: "one_to_many",
                        kind_config: { property_name: "orders.company_id" },
                    },
                ],
            },
        }),
        searchEntities: async () => ({
            data: [{ type: "company", id: "c-05", title: "Tallinn Tech" }],
        }),
        getRelations: async () => ({
            data: { entity: { type: "company", id: "c-05", title: "Tallinn Tech" }, groups: [] },
        }),
    },
}));

globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
};

function graphRouter() {
    return createRouter({
        history: createMemoryHistory(),
        routes: [
            {
                path: "/graph",
                name: "graph",
                component: GraphSectionView,
                redirect: (to) => ({ name: "graph-explore", query: to.query }),
                children: [
                    { path: "explore", name: "graph-explore", component: GraphView },
                    { path: "ontology", name: "graph-ontology", component: OntologyView },
                    {
                        path: "ontology/types/:key",
                        name: "graph-ontology-type",
                        component: OntologyTypeView,
                    },
                ],
            },
        ],
    });
}

async function mountAt(router, location) {
    await router.push(location);

    const wrapper = mount(GraphSectionView, { global: { plugins: [router] } });

    await flushPromises();
    await new Promise((resolve) => setTimeout(resolve, 450));
    await flushPromises();

    return wrapper;
}

it("opens a /graph address on the entity it names", async () => {
    setActivePinia(createPinia());

    const router = graphRouter();
    const wrapper = await mountAt(router, {
        path: "/graph",
        query: { type: "company", id: "c-05" },
    });

    expect(router.currentRoute.value.name).toBe("graph-explore");
    expect(wrapper.text()).toContain("Tallinn Tech");
    expect(wrapper.text()).toContain("Types: 3 · Relation types: 1");
});

it("shows the entity types as cards", async () => {
    setActivePinia(createPinia());

    const wrapper = await mountAt(graphRouter(), { name: "graph-ontology" });

    expect(wrapper.text()).toContain("Companies");
    expect(wrapper.text()).toContain("Customers and suppliers from the ERP.");
    expect(wrapper.text()).toContain("Dataset companies");
    expect(wrapper.text()).toContain("Built-in");
});

it("lists the relation types on their own tab", async () => {
    setActivePinia(createPinia());

    const wrapper = await mountAt(graphRouter(), {
        name: "graph-ontology",
        query: { tab: "relations" },
    });

    const text = wrapper.text().replace(/\s+/g, " ");

    expect(text).toContain("Company places Order");
    expect(text).toContain("One to many");
    expect(text).toContain("orders.company_id");
    expect(text).toContain("From data");
});

it("reads a relation from the other side on that type's page", async () => {
    setActivePinia(createPinia());

    const wrapper = await mountAt(graphRouter(), {
        name: "graph-ontology-type",
        params: { key: "order" },
        query: { tab: "relations" },
    });

    const text = wrapper.text().replace(/\s+/g, " ");

    expect(text).toContain("Order placed by Company");
    expect(text).toContain("Many to one");
});
