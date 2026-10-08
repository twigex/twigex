// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import { useUserStore } from "@/store/user";
import MentionTag from "@/components/Mentions/MentionTag.vue";

// The store's lazy loaders hit userService; stub it so an unknown handle does
// not make a real request during the test.
vi.mock("@/services/userService", () => ({
    default: {
        byUsernames: () => Promise.resolve({ data: [] }),
        byIds: () => Promise.resolve({ data: [] }),
    },
}));

let pinia;

beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
});

function mountMention(username) {
    return mount(MentionTag, {
        props: { username },
        global: { plugins: [pinia] },
    });
}

function mountIdMention(userId) {
    return mount(MentionTag, {
        props: { userId },
        global: { plugins: [pinia] },
    });
}

describe("Mention", () => {
    it("styles a known other member gray", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };
        store.addUsers([{ id: "u2", username: "bob", name: "Bob", lastname: "B" }]);

        const span = mountMention("bob").find("span.mention");

        expect(span.exists()).toBe(true);
        expect(span.classes()).toContain("bg-gray-200");
    });

    it("styles the current user yellow", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };

        expect(mountMention("me").find("span.mention").classes()).toContain("bg-yellow-300");
    });

    it("styles @all/@here indigo without a lookup", () => {
        expect(mountMention("all").find("span.mention").classes()).toContain("bg-indigo-200");
        expect(mountMention("here").find("span.mention").classes()).toContain("bg-indigo-200");
    });

    it("renders an unknown handle as plain text", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };

        const w = mountMention("ghost");

        expect(w.find("span.mention").exists()).toBe(false);
        expect(w.text()).toContain("@ghost");
    });
});

describe("Mention by id", () => {
    const ID = "0f8fad5bd9cb469fa16570867728950e";

    it("renders the name the id currently resolves to", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };
        store.addUsers([{ id: ID, username: "jane", name: "Jane" }]);

        expect(mountIdMention(ID).text()).toContain("@jane");
    });

    it("follows a rename without the message changing", async () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };
        store.addUsers([{ id: ID, username: "jane" }]);

        const w = mountIdMention(ID);

        expect(w.text()).toContain("@jane");

        store.addUsers([{ id: ID, username: "jsmith" }]);
        await w.vm.$nextTick();

        expect(w.text()).toContain("@jsmith");
        expect(w.text()).not.toContain("@jane");
    });

    it("renders a tombstone for an id that resolves to nobody", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };

        const w = mountIdMention(ID);

        expect(w.text()).not.toContain(ID);
        expect(w.find("span.mention").exists()).toBe(true);
    });

    it("leaves a tombstone unclickable", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "me" };

        const span = mountIdMention(ID).find("span.mention");

        expect(span.attributes("role")).toBeUndefined();
        expect(span.attributes("tabindex")).toBeUndefined();
    });

    it("styles the current user yellow when mentioned by id", () => {
        const store = useUserStore();

        store.user = { id: ID, username: "me" };
        store.addUsers([{ id: ID, username: "me" }]);

        expect(mountIdMention(ID).find("span.mention").classes()).toContain("bg-yellow-300");
    });
});
