// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, watch, onMounted, nextTick } from "vue";
import { useRoute } from "vue-router";
import { useFilesStore } from "@/store/files";
import fileService from "@/services/fileService";

export function useFileBrowserLoading({ props, state, fileItemsRefs, selectedFiles, sortedFiles }) {
    const route = useRoute();
    const filesStore = useFilesStore();

    const loaded = ref(false);
    const breadcrumbs = ref([]);

    function loadFiles() {
        fileService.getFilesFromFolder(route.params.id).then((res) => {
            filesStore.setFiles(res.data);
            loaded.value = true;

            setTimeout(() => {
                findAndHighlightFile();
            }, 500);
        });
    }

    function loadRecentFiles() {
        fileService.recentFiles().then((res) => {
            filesStore.setFiles(res.data);
            loaded.value = true;
        });
    }

    function loadSharedFiles() {
        fileService.sharedFiles().then((res) => {
            filesStore.setFiles(res.data);
            loaded.value = true;

            setTimeout(() => {
                findAndHighlightFile();
            }, 500);
        });
    }

    function loadFavoriteFiles() {
        fileService.favoriteFiles().then((res) => {
            filesStore.setFiles(res.data);
            loaded.value = true;
        });
    }

    function loadDeletedFiles() {
        fileService.deletedFiles().then((res) => {
            filesStore.setFiles(res.data);
            loaded.value = true;
        });
    }

    function loadBreadcrumbs() {
        if (props.view == "files") {
            fileService.getParent(route.params.id).then((res) => {
                breadcrumbs.value = res.data.reverse();
            });
        }
    }

    function findAndHighlightFile() {
        if (route.query.file != undefined) {
            const fileId = route.query.file;

            const fileIndex = sortedFiles.value.findIndex((file) => file.id === fileId);

            if (fileIndex !== -1) {
                const pageNumber = Math.floor(fileIndex / state.itemsPerPage) + 1;

                state.currentPage = pageNumber;

                nextTick().then(() => {
                    const file = fileItemsRefs.value.find((ref) => ref.$el.dataset.id === fileId);

                    if (file) {
                        file.$el.scrollIntoView({
                            behavior: "smooth",
                            block: "center",
                        });

                        file.$el.classList.add("bg-indigo-100");

                        setTimeout(() => {
                            file.$el.classList.remove("bg-indigo-100");
                        }, 3000);
                    }
                });
            }
        }
    }

    onMounted(() => {
        switch (props.view) {
            case "files":
                loadFiles();
                break;
            case "recents":
                loadRecentFiles();
                break;
            case "favorites":
                loadFavoriteFiles();
                break;
            case "shared":
                loadSharedFiles();
                break;
            case "deleted":
                loadDeletedFiles();
                break;

            default:
                break;
        }

        filesStore.setCurrentView({
            currentDirectory: route.params.id,
            currentName: props.view,
        });
    });

    watch(
        () => route.params.id,
        () => {
            fileService.getFilesFromFolder(route.params.id).then((res) => {
                loadBreadcrumbs();
                filesStore.setFiles(res.data);
                selectedFiles.value = [];

                filesStore.setCurrentView({
                    currentDirectory: route.params.id,
                    currentName: props.view,
                });

                setTimeout(() => {
                    findAndHighlightFile();
                }, 500);
            });
        },
    );

    onMounted(() => {
        loadBreadcrumbs();
    });

    return {
        loaded,
        breadcrumbs,
    };
}
