<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="flex flex-col h-full">
        <FilesToolbar
            @files-selected="handleFilesSelected"
            @trash="emptyTrash"
            @page-clicked="handlePageClicked"
            v-model:view-mode="viewMode"
            :pages="breadcrumbs"
        >
            <template v-slot:actions>
                <FileBrowserSelectionActions
                    :has-selection="selectedFiles.length > 0"
                    :deleting="deleting"
                    @select-all="selectedFiles = filesStore.getFiles"
                    @empty-trash="emptyTrash"
                />
            </template>
        </FilesToolbar>
        <EmptyFilesState
            v-if="filesStore.getFiles.length <= 0"
            @files-selected="handleFilesSelected"
            @drop.prevent.stop="onEmptyDrop($event, route.params.id)"
            @dragover.prevent=""
        />
        <div
            v-else
            class="relative h-full w-full overflow-hidden select-none"
            ref="container"
            @mousedown="onMouseDown"
            @click="onContainerClick"
            @drop.prevent="onEmptyDrop($event, route.params.id)"
            @dragover.prevent=""
        >
            <FileList
                @contextmenu.prevent="openEmptyContextMenu"
                :grid="viewMode === 'grid'"
                class="mt-1"
            >
                <template v-slot:default>
                    <component
                        :is="viewMode === 'grid' ? FileGridItem : FileListItem"
                        v-for="file in fileList"
                        @click.stop="onRowClick($event, file)"
                        @dblclick="openFile(file)"
                        @contextmenu.prevent="openFileContextMenu(file, $event)"
                        @pointerdown="pressFile($event, file)"
                        @dragover.prevent=""
                        @dragenter.prevent="onUploadDragEnter($event, file)"
                        @dragleave="onUploadDragLeave($event, file)"
                        @drop.prevent.stop="onUploadDrop($event, file)"
                        :key="file.id"
                        :file="file"
                        :data-id="file.id"
                        :data-move-target="file.isFolder ? file.id : undefined"
                        ref="fileItemsRefs"
                        :class="itemClass(file)"
                    >
                        <template v-slot:actions="{ item }">
                            <FileBrowserItemActions
                                :item="item"
                                :view="props.view"
                                :deleting="deleting"
                                @download="handleDownloadFile"
                                @restore="handleRestoreFile"
                                @delete-permanently="handleDeleteFilePermanently"
                            />
                        </template>
                    </component>
                </template>

                <template v-slot:pagination>
                    <BasePagination
                        class="px-2"
                        :currentPage="state.currentPage"
                        :totalItems="filesStore.getFiles.length"
                        :itemsPerPage="state.itemsPerPage"
                        @update:currentPage="updateCurrentPage"
                    />
                </template>
            </FileList>
            <div
                v-if="isSelecting"
                class="pointer-events-none absolute border border-dashed border-indigo-500 bg-indigo-500/20"
                :style="{
                    left: `${boxLeft}px`,
                    top: `${boxTop}px`,
                    width: `${boxWidth}px`,
                    height: `${boxHeight}px`,
                }"
            ></div>
        </div>

        <FileContextMenu
            class="z-50"
            :show="showFileContextMenu"
            :coordinates="fileContextMenuCoordinates"
            :caller="caller"
            :file="contextMenuFile"
            @open="openFile(contextMenuFile)"
            @details="showDetails(contextMenuFile)"
            @close="showFileContextMenu = false"
            @download="handleDownloadFile(contextMenuFile)"
            @rename="dialogStore.openRenameDialog"
            @delete="dialogStore.openDeleteDialog"
            @move="dialogStore.openMoveDialog"
            @share="dialogStore.openShareDialog(contextMenuFile)"
            @add-to-favorites="handleAddToFavorites"
            @new-folder="openNewFolderDialog"
            @document="openNewDocumentDialog"
            @spreadsheet="openNewSpreadsheetDialog"
            @presentation="openNewPresentationDialog"
            @text="openNewTextDialog"
        >
        </FileContextMenu>
        <FileCreateDialog
            :open="dialogStore.fileDialogOpen"
            :fileType="dialogStore.fileType"
            @close="dialogStore.closeFileDialog"
            @create="createFile"
        />
        <RenameDialog
            :open="dialogStore.renameDialogOpen"
            :file="contextMenuFile"
            @close="dialogStore.closeRenameDialog"
            @rename="handleRenameFile"
        />
        <DeleteDialog
            :open="dialogStore.deleteDialogOpen"
            :files="selectedFiles"
            @close="dialogStore.closeDeleteDialog"
            @delete="handleDeleteFile"
        />
        <MoveDialog
            ref="moveDialogRef"
            :open="dialogStore.moveDialogOpen"
            :files="selectedFiles"
            @close="dialogStore.closeMoveDialog"
            @move="handleMoveFile"
        />
        <ShareDialog
            v-model="dialogStore.shareDialogOpen"
            @close="dialogStore.closeShareDialog"
            @share="handleShareFile"
            @unshare="handleUnshareFile"
            @unshare-group="handleUnshareGroup"
            @update-share="handleShareFileUpdate"
            @link-share="handleLinkShare"
            @update-link="handleLinkShareUpdate"
        />
        <FileProgressCards />
    </div>
