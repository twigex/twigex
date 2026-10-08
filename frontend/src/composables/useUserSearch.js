// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import userService from "@/services/userService";

const DEBOUNCE_MS = 250;

export function useUserSearch(limit = 20) {
    const query = ref("");
    const results = ref([]);
    const loading = ref(false);

    let seq = 0;
    let timer = null;

    async function run(q) {
        const mySeq = ++seq;

        loading.value = true;
        try {
            const res = await userService.search(q, limit);

            if (mySeq !== seq) return; // a newer request superseded this one
            results.value = res.data ?? [];
        } catch {
            if (mySeq !== seq) return;
            results.value = [];
        } finally {
            if (mySeq === seq) loading.value = false;
        }
    }

    function search(q = "") {
        query.value = q;
        if (timer) clearTimeout(timer);
        timer = setTimeout(() => run(q.trim()), DEBOUNCE_MS);
    }

    function reset() {
        if (timer) clearTimeout(timer);
        seq++;
        query.value = "";
        results.value = [];
        loading.value = false;
    }

    return { query, results, loading, search, reset };
}
