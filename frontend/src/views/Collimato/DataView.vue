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
                                                {{ t("collimato.builder.new_cube_visual") }}
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
                                        <MenuItem v-slot="{ active }">
                                            <button
                                                @click="
                                                    fileType = 'cube';
                                                    newFileDialog = true;
                                                "
                                                type="button"
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900'
                                                        : 'text-gray-700',
                                                    'flex w-full items-center px-4 py-2 text-sm',
                                                ]"
                                            >
                                                {{ t("collimato.data_view.new_cube") }}
                                            </button>
                                        </MenuItem>
                                        <MenuItem v-slot="{ active }">
                                            <button
                                                @click="
                                                    fileType = 'view';
                                                    newFileDialog = true;
                                                "
                                                type="button"
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900'
                                                        : 'text-gray-700',
                                                    'flex w-full items-center px-4 py-2 text-sm',
                                                ]"
                                            >
                                                {{ t("collimato.data_view.new_view") }}
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
                                    <span
                                        :class="[
                                            item.type === 'cube'
                                                ? 'bg-indigo-100 text-indigo-700'
                                                : 'bg-emerald-100 text-emerald-700',
                                            'shrink-0 rounded-full px-1.5 py-0.5 text-xs font-medium',
                                        ]"
                                        >{{ item.type === "cube" ? "dataset" : item.type }}</span
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
                    <!-- Model summary (builder-made cubes/views): a collapsible
                         pane mirroring the editor disclosure. -->
                    <ModelSummary
                        v-if="selectedFile && selectedFile.builder_model"
                        v-model:open="modelOpen"
                        :file="selectedFile"
                        :editable="collimatoStore.hasPermissionToEditDataModels"
                        class="flex flex-col border rounded-md overflow-hidden"
                        :class="modelOpen ? 'flex-1 min-h-0' : 'shrink-0'"
                        @edit="editSelectedInBuilder"
                    />

                    <!-- Editor pane (hand-written / legacy files) -->
                    <div
                        v-show="selectedFile && !selectedFile.builder_model"
                        class="flex flex-col border rounded-md overflow-hidden"
                        :class="editorOpen ? 'flex-1 min-h-0' : 'shrink-0'"
                    >
                        <div
                            class="flex items-center justify-between px-3 py-2 bg-gray-50 hover:bg-gray-100 cursor-pointer select-none shrink-0"
                            @click="editorOpen = !editorOpen"
                        >
                            <div class="flex items-center gap-x-2">
                                <ChevronRightIcon
                                    class="h-4 w-4 text-gray-400 transition-transform duration-200"
                                    :class="editorOpen ? 'rotate-90' : ''"
                                    aria-hidden="true"
                                />
                                <span
                                    class="text-xs font-semibold uppercase tracking-wide text-gray-500"
                                >
                                    {{ t("collimato.data_view.editor") }}
                                </span>
                            </div>
                            <button
                                v-if="collimatoStore.hasPermissionToEditDataModels"
                                type="button"
                                :disabled="disableSave"
                                @click.stop="save"
                                :class="[
                                    disableSave
                                        ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                                        : 'bg-indigo-600 text-white hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600',
                                    'inline-flex items-center gap-x-1.5 rounded-md px-2.5 py-1 text-xs font-semibold shadow-sm transition-colors',
                                ]"
                            >
                                <div
                                    v-if="saving"
                                    class="loader h-3 w-3 rounded-full border-2 border-white border-t-transparent animate-spin"
                                ></div>
                                {{
                                    saving
                                        ? t("collimato.data_view.button.saving")
                                        : t("common.button.save")
                                }}
                            </button>
                        </div>
                        <div v-show="editorOpen" class="flex-1 min-h-0">
                            <div ref="aceEditor" class="h-full rounded overflow-hidden"></div>
                        </div>
                    </div>

                    <!-- Data preview disclosure -->
                    <div
                        v-if="collimatoStore.canQueryData"
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
                            {{ t("collimato.builder.new_cube_visual") }}
                        </button>
                    </div>
                </div>
            </div>

            <!-- New file dialog -->
            <Teleport to="body">
                <TransitionRoot as="template" :show="newFileDialog">
                    <Dialog class="relative z-10" @close="newFileDialog = false">
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
                                        class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6"
                                    >
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold text-gray-900"
                                        >
                                            {{ newFileDialogTitle }}
                                        </DialogTitle>
                                        <div class="mt-4">
                                            <label
                                                for="new-file-name"
                                                class="block text-sm font-medium leading-6 text-gray-900"
                                            >
                                                {{ t("collimato.data_view.filename") }}
                                            </label>
                                            <div class="mt-1">
                                                <input
                                                    v-model="filename"
                                                    type="text"
                                                    id="new-file-name"
                                                    class="block w-full rounded-md bg-white px-3 py-1.5 text-sm text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
                                                    :placeholder="
                                                        t(
                                                            'collimato.data_view.filename_placeholder',
                                                        )
                                                    "
                                                    @keydown.enter="createNewFile()"
                                                />
                                            </div>
                                        </div>
                                        <div class="mt-5 flex flex-row-reverse gap-2">
                                            <button
                                                :disabled="saving || !filename.trim()"
                                                type="button"
                                                @click="createNewFile()"
                                                :class="[
                                                    saving || !filename.trim()
                                                        ? 'bg-gray-300 cursor-not-allowed'
                                                        : 'bg-indigo-600 hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600',
                                                    'inline-flex items-center gap-x-1.5 rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm',
                                                ]"
                                            >
                                                <div
                                                    v-if="saving"
                                                    class="loader h-3 w-3 rounded-full border-2 border-white border-t-transparent animate-spin"
                                                ></div>
                                                {{
                                                    saving
                                                        ? t("collimato.data_view.button.creating")
                                                        : t("common.button.create")
                                                }}
                                            </button>
                                            <button
                                                type="button"
                                                class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                                @click="newFileDialog = false"
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

            <FileDeleteDialog
                v-model="deleteDialog"
                :deleting="saving"
                :dependents="deleteDependents"
                @delete="deleteFile(deleteItem)"
            />

            <CubeBuilder
                :show="builderOpen"
                :workspace-id="route.params.workspaceId"
                :existing-cubes="existingCubes"
                :edit-file="builderEditFile"
                @close="builderOpen = false"
                @saved="handleBuilderSaved"
            />

            <ViewBuilder
                :show="viewBuilderOpen"
                :workspace-id="route.params.workspaceId"
                :existing-cubes="existingCubes"
                :edit-file="viewBuilderEditFile"
                @close="viewBuilderOpen = false"
                @saved="handleBuilderSaved"
            />
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { t } from "@/i18n/index.js";
import { useRoute } from "vue-router";
import FileDeleteDialog from "@/components/Collimato/Dialogs/FileDeleteDialog.vue";
import Table from "@/components/Collimato/ResultTable.vue";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
} from "@headlessui/vue";
import { DocumentIcon, TrashIcon, CircleStackIcon } from "@heroicons/vue/24/outline";
import { PlusIcon, ChevronRightIcon } from "@heroicons/vue/20/solid";
import CubeBuilder from "@/components/Collimato/Builder/CubeBuilder.vue";
import ViewBuilder from "@/components/Collimato/Builder/ViewBuilder.vue";
import ModelSummary from "@/components/Collimato/Builder/ModelSummary.vue";
import { toTableData, MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";
import collimatoService from "@/services/collimatoService.js";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import ace from "ace-builds";
import "ace-builds/src-noconflict/mode-javascript";
import "ace-builds/src-noconflict/theme-dracula";
import "ace-builds/src-noconflict/ext-language_tools";

const loaded = ref(false);
const loading = ref(false);
const collimatoStore = useCollimatoStore();
const alertStore = useAlertStore();
const route = useRoute();
const aceEditor = ref(null);
let editor = null;
const files = ref([]);
const selectedFile = ref(null);
const disableSave = ref(true);
const saving = ref(false);
const deleteDialog = ref(false);
const deleteItem = ref(null);

// deleteDependents lists the builder-made cubes/views that reference the cube
// pending deletion, so the dialog can block a delete that would break Cube
// compilation (mirrors the backend guard).
const deleteDependents = computed(() => {
    const item = deleteItem.value;

    if (!item || item.type !== "cube") return [];
    const cubeName = item.name.replace(/\.(ya?ml|js)$/, "");
    const deps = [];

    for (const f of files.value) {
        if (f === item || !f.builder_model) continue;
        let m;

        try {
            m = JSON.parse(f.builder_model);
        } catch {
            continue;
        }

        if (f.type === "cube") {
            if ((m.joins || []).some((j) => j.name === cubeName)) deps.push(f.name);
        } else if (f.type === "view") {
            if ((m.cubes || []).some((c) => (c.join_path || "").split(".").includes(cubeName)))
                deps.push(f.name);
        }
    }

    return deps;
});
const newFileDialog = ref(false);
const filename = ref("");
const fileType = ref("");
const tableData = ref({});
const tableOpen = ref(false);
const editorOpen = ref(true);
const modelOpen = ref(true);

watch(editorOpen, (open) => {
    if (open && editor) nextTick(() => editor.resize());
});

// primaryOpen tracks whichever top pane is active for the selected file (the
// model summary for builder files, the code editor for legacy ones), so the
// data preview below expands when that pane is collapsed.
const primaryOpen = computed(() =>
    selectedFile.value?.builder_model ? modelOpen.value : editorOpen.value,
);

const newFileDialogTitle = computed(() =>
    fileType.value === "cube"
        ? t.value("collimato.data_view.new_cube")
        : t.value("collimato.data_view.new_view"),
);

let loadDataTimeout = null;
let loadAbort = null;

onUnmounted(() => {
    clearTimeout(loadDataTimeout);
    if (loadAbort) loadAbort.abort();
});

// cancelLoad stops the in-flight request and the retry polling, and clears the
// spinner. The DB query keeps running server-side until CUBEJS_DB_QUERY_TIMEOUT.
function cancelLoad() {
    clearTimeout(loadDataTimeout);
    if (loadAbort) loadAbort.abort();
    loading.value = false;
}

onMounted(() => {
    if (!collimatoStore.hasPermissionToViewDataModels) {
        alertStore.showError(t.value("collimato.data_view.error.no_permission_to_view_data"));

        return;
    }

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

// ensureEditor lazily creates the Ace instance the first time a code-editable
// (legacy) file is shown; builder-made files never mount it. Idempotent.
function ensureEditor() {
    if (editor) return true;
    if (!aceEditor.value) return false;
    editor = ace.edit(aceEditor.value);
    editor.setTheme("ace/theme/dracula");
    editor.getSession().setMode("ace/mode/javascript");
    editor.setFontSize(14);
    editor.setOptions({
        enableBasicAutocompletion: true,
        enableSnippets: true,
        enableLiveAutocompletion: true,
        useWorker: false,
    });
    editor.on("change", handleEditorChange);

    return true;
}

const builderOpen = ref(false);
const builderEditFile = ref(null);
const viewBuilderOpen = ref(false);
const viewBuilderEditFile = ref(null);

// Existing cubes offered as join targets in the builder, each with its parsed
// structured model (when built visually) so the builder can resolve the target's
// columns for the join dropdown.
const existingCubes = computed(() =>
    files.value
        .filter((f) => f.type === "cube")
        .map((f) => {
            let model;

            try {
                model = f.builder_model ? JSON.parse(f.builder_model) : null;
            } catch {
                model = null;
            }

            return { name: f.name.replace(/\.(yml|js)$/, ""), model };
        }),
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
    // Selecting a file only loads it. Builder-made files render as a read-only
    // model summary (edited via the visual builder); hand-written/legacy files
    // load into the Ace editor.
    selectedFile.value = file;
    disableSave.value = true;
    tableData.value = {};
    if (file.builder_model) return;

    nextTick(() => {
        if (!ensureEditor()) return;
        editor.setValue(file.content, 1);
        editor.clearSelection();
        editor.setReadOnly(!collimatoStore.hasPermissionToEditDataModels);
        editor.resize();
    });
}

function editSelectedInBuilder() {
    const file = selectedFile.value;

    if (!file) return;
    if (file.type === "view") openViewBuilder(file);
    else openBuilder(file);
}

function createNewFile() {
    if (!filename.value.trim()) return;
    saving.value = true;
    collimatoService
        .newCubeFile(route.params.workspaceId, {
            filetype: fileType.value,
            filename: filename.value,
        })
        .then((response) => {
            saving.value = false;
            newFileDialog.value = false;
            filename.value = "";
            files.value.push(response.data);
            selectedFile.value = files.value[files.value.length - 1];
            switchFile(files.value[files.value.length - 1]);
        })
        .catch((error) => {
            saving.value = false;
            newFileDialog.value = false;
            alertStore.showError(extractErrorMessage(error));
        });
}

function save() {
    saving.value = true;
    disableSave.value = true;

    collimatoService
        .saveFile(route.params.workspaceId, {
            name: selectedFile.value.name,
            content: editor.getValue(),
            type: selectedFile.value.type,
        })
        .then((response) => {
            for (const file of files.value) {
                if (file.name === selectedFile.value.name) {
                    file.content = response.data.content;
                    break;
                }
            }

            disableSave.value = true;
            saving.value = false;
            alertStore.showSuccess(t.value("collimato.data_view.success.file_saved"));
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
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
        })
        .catch((error) => {
            deleteDialog.value = false;
            saving.value = false;
            alertStore.showError(extractErrorMessage(error));
        });
}

// legacyQuery regex-parses a hand-written JS cube (all dimensions + count) for
// files that have no structured builder model.
function legacyQuery(content) {
    const tableMatch = content.match(/sql_table:\s*`(\w+)`/);
    const sqlTable = tableMatch ? tableMatch[1] : null;
    const dimensionBlockMatch = content.match(/dimensions:\s*{([\s\S]*?)},\n\s*measures:/);
    const dimensions = [];

    if (sqlTable && dimensionBlockMatch) {
        for (const m of dimensionBlockMatch[1].matchAll(/^\s*(\w+):\s*{/gm)) {
            dimensions.push(`${sqlTable}.${m[1]}`);
        }
    }

    return { dimensions, measures: sqlTable ? [`${sqlTable}.count`] : [] };
}

// buildPreviewQuery resolves a cube/view's queryable members from Cube meta
// (fully-qualified names, so views and prefixes resolve correctly). YAML files
// can't be regex-parsed like the old JS format, so builder files rely on this.
async function buildPreviewQuery() {
    const file = selectedFile.value;
    const base = {
        order: [],
        filters: [],
        timeDimensions: [],
        limit: MAX_QUERY_LIMIT,
    };

    let entityName = null;

    if (file.builder_model) {
        try {
            entityName = JSON.parse(file.builder_model)?.name || null;
        } catch {
            entityName = null;
        }
    } else {
        const m = file.content.match(/sql_table:\s*`(\w+)`/);

        entityName = m ? m[1] : null;
    }

    if (entityName) {
        const res = await collimatoService.meta(route.params.workspaceId);

        if (!res.data?.error && Array.isArray(res.data?.cubes)) {
            const entry = res.data.cubes.find((c) => c.name === entityName);

            if (entry) {
                return {
                    ...base,
                    dimensions: (entry.dimensions || []).map((d) => d.name),
                    measures: (entry.measures || []).map((m) => m.name),
                };
            }
        }
    }

    return { ...base, ...legacyQuery(file.content) };
}

async function loadData() {
    loading.value = true;
    loadAbort = new AbortController();

    let query;

    try {
        query = await buildPreviewQuery();
    } catch (error) {
        loading.value = false;
        alertStore.showError(extractErrorMessage(error));

        return;
    }

    if (!query.dimensions.length && !query.measures.length) {
        loading.value = false;
        alertStore.showError(t.value("collimato.data_view.error.load_data_failed"));

        return;
    }

    runQuery(query, 0);
}

function runQuery(query, retryCount) {
    const MAX_RETRIES = 30;

    collimatoService
        .loadData(route.params.workspaceId, query, loadAbort?.signal)
        .then((result) => {
            if (result.data.error === "Continue wait") {
                if (retryCount >= MAX_RETRIES) {
                    loading.value = false;
                    alertStore.showError(t.value("collimato.data_view.error.load_data_failed"));

                    return;
                }

                loadDataTimeout = setTimeout(() => runQuery(query, retryCount + 1), 2000);

                return;
            }

            // Cube reports real query errors in a 200 body via an "error" field;
            // surface it and stop the spinner instead of trying to render it.
            if (result.data.error !== undefined) {
                loading.value = false;
                alertStore.showError(
                    result.data.error || t.value("collimato.data_view.error.load_data_failed"),
                );

                return;
            }

            tableData.value = toTableData(result.data);
            loading.value = false;
            tableOpen.value = true;
        })
        .catch((error) => {
            if (error?.code === "ERR_CANCELED") return; // cancelled by the user
            loading.value = false;
            alertStore.showError(extractErrorMessage(error));
        });
}

function handleEditorChange() {
    if (!editor || !selectedFile.value) return;
    disableSave.value = editor.getValue() === selectedFile.value.content;
}
</script>
