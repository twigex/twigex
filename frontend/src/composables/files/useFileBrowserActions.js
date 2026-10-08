// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useFilesStore } from "@/store/files";
import { useOfficeStore } from "@/store/office";
import { useDetailsStore } from "@/store/details";
import { useAlertStore } from "@/store/alerts";
import { useMediaStore } from "@/store/media";
import { extractErrorMessage } from "@/utils/errors";
import fileService from "@/services/fileService";
import officeService from "@/services/officeService";
import useFileOperations from "@/composables/files/useFileOperations";
import { useMoveFiles } from "@/composables/files/useMoveFiles";

export function useFileBrowserActions({
    state,
    selectedFiles,
    contextMenuFile,
    moveDialogRef,
    sortedFiles,
}) {
    const route = useRoute();
    const router = useRouter();
    const filesStore = useFilesStore();
    const officeStore = useOfficeStore();
    const detailsStore = useDetailsStore();
    const mediaStore = useMediaStore();
    const { renameFile, addToFavorite, deleteFile, downloadFile } = useFileOperations();
    const { moveFiles } = useMoveFiles();

    const deleting = ref(false);

    function createFile({ name, type }) {
        let extension = "";

        switch (type) {
            case "document":
                extension = "docx";
                break;
            case "spreadsheet":
                extension = "xlsx";
                break;
            case "presentation":
                extension = "pptx";
                break;
            case "text":
                extension = "txt";
                break;
            default:
                break;
        }

        if (type !== "folder") {
            fileService
                .createFile({
                    id: route.params.id,
                    docName: name,
                    type: type,
                    extension: extension,
                })
                .then((response) => {
                    filesStore.addFile(response.data);
                    selectedFiles.value = [];
                });

            return;
        }

        fileService
            .createFolder({
                id: route.params.id,
                name: name,
            })
            .then((response) => {
                filesStore.addFile(response.data);
                selectedFiles.value = [];
            });
    }

    function handleDownloadFile(file) {
        downloadFile(fileService, file);
    }

    function openFile(file) {
        if (file.isFolder || file.type == "cloud#drive") {
            router.push({ path: `/files/${file.id}` });

            state.currentPage = 1;

            return;
        }

        if (
            file.type.includes("image") ||
            file.type.includes("video") ||
            file.type.includes("pdf")
        ) {
            let fileTypes = [];

            sortedFiles.value.forEach((file) => {
                if (
                    file.type.includes("image") ||
                    file.type.includes("video") ||
                    file.type.includes("pdf")
                ) {
                    fileTypes.push({
                        id: file.id,
                        name: file.name,
                        type: file.type,
                        url: window.location.origin + "/api/files/view/" + file.id,
                    });
                }
            });

            let initialIndex = fileTypes.findIndex((f) => f.id === file.id);

            setTimeout(() => {
                mediaStore.openViewer(fileTypes, initialIndex);
            }, 100);

            return;
        }

        officeService
            .supportFileType(file.type)
            .then((res) => {
                if (res.data) {
                    officeService.openFile(file.id, "files").then((res) => {
                        officeStore.openOffice(res.data);
                    });
                } else {
                    handleDownloadFile(file);
                }
            })
            .catch(() => {
                handleDownloadFile(file);
            });

        return;
    }

    function showDetails(file) {
        detailsStore.loadDetails(file);

        if (window.matchMedia("(min-width: 1024px)").matches) {
            detailsStore.open = true;
        } else {
            detailsStore.mobileOpen = true;
        }
    }

    async function handleRenameFile(event) {
        await renameFile(fileService, event)
            .then(() => {
                const fileIndex = filesStore.getFiles.findIndex((f) => f.id === event.file.id);

                if (fileIndex !== -1) {
                    filesStore.getFiles[fileIndex].name = event.name;
                }
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    function handleAddToFavorites() {
        addToFavorite(fileService, contextMenuFile)
            .then(() => {
                const fileIndex = filesStore.getFiles.findIndex((f) => f.id === contextMenuFile.id);

                if (fileIndex !== -1) {
                    filesStore.files[fileIndex].favourite = !filesStore.files[fileIndex].favourite;
                }
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    function handleDeleteFile() {
        fileService
            .trashFile({ id: selectedFiles.value.map((file) => file.id) })
            .then(() => {
                filesStore.setFiles(
                    filesStore.getFiles.filter(
                        (file) => !selectedFiles.value.some((selected) => selected.id === file.id),
                    ),
                );

                selectedFiles.value = [];
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    async function emptyTrash() {
        deleting.value = true;
        let filesToDelete = selectedFiles.value.map((file) => file.id);

        if (selectedFiles.value.length == 0) {
            filesToDelete = filesStore.getFiles.map((file) => file.id);
        }

        await deleteFile(fileService, filesToDelete)
            .then(() => {
                filesStore.setFiles(
                    filesStore.getFiles.filter((file) => !filesToDelete.includes(file.id)),
                );

                deleting.value = false;
            })
            .catch((error) => {
                deleting.value = false;
                useAlertStore().showError(extractErrorMessage(error));
            });

        return;
    }

    async function handleRestoreFile(file) {
        fileService
            .restoreFile({ id: [file.id] })
            .then(() => {
                filesStore.setFiles(filesStore.getFiles.filter((f) => f.id !== file.id));

                selectedFiles.value = [];
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    async function handleDeleteFilePermanently(file) {
        deleting.value = true;

        await deleteFile(fileService, [file.id])
            .then(() => {
                filesStore.setFiles(filesStore.getFiles.filter((f) => f.id !== file.id));

                deleting.value = false;
            })
            .catch((error) => {
                deleting.value = false;
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    async function handleMoveFile(event) {
        if (await moveFiles(selectedFiles.value, event.location)) {
            moveDialogRef.value?.closeDialog();
        }
    }

    return {
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
    };
}
