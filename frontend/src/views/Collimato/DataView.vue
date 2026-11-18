<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full h-full flex flex-col items-center justify-center">
        <div class="flex w-full h-full">
            <!-- Loading state -->
            <div v-if="!loaded" class="flex items-center justify-center h-full w-full">
                <div
                    class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
                ></div>
            </div>

            <!-- Main layout -->
            <div v-else-if="files.length > 0" class="flex-grow overflow-hidden flex">
                <!-- Left sidebar -->
                <div class="flex w-64 shrink-0 flex-col border-r overflow-hidden">
                    <div class="flex items-center justify-between px-3 py-2 border-b">
                        <span class="text-xs font-semibold uppercase tracking-wide text-gray-400">
                            {{ t("collimato.data_view.data_models") }}
                        </span>
                        <Menu
                            v-if="collimatoStore.hasPermissionToCreateDataModels"
                            as="div"
                            class="relative"
                        >
                            <MenuButton
                                class="flex items-center rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                            >
                                <PlusIcon class="h-4 w-4" aria-hidden="true" />
                            </MenuButton>
                            <transition
                                enter-active-class="transition ease-out duration-100"
                                enter-from-class="transform opacity-0 scale-95"
                                enter-to-class="transform opacity-100 scale-100"
                                leave-active-class="transition ease-in duration-75"
                                leave-from-class="transform opacity-100 scale-100"
                                leave-to-class="transform opacity-0 scale-95"
                            >
                                <MenuItems
                                    class="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                                >
                                    <div class="py-1">
                                        <MenuItem v-slot="{ active }">
                                            <button
                                                @click="openBuilder(null)"
                                                type="button"
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900'
                                                        : 'text-gray-700',
                                                    'flex w-full items-center px-4 py-2 text-sm font-medium',
                                                ]"
                                            >
                                                {{ t("collimato.builder.new_dataset_visual") }}
                                            </button>
                                        </MenuItem>
                                        <MenuItem v-slot="{ active }">
                                            <button
                                                @click="openViewBuilder(null)"
                                                type="button"
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900'
                                                        : 'text-gray-700',
                                                    'flex w-full items-center px-4 py-2 text-sm font-medium',
                                                ]"
                                            >
                                                {{ t("collimato.view_builder.new_view_visual") }}
                                            </button>
                                        </MenuItem>
                                    </div>
                                </MenuItems>
                            </transition>
                        </Menu>
                    </div>

                    <nav class="flex-1 overflow-y-auto px-2 py-2">
                        <ul role="list" class="space-y-0.5">
                            <li v-for="item in files" :key="item.name" class="group relative">
                                <button
                                    @click="switchFile(item)"
                                    type="button"
                                    :class="[
                                        item === selectedFile
                                            ? 'bg-indigo-50 text-indigo-600'
                                            : 'text-gray-700 hover:bg-gray-50 hover:text-indigo-600',
                                        'w-full flex items-center gap-x-2 rounded-md px-2 py-1.5 pr-8 text-sm font-medium text-left',
                                    ]"
                                >
                                    <span class="flex-1 truncate">{{ item.name }}</span>
                                    <ExclamationTriangleIcon
                                        v-if="warningOf(item)"
                                        class="h-4 w-4 shrink-0 text-amber-500"
                                        :title="warningOf(item)"
                                        aria-hidden="true"
                                    />
                                    <span
                                        :class="[
                                            item.type === 'dataset'
                                                ? 'bg-indigo-100 text-indigo-700'
                                                : 'bg-emerald-100 text-emerald-700',
                                            'shrink-0 rounded-full px-1.5 py-0.5 text-xs font-medium',
                                        ]"
                                        >{{ item.type }}</span
                                    >
                                </button>
                                <button
                                    v-if="collimatoStore.hasPermissionToDeleteDataModels"
                                    @click="
                                        deleteDialog = true;
                                        deleteItem = item;
                                    "
                                    type="button"
                                    class="absolute right-1 top-1/2 -translate-y-1/2 invisible group-hover:visible rounded p-0.5 text-red-400 hover:bg-red-100 hover:text-red-600"
                                >
                                    <TrashIcon class="h-3.5 w-3.5" aria-hidden="true" />
                                </button>
                            </li>
                        </ul>
                    </nav>

                    <div v-if="collimatoStore.hasPermissionToCreateDataModels" class="border-t p-2">
                        <button
                            type="button"
                            @click="openBuilder(null)"
                            class="inline-flex w-full items-center justify-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        >
                            {{ t("collimato.data_view.button.new_data_model") }}
                        </button>
                    </div>
                </div>

                <!-- Main content -->
                <div class="flex flex-col flex-1 min-w-0 overflow-hidden p-2 gap-2">
                    <!-- Model summary (datasets and views made in the builder): a collapsible
                         pane mirroring the editor disclosure. -->
                    <ModelSummary
                        v-if="selectedFile && selectedFile.builder_model"
                        v-model:open="modelOpen"
                        :file="selectedFile"
                        :editable="collimatoStore.hasPermissionToEditDataModels"
                        :model-error="modelErrorOf(selectedFile)"
                        :name-problems="selectedNameProblems"
                        class="flex flex-col border rounded-md overflow-hidden"
                        :class="modelOpen ? 'flex-1 min-h-0' : 'shrink-0'"
                        @edit="editSelectedInBuilder"
                        @fix-names="useDatabaseNames"
                    />

                    <!-- An old model written by hand: shown read-only, to copy from
                         when making it again in the builder. -->
                    <div
                        v-if="selectedFile && !selectedFile.builder_model"
                        class="flex flex-col flex-1 min-h-0 border rounded-md overflow-hidden"
                    >
                        <div class="shrink-0 border-b bg-amber-50 px-3 py-2 text-sm text-amber-800">
                            {{ modelErrorOf(selectedFile)?.message }}
                        </div>
                        <pre
                            class="min-h-0 flex-1 overflow-auto bg-gray-900 p-3 font-mono text-xs text-gray-100"
                            >{{ selectedFile.content }}</pre>
                    </div>

                    <!-- Data preview disclosure -->
                    <div
                        v-if="collimatoStore.canQueryData && selectedFile?.builder_model"
                        class="flex flex-col border rounded-md overflow-hidden"
                        :class="!primaryOpen && tableOpen ? 'flex-1 min-h-0' : 'shrink-0'"
                    >
                        <div
                            class="flex items-center justify-between px-3 py-2 bg-gray-50 hover:bg-gray-100 cursor-pointer select-none"
                            @click="tableOpen = !tableOpen"
                        >
                            <div class="flex items-center gap-x-2">
                                <ChevronRightIcon
                                    class="h-4 w-4 text-gray-400 transition-transform duration-200"
                                    :class="tableOpen ? 'rotate-90' : ''"
                                    aria-hidden="true"
                                />
                                <span
                                    class="text-xs font-semibold uppercase tracking-wide text-gray-500"
                                >
                                    {{ t("collimato.data_view.data_preview") }}
                                </span>
                                <span
                                    v-if="tableData.data?.length > 0"
                                    class="rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700"
                                >
                                    {{ tableData.data.length }}
                                </span>
                            </div>
                            <button
                                v-if="loading"
                                type="button"
                                @click.stop="cancelLoad()"
                                :title="t('common.button.cancel')"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-slate-600 px-2.5 py-1 text-xs font-semibold text-white transition-colors hover:bg-slate-700"
                            >
                                <div
                                    class="h-3 w-3 rounded-full border-2 border-white border-t-transparent animate-spin"
                                ></div>
                                {{ t("collimato.query.running") }}
                                <svg
                                    class="h-3.5 w-3.5 opacity-75"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                    aria-hidden="true"
                                >
                                    <path
                                        d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z"
                                    />
                                </svg>
                            </button>
                            <button
                                v-else
                                type="button"
                                @click.stop="loadData()"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-2.5 py-1 text-xs font-semibold text-white transition-colors hover:bg-indigo-500"
                            >
                                {{ t("collimato.data_view.button.load_data") }}
                            </button>
                        </div>
                        <div
                            v-show="tableOpen"
                            class="border-t overflow-auto"
                            :class="primaryOpen ? 'max-h-64' : 'flex-1 min-h-0'"
                        >
                            <Table
                                v-if="tableData.headers && tableData.headers.length > 0"
                                class="px-1 py-1"
                                :headers="tableData.headers"
                                :items="tableData.data"
                                :dense="true"
                            />
                            <div
                                v-else
                                class="flex flex-col items-center justify-center gap-2 py-6"
                            >
                                <CircleStackIcon class="h-8 w-8 text-gray-300" aria-hidden="true" />
                                <p class="text-sm text-gray-500">
                                    {{ t("collimato.data_view.button.no_results") }}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Empty state -->
            <div v-else class="flex flex-col items-center justify-center h-full w-full">
                <div class="text-center">
                    <DocumentIcon class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                    <h3 class="mt-3 text-sm font-semibold text-gray-900">
                        {{ t("collimato.data_view.no_files") }}
                    </h3>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("collimato.data_view.no_files_description") }}
                    </p>
                    <div v-if="collimatoStore.hasPermissionToCreateDataModels" class="mt-6">
                        <button
                            @click="openBuilder(null)"
                            type="button"
                            class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        >
                            <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                            {{ t("collimato.builder.new_dataset_visual") }}
                        </button>
                    </div>
                </div>
            </div>

            <FileDeleteDialog
                v-model="deleteDialog"
                :deleting="saving"
                :dependents="deleteDependents"
                @delete="deleteFile(deleteItem)"
            />

            <DatasetBuilder
                :show="builderOpen"
                :workspace-id="route.params.workspaceId"
                :existing-datasets="existingDatasets"
                :edit-file="builderEditFile"
                @close="builderOpen = false"
                @saved="handleBuilderSaved"
            />

            <ViewBuilder
                :show="viewBuilderOpen"
                :workspace-id="route.params.workspaceId"
                :existing-datasets="existingDatasets"
                :edit-file="viewBuilderEditFile"
                @close="viewBuilderOpen = false"
                @saved="handleBuilderSaved"
            />
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { t } from "@/i18n/index.js";
import { useRoute } from "vue-router";
import FileDeleteDialog from "@/components/Collimato/Dialogs/FileDeleteDialog.vue";
import Table from "@/components/Collimato/ResultTable.vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import {
    DocumentIcon,
    TrashIcon,
    CircleStackIcon,
    ExclamationTriangleIcon,
} from "@heroicons/vue/24/outline";
import { PlusIcon, ChevronRightIcon } from "@heroicons/vue/20/solid";
import DatasetBuilder from "@/components/Collimato/Builder/DatasetBuilder.vue";
import ViewBuilder from "@/components/Collimato/Builder/ViewBuilder.vue";
import ModelSummary from "@/components/Collimato/Builder/ModelSummary.vue";
import { toTableData } from "@/utils/collimato/tableData.js";
import { fetchChartData } from "@/utils/collimato/chartData.js";
import { modelNameProblems, withDatabaseNames } from "@/utils/collimato/modelNames.js";
import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartQuery.js";
import collimatoService from "@/services/collimatoService.js";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const loaded = ref(false);
const loading = ref(false);
const collimatoStore = useCollimatoStore();
const alertStore = useAlertStore();
const route = useRoute();
const files = ref([]);
const selectedFile = ref(null);
const saving = ref(false);
const deleteDialog = ref(false);
const deleteItem = ref(null);
const tableData = ref({});
const tableOpen = ref(false);
const modelOpen = ref(true);

