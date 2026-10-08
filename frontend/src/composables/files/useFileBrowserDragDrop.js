// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { usePointerDrag } from "@/composables/usePointerDrag";
import { useMoveFiles } from "@/composables/files/useMoveFiles";

export function useFileBrowserDragDrop({ props, selectedFiles, handleDroppedFiles }) {
    const route = useRoute();
    const { dragged, startPress } = usePointerDrag();
    const { moveFiles } = useMoveFiles();

    const uploadTargetId = ref(null);

    function startDrag(file) {
        if (!selectedFiles.value.some((f) => f.id === file.id)) {
            selectedFiles.value = [file];
        }

        const files = [...selectedFiles.value];

        return {
            item: files,
            label: files[0]?.name || t.value("files.drag.no_name"),
            count: files.length,
        };
    }

    function resolveTarget(id, files) {
        if (id === route.params.id) return null;
        if (files.some((file) => file.id === id)) return null;

        return { id };
    }

    async function dropFiles(files, target) {
        if (await moveFiles(files, target.id)) {
            selectedFiles.value = [];
        }
    }

    function pressFile(event, file) {
        if (props.view != "files") return;
        if (!event.target.closest("[data-drag-handle]")) return;

        startPress(event, {
            start: () => startDrag(file),
            resolveTarget,
            drop: dropFiles,
        });
    }

    function isDragged(file) {
        return dragged.value?.some((f) => f.id === file.id) ?? false;
    }

    function hasDesktopFiles(event) {
        return event.dataTransfer?.types.includes("Files");
    }

    function onUploadDragEnter(event, file) {
        if (file.isFolder && hasDesktopFiles(event)) {
            uploadTargetId.value = file.id;
        }
    }

    function onUploadDragLeave(event, file) {
        if (uploadTargetId.value === file.id) {
            uploadTargetId.value = null;
        }
    }

    function onUploadDrop(event, file) {
        uploadTargetId.value = null;

        if (event.dataTransfer.files.length == 0) return;

        handleDroppedFiles(event, file.isFolder ? file.id : route.params.id);
    }

    function onEmptyDrop(event, id) {
        uploadTargetId.value = null;
        handleDroppedFiles(event, id);
    }

    return {
        uploadTargetId,
        pressFile,
        isDragged,
        onUploadDragEnter,
        onUploadDragLeave,
        onUploadDrop,
        onEmptyDrop,
    };
}
