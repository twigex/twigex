<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span>
        <button
            ref="triggerRef"
            type="button"
            class="text-indigo-600 hover:text-indigo-800 font-medium tabular-nums"
            @click.stop="open"
        >
            <slot />
        </button>

        <Teleport to="body">
            <div v-if="show" class="fixed inset-0 z-40" @click="close" />

            <div
                v-if="show"
                ref="popoverRef"
                class="fixed z-40 w-64 rounded-md bg-white shadow-lg ring-1 ring-black/5"
                :style="popoverStyle"
                @click.stop
            >
                <div class="p-2 border-b border-gray-100">
                    <input
                        v-model="query"
                        type="text"
                        :placeholder="t('projects.workspace_members.search_members')"
                        class="block w-full rounded-md border-0 px-2 py-1.5 text-xs text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                        @input="onSearch"
                        autofocus
                    />
                </div>

                <div ref="listRef" class="max-h-56 overflow-y-auto">
                    <div
                        v-for="member in members"
                        :key="member.user_id || member.id"
                        class="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 text-xs text-gray-700"
                    >
                        <div class="h-6 w-6 flex-shrink-0">
                            <UserAvatar
                                :user="member.user_info || { id: member.user_id || member.id }"
                                :name="member.name || member.user_info?.name"
                            />
                        </div>
                        <span class="truncate">{{ fullName(member) }}</span>
                    </div>

                    <div
                        v-if="members.length === 0 && !loading"
                        class="px-3 py-4 text-xs text-gray-400 text-center"
                    >
                        {{ t("projects.workspace_members.no_members") }}
                    </div>

                    <div v-if="loading" class="px-3 py-2 text-xs text-gray-400 text-center">
                        {{ t("common.label.loading") }}
                    </div>

                    <div ref="sentinelRef" class="h-1" />
                </div>
            </div>
        </Teleport>
    </span>
</template>

<script setup>
import { ref, nextTick, onBeforeUnmount } from "vue";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import groupService from "@/services/groupService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors.js";

const props = defineProps({
    groupId: { type: String, required: true },
});

const triggerRef = ref(null);
const sentinelRef = ref(null);
const listRef = ref(null);

const show = ref(false);
const query = ref("");
const members = ref([]);
const offset = ref(0);
const hasMore = ref(true);
const loading = ref(false);
const { floating: popoverRef, floatingStyles: popoverStyle } = useAnchoredPopup({
    anchor: triggerRef,
});

const LIMIT = 30;
let searchTimeout = null;
let observer = null;

function fullName(m) {
    const name = m.name || m.user_info?.name || "";
    const last = m.lastname || m.user_info?.lastname || "";

    return `${name} ${last}`.trim() || m.user_id || "—";
}

let latestRequest = 0;

async function fetchMembers(reset = false) {
    if (!reset && (loading.value || !hasMore.value)) return;
    const seq = ++latestRequest;

    loading.value = true;
    if (reset) {
        members.value = [];
        offset.value = 0;
        hasMore.value = true;
    }

    try {
        const res = await groupService.members(props.groupId, query.value, LIMIT, offset.value);

        if (seq !== latestRequest) return;
        const data = res.data?.items || [];

        members.value.push(...data);
        offset.value += data.length;
        hasMore.value = data.length === LIMIT;
    } catch (error) {
        if (seq !== latestRequest) return;
        hasMore.value = false;
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        if (seq === latestRequest) loading.value = false;
    }
}

function onSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => fetchMembers(true), 300);
}

function setupObserver() {
    if (observer) observer.disconnect();
    if (!sentinelRef.value) return;
    observer = new IntersectionObserver(
        (entries) => {
            if (entries[0]?.isIntersecting && hasMore.value && !loading.value) {
                fetchMembers();
            }
        },
        { root: listRef.value, threshold: 0.1 },
    );
    observer.observe(sentinelRef.value);
}

async function open() {
    show.value = true;
    query.value = "";
    await fetchMembers(true);
    await nextTick();
    setupObserver();
}

function close() {
    show.value = false;
    if (observer) {
        observer.disconnect();
        observer = null;
    }

    clearTimeout(searchTimeout);
}

onBeforeUnmount(close);
</script>
