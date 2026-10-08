// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The mention list is driven by keys in the comment box, which only a mounted
// component shows working end to end.

import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import { useUserStore } from "@/store/user";

const members = [
    { id: "u2", username: "anna", name: "Anna", lastname: "Ozola" },
    { id: "u3", username: "andris", name: "Andris", lastname: "Kalns" },
];

vi.mock("@/services/workspaceService", () => ({
    default: new Proxy(
        {
            searchWorkspaceMembers: async (_ws, q) => ({
                data: members.filter((m) => m.username.startsWith(q)),
            }),
        },
        { get: (target, key) => target[key] || (() => new Promise(() => {})) },
    ),
}));
globalThis.IntersectionObserver = class {
    observe() {}
    disconnect() {}
};

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

async function mountComments() {
    const pinia = createPinia();

    setActivePinia(pinia);
    useUserStore().setUser({ id: "u1", name: "Test", lastname: "User", username: "test" });
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [{ path: "/p/:id/grid/:tid", component: { template: "<div/>" } }],
    });

    await router.push("/p/w1/grid/t1");
    await router.isReady();
    const { default: TaskComments } = await import("../TaskComments.vue");

    return mount(TaskComments, { global: { plugins: [pinia, router] }, attachTo: document.body });
}

async function type(box, value) {
    box.element.value = value;
    box.element.selectionStart = box.element.selectionEnd = value.length;
    await box.trigger("input");
    await vi.runAllTimersAsync();
    await flushPromises();
}

it("lists matching members after an @ and puts the one picked by keys in place", async () => {
    const w = await mountComments();
    const box = w.find("textarea#comment");

    await type(box, "thanks @an");
    expect(w.text()).toContain("Anna Ozola@anna");
    expect(w.text()).toContain("Andris Kalns@andris");
    expect(w.text()).not.toContain("@all");

    await box.trigger("keydown", { key: "ArrowDown" });
    await box.trigger("keydown", { key: "Enter" });
    await flushPromises();

    expect(box.element.value).toBe("thanks @andris ");
    expect(w.text()).not.toContain("Andris Kalns@andris");
    w.unmount();
});

it("offers @all for a bare @ and leaves the arrow keys alone when no list is open", async () => {
    const w = await mountComments();
    const box = w.find("textarea#comment");

    await type(box, "@");
    expect(w.text()).toContain("@all");

    await type(box, "no mention here");
    const down = new KeyboardEvent("keydown", { key: "ArrowDown", cancelable: true });

    box.element.dispatchEvent(down);
    expect(down.defaultPrevented).toBe(false);
    w.unmount();
});
