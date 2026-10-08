// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useUserStore } from "@/store/user";
import userService from "@/services/userService";

vi.mock("@/services/userService", () => ({
    default: { byIds: vi.fn(), byUsernames: vi.fn() },
}));

const ada = { id: "u1", username: "ada", name: "Ada", lastname: "Abele" };

function deferred() {
    let resolve;
    const promise = new Promise((r) => (resolve = r));

    return { promise, resolve };
}

const flush = () => new Promise((r) => setTimeout(r, 0));

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
});

describe("ensureUsers", () => {
    it("waits for a request another caller already started", async () => {
        const first = deferred();

        userService.byIds.mockReturnValueOnce(first.promise);

        const store = useUserStore();

        store.ensureUsers(["u1"]);

        let settled = false;
        const second = store.ensureUsers(["u1"]).then(() => (settled = true));

        await flush();
        expect(settled).toBe(false);
        expect(store.getUserById("u1")).toBeUndefined();

        first.resolve({ data: [ada] });
        await second;

        expect(store.getUserById("u1")).toEqual(ada);
        expect(userService.byIds).toHaveBeenCalledTimes(1);
    });

    it("resolves immediately for an already cached id", async () => {
        const store = useUserStore();

        store.addUsers([ada]);

        await store.ensureUsers(["u1"]);

        expect(userService.byIds).not.toHaveBeenCalled();
    });

    it("frees the id again once the request settles", async () => {
        userService.byIds.mockResolvedValue({ data: [] });

        const store = useUserStore();

        await store.ensureUsers(["ghost"]);
        await store.ensureUsers(["ghost"]);

        expect(userService.byIds).toHaveBeenCalledTimes(1);
    });

    it("marks an id the server did not return as failed", async () => {
        userService.byIds.mockResolvedValue({ data: [ada] });

        const store = useUserStore();

        await store.ensureUsers(["u1", "missing-1"]);

        expect(store.isUserFailed("u1")).toBe(false);
        expect(store.isUserFailed("missing-1")).toBe(true);
    });

    it("marks ids failed when the request errors, and clears them once a retry loads them", async () => {
        const bob = { id: "flaky-1", name: "Bob", lastname: "Berg" };

        userService.byIds
            .mockRejectedValueOnce(new Error("offline"))
            .mockResolvedValueOnce({ data: [bob] });

        const store = useUserStore();

        await store.ensureUsers(["flaky-1"]);
        expect(store.isUserFailed("flaky-1")).toBe(true);

        await store.ensureUsers(["flaky-1"]);
        expect(store.isUserFailed("flaky-1")).toBe(false);
        expect(store.getUserById("flaky-1")).toEqual(bob);
    });
});

describe("ensureUsersByUsernames", () => {
    it("waits for a request another caller already started", async () => {
        const first = deferred();

        userService.byUsernames.mockReturnValueOnce(first.promise);

        const store = useUserStore();

        store.ensureUsersByUsernames(["ada"]);

        let settled = false;
        const second = store.ensureUsersByUsernames(["ada"]).then(() => (settled = true));

        await flush();
        expect(settled).toBe(false);

        first.resolve({ data: [ada] });
        await second;

        expect(store.getUserByUsername("ada")).toEqual(ada);
        expect(userService.byUsernames).toHaveBeenCalledTimes(1);
    });
});

describe("getUserIdByUsername", () => {
    it("resolves a cached user", () => {
        const store = useUserStore();

        store.addUsers([ada]);

        expect(store.getUserIdByUsername("ada")).toBe("u1");
    });

    it("ignores case", () => {
        const store = useUserStore();

        store.addUsers([ada]);

        expect(store.getUserIdByUsername("ADA")).toBe("u1");
    });

    it("resolves the logged-in user, who is not in the list", () => {
        const store = useUserStore();

        store.user = { id: "me", username: "grace" };

        expect(store.getUserIdByUsername("grace")).toBe("me");
    });

    // Reading only `users` is what broke this the first time.
    it("resolves a user held only in usersMap", () => {
        const store = useUserStore();

        store.usersMap[ada.id] = ada;

        expect(store.getUserIdByUsername("ada")).toBe("u1");
    });

    it("returns null for an unknown handle", () => {
        const store = useUserStore();

        store.addUsers([ada]);

        expect(store.getUserIdByUsername("nobody")).toBeNull();
    });

    it("returns null for an empty handle", () => {
        expect(useUserStore().getUserIdByUsername("")).toBeNull();
    });
});
