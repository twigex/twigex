// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import groupService from "@/services/groupService";

const DEBOUNCE_MS = 250;

export function useGroupSearch(limit = 20) {
    const results = ref([]);
    const loading = ref(false);

    let seq = 0;
    let timer = null;

    async function run(q) {
        const mySeq = ++seq;

        loading.value = true;

        try {
            const res = await groupService.search(q, limit);

            if (mySeq !== seq) return;

            results.value = res.data ?? [];
        } catch {
            if (mySeq !== seq) return;

            results.value = [];
        } finally {
            if (mySeq === seq) loading.value = false;
        }
    }

    function search(q = "", delay = DEBOUNCE_MS) {
        if (timer) clearTimeout(timer);

        loading.value = true;
        timer = setTimeout(() => run(q.trim()), delay);
    }

    function reset() {
        if (timer) clearTimeout(timer);

        seq++;
        results.value = [];
        loading.value = false;
    }

    return { results, loading, search, reset };
}