// The field list says which models cannot be used, and which fields the
// preview can show.
const fieldList = ref({ datasets: [], modelErrors: [] });

// The database's tables, to check the names in saved models. Loaded only for
// people who can edit models, who can fix them.
const databases = ref([]);

// A saved model's name without the extension an old one written by hand has.
function dataModelName(fileName) {
    return fileName.replace(/\.(ya?ml|js)$/, "");
}

function parsedModel(file) {
    try {
        return file?.builder_model ? JSON.parse(file.builder_model) : null;
    } catch {
        return null;
    }
}

// modelErrorOf finds why a model cannot be used: by its file name for a file
// written by hand or one that cannot be read, by its own name for one with
// errors.
function modelErrorOf(file) {
    const ownName = parsedModel(file)?.name;

    return (fieldList.value.modelErrors ?? []).find(
        (modelError) => modelError.name === file.name || modelError.name === ownName,
    );
}

function nameProblemsOf(file) {
    const model = parsedModel(file);

    if (!model || file.type !== "dataset") return [];

    const database = databases.value.find((d) => d.database === model.connection_id);

    if (!database) return [];

    return modelNameProblems(model, database.tables, database.type);
}

const selectedNameProblems = computed(() => nameProblemsOf(selectedFile.value));

// warningOf says why a model in the list is marked: it cannot be used, or
// some of its names are not in the database as written.
function warningOf(file) {
    const modelError = modelErrorOf(file);

    if (modelError) {
        return modelError.message;
    }

    if (nameProblemsOf(file).length > 0) {
        return t.value("collimato.data_view.name_problem_mark");
    }

    return null;
}

