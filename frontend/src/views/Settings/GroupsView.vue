<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div
            v-if="!isAdmin"
            class="flex h-full flex-col items-center justify-center px-6 text-center"
        >
            <LockClosedIcon class="size-12 text-gray-300" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.access_restricted.title") }}
            </h3>
            <p class="mt-1 max-w-sm text-sm text-gray-500">
                {{ t("settings.access_restricted.groups") }}
            </p>
        </div>

        <div v-else-if="!loaded" class="flex flex-1 items-center justify-center">
            <BaseSpinner />
        </div>

        <div
            v-else-if="groupCount === 0 && !hasGroupsLicense"
            class="flex flex-1 items-center justify-center p-6"
        >
            <div class="w-full max-w-lg">
                <LicensePurchase :feature="'groups'" />
            </div>
        </div>

        <div
            v-else-if="groupCount === 0"
            class="flex flex-1 flex-col items-center justify-center px-6 text-center"
        >
            <UserGroupIcon class="size-12 text-gray-300" aria-hidden="true" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.groups.empty.title") }}
            </h3>
            <p class="mt-1 max-w-sm text-sm text-gray-500">
                {{ t("settings.groups.empty.description") }}
            </p>
            <BaseButton
                class="mt-6"
                :prepend-icon="PlusIcon"
                @click="router.push({ name: 'new-group' })"
            >
                {{ t("settings.groups.empty.cta") }}
            </BaseButton>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="space-y-6 px-6 py-6">
                <div v-if="hasLapsedGroups" class="rounded-md bg-yellow-50 p-4">
                    <div class="flex">
                        <LockClosedIcon
                            class="size-5 shrink-0 text-yellow-400"
                            aria-hidden="true"
                        />
                        <div class="ml-3">
                            <h3 class="text-sm font-medium text-yellow-800">
                                {{ t("settings.groups.lapsed.title") }}
                            </h3>
                            <p class="mt-1 text-sm text-yellow-700">
                                {{ t("settings.groups.lapsed.description") }}
                            </p>
                        </div>
                    </div>
                </div>

                <DataTable
                    v-model:sort="sort"
                    :columns="columns"
                    :items="groups"
                    :loading="loading"
                    :itemsPerPage="PAGE_SIZE"
                    :totalItems="total"
                    :currentPage="page"
                    :clickable="hasGroupsLicense"
                    @update:currentPage="changePage"
                    @row-click="openGroup"
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
                                {{ t("settings.groups.count_summary", { count: total }) }}
                            </p>
                        </div>
                    </template>

                    <template #name="{ item }">
                        <div class="flex items-center">
                            <div
                                class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50"
                            >
                                <UserGroupIcon class="size-5 text-indigo-600" aria-hidden="true" />
                            </div>
                            <div class="ml-4 min-w-0">
                                <div class="truncate font-medium text-gray-900">
                                    {{ item.name }}
                                </div>
                                <div
                                    :class="[
                                        item.description ? 'text-gray-500' : 'italic text-gray-400',
                                        'mt-1 max-w-md truncate',
                                    ]"
                                >
                                    {{ item.description || t("settings.groups.no_description") }}
                                </div>
                            </div>
                        </div>
                    </template>

                    <template #member_count="{ item }">
                        <span class="inline-flex items-center gap-x-1.5">
                            <UsersIcon class="size-4 text-gray-400" aria-hidden="true" />
                            {{ item.member_count ?? 0 }}
                        </span>
                    </template>

                    <template #created_at="{ item }">
                        {{ getDate(item.created_at) }}
                    </template>

                    <template #actions="{ item }">
                        <div class="flex items-center justify-end gap-x-4 font-medium">
                            <RouterLink
                                v-if="hasGroupsLicense"
                                :to="{ name: 'edit-group', params: { id: item.id } }"
                                class="text-indigo-600 hover:text-indigo-900"
                                @click.stop
                            >
                                {{ t("common.button.edit") }}
                                <span class="sr-only">, {{ item.name }}</span>
                            </RouterLink>
                            <button
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click.stop="askDelete(item)"
                            >
                                {{ t("common.button.delete") }}
                                <span class="sr-only">, {{ item.name }}</span>
                            </button>
                        </div>
                    </template>

                    <template #empty>{{ t("settings.groups.no_match") }}</template>
                </DataTable>
            </div>
        </div>

        <ConfirmDialog
            :open="deleteDialogOpen"
            :title="t('settings.groups.confirm_delete_title', { name: groupToDelete?.name ?? '' })"
            :message="t('settings.groups.confirm_delete')"
            :confirm-label="t('common.button.delete')"
            @confirm="confirmDelete"
            @close="deleteDialogOpen = false"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import groupService from "@/services/groupService";
