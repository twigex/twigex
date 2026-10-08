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
                        {{ t("collimato.dashboards.table.title") }}
                    </h1>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("collimato.dashboards.table.description") }}
                    </p>
                </div>
                <button
                    v-if="collimatoStore.hasPermissionToCreateDashboards"
                    type="button"
                    @click="open = true"
                    class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.dashboards.button.new_dashboard") }}
                </button>
            </div>

            <!-- Empty state -->
            <div
                v-if="dashboards.length === 0"
                class="flex flex-1 flex-col items-center justify-center"
            >
                <Squares2X2Icon class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                <h3 class="mt-3 text-sm font-semibold text-gray-900">
                    {{ t("collimato.dashboards.empty.title") }}
                </h3>
                <p class="mt-1 text-sm text-gray-500">
                    {{ t("collimato.dashboards.empty.description") }}
                </p>
                <button
                    v-if="collimatoStore.hasPermissionToCreateDashboards"
                    type="button"
                    @click="open = true"
                    class="mt-6 inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.dashboards.button.new_dashboard") }}
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
                            :placeholder="t('collimato.dashboards.search_placeholder')"
                            class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                        />
                    </div>
                </div>

                <!-- Card grid -->
                <div class="flex-1 overflow-y-auto p-6">
                    <!-- No search results -->
                    <div
                        v-if="filteredDashboards.length === 0"
                        class="flex h-full flex-col items-center justify-center"
                    >
                        <MagnifyingGlassIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                        <p class="mt-3 text-sm font-medium text-gray-900">
                            {{ t("collimato.dashboards.search_no_results") }}
                        </p>
                        <button
                            type="button"
                            @click="searchQuery = ''"
                            class="mt-3 text-sm text-indigo-600 hover:text-indigo-500"
                        >
                            {{ t("collimato.dashboards.search_clear") }}
                        </button>
                    </div>

                    <ul
                        v-else
                        role="list"
                        class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4 content-start"
                    >
                        <ItemCard
                            v-for="dashboard in filteredDashboards"
                            :key="dashboard.id"
                            :title="dashboard.title"
                            :description="dashboard.description"
                            :description-fallback="t('common.label.no_description')"
                            @click="rowclick(dashboard)"
                        >
                            <template #icon>
                                <LetterAvatar
                                    :id="dashboard.id"
                                    :name="dashboard.title"
                                    class="h-10 w-10 rounded-lg text-sm"
                                />
                            </template>
                            <template
                                v-if="
                                    collimatoStore.hasPermissionToEditDashboards ||
                                    collimatoStore.hasPermissionToDeleteDashboards
                                "
                                #actions
                            >
                                <button
                                    v-if="collimatoStore.hasPermissionToEditDashboards"
                                    @click="openEditDialog(dashboard)"
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-indigo-600"
                                    :title="t('common.button.edit')"
                                >
                                    <PencilIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                                <button
                                    v-if="collimatoStore.hasPermissionToDeleteDashboards"
                                    @click="openDeleteDialog(dashboard)"
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    :title="t('common.button.delete')"
                                >
                                    <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                            </template>
                            <template #footer>
                                <UserAvatarWithText
                                    class="gap-x-2"
                                    :user-id="dashboard.owner_id"
                                    avatar-class="size-6 shrink-0"
                                    text-class="truncate text-xs text-gray-500"
                                />
                                <span class="shrink-0 text-xs text-gray-400">
                                    {{ getDateAndTime(dashboard.updated_at) }}
                                </span>
                            </template>
                        </ItemCard>
                    </ul>
                </div>
            </template>
        </template>

        <!-- Create / Edit dialog -->
        <Teleport to="body">
            <TransitionRoot as="template" :show="open">
                <Dialog as="div" class="relative z-50" @close="close">
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0"
                        enter-to="opacity-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100"
                        leave-to="opacity-0"
                    >
                        <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
                    </TransitionChild>
                    <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                        <div
                            class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                        >
                            <TransitionChild
                                as="template"
                                enter="ease-out duration-300"
                                enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                                enter-to="opacity-100 translate-y-0 sm:scale-100"
                                leave="ease-in duration-200"
                                leave-from="opacity-100 translate-y-0 sm:scale-100"
                                leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                            >
                                <DialogPanel
                                    class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                                >
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                    >
                                        {{
                                            state.edit
                                                ? t(
                                                      "collimato.dashboards.new_dashboard_dialog.edit_title",
                                                  )
                                                : t(
                                                      "collimato.dashboards.new_dashboard_dialog.title",
                                                  )
                                        }}
                                    </DialogTitle>

                                    <div class="mt-4 space-y-4">
                                        <div>
                                            <label
                                                for="dashboard-title"
                                                class="block text-sm font-medium leading-6 text-gray-900"
                                            >
                                                {{
                                                    t(
                                                        "collimato.dashboards.new_dashboard_dialog.dashboard_title",
                                                    )
                                                }}
                                            </label>
                                            <div class="mt-2">
                                                <input
                                                    v-model="state.title"
                                                    type="text"
                                                    id="dashboard-title"
                                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6"
                                                    :class="
                                                        v$.state.title.$error
                                                            ? 'ring-red-600 focus:ring-red-600'
                                                            : 'ring-gray-300 focus:ring-indigo-600'
                                                    "
                                                    :placeholder="
                                                        t(
                                                            'collimato.dashboards.new_dashboard_dialog.dashboard_title',
                                                        )
                                                    "
                                                />
                                                <p
                                                    v-if="v$.state.title.$error"
                                                    class="mt-1.5 text-sm text-red-600"
                                                >
                                                    {{ t("common.error.required_field") }}
                                                </p>
                                            </div>
                                        </div>
                                        <div>
                                            <label
                                                for="dashboard-description"
                                                class="block text-sm font-medium leading-6 text-gray-900"
                                            >
                                                {{
                                                    t(
                                                        "collimato.dashboards.new_dashboard_dialog.dashboard_description",
                                                    )
                                                }}
                                            </label>
                                            <div class="mt-2">
                                                <textarea
                                                    v-model="state.description"
                                                    id="dashboard-description"
                                                    rows="3"
                                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6"
                                                    :placeholder="
                                                        t(
                                                            'collimato.dashboards.new_dashboard_dialog.dashboard_description',
                                                        )
                                                    "
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <div class="mt-5 flex flex-row-reverse gap-2">
                                        <button
                                            type="button"
                                            class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                            @click="create"
                                        >
                                            {{
                                                state.edit
                                                    ? t("common.button.save")
                                                    : t("common.button.create")
                                            }}
                                        </button>
                                        <button
                                            type="button"
                                            class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                            @click="close"
                                        >
                                            {{ t("common.button.cancel") }}
                                        </button>
                                    </div>
                                </DialogPanel>
                            </TransitionChild>
                        </div>
                    </div>
                </Dialog>
            </TransitionRoot>
        </Teleport>

        <DashboardDeleteDialog
            v-model="deleteDialog"
            @update:modelValue="deleteDialog = $event"
            @delete="deleteDashboard"
        />
    </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { PlusIcon, PencilIcon, TrashIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { Squares2X2Icon } from "@heroicons/vue/24/outline";
import collimatoService from "@/services/collimatoService.js";
import useDateOperations from "@/composables/useDateOperations.js";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useUsers } from "@/composables/useUser";
import { useCollimatoStore } from "@/store/collimato";
import DashboardDeleteDialog from "@/components/Collimato/Dialogs/DashboardDeleteDialog.vue";
import ItemCard from "@/components/ItemCard.vue";
import LetterAvatar from "@/components/LetterAvatar.vue";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";

