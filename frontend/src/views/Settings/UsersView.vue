<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div
            v-if="!can('create_users') && !can('delete_users') && !can('edit_users')"
            class="flex h-full flex-col items-center justify-center px-6 text-center"
        >
            <LockClosedIcon class="size-12 text-gray-300" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.access_restricted.title") }}
            </h3>
            <p class="mt-1 max-w-sm text-sm text-gray-500">
                {{ t("settings.access_restricted.users") }}
            </p>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="space-y-6 px-6 py-6">
                <div v-if="seatsOverLimit" class="rounded-md bg-yellow-50 p-4">
                    <div class="flex">
                        <LockClosedIcon
                            class="size-5 shrink-0 text-yellow-400"
                            aria-hidden="true"
                        />
                        <div class="ml-3">
                            <h3 class="text-sm font-medium text-yellow-800">
                                {{
                                    t("settings.license.seats_over.title", {
                                        used: activeUsers,
                                        licensed: seatLimit,
                                    })
                                }}
                            </h3>
                            <p class="mt-1 text-sm text-yellow-700">
                                {{
                                    t("settings.license.seats_over.description", {
                                        hard: seatHardLimit,
                                    })
                                }}
                            </p>
                        </div>
                    </div>
                </div>

                <DataTable
                    v-model:sort="sort"
                    :columns="columns"
                    :items="users"
                    :loading="loading"
                    :itemsPerPage="PAGE_SIZE"
                    :totalItems="totalUsers"
                    :currentPage="page"
                    clickable
                    @update:currentPage="loadUsers"
                    @row-click="editUser"
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
                                    :placeholder="t('settings.users.placeholder.search')"
                                    class="block w-full rounded-md border-0 bg-white py-1.5 pl-9 pr-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                />
                            </div>
                            <BaseButton
                                v-if="can('create_users')"
                                :prepend-icon="PlusIcon"
                                @click="router.push({ name: 'new-user' })"
                            >
                                {{ t("settings.users.button.add_user") }}
                            </BaseButton>
                        </div>
                    </template>

                    <template #name="{ item }">
                        <div class="flex items-center">
                            <div class="size-10 shrink-0">
                                <UserAvatar :user="item" />
                            </div>
                            <div class="ml-4 min-w-0">
                                <div class="truncate font-medium text-gray-900">
                                    {{ item.name }} {{ item.lastname }}
                                </div>
                                <div class="mt-1 truncate text-gray-500">{{ item.email }}</div>
                            </div>
                        </div>
                    </template>

                    <template #storage_limit="{ item }">
                        {{ storageLabel(item.storage_limit) }}
                    </template>

                    <template #status="{ item }">
                        <span
                            v-if="item.deactivated_at == 0"
                            class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
                        >
                            {{ t("settings.users.status.active") }}
                        </span>
                        <span
                            v-else
                            class="inline-flex items-center rounded-md bg-red-50 px-2 py-1 text-xs font-medium text-red-700 ring-1 ring-inset ring-red-600/20"
                        >
                            {{ t("settings.users.status.deactivated") }}
                        </span>
                    </template>

                    <template #role="{ item }">
                        {{ roleLabel(item.role) }}
                    </template>

                    <template #actions="{ item }">
                        <RouterLink
                            :to="{ name: 'edit-user', params: { id: item.id } }"
                            class="font-medium text-indigo-600 hover:text-indigo-900"
                            @click.stop
                        >
                            {{ t("settings.users.edit") }}
                            <span class="sr-only">, {{ item.name }} {{ item.lastname }}</span>
                        </RouterLink>
                    </template>
                </DataTable>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";
import { storageLabel } from "@/constants/storage";

import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRouter } from "vue-router";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { useSettingsStore } from "@/store/settings";
import { usePermissions } from "@/composables/usePermissions";
import userService from "@/services/userService";
import roleService from "@/services/roleService";
import { LockClosedIcon } from "@heroicons/vue/24/outline";
import { MagnifyingGlassIcon, PlusIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import { roleLabel as systemRoleLabel } from "@/utils/roleLabels";

const columns = computed(() => [
    { key: "name", label: t.value("settings.users.table.name"), sortable: true },
    {
        key: "storage_limit",
        label: t.value("settings.users.table.storage_space"),
        sortable: true,
        hiddenBelow: "lg",
    },
    {
        key: "status",
        label: t.value("settings.users.table.status"),
        sortable: true,
        hiddenBelow: "md",
    },
    {
        key: "role",
        label: t.value("settings.users.table.role"),
        sortable: true,
        hiddenBelow: "sm",
    },
    {
        key: "actions",
        label: t.value("settings.users.table.actions"),
        srOnly: true,
        align: "right",
    },
]);

const userStore = useUserStore();
const settingsStore = useSettingsStore();

// Inside the grace band the instance keeps working, so the admin only learns
// they are over their licence if we say so.
const seatsOverLimit = computed(() => settingsStore.getLicenseFeature("seats_over_limit"));
const activeUsers = computed(() => settingsStore.getLicenseFeature("active_users"));
const seatLimit = computed(() => settingsStore.getLicenseFeature("seat_limit"));
const seatHardLimit = computed(() => settingsStore.getLicenseFeature("seat_hard_limit"));
const router = useRouter();
const { can } = usePermissions();

const PAGE_SIZE = 20;
const SEARCH_DEBOUNCE_MS = 250;

const users = ref([]);
const totalUsers = ref(0);
const loading = ref(true);
const sort = ref({ key: "name", desc: false });
const roleLabels = ref({});
const searchQuery = ref("");
const page = ref(1);

function roleLabel(name) {
    if (!name) return "";

    return roleLabels.value[name] ?? name;
}

// Guards against a slow response for an earlier search term landing last.
let requestSeq = 0;

async function loadUsers(requestedPage = 1) {
    page.value = requestedPage;
    loading.value = true;

    const seq = ++requestSeq;

    try {
        const res = await userService.users({
            limit: PAGE_SIZE,
            offset: (requestedPage - 1) * PAGE_SIZE,
            includeDeactivated: true,
            query: searchQuery.value.trim(),
            sort: sort.value,
        });

        if (seq !== requestSeq) return;

        users.value = res.data.items;
        totalUsers.value = res.data.total;
        userStore.addUsers(res.data.items.filter((u) => u.deactivated_at === 0));
    } catch {
        if (seq !== requestSeq) return;
        useAlertStore().showError(t.value("settings.users.error.load_failed"));
    } finally {
        if (seq === requestSeq) loading.value = false;
    }
}

function editUser(user) {
    router.push({ name: "edit-user", params: { id: user.id } });
}

let searchTimer = null;

watch(sort, () => loadUsers(1));

watch(searchQuery, () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => loadUsers(1), SEARCH_DEBOUNCE_MS);
});

onMounted(async () => {
    try {
        const res = await roleService.getRoles();

        roleLabels.value = Object.fromEntries(
            res.data.map((r) => [r.name, systemRoleLabel(r, "system")]),
        );
    } catch {
        // delete_users can reach this list but not /settings/roles.
        roleLabels.value = {};
    }

    await loadUsers();
});
</script>