// deleteDependents lists the builder models that use the dataset pending
// deletion, so the dialog can refuse a delete that would leave them broken
// (as the server does).
const deleteDependents = computed(() => {
    const item = deleteItem.value;

    if (!item || item.type !== "dataset") return [];
    const datasetName = dataModelName(item.name);
    const dependents = [];

    for (const file of files.value) {
        if (file === item || !file.builder_model) continue;
        const model = parsedModel(file);

        if (!model) continue;

        if (file.type === "dataset") {
            if ((model.joins || []).some((j) => j.name === datasetName)) {
                dependents.push(file.name);
            }
        } else if (file.type === "view") {
            const usesIt = (model.parts || []).some((part) =>
                (part.join_path || "").split(".").includes(datasetName),
            );

            if (usesIt) dependents.push(file.name);
        }
    }

    return dependents;
});

// primaryOpen is whether the model summary above the data preview is open,
// so the preview takes the room when it is closed.
const primaryOpen = computed(() => modelOpen.value);

let loadAbort = null;

onUnmounted(() => {
    if (loadAbort) loadAbort.abort();
});

// cancelLoad stops the request, which also stops the query in the database.
function cancelLoad() {
    if (loadAbort) loadAbort.abort();
    loading.value = false;
}

function loadFieldList() {
    return collimatoService
        .fieldList(route.params.workspaceId)
        .then((response) => {
            fieldList.value = response.data;
        })
        .catch((error) => alertStore.showError(extractErrorMessage(error)));
}