</template>

<script setup>
import FilesToolbar from "@/components/Files/Toolbar/FilesToolbar.vue";
import FileCreateDialog from "@/components/Files/Dialogs/CreateFileDialog.vue";
import RenameDialog from "@/components/Files/Dialogs/RenameDialog.vue";
import DeleteDialog from "@/components/Files/Dialogs/DeleteDialog.vue";
import MoveDialog from "@/components/Files/Dialogs/MoveDialog.vue";
import ShareDialog from "@/components/Files/Dialogs/ShareDialog/ShareDialog.vue";
import EmptyFilesState from "@/components/Files/List/EmptyFilesState.vue";
import FileList from "@/components/Files/List/FileList.vue";
import FileListItem from "@/components/Files/List/FileListItem.vue";
import FileGridItem from "@/components/Files/List/FileGridItem.vue";
import BasePagination from "@/components/BasePagination.vue";
import FileProgressCards from "@/components/Files/FileProgressCards.vue";
import FileContextMenu from "@/components/Files/FileContextMenu.vue";
import FileBrowserSelectionActions from "@/components/Files/FileBrowserSelectionActions.vue";
import FileBrowserItemActions from "@/components/Files/FileBrowserItemActions.vue";
import { ref, watch, reactive } from "vue";
import { useRoute } from "vue-router";
import { useDialogStore } from "@/store/dialogs";
import { useFilesStore } from "@/store/files";
import { useFileBrowserFileList } from "@/composables/files/useFileBrowserFileList";
import { useFileBrowserLoading } from "@/composables/files/useFileBrowserLoading";
import { useFileBrowserContextMenu } from "@/composables/files/useFileBrowserContextMenu";
import { useFileBrowserActions } from "@/composables/files/useFileBrowserActions";
import { useFileBrowserSharing } from "@/composables/files/useFileBrowserSharing";
import { useFileBrowserUpload } from "@/composables/files/useFileBrowserUpload";
import { useFileBrowserBoxSelection } from "@/composables/files/useFileBrowserBoxSelection";
import { useFileBrowserDragDrop } from "@/composables/files/useFileBrowserDragDrop";

const props = defineProps({
    view: String,
});

const route = useRoute();
const dialogStore = useDialogStore();
const filesStore = useFilesStore();

const viewMode = ref(localStorage.getItem("files_view") || "list");
const fileItemsRefs = ref([]);
const selectedFiles = ref([]);
const container = ref(null);
const moveDialogRef = ref(null);

const state = reactive({
    currentPage: 1,
    itemsPerPage: 20,
});

const { sortedFiles, fileList } = useFileBrowserFileList({ state });

const { loaded, breadcrumbs } = useFileBrowserLoading({
    props,
    state,
    fileItemsRefs,
    selectedFiles,
    sortedFiles,
});

const {
    showFileContextMenu,
    fileContextMenuCoordinates,
    caller,
    contextMenuFile,
    openEmptyContextMenu,
    openFileContextMenu,
    openNewDocumentDialog,
    openNewSpreadsheetDialog,
    openNewPresentationDialog,
    openNewTextDialog,
    openNewFolderDialog,
} = useFileBrowserContextMenu({ props, selectedFiles });

