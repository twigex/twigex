// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useFilesStore } from "@/store/files";
import { useJobsStore } from "@/store/jobs";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import fileService from "@/services/fileService";

export function useMoveFiles() {
    const filesStore = useFilesStore();
    const jobsStore = useJobsStore();

    async function moveFiles(files, targetId) {
        try {
            const res = await fileService.moveFile({
                id: files.map((file) => file.id),
                path: targetId,
            });

            if (res.data?.job) {
                jobsStore.startPolling();
            }

            if (res.data?.file_move?.immediate) {
                await filesStore.completeMoveJob(res.data.file_move.payload);
            }

            return true;
        } catch (error) {
            useAlertStore().showError(extractErrorMessage(error));

            return false;
        }
    }

    return {
        moveFiles,
    };
}
