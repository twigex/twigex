// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { useLatestRequest } from "../useLatestRequest";

describe("useLatestRequest", () => {
    it("keeps the latest request current", () => {
        const { start } = useLatestRequest();
        const req = start();

        expect(req.isCurrent()).toBe(true);
        expect(req.signal.aborted).toBe(false);
    });

    it("cancels a request once a newer one starts", () => {
        const { start } = useLatestRequest();
        const first = start();
        const second = start();

        expect(first.isCurrent()).toBe(false);
        expect(first.signal.aborted).toBe(true);
        expect(second.isCurrent()).toBe(true);
    });

    it("cancels the request in flight", () => {
        const { start, cancel } = useLatestRequest();
        const req = start();

        cancel();
        expect(req.isCurrent()).toBe(false);
        expect(req.signal.aborted).toBe(true);
    });
});
