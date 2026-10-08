// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { beforeEach, describe, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useProjectRealtime } from "../useProjectRealtime";

let deliver;

vi.mock("@/js/websocket", () => ({
    onProjectChange: (listener) => {
        deliver = listener;

        return () => {};
    },
    onReconnect: () => () => {},
}));

// A change as the server sends it for the open workspace, from another tab.
const change = (fields) => ({ workspace_id: "w1", origin_client_id: "other-tab", ...fields });

async function setup() {
    const pinia = createPinia();

    setActivePinia(pinia);
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [
            {
                path: "/p/:id/grid/:tid/view/:fid?",
                name: "grid-view",
                component: { template: "<div/>" },
            },
            { path: "/:pathMatch(.*)*", component: { template: "<div/>" } },
        ],
    });

    await router.push("/p/w1/grid/t1/view/v1");
    await router.isReady();

    const Host = defineComponent({ setup: () => useProjectRealtime(), template: "<div/>" });

    mount(Host, { global: { plugins: [pinia, router] } });

    return useWorkspaceStore();
}

// One change of each kind the composable hands to a different file, so a
// handler that is not wired up shows here rather than when that change
// arrives in the app.
describe("useProjectRealtime", () => {
    let store;

    beforeEach(async () => {
        store = await setup();
    });

    it("takes a deleted task off the grid page", () => {
        store.setTableData([{ id: "a" }, { id: "b" }]);
        deliver(change({ type: "task_deleted", table_id: "t1", task: { id: "a" } }));

        expect(store.getTableData.map((r) => r.id)).toEqual(["b"]);
    });

    it("puts a field's new formula on the table's headers", () => {
        store.setTableHeaders([{ name: "c1" }, { name: "c2" }]);
        deliver(
            change({
                type: "field_formula_updated",
                table_id: "t1",
                field_name: "c1",
                formula: { name: "SUM" },
            }),
        );

        expect(store.getTableHeaders[0].formula).toEqual({ name: "SUM" });
        expect(store.getTableHeaders[1].formula).toBeUndefined();
    });

    it("takes a deleted view out of the view menu", () => {
        store.setViewTypes([{ id: "v2" }, { id: "v3" }]);
        deliver(change({ type: "DELETE_VIEW", table_id: "t1", delete_id: "v2" }));

        expect(store.getViewTypes.map((v) => v.id)).toEqual(["v3"]);
    });

    it("moves a card another tab moved on the open board", () => {
        store.setKanbanTaskList([
            { id: "todo", tasks: [{ id: "a" }] },
            { id: "done", tasks: [] },
        ]);
        deliver(
            change({
                type: "KANBAN_CARD_MOVED",
                table_id: "t1",
                view_id: "v1",
                task_id: "a",
                column: "done",
                after_id: "",
            }),
        );

        expect(store.getKanbanTaskList.map((c) => c.tasks.map((t) => t.id))).toEqual([[], ["a"]]);
    });

    it("renames a folder in the navigation", async () => {
        store.setWorkspaceFolders([{ id: "f1", name: "Old" }]);
        deliver(change({ type: "UPDATE_FOLDER_NAME", folder_id: "f1", data: { name: "New" } }));
        await flushPromises();

        expect(store.getWorkspaceFolders[0].name).toBe("New");
    });

    it("removes a deleted folder but not a table with its id", () => {
        store.setWorkspaceFolders([
            { id: "f1", isFolder: true, children: [] },
            { id: "f2", isFolder: true, children: [] },
        ]);
        deliver(change({ type: "DELETE_FOLDER", folder_id: "f1" }));

        expect(store.getWorkspaceFolders.map((f) => f.id)).toEqual(["f2"]);
    });

    it("ignores changes to another workspace", () => {
        store.setTableData([{ id: "a" }]);
        deliver(
            change({ type: "task_deleted", workspace_id: "w2", table_id: "t1", task: { id: "a" } }),
        );

        expect(store.getTableData.map((r) => r.id)).toEqual(["a"]);
    });
});
