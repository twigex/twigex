// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { useFilesStore } from "@/store/files";

export function useFileBrowserFileList({ state }) {
    const filesStore = useFilesStore();

    const sortedFiles = computed(() => {
        const files = [...filesStore.getFiles];

        if (filesStore.fileOrder == null || filesStore.fileOrder.order == "none") {
            sortDesc(files, "date");
        } else if (filesStore.fileOrder.order == "asc") {
            sortAsc(files, filesStore.fileOrder.type);
        } else {
            sortDesc(files, filesStore.fileOrder.type);
        }

        return files;
    });

    const fileList = computed(() =>
        sortedFiles.value.slice(
            (state.currentPage - 1) * state.itemsPerPage,
            state.currentPage * state.itemsPerPage,
        ),
    );

    function sortDesc(arr, type) {
        switch (type) {
            case "date":
                arr.sort((a, b) => new Date(b.created) - new Date(a.created));
                break;
            case "modified":
                arr.sort((a, b) => new Date(b.modified) - new Date(a.modified));
                break;
            case "size":
                arr.sort((a, b) => b.size - a.size);
                break;
            case "name":
                arr.sort((a, b) => b.name.localeCompare(a.name));
                break;
            case "type":
                arr.sort((a, b) => b.type.localeCompare(a.type));
                break;
            default:
                break;
        }
    }

    function sortAsc(arr, type) {
        switch (type) {
            case "date":
                arr.sort((a, b) => new Date(a.created) - new Date(b.created));
                break;
            case "modified":
                arr.sort((a, b) => new Date(a.modified) - new Date(b.modified));
                break;
            case "size":
                arr.sort((a, b) => a.size - b.size);
                break;
            case "name":
                arr.sort((a, b) => a.name.localeCompare(b.name));
                break;
            case "type":
                arr.sort((a, b) => a.type.localeCompare(b.type));
                break;
            default:
                break;
        }
    }

    return {
        sortedFiles,
        fileList,
    };
}