import { useSettingsStore } from "@/store/settings";
import LicensePurchase from "@/components/Settings/LicensePurchase.vue";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { LockClosedIcon, UserGroupIcon } from "@heroicons/vue/24/outline";
import { MagnifyingGlassIcon, PlusIcon, UsersIcon } from "@heroicons/vue/20/solid";

const router = useRouter();
const userStore = useUserStore();
const alertStore = useAlertStore();

const settingsStore = useSettingsStore();

const isAdmin = computed(() => userStore.user?.role === "system_admin");
const hasGroupsLicense = computed(() => settingsStore.getLicenseFeature("groups"));

const PAGE_SIZE = 20;
const SEARCH_DEBOUNCE_MS = 250;

const { getDate } = useDateOperations();

const columns = computed(() => [
    { key: "name", label: t.value("data_table.name"), sortable: true },
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

const groups = ref([]);
const sort = ref({ key: "name", desc: false });
const loading = ref(false);
const total = ref(0);
const groupCount = ref(0);
const page = ref(1);
const loaded = ref(false);
const searchQuery = ref("");
const deleteDialogOpen = ref(false);
const groupToDelete = ref(null);

// Groups outlive the licence, so an unlicensed instance with groups is a lapse
// rather than someone who never had the feature.
const hasLapsedGroups = computed(() => !hasGroupsLicense.value && groupCount.value > 0);

let request = 0;
let searchTimer = null;

watch(sort, () => {
    page.value = 1;
    loadGroups();
});

watch(searchQuery, () => {
    clearTimeout(searchTimer);

    searchTimer = setTimeout(() => {
        page.value = 1;
        loadGroups();
    }, SEARCH_DEBOUNCE_MS);
});

async function loadGroups() {
    const current = ++request;
    const query = searchQuery.value.trim();

    loading.value = true;

    try {
        const { data } = await groupService.page(
            query,
            PAGE_SIZE,
            (page.value - 1) * PAGE_SIZE,
            sort.value,
        );

        if (current !== request) return;

        groups.value = data.items ?? [];
        total.value = data.total ?? 0;

        if (!query) groupCount.value = total.value;
    } catch {
        if (current !== request) return;

        alertStore.showError(t.value("settings.groups.list_failed"));
    } finally {
        if (current === request) loading.value = false;
    }

    if (groups.value.length === 0 && page.value > 1) {
        page.value -= 1;
        await loadGroups();
    }
}

function changePage(newPage) {
    page.value = newPage;
    loadGroups();
}

function openGroup(group) {
    router.push({ name: "edit-group", params: { id: group.id } });
}

function askDelete(group) {
    groupToDelete.value = group;
    deleteDialogOpen.value = true;
}

async function confirmDelete() {
    deleteDialogOpen.value = false;

    try {
        await groupService.delete(groupToDelete.value.id);

        groupCount.value -= 1;

        await loadGroups();

        alertStore.showSuccess(t.value("settings.groups.deleted"));
    } catch {
        alertStore.showError(t.value("settings.groups.delete_failed"));
    }
}

onMounted(async () => {
    if (!isAdmin.value) return;

    await loadGroups();

    loaded.value = true;
});
</script>
