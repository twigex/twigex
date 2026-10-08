// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { pollQuery, QUERY_FAILED, QUERY_RETRY_LIMIT } from "@/utils/collimato/queryPolling.js";

const waiting = { data: { error: "Continue wait" } };
const rows = { data: { data: [{ a: 1 }] } };

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

async function settle(promise) {
    const outcome = promise.then(
        (value) => ({ value }),
        (error) => ({ error }),
    );

    await vi.runAllTimersAsync();
    const result = await outcome;

    if ("error" in result) {
        throw result.error;
    }

    return result.value;
}

describe("pollQuery", () => {
    it("returns the result when the engine answers at once", async () => {
        const loadData = vi.fn().mockResolvedValue(rows);

        await expect(pollQuery(loadData, {})).resolves.toBe(rows);
        expect(loadData).toHaveBeenCalledTimes(1);
    });

    it("waits and retries while the engine says continue wait", async () => {
        const loadData = vi
            .fn()
            .mockResolvedValueOnce(waiting)
            .mockResolvedValueOnce(waiting)
            .mockResolvedValue(rows);

        await expect(settle(pollQuery(loadData, {}))).resolves.toBe(rows);
        expect(loadData).toHaveBeenCalledTimes(3);
    });

    it("gives up after the retry limit", async () => {
        const loadData = vi.fn().mockResolvedValue(waiting);

        await expect(settle(pollQuery(loadData, {}))).rejects.toThrow(QUERY_FAILED);
        expect(loadData).toHaveBeenCalledTimes(QUERY_RETRY_LIMIT + 1);
    });

    it("throws the engine's own error", async () => {
        const loadData = vi.fn().mockResolvedValue({ data: { error: "bad column" } });

        await expect(pollQuery(loadData, {})).rejects.toThrow("bad column");
    });

    it("stops when the caller aborts mid flight", async () => {
        const controller = new AbortController();
        const loadData = vi.fn(() => new Promise(() => {}));

        const pending = pollQuery(loadData, {}, { signal: controller.signal });

        controller.abort();

        await expect(pending).rejects.toThrow("Aborted");
    });

    it("stops while waiting between retries", async () => {
        const controller = new AbortController();
        const loadData = vi.fn().mockResolvedValue(waiting);

        const pending = pollQuery(loadData, {}, { signal: controller.signal });

        await Promise.resolve();
        controller.abort();

        await expect(pending).rejects.toThrow("Aborted");
    });

    it("does not start when the signal is already aborted", async () => {
        const controller = new AbortController();

        controller.abort();
        const loadData = vi.fn(() => new Promise(() => {}));

        await expect(pollQuery(loadData, {}, { signal: controller.signal })).rejects.toThrow(
            "Aborted",
        );
    });

    it("passes the signal to the loader so the request can be cancelled", async () => {
        const controller = new AbortController();
        const loadData = vi.fn().mockResolvedValue(rows);

        await pollQuery(loadData, { q: 1 }, { signal: controller.signal });

        expect(loadData).toHaveBeenCalledWith({ q: 1 }, controller.signal);
    });

    it("honours a shorter retry limit", async () => {
        const loadData = vi.fn().mockResolvedValue(waiting);

        await expect(settle(pollQuery(loadData, {}, { retryLimit: 2 }))).rejects.toThrow(
            QUERY_FAILED,
        );
        expect(loadData).toHaveBeenCalledTimes(3);
    });
});
