<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="relative">
        <div
            ref="reference"
            class="flex flex-wrap items-center gap-2 rounded-md border border-gray-300 bg-white px-3 py-2 shadow-sm focus-within:border-indigo-600 focus-within:ring-1 focus-within:ring-indigo-600"
            @click="focusInput"
        >
            <span
                v-for="user in selectedUsers"
                :key="'u-' + user.id"
                class="inline-flex items-center gap-1 rounded-full bg-gray-100 py-1 pl-1 pr-2 text-xs font-medium text-gray-700"
            >
                <div class="h-5 w-5 shrink-0">
                    <UserAvatar :user="user" />
                </div>
                <span class="max-w-[140px] truncate"> {{ user.name }} {{ user.lastname }} </span>
                <button
                    type="button"
                    class="ml-0.5 text-gray-400 hover:text-gray-700"
                    @click.stop="removeUser(user)"
                >
                    <XMarkIcon class="h-3.5 w-3.5" aria-hidden="true" />
                </button>
            </span>

            <span
                v-for="group in selectedGroups"
                :key="'g-' + group.id"
                class="inline-flex items-center gap-1 rounded-full bg-indigo-50 py-1 pl-1.5 pr-2 text-xs font-medium text-indigo-700"
            >
                <UserGroupIcon class="h-4 w-4 text-indigo-600" aria-hidden="true" />
                <span class="max-w-[140px] truncate">{{ group.name }}</span>
                <button
                    type="button"
                    class="ml-0.5 text-indigo-400 hover:text-indigo-700"
                    @click.stop="removeGroup(group)"
                >
                    <XMarkIcon class="h-3.5 w-3.5" aria-hidden="true" />
                </button>
            </span>

            <input
                ref="input"
                class="min-w-[12ch] flex-1 border-0 bg-transparent p-0 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-0"
                spellcheck="false"
                :placeholder="selectedUsers.length || selectedGroups.length ? '' : placeholder"
                :value="query"
                @input="onInput"
                @blur="open = false"
                @keydown.esc="closeAndBlur"
            />
        </div>

        <Teleport to="body">
            <div
                v-if="showDropdown"
                ref="floating"
                :style="floatingStyles"
                class="z-[100] overflow-y-auto rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5"
            >
                <button
                    v-for="item in results"
                    :key="item.type + '-' + item.id"
                    type="button"
                    class="flex w-full items-center gap-3 px-3 py-2 text-left hover:bg-gray-100"
                    @mousedown.prevent="selectItem(item)"
                >
                    <template v-if="item.type === 'user'">
                        <div class="h-8 w-8 shrink-0">
                            <UserAvatar :user="item" />
                        </div>
                        <div class="min-w-0 flex-1">
                            <div class="truncate text-sm text-gray-900">
                                {{ item.name }} {{ item.lastname }}
                            </div>
                            <div v-if="item.email" class="truncate text-xs text-gray-500">
                                {{ item.email }}
                            </div>
                        </div>
                    </template>
                    <template v-else>
                        <span
                            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-indigo-100"
                        >
                            <UserGroupIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                        </span>
                        <div class="min-w-0 flex-1">
                            <div class="truncate text-sm font-medium text-gray-900">
                                {{ item.name }}
                            </div>
                            <div class="truncate text-xs text-gray-500">
                                {{
                                    t("common.label.group_meta", {
                                        count: item.member_count ?? 0,
                                    })
                                }}
                            </div>
                        </div>
                    </template>
                </button>

                <div
                    v-if="results.length === 0"
                    class="px-3 py-6 text-center text-sm text-gray-500"
                >
                    {{ t("common.label.no_results") }}
                </div>
            </div>
        </Teleport>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import { XMarkIcon } from "@heroicons/vue/20/solid";
import { useFloating, flip, shift, offset, size, autoUpdate } from "@floating-ui/vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { useGroupSearch } from "@/composables/useGroupSearch";
import { useUserSearch } from "@/composables/useUserSearch";

const props = defineProps({
    selectedUsers: { type: Array, default: () => [] },
    selectedGroups: { type: Array, default: () => [] },
    excludeUserIds: { type: Array, default: () => [] },
    excludeGroupIds: { type: Array, default: () => [] },
    placeholder: { type: String, default: "" },
});

const emit = defineEmits(["update:selectedUsers", "update:selectedGroups"]);

const groupSearch = useGroupSearch();
const userSearch = useUserSearch();
const query = ref("");
const open = ref(false);
const input = ref(null);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [
        offset(4),
        flip(),
        shift({ padding: 8 }),
        size({
            apply({ rects, elements, availableHeight }) {
                Object.assign(elements.floating.style, {
                    width: `${rects.reference.width}px`,
                    maxHeight: `${Math.min(availableHeight - 8, 320)}px`,
                });
            },
        }),
    ],
    whileElementsMounted: autoUpdate,
});

const showDropdown = computed(() => open.value && query.value.trim().length > 0);

function onDocumentPointerDown(e) {
    if (!open.value) return;
    if (reference.value?.contains(e.target)) return;
    if (floating.value?.contains(e.target)) return;
    open.value = false;
}

onMounted(() => {
    document.addEventListener("pointerdown", onDocumentPointerDown, true);
    userSearch.search("");
    groupSearch.search("");
});

onBeforeUnmount(() => {
    document.removeEventListener("pointerdown", onDocumentPointerDown, true);
    userSearch.reset();
    groupSearch.reset();
});

const results = computed(() => {
    const q = query.value.toLowerCase().trim();
    const selectedUserIds = new Set(props.selectedUsers.map((u) => u.id));
    const selectedGroupIds = new Set(props.selectedGroups.map((g) => g.id));

    const users = userSearch.results.value
        .filter((u) => !props.excludeUserIds.includes(u.id) && !selectedUserIds.has(u.id))
        .map((u) => ({ ...u, type: "user" }));

    // A leading "@" means searching people by handle, no group matches that.
    const groups = q.startsWith("@")
        ? []
        : groupSearch.results.value
              .filter((g) => !props.excludeGroupIds.includes(g.id) && !selectedGroupIds.has(g.id))
              .map((g) => ({ ...g, type: "group" }));

    return [...users, ...groups];
});

function onInput(e) {
    query.value = e.target.value;
    userSearch.search(e.target.value);

    if (!e.target.value.trim().startsWith("@")) groupSearch.search(e.target.value);

    open.value = true;
}

function focusInput() {
    input.value?.focus();
    open.value = true;
}

function closeAndBlur() {
    open.value = false;
    input.value?.blur();
}

function selectItem(item) {
    if (item.type === "user") {
        emit("update:selectedUsers", [...props.selectedUsers, item]);
    } else {
        emit("update:selectedGroups", [...props.selectedGroups, item]);
    }

    query.value = "";
    userSearch.search("");
    open.value = true;
    input.value?.focus();
}

function removeUser(user) {
    emit(
        "update:selectedUsers",
        props.selectedUsers.filter((u) => u.id !== user.id),
    );
}

function removeGroup(group) {
    emit(
        "update:selectedGroups",
        props.selectedGroups.filter((g) => g.id !== group.id),
    );
}
</script>
