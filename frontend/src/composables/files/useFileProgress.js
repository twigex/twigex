// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { useUploadStore } from "@/store/upload";
import { useJobsStore } from "@/store/jobs";

const JOB_STATUS_MAP = {
    running: "uploading",
    pending: "pending",
    completed: "completed",
    failed: "error",
    cancelled: "cancelled",
};

export function moveJobName(job) {
    if (job.type !== "move_files") {
        return t.value("files.progress.processing");
    }

    try {
        const payload = JSON.parse(atob(job.payload));
        const count = (payload.root_ids?.length || 0) + (payload.subfile_ids?.length || 0);

        if (count > 0) {
            return t.value("files.progress.moving_items", { count });
        }
    } catch {}

    return t.value("files.progress.moving");
}

export function useFileProgress() {
    const uploadStore = useUploadStore();
    const jobsStore = useJobsStore();

    const uploads = computed(() =>
        uploadStore.files.map((file) => ({
            id: file.id,
            name: file.name,
            type: file.type,
            progress: file.progress ?? 0,
            status: file.status,
            error: file.status === "error" ? file.error : null,
            cancel: () => uploadStore.cancelUpload(file.id),
        })),
    );

    const moves = computed(() => {
        if (jobsStore.hidden) return [];

        return jobsStore.jobs.map((job) => {
            const status = JOB_STATUS_MAP[job.status] ?? "error";

            return {
                id: job.id,
                name: moveJobName(job),
                progress: job.progress ?? 0,
                status,
                error: status === "error" ? t.value("files.progress.move_failed_detail") : null,
                cancel: () => jobsStore.cancelJob(job.id),
            };
        });
    });

    const uploadTitle = computed(() => {
        const done = uploads.value.filter((upload) => upload.status === "completed").length;
        const failed = uploads.value.filter((upload) => upload.status === "error").length;
        const count = uploads.value.length;

        if (failed > 0) {
            return t.value("files.progress.uploaded_with_failed", { done, count, failed });
        }

        return t.value("files.progress.uploaded", { done, count });
    });

    const moveTitle = computed(() => {
        const statuses = moves.value.map((move) => move.status);

        if (statuses.some((status) => status === "pending" || status === "uploading")) {
            return t.value("files.progress.moving");
        }

        if (statuses.some((status) => status === "error")) {
            return t.value("files.progress.move_failed");
        }

        return t.value("files.progress.move_complete");
    });

    return {
        uploads,
        moves,
        uploadTitle,
        moveTitle,
        closeUploads: () => uploadStore.close(),
        closeMoves: () => jobsStore.dismiss(),
    };
}