function loadDatabases() {
    if (!collimatoStore.hasPermissionToEditDataModels) return;

    collimatoService
        .tables(route.params.workspaceId)
        .then((response) => {
            databases.value = response.data || [];
        })
        .catch(() => {
            databases.value = [];
        });
}

onMounted(() => {
    if (!collimatoStore.hasPermissionToViewDataModels) {
        alertStore.showError(t.value("collimato.data_view.error.no_permission_to_view_data"));

        return;
    }

    loadFieldList();
    loadDatabases();

    collimatoService
        .workspaceFiles(route.params.workspaceId)
        .then((response) => {
            files.value = response.data;
            if (files.value.length > 0) {
                switchFile(files.value[0]);
            }

            loaded.value = true;
        })
        .catch((error) => {
            loaded.value = true; // never leave the page stuck on the spinner
            alertStore.showError(extractErrorMessage(error));
        });
});

const builderOpen = ref(false);
const builderEditFile = ref(null);
const viewBuilderOpen = ref(false);
const viewBuilderEditFile = ref(null);

// Existing datasets offered as join targets in the builder, each with its
// builder model when it has one, so the builder can list the target's columns.
const existingDatasets = computed(() =>
    files.value
        .filter((file) => file.type === "dataset")
        .map((file) => ({ name: dataModelName(file.name), model: parsedModel(file) })),
);

