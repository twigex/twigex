// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, reactive } from "vue";
import { useDialogStore } from "@/store/dialogs";
import { useDetailsStore } from "@/store/details";

export function useFileBrowserContextMenu({ props, selectedFiles }) {
    const dialogStore = useDialogStore();
    const detailsStore = useDetailsStore();

    const showFileContextMenu = ref(false);
    const fileContextMenuCoordinates = reactive({ x: 0, y: 0 });
    const caller = ref("file");
    const contextMenuFile = reactive({});

    function openEmptyContextMenu(event) {
        if (props.view !== "files") {
            return;
        }

        event.stopPropagation();

        fileContextMenuCoordinates.x = event.x;
        fileContextMenuCoordinates.y = event.y;

        caller.value = "empty";

        showFileContextMenu.value = true;
    }

    function openFileContextMenu(file, event) {
        if (props.view == "deleted") {
            return;
        }

        if (!selectedFiles.value.includes(file)) {
            selectedFiles.value = [file];
        }

        detailsStore.setFile(file);

        event.stopPropagation();

        fileContextMenuCoordinates.x = event.x;
        fileContextMenuCoordinates.y = event.y;

        caller.value = "file";

        Object.assign(contextMenuFile, file);

        showFileContextMenu.value = true;
    }

    function openNewDocumentDialog() {
        dialogStore.openFileDialog("document");
    }

    function openNewSpreadsheetDialog() {
        dialogStore.openFileDialog("spreadsheet");
    }

    function openNewPresentationDialog() {
        dialogStore.openFileDialog("presentation");
    }

    function openNewTextDialog() {
        dialogStore.openFileDialog("text");
    }

    function openNewFolderDialog() {
        dialogStore.openFileDialog("folder");
    }

    return {
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
    };
}
