// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The graph page sets up and mounts without an error, and starts from the
// entity in its address. Sigma needs WebGL, which the test browser lacks, so
// it is replaced by a stand-in.

import { it, expect, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import GraphView from "@/views/Graph/GraphView.vue";

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
                    { id: "t-company", key: "company", name: "Company", source: "dataset" },
                ],
                relation_types: [],
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

it("mounts and starts from the entity in the address", async () => {
    setActivePinia(createPinia());

    const router = createRouter({
        history: createMemoryHistory(),
        routes: [{ path: "/graph", name: "graph", component: GraphView }],
    });

    await router.push({ name: "graph", query: { type: "company", id: "c-05" } });

    const wrapper = mount(GraphView, { global: { plugins: [router] } });

    await flushPromises();
    await new Promise((resolve) => setTimeout(resolve, 350));
    await flushPromises();

    expect(wrapper.text()).toContain("Tallinn Tech");
    expect(wrapper.text()).toContain("Company");
});
