<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div
            v-if="!can('manage_roles')"
            class="flex h-full flex-col items-center justify-center px-6 text-center"
        >
            <LockClosedIcon class="size-12 text-gray-300" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.access_restricted.title") }}
            </h3>
            <p class="mt-1 max-w-sm text-sm text-gray-500">
                {{ t("settings.access_restricted.roles") }}
            </p>
        </div>

        <div v-else-if="!loaded" class="flex flex-1 items-center justify-center">
            <BaseSpinner />
        </div>

        <div
            v-else-if="roleStore.roles.length === 0"
            class="flex flex-1 flex-col items-center justify-center px-6 text-center"
        >
            <IdentificationIcon class="size-12 text-gray-300" aria-hidden="true" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.roles.empty_title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("settings.roles.empty_description") }}
            </p>
            <BaseButton
                class="mt-6"
                :prepend-icon="PlusIcon"
                @click="router.push({ name: 'new-role' })"
            >
                {{ t("settings.roles.new_role") }}
            </BaseButton>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="px-6 py-6">
                <DataTable
                    v-model:sort="sort"
                    :columns="columns"
                    :items="roleRows"
                    clickable
                    @row-click="editRole"
                >
                    <template #display_name="{ item }">
                        <div class="min-w-0">
                            <div class="flex items-center gap-x-2">
                                <span class="truncate font-medium text-gray-900">
                                    {{ item.display_name }}
                                </span>
                                <span
                                    v-if="item.built_in"
                                    class="inline-flex shrink-0 items-center rounded-md bg-gray-50 px-2 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                >
                                    {{ t("settings.roles.built_in") }}
                                </span>
                            </div>
                            <div class="mt-1 flex items-center gap-x-2 text-gray-500">
                                <span class="font-mono text-xs">{{ item.name }}</span>
                                <span v-if="item.description" class="max-w-md truncate">
                                    · {{ item.description }}
                                </span>
                            </div>
                        </div>
                    </template>

                    <template #actions="{ item }">
                        <div class="flex items-center justify-end gap-x-4 font-medium">
                            <RouterLink
                                :to="{ name: 'edit-role', params: { id: item.id } }"
                                class="text-indigo-600 hover:text-indigo-900"
                                @click.stop
                            >
                                {{ t("common.button.edit") }}
                                <span class="sr-only">, {{ item.display_name }}</span>
                            </RouterLink>
                            <button
                                v-if="!item.built_in"
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click.stop="pendingDelete = item.source"
                            >
                                {{ t("common.button.delete") }}
                                <span class="sr-only">, {{ item.display_name }}</span>
                            </button>
                        </div>
                    </template>
                </DataTable>
            </div>
        </div>

        <ConfirmDialog
            :open="pendingDelete !== null"
            :title="
                t('settings.roles.delete_confirm', {
                    name: pendingDelete ? roleLabel(pendingDelete, 'system') : '',
                })
            "
            :confirm-label="t('common.button.delete')"
            @confirm="deleteRole(pendingDelete)"
            @close="pendingDelete = null"
        />
    </div>
</template>

<script setup>
import { computed, ref, onMounted } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useRoleStore } from "@/store/roles";
import { useAlertStore } from "@/store/alerts";
import roleService from "@/services/roleService";
import { usePermissions } from "@/composables/usePermissions";
import { t } from "@/i18n/index";
import { PlusIcon } from "@heroicons/vue/20/solid";
import { IdentificationIcon, LockClosedIcon } from "@heroicons/vue/24/outline";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import DataTable from "@/components/DataTable.vue";
import { roleLabel, roleDescription } from "@/utils/roleLabels";

const router = useRouter();
const roleStore = useRoleStore();
const { can } = usePermissions();
const loaded = ref(false);
const pendingDelete = ref(null);
const sort = ref({ key: "display_name", desc: false });

const columns = computed(() => [
    { key: "display_name", label: t.value("data_table.role"), sortable: true },
    {
        key: "permission_count",
        label: t.value("data_table.permissions"),
        sortable: true,
        hiddenBelow: "sm",
    },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const roleRows = computed(() =>
    roleStore.roles.map((role) => ({
        ...role,
        display_name: roleLabel(role, "system"),
        description: roleDescription(role, "system"),
        permission_count: role.permissions?.length ?? 0,
        source: role,
    })),
);

function editRole(role) {
    router.push({ name: "edit-role", params: { id: role.id } });
}

onMounted(() => {
    roleService.getRoles().then((res) => {
        roleStore.setRoles(res.data);
        loaded.value = true;
    });
});

function deleteRole(role) {
    pendingDelete.value = null;

    roleService.deleteRole(role.id).then(() => {
        roleStore.removeRole(role.id);
        useAlertStore().showSuccess(t.value("settings.roles.role_deleted"));
    });
}
</script>
