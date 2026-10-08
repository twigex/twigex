// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, watchEffect, toValue } from "vue";
import { useUserStore } from "@/store/user";

// DataLoader-style batching: coalesce every id requested within the same tick
// into a single ensureUsers() call, so a list of N components triggers one
// request instead of N. Lives here (not in the store getter) so the store's
// getters stay pure and the fetch is a proper effect.
//
// The request runs in a microtask so it falls outside the tracked scope of the
// effect that asked for it. Inline, the "do we have this one?" check reads the
// cache the fetch then fills, so the fetch re-triggers its own caller.
let queued = new Set();
let flushScheduled = false;

function requestUser(store, id) {
    if (!id || store.usersMap[id] || store.user?.id === id) return;

    queued.add(id);

    if (flushScheduled) return;
    flushScheduled = true;
    queueMicrotask(() => {
        flushScheduled = false;
        const ids = [...queued];

        queued = new Set();
        if (ids.length) store.ensureUsers(ids);
    });
}

// Reactively resolve a single user, loading it on demand. `id` may be a ref,
// getter, or plain value. Returns a computed that fills in once loaded.
export function useUser(id) {
    const store = useUserStore();

    watchEffect(() => {
        const wanted = toValue(id);

        queueMicrotask(() => requestUser(store, wanted));
    });

    return computed(() => store.getUserById(toValue(id)));
}

// useUser plus where the load stands, so a caller can tell a user still on
// its way from one that will never arrive.
export function useUserState(id) {
    const store = useUserStore();
    const user = useUser(id);

    const failed = computed(() => store.isUserFailed(toValue(id)));
    const loading = computed(() => Boolean(toValue(id)) && !user.value && !failed.value);

    return { user, loading, failed };
}

// Same DataLoader batching as requestUser, but keyed by @mention handle so a
// screenful of <Mention>s coalesces into one /users/by-usernames request.
let queuedNames = new Set();
let flushNamesScheduled = false;

function requestUsername(store, username) {
    if (!username || store.getUserByUsername(username)) return;

    queuedNames.add(username);

    if (flushNamesScheduled) return;
    flushNamesScheduled = true;
    queueMicrotask(() => {
        flushNamesScheduled = false;
        const names = [...queuedNames];

        queuedNames = new Set();
        if (names.length) store.ensureUsersByUsernames(names);
    });
}

// Reactively resolve a user from an @mention handle, loading on demand.
// `username` may be a ref, getter, or plain value.
export function useUserByUsername(username) {
    const store = useUserStore();

    watchEffect(() => {
        const wanted = toValue(username);

        queueMicrotask(() => requestUsername(store, wanted));
    });

    return computed(() => store.getUserByUsername(toValue(username)));
}

// Folded in as the post renders rather than as it arrives: a post reaches the
// client through many paths, but only one draws it.
export function useMentionUsers(post) {
    const store = useUserStore();

    watchEffect(() => {
        const mentions = toValue(post)?.mentions;

        if (mentions) queueMicrotask(() => store.addMentionUsers(mentions));
    });
}

// Reactively resolve a list of users, loading any that aren't cached.
export function useUsers(ids) {
    const store = useUserStore();

    watchEffect(() => {
        const wanted = toValue(ids) ?? [];

        queueMicrotask(() => {
            for (const id of wanted) requestUser(store, id);
        });
    });

    return computed(() => store.getUsers(toValue(ids) ?? []));
}
