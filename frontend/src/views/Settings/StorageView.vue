<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="w-full h-full flex flex-col items-center">
        <StorageList
            v-if="storages.length > 0"
            :storages="storages"
            :has-multiple-storages-license="hasMultipleStoragesLicense"
            @create="openCreateDialog"
            @edit="openEditDialog"
            @delete="openDeleteDialog"
            @set-primary="setPrimary"
        />

        <StorageEmptyState v-else @create="openCreateDialog" />

        <StorageCreateDialog ref="createDialog" @created="handleCreated" />

        <StorageEditDialog ref="editDialog" @updated="handleUpdated" />

        <StorageDeleteDialog ref="deleteDialog" @deleted="handleDeleted" />
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { t } from "@/i18n/index.js";
import settingsService from "@/services/settingsService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useSettingsStore } from "@/store/settings";
import StorageList from "@/components/Settings/StorageList.vue";
import StorageEmptyState from "@/components/Settings/StorageEmptyState.vue";
import StorageCreateDialog from "@/components/Settings/StorageCreateDialog.vue";
import StorageEditDialog from "@/components/Settings/StorageEditDialog.vue";
import StorageDeleteDialog from "@/components/Settings/StorageDeleteDialog.vue";

const settingsStore = useSettingsStore();
const hasMultipleStoragesLicense = computed(() =>
    settingsStore.getLicenseFeature("multiple_storages"),
);

const loaded = ref(false);
const storages = ref([]);
const createDialog = ref(null);
const editDialog = ref(null);
const deleteDialog = ref(null);

function openCreateDialog() {
    createDialog.value.open();
}

function handleCreated(storage) {
    storages.value.push(storage);
}

function openEditDialog(storage) {
    editDialog.value.open(storage);
}

function handleUpdated(id, storage) {
    const index = storages.value.findIndex((s) => s.id === id);

    if (index !== -1) storages.value[index] = storage;
}

function openDeleteDialog(storage) {
    deleteDialog.value.open(storage);
}

function handleDeleted(id) {
    storages.value = storages.value.filter((s) => s.id !== id);
}

function setPrimary(storage) {
    settingsService
        .setPrimaryStorage(storage.id)
        .then(() => {
            storages.value.forEach((s) => (s.is_primary = s.id === storage.id));
            useAlertStore().showSuccess(t.value("settings.storage.primary_set_success"));
        })
        .catch((err) => {
            useAlertStore().showError(extractErrorMessage(err));
        });
}

onMounted(() => {
    settingsService
        .getStorages()
        .then((res) => {
            storages.value = res.data;
        })
        .catch((err) => {
            useAlertStore().showError(extractErrorMessage(err));
        })
        .finally(() => {
            loaded.value = true;
        });
});
</script>
