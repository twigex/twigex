// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import { useTaskCompletion } from "@/composables/projects/useTaskCompletion";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";

vi.mock("@/services/workspaceService", () => ({
    default: {
        getTableStatusTypes: vi.fn(),
        getTaskCompletion: vi.fn(),
        completeSubtasks: vi.fn(),
        updateTask: vi.fn(),
    },
}));

let table = 0;
let completions;

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    table += 1;
    completions = {};
    workspaceService.getTableStatusTypes.mockResolvedValue({
        data: {
            done_ids: ["completed", "cancelled"],
            options: [
                { id: "todo", status_type: "Not started" },
                { id: "completed", status_type: "Done" },
                { id: "cancelled", status_type: "Closed" },
            ],
        },
    });
    workspaceService.getTaskCompletion.mockImplementation(({ task_id }) =>
        Promise.resolve({ data: completions[task_id] || { open_subtasks: 0, parent: null } }),
    );
    workspaceService.completeSubtasks.mockResolvedValue({ data: { completed: 2 } });
    workspaceService.updateTask.mockResolvedValue({ data: {} });
});

function change(statusId, taskId = "task") {
    return useTaskCompletion().beforeStatusChange({
        workspaceId: "ws",
        tableId: `t${table}`,
        taskId,
        statusId,
    });
}

async function answerWith(choice) {
    await flushPromises();
    const { question, answer } = useTaskCompletion();

    expect(question.value).not.toBeNull();
    const asked = { ...question.value };

    answer(choice);

    return asked;
}

describe("beforeStatusChange", () => {
    it("asks nothing for a status that does not close the task", async () => {
        completions.task = { open_subtasks: 3, parent: null };

        const completion = await change("todo");

        await completion.afterSave();

        expect(workspaceService.getTaskCompletion).not.toHaveBeenCalled();
        expect(useTaskCompletion().question.value).toBeNull();
    });

    it("asks nothing when the task has no open subtasks and no parent to close", async () => {
        const completion = await change("completed");

        await completion.afterSave();

        expect(useTaskCompletion().question.value).toBeNull();
        expect(workspaceService.completeSubtasks).not.toHaveBeenCalled();
    });

    it("cancels the change when asked about open subtasks", async () => {
        completions.task = { open_subtasks: 2, parent: null };

        const pending = change("completed");
        const asked = await answerWith("cancel");

        expect(asked).toMatchObject({ kind: "subtasks", count: 2 });
        expect(await pending).toBeNull();
    });

    it("gives the open subtasks the same status after the task is saved", async () => {
        completions.task = { open_subtasks: 2, parent: null };
        const reloads = useWorkspaceStore().getGridReloadToken;

        const pending = change("cancelled");

        await answerWith("all");
        const completion = await pending;

        expect(workspaceService.completeSubtasks).not.toHaveBeenCalled();

        await completion.afterSave();
        expect(workspaceService.completeSubtasks).toHaveBeenCalledWith({
            workspace_id: "ws",
            table_id: `t${table}`,
            task_id: "task",
            status: "cancelled",
        });
        expect(useWorkspaceStore().getGridReloadToken).not.toBe(reloads);
    });

    it("leaves the subtasks open when only the task is closed", async () => {
        completions.task = { open_subtasks: 2, parent: null };

        const pending = change("completed");

        await answerWith("only");
        await (await pending).afterSave();

        expect(workspaceService.completeSubtasks).not.toHaveBeenCalled();
    });
});

describe("closing the last open subtask", () => {
    it("offers to complete the parent, and then the parent's parent", async () => {
        completions.task = {
            open_subtasks: 0,
            parent: { id: "parent", name: "Launch", open: true, open_subtasks: 0 },
        };
        completions.parent = {
            open_subtasks: 0,
            parent: { id: "top", name: "Roadmap", open: true, open_subtasks: 0 },
        };

        const completion = await change("cancelled");
        const done = completion.afterSave();

        expect(await answerWith("complete")).toMatchObject({ kind: "parent", name: "Launch" });
        expect(await answerWith("not_now")).toMatchObject({ kind: "parent", name: "Roadmap" });
        await done;

        expect(workspaceService.updateTask).toHaveBeenCalledTimes(1);
        expect(workspaceService.updateTask).toHaveBeenCalledWith(
            expect.objectContaining({ task_id: "parent", field: "status", value: "completed" }),
        );
    });

    it("does not offer a parent that is already closed or still has open subtasks", async () => {
        completions.task = {
            open_subtasks: 0,
            parent: { id: "parent", name: "Launch", open: true, open_subtasks: 1 },
        };
        await (await change("completed")).afterSave();

        completions.task = {
            open_subtasks: 0,
            parent: { id: "parent", name: "Launch", open: false, open_subtasks: 0 },
        };
        await (await change("completed")).afterSave();

        expect(useTaskCompletion().question.value).toBeNull();
        expect(workspaceService.updateTask).not.toHaveBeenCalled();
    });
});
