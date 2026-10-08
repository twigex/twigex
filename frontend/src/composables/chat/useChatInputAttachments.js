// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, reactive, toValue } from "vue";
import { useRoute } from "vue-router";
import chatService from "@/services/chatService";

export function useChatInputAttachments({ message }) {
    const route = useRoute();

    const filePreviews = ref([]);
    const error = ref([]);
    const uploadErrors = ref([]);
    const removedFiles = ref([]);
    const fileSize = ref("100MB");

    function clearFileErrors() {
        error.value = [];
        uploadErrors.value = [];
    }

    function processFiles(fileList) {
        const sizeLimit = 100 * 1024 * 1024;

        clearFileErrors();

        Array.from(fileList).forEach((file) => {
            if (file.size > sizeLimit) {
                error.value.push(file.name);

                return;
            }

            const previewFile = reactive({
                id: null,
                name: file.name,
                type: file.type,
                size: file.size,
                progress: 0,
                url: null,
            });

            if (file.type.includes("image")) {
                const reader = new FileReader();

                reader.onload = (e) => {
                    previewFile.url = e.target.result;
                };

                reader.readAsDataURL(file);
            }

            filePreviews.value.push(previewFile);

            chatService
                .uploadFile(file, route.params.chatId, (progress) => {
                    previewFile.progress = Math.min(progress, 99);
                })
                .then((response) => {
                    if (!filePreviews.value.includes(previewFile)) {
                        return;
                    }

                    previewFile.progress = 100;
                    previewFile.id = response.data.id;
                    previewFile.storage_id = response.data.storage_id;
                })
                .catch(() => {
                    if (!filePreviews.value.includes(previewFile)) {
                        return;
                    }

                    filePreviews.value = filePreviews.value.filter(
                        (preview) => preview !== previewFile,
                    );
                    uploadErrors.value.push(file.name);
                });
        });
    }

    function removeFilePreview(index) {
        if (toValue(message) != null) {
            removedFiles.value.push(filePreviews.value[index].id);
        }

        filePreviews.value.splice(index, 1);
    }

    function loadMessageFiles(post) {
        if (post.metadata && post.metadata.files && post.metadata.files.length > 0) {
            post.metadata.files.forEach((file) => {
                let previewFile = reactive({
                    id: file.id,
                    name: file.name,
                    type: file.mime_type,
                    size: file.size,
                    progress: 100,
                    url: file.kind == "gifv" ? file.url : null,
                    kind: file.kind,
                });

                if (file.mime_type.includes("image")) {
                    previewFile.url = `/api/channels/${post.channel_id}/posts/${post.id}/file/${file.id}`;
                }

                filePreviews.value.push(previewFile);
            });
        }
    }

    return {
        filePreviews,
        error,
        uploadErrors,
        removedFiles,
        fileSize,
        clearFileErrors,
        processFiles,
        removeFilePreview,
        loadMessageFiles,
    };
}
