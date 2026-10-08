// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export const QUERY_RETRY_LIMIT = 30;
export const QUERY_RETRY_MS = 2000;
export const QUERY_FAILED = "collimato.data_view.error.load_data_failed";

function aborted(signal) {
    return new Promise((_, reject) => {
        if (signal?.aborted) {
            reject(new DOMException("Aborted", "AbortError"));

            return;
        }

        signal?.addEventListener("abort", () => reject(new DOMException("Aborted", "AbortError")));
    });
}

function wait(ms, signal) {
    return new Promise((resolve, reject) => {
        if (signal?.aborted) {
            reject(new DOMException("Aborted", "AbortError"));

            return;
        }

        const timer = setTimeout(resolve, ms);

        signal?.addEventListener("abort", () => {
            clearTimeout(timer);
            reject(new DOMException("Aborted", "AbortError"));
        });
    });
}

export async function pollQuery(loadData, query, options = {}) {
    const { signal, retryLimit = QUERY_RETRY_LIMIT, retryMs = QUERY_RETRY_MS } = options;

    for (let attempt = 0; attempt <= retryLimit; attempt++) {
        const result = signal
            ? await Promise.race([loadData(query, signal), aborted(signal)])
            : await loadData(query, signal);

        if (result.data.error === "Continue wait") {
            await wait(retryMs, signal);
            continue;
        }

        if (result.data.error) {
            throw new Error(result.data.error);
        }

        return result;
    }

    throw new Error(QUERY_FAILED);
}
