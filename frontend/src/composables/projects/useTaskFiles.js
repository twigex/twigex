// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useMediaStore } from "@/store/media";
import { useOfficeStore } from "@/store/office";
import { useAlertStore } from "@/store/alerts";
import officeService from "@/services/officeService";
import workspaceService from "@/services/workspaceService";
import { isViewable } from "@/utils/projects/files";

// Attaching, opening and downloading the files of a task, as the grid and
// the task panel do it.
export function useTaskFiles() {
    const route = useRoute();
    const mediaStore = useMediaStore();
    const officeStore = useOfficeStore();

    const getFileExtension = (filename) => {
        return filename.split(".").pop().toLowerCase();
    };

    const getFileName = (file) => {
        return file.name || file.filename || t.value("projects.full_task_navigation.unknow_file");
    };

    const openFile = async (file, allFiles) => {
        if (isViewable(file)) {
            const viewable = allFiles.filter(isViewable);
            const mediaFiles = await Promise.all(
                viewable.map(async (f) => {
                    const fileUrl = await getFileUrl(f);

                    return {
                        ...f,
                        url: fileUrl,
                        type: f.type,
                    };
                }),
            );

            const fileIndex = viewable.findIndex(
                (f) => f.id === file.id || f.name === file.name || f.url === file.url,
            );

            mediaStore.files = mediaFiles;
            mediaStore.index = fileIndex >= 0 ? fileIndex : 0;
            mediaStore.show = true;
        } else {
            officeService
                .supportFileType(file.type || getFileExtension(file.name))
                .then((res) => {
                    if (res.data) {
                        return officeService.openFile(file.id, "projects").then((res) => {
                            officeStore.openOffice(res.data);
                        });
                    } else {
                        downloadFile(file);
                    }
                })
                .catch(() => {
                    downloadFile(file);
                });
        }
    };

    const getFileUrl = async (file) => {
        if (file.url) return file.url;

        if (file.id) {
            return `/api/workspaces/${route.params.id}/file/${file.id}`;
        }

        if (file instanceof File || (file.constructor && file.constructor.name === "File")) {
            return URL.createObjectURL(file);
        }

        return "#";
    };

    const downloadFile = async (file) => {
        try {
            if (!file.id) {
                throw new Error("File ID not found");
            }

            const fileUrl = `/api/workspaces/${route.params.id}/file/${file.id}`;
            const response = await fetch(fileUrl);

            if (!response.ok) {
                throw new Error(`Download failed: ${response.statusText}`);
            }

            const blob = await response.blob();
            const downloadUrl = window.URL.createObjectURL(blob);
            const link = document.createElement("a");

            link.href = downloadUrl;
            link.download = getFileName(file);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            window.URL.revokeObjectURL(downloadUrl);
        } catch (error) {
            console.error("Download error:", error);
            useAlertStore().showError(
                t.value("projects.full_task_navigation.failed_to_download_file"),
            );
        }
    };

    // chooseFiles opens the file picker of the hidden input the selector names.
    const chooseFiles = (selector, canEdit) => {
        if (!canEdit) {
            useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

            return;
        }

        const input = document.querySelector(selector);

        if (input) input.click();
        else console.error("File input element not found with selector:", selector);
    };

    // uploadTaskFiles attaches files to a task's file field and returns the
    // attachments the server made of them.
    const uploadTaskFiles = async (files, { tableId, taskId, header }) => {
        const formData = new FormData();

        files.forEach((file) => formData.append("files", file));
        formData.append("workspace_id", route.params.id);
        formData.append("table_id", tableId);
        formData.append("task_id", taskId);
        formData.append("field", header.name);
        formData.append("field_id", header.id);

        const response = await workspaceService.uploadFiles(formData);

        return Array.isArray(response.data) ? response.data : null;
    };

    return {
        getFileExtension,
        getFileName,
        openFile,
        getFileUrl,
        downloadFile,
        chooseFiles,
        uploadTaskFiles,
    };
}
