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
            <ClipboardDocumentListIcon class="size-12 text-gray-300" aria-hidden="true" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.projects.empty.title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("settings.projects.empty.description") }}
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
                                    t("settings.projects.count_summary", {
                                        count: filteredWorkspaces.length,
                                    })
                                }}
                            </p>
                        </div>
                    </template>

                    <template #title="{ item }">
                        <div class="flex items-center">
                            <LetterAvatar
                                :id="item.id"
                                :name="item.title"
                                class="size-10 shrink-0 rounded-lg text-sm"
                            />
                            <div class="ml-4 min-w-0">
                                <div class="truncate font-medium text-gray-900">
                                    {{ item.title }}
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

                    <template #member_count="{ item }">
                        <div class="flex items-center">
                            <div class="flex -space-x-1.5">
                                <div
                                    v-for="memberId in item.member_ids"
                                    :key="memberId"
                                    class="size-6 rounded-full ring-2 ring-white"
                                >
                                    <UserAvatar :user-id="memberId" />
                                </div>
                            </div>
                            <span
                                v-if="item.member_count > item.member_ids.length"
                                class="ml-1.5 text-xs text-gray-500"
                            >
                                +{{ item.member_count - item.member_ids.length }}
                            </span>
                        </div>
                    </template>

                    <template #created_at="{ item }">
                        {{ getDate(item.created_at) }}
                    </template>

                    <template #actions="{ item }">
                        <RouterLink
                            :to="{ name: 'projects-settings-workspace', params: { id: item.id } }"
                            class="font-medium text-indigo-600 hover:text-indigo-900"
                            @click.stop
                        >
                            {{ t("common.button.edit") }}
                            <span class="sr-only">, {{ item.title }}</span>
                        </RouterLink>
                    </template>

                    <template #empty>{{ t("settings.projects.no_match") }}</template>
                </DataTable>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import workspaceService from "@/services/workspaceService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import BaseSpinner from "@/components/BaseSpinner.vue";
import DataTable from "@/components/DataTable.vue";
import LetterAvatar from "@/components/LetterAvatar.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { ClipboardDocumentListIcon } from "@heroicons/vue/24/outline";
import { MagnifyingGlassIcon } from "@heroicons/vue/20/solid";

const router = useRouter();
const alertStore = useAlertStore();
const { getDate } = useDateOperations();

const workspaces = ref([]);
const loaded = ref(false);
const searchQuery = ref("");
const sort = ref({ key: "title", desc: false });

const columns = computed(() => [
    { key: "title", label: t.value("data_table.name"), sortable: true },
    {
        key: "member_count",
        label: t.value("data_table.members"),
        sortable: true,
        hiddenBelow: "sm",
    },
    {
        key: "group_count",
        label: t.value("settings.projects.groups"),
        sortable: true,
        hiddenBelow: "lg",
    },
    {
        key: "table_count",
        label: t.value("settings.projects.tables"),
        sortable: true,
        hiddenBelow: "lg",
    },
    {
        key: "created_at",
        label: t.value("data_table.created"),
        sortable: true,
        hiddenBelow: "md",
    },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const rows = computed(() =>
    workspaces.value.map((workspace) => ({
        ...workspace,
        member_count: workspace.member_count ?? 0,
        member_ids: workspace.members.map((m) => m.user_id),
        group_count: workspace.group_count ?? 0,
        table_count: workspace.table_count ?? 0,
    })),
);

const filteredWorkspaces = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();

    if (!q) return rows.value;

    return rows.value.filter((w) => {
        return w.title.toLowerCase().includes(q) || (w.description ?? "").toLowerCase().includes(q);
    });
});

function openWorkspace(workspace) {
    router.push({ name: "projects-settings-workspace", params: { id: workspace.id } });
}

onMounted(async () => {
    try {
        const response = await workspaceService.listAllManaged();

        workspaces.value = response.data ?? [];
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        loaded.value = true;
    }
});
</script>