const {
    deleting,
    createFile,
    handleDownloadFile,
    openFile,
    showDetails,
    handleRenameFile,
    handleAddToFavorites,
    handleDeleteFile,
    emptyTrash,
    handleRestoreFile,
    handleDeleteFilePermanently,
    handleMoveFile,
} = useFileBrowserActions({
    state,
    selectedFiles,
    contextMenuFile,
    moveDialogRef,
    sortedFiles,
});

const {
    handleShareFile,
    handleUnshareFile,
    handleUnshareGroup,
    handleShareFileUpdate,
    handleLinkShare,
    handleLinkShareUpdate,
} = useFileBrowserSharing();

const { handleDroppedFiles, handleFilesSelected } = useFileBrowserUpload();

const { isSelecting, boxLeft, boxTop, boxWidth, boxHeight, onMouseDown, onContainerClick } =
    useFileBrowserBoxSelection({
        container,
        fileItemsRefs,
        fileList,
        selectedFiles,
    });

const {
    uploadTargetId,
    pressFile,
    isDragged,
    onUploadDragEnter,
    onUploadDragLeave,
    onUploadDrop,
    onEmptyDrop,
} = useFileBrowserDragDrop({
    props,
    selectedFiles,
    handleDroppedFiles,
});

watch(viewMode, (mode) => localStorage.setItem("files_view", mode));

function itemClass(file) {
    const selected = selectedFiles.value.includes(file);
    const uploadOver = uploadTargetId.value === file.id;
    const dragged = isDragged(file) ? "opacity-50" : "";

    if (viewMode.value === "grid") {
        return [
            selected ? "ring-1 ring-indigo-500 border-indigo-500 bg-indigo-50" : "",
            uploadOver ? "bg-indigo-50 ring-1 ring-indigo-500 border-indigo-500" : "",
            "data-[drop-target]:bg-indigo-50 data-[drop-target]:ring-1 data-[drop-target]:ring-indigo-500 data-[drop-target]:border-indigo-500",
            dragged,
        ];
    }

    let background = "hover:bg-gray-50 data-[drop-target]:bg-indigo-50";

    if (selected) {
        background = "bg-indigo-100 text-white";
    } else if (uploadOver) {
        background = "bg-indigo-50";
    }

    return [
        "group rounded",
        dragged,
        background,
        uploadOver ? "ring-1 ring-inset ring-indigo-500" : "",
        "data-[drop-target]:ring-1 data-[drop-target]:ring-inset data-[drop-target]:ring-indigo-500",
        "border-r border-l border-l-transparent border-b border-b-transparent border-r-transparent",
        fileList.value.indexOf(file) == 0 ? "border-t border-t-transparent" : "border-t",
    ];
}

function handlePageClicked(page) {
    openFile(page);
}

function updateCurrentPage(newPage) {
    state.currentPage = newPage;
}

function onRowClick(event, file) {
    // Touch has no dblclick, so below lg a tap opens; pointer devices keep click-to-select + dblclick-to-open.
    if (window.matchMedia("(max-width: 1023px)").matches) {
        openFile(file);

        return;
    }

    selectFile(event, file);
}

function selectFile(event, file) {
    if (event.shiftKey) {
        let index = fileList.value.indexOf(file);

        let firstIndex = fileList.value.indexOf(selectedFiles.value[0]);

        if (selectedFiles.value.length == 0) {
            selectedFiles.value.push(file);
        }

        selectedFiles.value = [selectedFiles.value[0]];

        if (index < firstIndex) {
            for (let i = index; i <= firstIndex; i++) {
                if (!selectedFiles.value.some((file) => file.id == fileList.value[i].id)) {
                    selectedFiles.value.push(fileList.value[i]);
                }
            }

            return;
        }

        firstIndex = fileList.value.indexOf(selectedFiles.value[selectedFiles.value.length - 1]);

        for (let i = firstIndex; i <= index; i++) {
            if (!selectedFiles.value.some((file) => file.id == fileList.value[i].id)) {
                selectedFiles.value.push(fileList.value[i]);
            }
        }

        return;
    }

    if (event.ctrlKey || event.metaKey) {
        selectedFiles.value.includes(file)
            ? selectedFiles.value.splice(selectedFiles.value.indexOf(file), 1)
            : selectedFiles.value.push(file);

        return;
    }

    if (selectedFiles.value.some((f) => f.id == file.id) && selectedFiles.value.length == 1) {
        return;
    }

    selectedFiles.value = [];
    selectedFiles.value.push(file);
}
</script>
