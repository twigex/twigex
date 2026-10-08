// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import groupService from "@/services/groupService";

const DEBOUNCE_MS = 250;
const PAGE_SIZE = 30;

export function useGroupMembers(getGroupId) {
    const members = ref([]);
    const query = ref("");
    const loading = ref(false);
    const done = ref(false);

    let seq = 0;
    let offset = 0;
    let timer = null;

    async function fetchPage(reset) {
        const gid = getGroupId();

        if (!gid) return;

        const mySeq = ++seq;

        loading.value = true;
        try {
            const res = await groupService.members(
                gid,
                query.value.trim(),
                PAGE_SIZE,
                reset ? 0 : offset,
            );

            if (mySeq !== seq) return; // superseded by a newer request

            const batch = res.data?.items ?? [];

            if (reset) {
                members.value = batch;
                offset = batch.length;
            } else {
                // Replace the array reference rather than push() in place:
                // RecycleScroller watches its `items` prop by reference and
                // won't re-render on an in-place mutation.
                members.value = [...members.value, ...batch];
                offset += batch.length;
            }

            // While searching we only load the first page, no scroll-loading.
            done.value = query.value.trim() !== "" || batch.length < PAGE_SIZE;
        } catch {
            if (mySeq !== seq) return;
            if (reset) members.value = [];
            done.value = true;
        } finally {
            if (mySeq === seq) loading.value = false;
        }
    }

    function start() {
        if (timer) clearTimeout(timer);
        query.value = "";
        offset = 0;
        done.value = false;
        members.value = [];
        fetchPage(true);
    }

    function search(q) {
        query.value = q;
        offset = 0;
        done.value = false;
        if (timer) clearTimeout(timer);
        timer = setTimeout(() => fetchPage(true), DEBOUNCE_MS);
    }

    function loadMore() {
        if (query.value.trim() !== "" || loading.value || done.value) return;
        fetchPage(false);
    }

    function reset() {
        if (timer) clearTimeout(timer);
        seq++;
        members.value = [];
        query.value = "";
        offset = 0;
        loading.value = false;
        done.value = false;
    }

    return { members, query, loading, done, start, search, loadMore, reset };
}
