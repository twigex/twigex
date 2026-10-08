<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col overflow-hidden">
        <!-- Loading -->
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <template v-else>
            <!-- Header -->
            <div class="shrink-0 flex items-center justify-between border-b px-6 py-4">
                <div>
                    <h1 class="text-base font-semibold leading-6 text-gray-900">
                        {{ t("collimato.connections.table.title") }}
                    </h1>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("collimato.connections.table.description") }}
                    </p>
                </div>
                <button
                    v-if="collimatoStore.hasPermissionToCreateConnections"
                    type="button"
                    @click="router.push({ name: 'new-connection' })"
                    class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.connections.table.button.new_connection") }}
                </button>
            </div>

            <!-- Empty state -->
            <div
                v-if="connections.length === 0"
                class="flex flex-1 flex-col items-center justify-center"
            >
                <WifiIcon class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                <h3 class="mt-3 text-sm font-semibold text-gray-900">
                    {{ t("collimato.connections.table.no_connections") }}
                </h3>
                <p class="mt-1 text-sm text-gray-500">
                    {{ t("collimato.connections.table.no_connections_description") }}
                </p>
                <button
                    v-if="collimatoStore.hasPermissionToCreateConnections"
                    type="button"
                    @click="router.push({ name: 'new-connection' })"
                    class="mt-6 inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.connections.table.button.new_connection") }}
                </button>
            </div>

            <template v-else>
                <!-- Search bar -->
                <div class="shrink-0 border-b px-6 py-3">
                    <div class="relative min-w-48 sm:max-w-xs">
                        <MagnifyingGlassIcon
                            class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                            aria-hidden="true"
                        />
                        <input
                            v-model="searchQuery"
                            type="search"
                            :placeholder="t('collimato.connections.search_placeholder')"
                            class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                        />
                    </div>
                </div>

                <!-- Card grid -->
                <div class="flex-1 overflow-y-auto p-6">
                    <!-- No search results -->
                    <div
                        v-if="filteredConnections.length === 0"
                        class="flex h-full flex-col items-center justify-center"
                    >
                        <MagnifyingGlassIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                        <p class="mt-3 text-sm font-medium text-gray-900">
                            {{ t("collimato.connections.search_no_results") }}
                        </p>
                        <button
                            type="button"
                            @click="searchQuery = ''"
                            class="mt-3 text-sm text-indigo-600 hover:text-indigo-500"
                        >
                            {{ t("collimato.connections.search_clear") }}
                        </button>
                    </div>

                    <ul
                        v-else
                        role="list"
                        class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4 content-start"
                    >
                        <ItemCard
                            v-for="connection in filteredConnections"
                            :key="connection.id"
                            :title="connection.displayname"
                            @click="editConnection(connection)"
                        >
                            <template #icon>
                                <span
                                    :class="[
                                        getTypeConfig(connection.type).bg,
                                        'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg',
                                    ]"
                                    aria-hidden="true"
                                >
                                    <CircleStackIcon class="h-5 w-5 text-white" />
                                </span>
                            </template>
                            <template #subtitle>
                                <span
                                    class="inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset"
                                    :class="getTypeConfig(connection.type).badge"
                                >
                                    {{ databaseLabel(connection.type) }}
                                </span>
                            </template>
                            <template
                                v-if="
                                    collimatoStore.hasPermissionToEditConnections ||
                                    collimatoStore.hasPermissionToDeleteConnections
                                "
                                #actions
                            >
                                <button
                                    v-if="collimatoStore.hasPermissionToEditConnections"
                                    @click="editConnection(connection)"
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-indigo-600"
                                    :title="t('common.button.edit')"
                                >
                                    <PencilIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                                <button
                                    v-if="collimatoStore.hasPermissionToDeleteConnections"
                                    @click="openDeleteDialog(connection)"
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    :title="t('common.button.delete')"
                                >
                                    <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                            </template>
                            <template #footer>
                                <span class="truncate text-xs text-gray-500">
                                    {{ connection.host
                                    }}{{ connection.port ? `:${connection.port}` : "" }}
                                </span>
                                <span class="shrink-0 truncate text-xs text-gray-400">
                                    {{ connection.database }}
                                </span>
                            </template>
                        </ItemCard>
                    </ul>
                </div>
            </template>
        </template>

        <DeleteDialog
            v-model="deleteDialog"
            :title="t('collimato.connections.delete_dialog.title')"
            :message="t('collimato.connections.delete_dialog.confirm')"
            @delete="deleteConnection"
        />
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import collimatoService from "@/services/collimatoService";
import DeleteDialog from "@/components/Collimato/Dialogs/DeleteDialog.vue";
import ItemCard from "@/components/ItemCard.vue";
import { databaseLabel } from "@/utils/collimato/databaseTypes";
import { PlusIcon, PencilIcon, TrashIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { WifiIcon, CircleStackIcon } from "@heroicons/vue/24/outline";

const router = useRouter();
const route = useRoute();
const collimatoStore = useCollimatoStore();
const alertStore = useAlertStore();

const connections = ref([]);
const itemToDelete = ref(null);
const deleteDialog = ref(false);
const loaded = ref(false);
const searchQuery = ref("");

const typeConfigs = {
    postgres: {
        bg: "bg-blue-500",
        badge: "bg-blue-50 text-blue-700 ring-blue-700/10",
    },
    mysql: {
        bg: "bg-orange-500",
        badge: "bg-orange-50 text-orange-700 ring-orange-700/10",
    },
};

function getTypeConfig(type) {
    return (
        typeConfigs[type] ?? {
            bg: "bg-indigo-500",
            badge: "bg-indigo-50 text-indigo-700 ring-indigo-700/10",
        }
    );
}

const filteredConnections = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();

    if (!query) return connections.value;

    return connections.value.filter(
        (c) =>
            c.displayname?.toLowerCase().includes(query) ||
            c.host?.toLowerCase().includes(query) ||
            databaseLabel(c.type)?.toLowerCase().includes(query) ||
            c.database?.toLowerCase().includes(query),
    );
});

function editConnection(connection) {
    if (!collimatoStore.hasPermissionToEditConnections) return;

    router.push({ name: "edit-connection", params: { id: connection.id } });
}

function openDeleteDialog(item) {
    itemToDelete.value = item;
    deleteDialog.value = true;
}

function deleteConnection() {
    collimatoService
        .deleteConnection(route.params.workspaceId, itemToDelete.value.id)
        .then(() => {
            connections.value = connections.value.filter((c) => c.id !== itemToDelete.value.id);
            itemToDelete.value = null;
        })
        .catch(() => {
            alertStore.showError(t.value("collimato.connections.error.no_permission_delete"));
        });
    deleteDialog.value = false;
}

onMounted(() => {
    if (!collimatoStore.hasPermissionToViewConnections) {
        alertStore.showError(t.value("collimato.connections.error.no_permission_view"));

        return;
    }

    collimatoService.getWorkspaceConnections(route.params.workspaceId).then((response) => {
        connections.value = response.data;
        loaded.value = true;
    });
});
</script>
