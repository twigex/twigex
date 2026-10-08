// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import { reactive } from "vue";
import userService from "@/services/userService";

// The in-flight maps hold the request promise, not just the key, so a caller
// asking for something already being fetched can await it rather than return
// before it lands.
const USER_FETCH_CHUNK = 100; // must stay <= the server cap in /users/by-ids
const inflightUserIds = new Map();
const notFoundUserIds = new Set();
const inflightUsernames = new Map();
const notFoundUsernames = new Set();

export const useUserStore = defineStore("user", {
    state: () => ({
        user: null,
        users: [],
        usersMap: reactive({}),
        failedUserIds: reactive({}),
        statuses: [],
        preferences: null,
    }),

    getters: {
        // usersMap holds every cached user, so a miss is a user not loaded
        // yet, not one to look for in `users`.
        getUserById() {
            return (id) => {
                if (!id) return undefined;
                if (this.user?.id === id) return this.user;

                return this.usersMap[id];
            };
        },

        isUserFailed() {
            return (id) => Boolean(id && this.failedUserIds[id]);
        },

        // Grows as users load, so rows drawn before a user arrived can be
        // drawn again with the name.
        knownUserCount() {
            return Object.keys(this.usersMap).length;
        },

        getUsers() {
            return (ids = []) => {
                const resolved = [];

                for (const id of ids) {
                    const user = this.getUserById(id);

                    if (user) resolved.push(user);
                }

                return resolved;
            };
        },

        // Reads usersMap, which is the complete cache, not `users`, which a
        // view can replace. Case-insensitive, like the column.
        getUserIdByUsername() {
            return (username) => {
                if (!username) return null;

                const wanted = username.toLowerCase();

                if (this.user?.username?.toLowerCase() === wanted) {
                    return this.user.id;
                }

                const match =
                    Object.values(this.usersMap).find(
                        (u) => u.username?.toLowerCase() === wanted,
                    ) ?? this.users.find((u) => u.username?.toLowerCase() === wanted);

                return match?.id ?? null;
            };
        },

        getUserByUsername() {
            return (username) => {
                if (!username) return undefined;
                // The logged-in user is kept in `user`, not `users`, so a
                // self-mention only resolves if we check it explicitly.
                if (this.user?.username === username) return this.user;

                return this.users.find((user) => user.username === username);
            };
        },

        getUserFullName() {
            return (id) => {
                const user = this.getUserById(id);

                return user ? `${user.name} ${user.lastname}` : "Unknown User";
            };
        },

        getTimezone(state) {
            if (state.user?.timezone?.useAutomaticTimezone === "true") {
                return Intl.DateTimeFormat().resolvedOptions().timeZone;
            }

            return (
                state.user?.timezone?.manualTimezone ||
                Intl.DateTimeFormat().resolvedOptions().timeZone
            );
        },

        getClockDisplay(state) {
            if (!state.preferences?.DisplaySettings) return "24h";
            const clock = state.preferences.DisplaySettings.find(
                (item) => item.name === "clock_display",
            );

            return clock ? clock.value : "24h";
        },

        getLanguage: (state) => {
            if (!state.preferences?.DisplaySettings) return null;
            const language = state.preferences.DisplaySettings.find(
                (item) => item.name === "language",
            );

            return language?.value || null;
        },

        getUserStatus: (state) => (userId) => {
            let status = state.statuses.find((status) => status.user_id === userId);

            if (!status) {
                return {
                    user_id: userId,
                    status: "offline",
                    last_activity: null,
                };
            }

            return status;
        },

        getPhotoSrc: (state) => (userId, photoId) => {
            let currentPhoto = photoId;

            if (currentPhoto === undefined) {
                const user = state.user?.id === userId ? state.user : state.usersMap[userId];

                currentPhoto = user?.photo;
            }

            return currentPhoto
                ? `/api/users/photo/${userId}?v=${encodeURIComponent(currentPhoto)}`
                : `/api/users/photo/${userId}`;
        },
    },

    actions: {
        setUser(user) {
            this.user = user;
        },

        addUsers(users) {
            for (const u of users) {
                if (!this.usersMap[u.id]) {
                    this.users.push(u);
                }

                this.usersMap[u.id] = u;
                delete this.failedUserIds[u.id];
            }
        },

        // These carry fewer fields than a fetched user but are resolved
        // fresher, so overlay rather than replace.
        // Writes only what is new: rebuilding every entry looked like a change
        // even when nothing had changed, re-triggering the effect that read it.
        addMentionUsers(mentions) {
            if (!mentions) return;

            const fresh = [];

            for (const [id, user] of Object.entries(mentions)) {
                const cached = this.usersMap[id];

                if (cached && Object.keys(user).every((k) => cached[k] === user[k])) {
                    continue;
                }

                fresh.push({ ...(cached ?? {}), id, ...user });
            }

            if (fresh.length) this.addUsers(fresh);
        },

        async ensureUsers(ids) {
            const wanted = [
                ...new Set(
                    ids.filter((id) => id && !this.usersMap[id] && !notFoundUserIds.has(id)),
                ),
            ];

            if (wanted.length === 0) return;

            const waitFor = new Set();
            const missing = [];

            for (const id of wanted) {
                const pending = inflightUserIds.get(id);

                if (pending) waitFor.add(pending);
                else missing.push(id);
            }

            if (missing.length) {
                const request = this.fetchUsersByIds(missing);

                missing.forEach((id) => inflightUserIds.set(id, request));
                request.finally(() => missing.forEach((id) => inflightUserIds.delete(id)));
                waitFor.add(request);
            }

            await Promise.all(waitFor);
        },

        async fetchUsersByIds(ids) {
            try {
                // The server caps /users/by-ids at 100 ids per request, so
                // split larger batches to avoid silently dropping the overflow.
                for (let i = 0; i < ids.length; i += USER_FETCH_CHUNK) {
                    const chunk = ids.slice(i, i + USER_FETCH_CHUNK);
                    const res = await userService.byIds(chunk);
                    const fetched = res.data || [];

                    this.addUsers(fetched);

                    const returned = new Set(fetched.map((u) => u.id));

                    chunk.forEach((id) => {
                        if (returned.has(id)) return;
                        notFoundUserIds.add(id);
                        this.failedUserIds[id] = true;
                    });
                }
            } catch {
                // Non-fatal, photos/names may just be empty
                ids.forEach((id) => {
                    if (!this.usersMap[id]) this.failedUserIds[id] = true;
                });
            }
        },

        async ensureUsersByUsernames(usernames) {
            const wanted = [
                ...new Set(
                    usernames.filter(
                        (name) =>
                            name && !this.getUserByUsername(name) && !notFoundUsernames.has(name),
                    ),
                ),
            ];

            if (wanted.length === 0) return;

            const waitFor = new Set();
            const missing = [];

            for (const name of wanted) {
                const pending = inflightUsernames.get(name);

                if (pending) waitFor.add(pending);
                else missing.push(name);
            }

            if (missing.length) {
                const request = this.fetchUsersByUsernames(missing);

                missing.forEach((name) => inflightUsernames.set(name, request));
                request.finally(() => missing.forEach((name) => inflightUsernames.delete(name)));
                waitFor.add(request);
            }

            await Promise.all(waitFor);
        },

        async fetchUsersByUsernames(names) {
            try {
                for (let i = 0; i < names.length; i += USER_FETCH_CHUNK) {
                    const chunk = names.slice(i, i + USER_FETCH_CHUNK);
                    const res = await userService.byUsernames(chunk);
                    const fetched = res.data || [];

                    this.addUsers(fetched);

                    const returned = new Set(fetched.map((u) => u.username));

                    chunk.forEach((name) => {
                        if (!returned.has(name)) notFoundUsernames.add(name);
                    });
                }
            } catch {
                // Non-fatal; the mention just stays a plain @handle.
            }
        },

        setUserStatus(status) {
            const index = this.statuses.findIndex((s) => s.user_id === status.user_id);

            if (index !== -1) {
                this.statuses[index] = status;
            } else {
                this.statuses.push(status);
            }
        },

        setStatuses(statuses) {
            this.statuses = statuses;
        },

        setPreferences(preferences) {
            this.preferences = preferences;
        },

        setPhoto(userId, newPhotoId) {
            const userInList = this.users.find((u) => u.id === userId);

            if (userInList) {
                userInList.photo = newPhotoId;
            }

            if (this.usersMap[userId]) {
                this.usersMap[userId].photo = newPhotoId;
            }

            if (this.user?.id === userId) {
                this.user.photo = newPhotoId;
            }
        },
    },
});
