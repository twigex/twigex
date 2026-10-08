// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The mention list is driven by keys in the message box, which only a
// mounted composer shows working end to end.

import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createRouter, createMemoryHistory } from "vue-router";
import { useUserStore } from "@/store/user";
import { usePermissionsStore } from "@/store/permissions";

const members = [
    { id: "u2", username: "anna", name: "Anna", lastname: "Ozola" },
    { id: "u3", username: "andris", name: "Andris", lastname: "Kalns" },
];

vi.mock("@/services/chatService", () => ({
    default: new Proxy(
        {
            searchChannelMembers: async (_channel, q) => ({
                data: members.filter((m) => m.username.startsWith(q)),
            }),
        },
        { get: (target, key) => target[key] || (() => new Promise(() => {})) },
    ),
}));
vi.mock("@/js/websocket", () => ({ connection: { readyState: 0, send() {} } }));

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

async function mountInput() {
    const pinia = createPinia();

    setActivePinia(pinia);
    useUserStore().setUser({ id: "u1", name: "Test", lastname: "User", username: "test" });
    usePermissionsStore().setPermissions(["send_message"]);
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [{ path: "/chat/:chatId", component: { template: "<div/>" } }],
    });

    await router.push("/chat/c1");
    await router.isReady();
    const { default: ChatInput } = await import("@/components/Chat/Messages/ChatInput.vue");

    return mount(ChatInput, { global: { plugins: [pinia, router] }, attachTo: document.body });
}

async function type(box, value) {
    box.element.value = value;
    box.element.selectionStart = box.element.selectionEnd = value.length;
    await box.trigger("input");
    await vi.runAllTimersAsync();
    await flushPromises();
}

it("takes the highlighted member on Enter instead of sending a half-typed mention", async () => {
    const w = await mountInput();
    const box = w.find("textarea");

    await type(box, "hi @an");
    expect(w.text()).toContain("@anna");
    expect(w.text()).toContain("@andris");

    await box.trigger("keydown", { key: "ArrowDown" });
    await box.trigger("keydown", { key: "Enter" });
    await flushPromises();

    expect(box.element.value).toBe("hi @andris ");
    expect(w.emitted("send-message")).toBeUndefined();
    w.unmount();
});

it("offers only the special mentions that match what was typed", async () => {
    const w = await mountInput();
    const box = w.find("textarea");

    await type(box, "@he");
    expect(w.text()).toContain("@here");
    expect(w.text()).not.toContain("@all");
    w.unmount();
});

it("still sends on Enter when no mention is being typed", async () => {
    const w = await mountInput();
    const box = w.find("textarea");

    await type(box, "hello");
    await box.trigger("keydown", { key: "Enter" });

    expect(w.emitted("send-message")).toHaveLength(1);
    w.unmount();
});
