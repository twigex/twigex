// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { beforeEach, describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";

let pinia;

beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
});

function mountAvatar(user) {
    return mount(UserAvatar, {
        props: { user },
        global: { plugins: [pinia] },
    });
}

describe("UserAvatar", () => {
    it("shows offline status before statuses load", () => {
        const avatar = mount(UserAvatar, {
            props: { user: { id: "u1" }, status: true },
            global: { plugins: [pinia] },
        });

        expect(avatar.find(".bg-gray-300").exists()).toBe(true);
    });

    it("uses a supplied photo when the cached record has none", () => {
        useUserStore().addUsers([{ id: "u1", name: "Old", photo: "" }]);

        const avatar = mountAvatar({ id: "u1", name: "Ada", photo: "s1" });

        expect(avatar.find("img").exists()).toBe(true);
        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s1");
    });

    it("uses supplied initials when the cached record has an old photo", () => {
        useUserStore().addUsers([{ id: "u1", name: "Old", photo: "s1" }]);

        const avatar = mountAvatar({ id: "u1", name: "Ada", photo: "" });

        expect(avatar.find("img").exists()).toBe(false);
        expect(avatar.find("svg text").text()).toBe("A");
    });

    it("uses the stored record for an id-only avatar", () => {
        useUserStore().addUsers([{ id: "u1", name: "Ada", photo: "s1" }]);

        const avatar = mountAvatar({ id: "u1" });

        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s1");
    });

    it("uses the current user's photo when a supplied record is stale", () => {
        useUserStore().setUser({ id: "u1", name: "Ada", photo: "s2" });

        const avatar = mountAvatar({ id: "u1", name: "Ada", photo: "s1" });

        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s2");
    });

    it("shows initials after a load error and retries when the photo changes", async () => {
        const avatar = mountAvatar({ id: "u1", name: "Ada", photo: "s1" });

        await avatar.find("img").trigger("error");

        expect(avatar.find("img").exists()).toBe(false);
        expect(avatar.find("svg text").text()).toBe("A");

        await avatar.setProps({ user: { id: "u1", name: "Ada", photo: "s2" } });

        expect(avatar.find("img").exists()).toBe(true);
        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s2");
    });

    it("retries after a failed photo is replaced in the store", async () => {
        const store = useUserStore();

        store.addUsers([{ id: "u1", name: "Ada", photo: "s1" }]);
        const avatar = mountAvatar({ id: "u1" });

        await avatar.find("img").trigger("error");
        store.addUsers([{ id: "u1", name: "Ada", photo: "s2" }]);
        await avatar.vm.$nextTick();

        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s2");
    });

    it("keeps the new photo URL after a store reset", async () => {
        const store = useUserStore();

        store.setUser({ id: "u1", name: "Ada", photo: "s1" });
        const avatar = mountAvatar({ id: "u1" });

        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s1");

        store.setPhoto("u1", "s2");
        await avatar.vm.$nextTick();

        expect(avatar.find("img").attributes("src")).toBe("/api/users/photo/u1?v=s2");

        pinia = createPinia();
        setActivePinia(pinia);
        useUserStore().setUser({ id: "u1", name: "Ada", photo: "s2" });

        expect(mountAvatar({ id: "u1" }).find("img").attributes("src")).toBe(
            "/api/users/photo/u1?v=s2",
        );
    });
});
