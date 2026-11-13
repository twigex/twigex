// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import axiosInstance, { setupInterceptors } from "@/services/api";

function fakeRouter() {
    return {
        currentRoute: { value: { name: "public-share", fullPath: "/share/tok" } },
        replace: vi.fn(),
    };
}

function unauthorized(url) {
    return { response: { status: 401 }, config: { url } };
}

describe("a 401 answer", () => {
    let router;
    let rejected;

    beforeEach(() => {
        setActivePinia(createPinia());
        axiosInstance.interceptors.response.clear();
        router = fakeRouter();
        setupInterceptors(router);
        rejected = axiosInstance.interceptors.response.handlers.at(-1).rejected;
    });

    it("from a share link asking for its password does not sign anyone out", async () => {
        await expect(rejected(unauthorized("/public/tok"))).rejects.toBeTruthy();

        expect(router.replace).not.toHaveBeenCalled();
    });

    it("from a signed-in page still sends the user to sign in", async () => {
        await expect(rejected(unauthorized("/files/list"))).rejects.toBeTruthy();

        expect(router.replace).toHaveBeenCalledWith(expect.objectContaining({ name: "login" }));
    });
});
