// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Each Projects view sets up and mounts without an error. Only running it
// shows a composable given a ref declared after it, or a name the template
// draws with that nothing defines.

import { it, expect, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import { ref } from "vue";
import { POPULAR_FORMULAS } from "@/constants/formulasCatalog.js";
import { useFormulaEditor } from "@/composables/projects/useFormulaEditor";
import { useUserStore } from "@/store/user";

vi.mock("@/services/workspaceService", () => ({
    default: new Proxy({}, { get: () => () => new Promise(() => {}) }),
}));
globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
};

const cases = [
    [
        "GridView",
        () => import("@/components/Projects/Grid/GridView.vue"),
        { isDialog: false, tableHeaders: [], tableData: [], columnWidths: [] },
    ],
    [
        "GridView",
        () => import("@/components/Projects/Grid/GridView.vue"),
        { isDialog: true, tableHeaders: [], tableData: [], columnWidths: [] },
    ],
    ["AssignedToMe", () => import("@/components/Projects/Reports/AssignedToMe.vue"), {}],
    [
        "DetailedTaskReport",
        () => import("@/components/Projects/Reports/DetailedTaskReport.vue"),
        {},
    ],
    ["FullTaskNavigation", () => import("@/components/Projects/Task/FullTaskNavigation.vue"), {}],
    ["KanbanView", () => import("@/components/Projects/Kanban/KanbanView.vue"), {}],
    ["GanttView", () => import("@/components/Projects/GanttView.vue"), {}],
    ["CalendarView", () => import("@/components/Projects/CalendarView.vue"), {}],
    [
        "GridViewDialog",
        () => import("@/components/Projects/Dialogs/GridViewDialog.vue"),
        { isOpen: true, workspaceId: "w1", tableId: "t1", field: "links" },
    ],
    ["CreateRole", () => import("@/components/Projects/Members/CreateRole.vue"), {}],
    ["TopNavigation", () => import("@/components/Projects/TopNavigation.vue"), {}],
    [
        "FieldEditorPopover",
        () => import("@/components/Projects/Grid/FieldEditorPopover.vue"),
        { tableHeaders: [{ name: "c1", header_usage: "number" }] },
    ],
    ["FilterBuilder", () => import("@/components/Projects/Filters/FilterBuilder.vue"), {}],
    ["TaskComments", () => import("@/components/Projects/Task/TaskComments.vue"), {}],
    [
        "TaskActivityList",
        () => import("@/components/Projects/Task/TaskActivityList.vue"),
        {
            activity: [
                {
                    id: 1,
                    type: "commented",
                    userId: "u1",
                    person: { name: "Test" },
                    comment: "hi <@u1> :smile:",
                    date: "now",
                    dateTime: "2026-10-02",
                },
                {
                    id: 2,
                    type: "updated",
                    person: { name: "Test" },
                    comment: "changed status",
                    date: "now",
                    dateTime: "2026-10-02",
                },
            ],
        },
    ],
    ["TaskCommentComposer", () => import("@/components/Projects/Task/TaskCommentComposer.vue"), {}],
    [
        "WorkspaceMembers",
        () => import("@/components/Projects/Members/WorkspaceMembers.vue"),
        { workspaceId: "w1" },
    ],
    [
        "MemberRolesDialog",
        () => import("@/components/Projects/Members/MemberRolesDialog.vue"),
        { open: true, roles: ["user"], options: ["user", "admin"], roleName: (r) => r },
    ],
    ...["SUM", "ROUND"].map((name) => [
        "FormulaFieldEditor",
        () => import("@/components/Projects/Grid/FormulaFieldEditor.vue"),
        { editor: formulaEditor(name), isDialog: name },
    ]),
];

function formulaEditor(name) {
    const editor = useFormulaEditor(ref([{ name: "c1", header_usage: "number" }]));

    editor.chosenFormula.value = POPULAR_FORMULAS.find((f) => f.name === name);
    editor.useMultipleFields.value = true;

    return editor;
}

// A view's first case also compiles it, which during a full parallel run can
// take longer than the default limit.
for (const [name, load, props] of cases) {
    it(`sets ${name} up without errors ${JSON.stringify(props.isDialog ?? "")}`, async () => {
        const pinia = createPinia();

        setActivePinia(pinia);
        useUserStore().setUser({ id: "u1", name: "Test", lastname: "User", username: "test" });
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
        const errors = [];
        const C = (await load()).default;
        const w = mount(C, {
            props,
            global: {
                plugins: [pinia, router],
                config: {
                    errorHandler: (e) => errors.push(e),
                    warnHandler: (m) => {
                        if (/Unhandled error/.test(m)) errors.push(new Error(m));
                    },
                },
            },
            attachTo: document.body,
        });

        await flushPromises();
        w.unmount();
        expect(errors.map((e) => e.message)).toEqual([]);
    }, 30000);
}
