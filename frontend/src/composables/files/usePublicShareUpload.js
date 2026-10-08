// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed } from "vue";
import { t } from "@/i18n";
import { extractErrorMessage } from "@/utils/errors";
import publicShareService from "@/services/publicShareService";
import { useAlertStore } from "@/store/alerts";

const CHUNK_SIZE = 10 * 1024 * 1024;

export function usePublicShareUpload({ token, share, currentChildId, load }) {
    const alertStore = useAlertStore();

    const dragOver = ref(false);
    const uploading = ref(false);
    const uploadProgress = ref(0);
    const fileInput = ref(null);

    const uploadLabel = computed(() => {
        if (!uploading.value) return t.value("public_share.upload_button");

        return uploadProgress.value
            ? `${uploadProgress.value}%`
            : t.value("public_share.uploading");
    });

    function triggerUpload() {
        fileInput.value?.click();
    }

    async function uploadOne(file) {
        const folderId = currentChildId.value;

        if (file.size < CHUNK_SIZE) {
            await publicShareService.upload(token, file, folderId);

            return;
        }

        const { data: session } = await publicShareService.startUpload(token, file, folderId);
        let start = 0;
        let partNumber = 1;

        while (start < file.size) {
            const end = Math.min(start + CHUNK_SIZE, file.size);

            await publicShareService.uploadChunk(
                token,
                session.id,
                file.slice(start, end),
                partNumber,
                folderId,
            );
            uploadProgress.value = Math.round((end / file.size) * 100);
            partNumber++;
            start = end;
        }

        await publicShareService.finishUpload(token, session.id, folderId);
    }

    async function runUpload(files) {
        if (!files.length) return;
        uploading.value = true;
        uploadProgress.value = 0;
        try {
            for (const file of files) {
                await uploadOne(file);
            }

            alertStore.showSuccess(t.value("public_share.upload_success"));
            await load();
        } catch (error) {
            alertStore.showError(extractErrorMessage(error));
        } finally {
            uploading.value = false;
            uploadProgress.value = 0;
        }
    }

    async function handleUpload(event) {
        const file = event.target.files?.[0];

        if (!file) return;
        await runUpload([file]);
        if (fileInput.value) fileInput.value.value = "";
    }

    async function handleDrop(event) {
        dragOver.value = false;
        if (!share.value?.AllowUpload) return;
        const files = [...(event.dataTransfer?.files || [])];

        await runUpload(files);
    }

    return {
        dragOver,
        uploading,
        fileInput,
        uploadLabel,
        triggerUpload,
        handleUpload,
        handleDrop,
    };
}