function openBuilder(file) {
    builderEditFile.value = file;
    builderOpen.value = true;
}

function openViewBuilder(file) {
    viewBuilderEditFile.value = file;
    viewBuilderOpen.value = true;
}

function handleBuilderSaved(saved) {
    loadFieldList();

    collimatoService
        .workspaceFiles(route.params.workspaceId)
        .then((response) => {
            files.value = response.data;
            // Re-select the just-saved file so its summary and preview refresh.
            const name = saved?.name || selectedFile.value?.name;
            const updated = files.value.find((f) => f.name === name);

            if (updated) switchFile(updated);
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });
}

function switchFile(file) {
    selectedFile.value = file;
    tableData.value = {};
}

function editSelectedInBuilder() {
    const file = selectedFile.value;

    if (!file) return;
    if (file.type === "view") openViewBuilder(file);
    else openBuilder(file);
}

// useDatabaseNames saves the selected dataset with the names the database has
// in another case, such as orderdate for OrderDate.
function useDatabaseNames() {
    const model = parsedModel(selectedFile.value);
    const fixed = withDatabaseNames(model, selectedNameProblems.value);

    saving.value = true;

    collimatoService
        .build(route.params.workspaceId, { model: fixed, overwrite: true })
        .then((response) => handleBuilderSaved(response.data))
        .catch((error) => alertStore.showError(extractErrorMessage(error)))
        .finally(() => {
            saving.value = false;
        });
}

function deleteFile(file) {
    saving.value = true;

    collimatoService
        .deleteFile(route.params.workspaceId, {
            name: file.name,
            filetype: file.type,
        })
        .then(() => {
            files.value = files.value.filter((f) => f.name !== file.name);

            if (files.value.length > 0) {
                switchFile(files.value[0]);
            } else {
                selectedFile.value = null;
                tableData.value = {};
            }

            deleteDialog.value = false;
            saving.value = false;
            loadFieldList();
        })
        .catch((error) => {
            deleteDialog.value = false;
            saving.value = false;
            alertStore.showError(extractErrorMessage(error));
        });
}

// previewQuery asks for every field of the selected dataset or view, as the
// field list names them.
function previewQuery() {
    const name = parsedModel(selectedFile.value)?.name;
    const dataset = fieldList.value.datasets.find((d) => d.name === name);

    if (!dataset) return null;

    const fieldsOfKind = (kind) =>
        dataset.fields.filter((field) => field.kind === kind).map((field) => field.key);

    return {
        dimensions: fieldsOfKind("dimension"),
        measures: fieldsOfKind("measure"),
        order: [],
        filters: [],
        timeDimensions: [],
        limit: MAX_QUERY_LIMIT,
    };
}

async function loadData() {
    const query = previewQuery();

    if (!query || (!query.dimensions.length && !query.measures.length)) {
        alertStore.showError(
            modelErrorOf(selectedFile.value)?.message ??
                t.value("collimato.data_view.error.load_data_failed"),
        );

        return;
    }

    loading.value = true;
    loadAbort = new AbortController();

    try {
        const data = await fetchChartData(
            (q, signal) => collimatoService.loadData(route.params.workspaceId, q, signal),
            query,
            { signal: loadAbort.signal },
        );

        tableData.value = toTableData(data);
        tableOpen.value = true;
    } catch (error) {
        const cancelled = error?.code === "ERR_CANCELED" || error?.name === "AbortError";

        if (cancelled) return;
        alertStore.showError(extractErrorMessage(error));
    } finally {
        loading.value = false;
    }
}
</script>
