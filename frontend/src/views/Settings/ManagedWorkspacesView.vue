<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div v-if="!loaded" class="flex flex-1 items-center justify-center">
            <BaseSpinner />
        </div>

        <div
            v-else-if="workspaces.length === 0"
            class="flex flex-1 flex-col items-center justify-center px-6 text-center"
        >
            <PresentationChartBarIcon class="size-12 text-gray-300" aria-hidden="true" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.collimato_workspaces.empty.title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("settings.collimato_workspaces.empty.description") }}
            </p>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="px-6 py-6">
                <DataTable
                    v-model:sort="sort"
                    :columns="columns"
                    :items="filteredWorkspaces"
                    clickable
                    @row-click="openWorkspace"
                >
                    <template #toolbar>
                        <div class="flex items-center justify-between gap-x-4">
                            <div class="relative w-full max-w-xs">
                                <MagnifyingGlassIcon
                                    class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
                                    aria-hidden="true"
                                />
                                <input
                                    v-model="searchQuery"
                                    type="search"
                                    :placeholder="t('common.placeholder.search')"
                                    class="block w-full rounded-md border-0 bg-white py-1.5 pl-9 pr-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                />
                            </div>
                            <p class="shrink-0 text-sm text-gray-500">
                                {{
                                    t("settings.collimato_workspaces.count_summary", {
                                        count: filteredWorkspaces.length,
                                    })
                                }}
                            </p>
                        </div>
                    </template>

                    <template #name="{ item }">
                        <div class="flex items-center">
                            <LetterAvatar
                                :id="item.id"
                                :name="item.name"
                                class="size-10 shrink-0 rounded-lg text-sm"
                            />
                            <div class="ml-4 min-w-0">
                                <div class="flex items-center gap-x-2">
                                    <span class="truncate font-medium text-gray-900">
                                        {{ item.name }}
                                    </span>
                                    <span
                                        v-if="item.status === 'draft'"
                                        class="inline-flex shrink-0 items-center rounded-md bg-yellow-50 px-2 py-0.5 text-xs font-medium text-yellow-800 ring-1 ring-inset ring-yellow-600/20"
                                    >
                                        {{ t("settings.collimato_workspaces.draft") }}
                                    </span>
                                </div>
                                <div
                                    v-if="item.description"
                                    class="mt-1 max-w-md truncate text-gray-500"
                                >
                                    {{ item.description }}
                                </div>
                            </div>
                        </div>
                    </template>

                    <template #created_by="{ item }">
                        <UserAvatarWithText
                            :user-id="item.created_by"
                            avatar-class="size-6 shrink-0"
                            text-class="ml-2 truncate text-sm text-gray-700"
                        />
                    </template>

                    <template #member_count="{ item }">
                        <div class="flex items-center">
                            <div class="flex -space-x-1.5">
                                <div
                                    v-for="memberId in item.member_ids || []"
                                    :key="memberId"
                                    class="size-6 rounded-full ring-2 ring-white"
                                >
                                    <UserAvatar :user-id="memberId" />
                                </div>
                            </div>
                            <span
                                v-if="extraMembers(item) > 0"
                                class="ml-1.5 text-xs text-gray-500"
                            >
                                +{{ extraMembers(item) }}
                            </span>
                        </div>
                    </template>

                    <template #created_at="{ item }">
                        {{ getDate(item.created_at) }}
                    </template>

                    <template #actions="{ item }">
                        <RouterLink
                            :to="{ name: 'collimato-settings-workspace', params: { id: item.id } }"
                            class="font-medium text-indigo-600 hover:text-indigo-900"
                            @click.stop
                        >
                            {{ t("common.button.edit") }}
                            <span class="sr-only">, {{ item.name }}</span>
                        </RouterLink>
                    </template>

                    <template #empty>{{ t("settings.collimato_workspaces.no_match") }}</template>
                </DataTable>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import collimatoService from "@/services/collimatoService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import BaseSpinner from "@/components/BaseSpinner.vue";
import DataTable from "@/components/DataTable.vue";
import LetterAvatar from "@/components/LetterAvatar.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { PresentationChartBarIcon } from "@heroicons/vue/24/outline";
import { MagnifyingGlassIcon } from "@heroicons/vue/20/solid";

const router = useRouter();
const alertStore = useAlertStore();
const { getDate } = useDateOperations();

const workspaces = ref([]);
const loaded = ref(false);
const searchQuery = ref("");
const sort = ref({ key: "name", desc: false });

const columns = computed(() => [
    { key: "name", label: t.value("data_table.name"), sortable: true },
    {
        key: "created_by",
        label: t.value("data_table.owner"),
        hiddenBelow: "lg",
    },
    {
        key: "member_count",
        label: t.value("data_table.members"),
        sortable: true,
        hiddenBelow: "sm",
    },
    {
        key: "created_at",
        label: t.value("data_table.created"),
        sortable: true,
        hiddenBelow: "md",
    },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const filteredWorkspaces = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();

    if (!q) return workspaces.value;

    return workspaces.value.filter((w) => {
        return w.name.toLowerCase().includes(q) || (w.description ?? "").toLowerCase().includes(q);
    });
});

function openWorkspace(workspace) {
    router.push({ name: "collimato-settings-workspace", params: { id: workspace.id } });
}

function extraMembers(workspace) {
    return (workspace.member_count ?? 0) - (workspace.member_ids?.length ?? 0);
}

onMounted(async () => {
    try {
        const response = await collimatoService.listAllWorkspaces();

        workspaces.value = response.data ?? [];
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        loaded.value = true;
    }
});
</script>