const router = useRouter();
const route = useRoute();
const alertStore = useAlertStore();
const collimatoStore = useCollimatoStore();
const { getDateAndTime } = useDateOperations();

const loaded = ref(false);
const dashboards = ref([]);

// Trigger a lazy load so owner names resolve in the list.
useUsers(() => dashboards.value.map((d) => d.owner_id));
const open = ref(false);
const deleteDialog = ref(false);
const itemToDelete = ref(null);
const searchQuery = ref("");

const state = reactive({
    title: "",
    description: "",
    edit: false,
    editId: null,
});

const rules = { state: { title: { required } } };
const v$ = useVuelidate(rules, { state });

const filteredDashboards = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();

    if (!query) return dashboards.value;

    return dashboards.value.filter((d) => d.title.toLowerCase().includes(query));
});

function rowclick(item) {
    router.push({ name: "dashboard", params: { id: item.id } });
}

function openEditDialog(item) {
    state.title = item.title;
    state.description = item.description;
    state.edit = true;
    state.editId = item.id;
    open.value = true;
}

function openDeleteDialog(item) {
    itemToDelete.value = item;
    deleteDialog.value = true;
}

function close() {
    open.value = false;
    state.title = "";
    state.description = "";
    state.edit = false;
    state.editId = null;
    v$.value.$reset();
}

async function create() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    if (state.edit) {
        collimatoService
            .updateDashboard(route.params.workspaceId, state.editId, {
                title: state.title,
                description: state.description,
            })
            .then(() => {
                const index = dashboards.value.findIndex((d) => d.id === state.editId);

                dashboards.value[index].title = state.title;
                dashboards.value[index].description = state.description;
                close();
                alertStore.showSuccess(t.value("collimato.dashboards.success.updated"));
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });

        return;
    }

    collimatoService
        .createDashboard(route.params.workspaceId, {
            title: state.title,
            description: state.description,
        })
        .then((response) => {
            dashboards.value.push(response.data);
            close();
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });
}

function deleteDashboard() {
    collimatoService
        .deleteDashboard(route.params.workspaceId, itemToDelete.value.id)
        .then(() => {
            dashboards.value = dashboards.value.filter((d) => d.id !== itemToDelete.value.id);
            itemToDelete.value = null;
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });
    deleteDialog.value = false;
}

onMounted(() => {
    if (!collimatoStore.hasPermissionToViewDashboards) {
        alertStore.showError(t.value("collimato.dashboards.error.no_permission_view"));

        return;
    }

    collimatoService.getDashboards(route.params.workspaceId).then((response) => {
        dashboards.value = response.data;
        loaded.value = true;
    });
});
</script>
