// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";

vi.mock("@/utils/utils", async (importOriginal) => {
    const original = await importOriginal();

    return { ...original, hasSessionCookie: () => true };
});

import router from "@/router";
import { useUserStore } from "@/store/user";

describe("admin pages", () => {
    beforeEach(async () => {
        setActivePinia(createPinia());
        await router.push("/");
    });

    it("open on a fresh page load, before the user has loaded", async () => {
        await router.push({ name: "smtp" });

        expect(router.currentRoute.value.name).toBe("smtp");
    });

    it("still turn away a user known not to be an admin", async () => {
        const userStore = useUserStore();

        userStore.user = { id: "u1", role: "system_user" };
        await router.push({ name: "smtp" });

        expect(router.currentRoute.value.name).toBe("profile");
    });
});
