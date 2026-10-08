// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import { useFilesStore } from "@/store/files";
import axios from "axios";

import fileService from "@/services/fileService";
import { extractErrorMessage } from "@/utils/errors";
import { t } from "@/i18n/index.js";

export const useUploadStore = defineStore("upload", {
    state: () => {
        return {
            files: [],
            type: "",
            show: false,
            chunkSize: 10 * 1024 * 1024, // 10MB
        };
    },

    getters: {},

    actions: {
        setFiles(files) {
            this.files = files.map((f) => this.initNode(f));
        },

        addFiles(files) {
            const newFiles = files.map((f) => this.initNode(f));

            this.files.push(...newFiles);

            let isUploading = this.files.some((f) => f.status === "uploading");

            if (!isUploading) {
                this.startUpload();
            }
        },

        computeRootProgress(root) {
            if (!root.totalFiles || root.totalFiles === 0) return 100;

            let sum = 0;
            const stack = [...(root.children || [])];

            while (stack.length) {
                const node = stack.pop();

                if (node.isFile) {
                    sum += (node.progress || 0) / 100;
                } else if (node.children) {
                    stack.push(...node.children);
                }
            }

            return Math.round((sum / root.totalFiles) * 100);
        },

        countFiles(node) {
            if (!node) return 0;
            if (node.isFile) return 1;

            let count = 0;
            const stack = [...(node.children || [])];

            while (stack.length) {
                const n = stack.pop();

                if (n.isFile) count++;
                else if (n.children) stack.push(...n.children);
            }

            return count;
        },

        initNode(node, controller = null) {
            let newNode = {
                ...node,
                controller: controller || new AbortController(),
            };

            if (!node.isFile && Array.isArray(node.children)) {
                newNode.children = node.children.map((child) =>
                    this.initNode(child, newNode.controller),
                );

                if (newNode.root) {
                    newNode.totalFiles = this.countFiles(node);
                }
            }

            return newNode;
        },

        findNode(root, predicate) {
            if (!root) return null;

            if (predicate(root)) return root;

            if (Array.isArray(root.children)) {
                for (const child of root.children) {
                    const found = this.findNode(child, predicate);

                    if (found) return found;
                }
            }

            return null;
        },

        cancelAllUploads() {
            for (const file of this.files) {
                if (file.controller && !file.controller.signal.aborted) {
                    file.controller.abort();
                }

                if (file.status !== "completed" && file.status !== "error") {
                    file.status = "cancelled";
                }
            }
        },

        async cancelUploadSession(file) {
            if (!file?.uploadSessionId) return;

            try {
                await fileService.cancelUpload(file.uploadSessionId);
            } catch (err) {
                // backend may already consider it aborted, safe to ignore
                console.warn("Upload session cancel failed:", err);
            } finally {
                file.uploadSessionId = null;
            }
        },

        close() {
            this.cancelAllUploads();
            this.show = false;
            this.files = [];
        },

        open() {
            this.show = true;
        },

        async cancelUpload(fileId) {
            const file = this.files.find((f) => f.id === fileId);

            if (!file) return;

            this.markCancelled(fileId);

            if (file.uploadSessionId) {
                await this.cancelUploadSession(file);
            }

            if (!file.isFile && Array.isArray(file.children)) {
                const stack = [...file.children];

                while (stack.length) {
                    const node = stack.pop();

                    this.markCancelled(node.id);
                    if (node.uploadSessionId) {
                        await this.cancelUploadSession(node);
                    }

                    if (node.children) stack.push(...node.children);
                }
            }
        },

        setFileProgress(fileId, progress, rootId = null) {
            let file;

            // if rootId is provided, we are updating a file inside a folder upload
            if (rootId) {
                const root = this.files.find((f) => f.id === rootId);

                if (!root) return;

                file = this.findNode(root, (n) => n.id === fileId);
                if (!file) return;

                if (file.status === "cancelled") return;

                file.progress = progress;
                file.status = "uploading";

                // recompute folder progress
                root.progress = this.computeRootProgress(root);

                return;
            }

            file = this.files.find((f) => f.id === fileId);
            if (!file) return;

            if (file.status === "cancelled") return;

            file.progress = progress;
            file.status = "uploading";
        },

        markCompleted(fileId) {
            const file = this.files.find((f) => f.id === fileId);

            if (!file) return;

            file.progress = 100;
            file.status = "completed";
        },

        markError(fileId, error) {
            const file = this.files.find((f) => f.id === fileId);

            if (!file) return;

            file.status = "error";
            file.error = extractErrorMessage(error, t.value("files.progress.upload_failed"));
        },

        markCancelled(fileId) {
            const file = this.files.find((f) => f.id === fileId);

            if (!file) return;

            file.status = "cancelled";
            file.controller.abort();
        },

        async startUpload() {
            for (let i = 0; i < this.files.length; i++) {
                if (this.files[i].status === "pending") {
                    if (this.files[i].isFile) {
                        this.files[i].status = "uploading";

                        if (this.files[i].size < this.chunkSize) {
                            await this.uploadFileNotInChunks(this.files[i], this.files[i].targetId);

                            continue;
                        }

                        await this.uploadFileInChunks(this.files[i], this.files[i].targetId).then(
                            () => {},
                        );
                    } else {
                        await this.createFoldersAndUploadFiles(
                            this.files[i],
                            this.files[i].targetId,
                            this.files[i],
                        ).then(() => {
                            if (this.files[i].status == "cancelled") {
                                return;
                            }

                            this.markCompleted(this.files[i].id);
                        });
                    }
                }
            }
        },

        async uploadFileNotInChunks(file, parentId) {
            if (file.controller.signal.aborted) {
                this.markCancelled(file.id);

                return;
            }

            const filesStore = useFilesStore();

            try {
                let formData = new FormData();

                formData.append("parent_id", parentId);
                formData.append("file", file.file);
                formData.append("name", file.name);
                formData.append("type", file.type);

                await fileService
                    .upload(
                        formData,
                        "upload",
                        (progress) => {
                            this.setFileProgress(file.id, progress, file.rootId);
                        },
                        file.controller.signal,
                    )
                    .then((response) => {
                        this.markCompleted(file.id);
                        if (filesStore.view.currentDirectory === parentId) {
                            filesStore.addFile(response.data);
                        }
                    })
                    .catch((err) => {
                        if (axios.isCancel?.(err) || err.code === "ERR_CANCELED") {
                            this.markCancelled(file.id);

                            return;
                        }

                        this.markError(file.id, err);

                        return;
                    });
            } catch (err) {
                if (file.controller.signal.aborted) {
                    this.markCancelled(file.id);

                    return;
                }

                this.markError(file.id, err);

                return;
            }
        },

        async uploadFileInChunks(file, parentId) {
            if (file.controller.signal.aborted) {
                this.markCancelled(file.id);

                return;
            }

            let uploadSession;

            try {
                uploadSession = await this.startFileUpload(file, parentId);
            } catch (err) {
                if (axios.isCancel(err) || err.code === "ERR_CANCELED") {
                    this.markCancelled(file.id);

                    return;
                }

                this.markError(file.id, err);

                return;
            }

            let start = 0;
            const uploadSessionId = uploadSession.id;
            let partNumber = 1;

            while (start < file.size) {
                if (file.status === "cancelled" || file.controller.signal.aborted) {
                    this.markCancelled(file.id);

                    return;
                }

                const end = Math.min(start + this.chunkSize, file.size);
                const chunk = file.file.slice(start, end);

                const formData = new FormData();

                formData.append("upload_session_id", uploadSessionId);
                formData.append("file", chunk);
                formData.append("part_number", String(partNumber));

                try {
                    await fileService.upload(formData, "chunk", null, file.controller.signal);

                    const progress = Math.min((end / file.size) * 100, 100);

                    this.setFileProgress(file.id, progress, file.rootId);
                } catch (err) {
                    if (axios.isCancel?.(err) || err.code === "ERR_CANCELED") {
                        this.markCancelled(file.id);

                        return;
                    }

                    this.markError(file.id, err);

                    return;
                }

                partNumber++;
                start = end;
            }

            try {
                await this.finishMultiUpload(file, uploadSessionId, parentId);
            } catch {
                return;
            }
        },

        async startFileUpload(file, parentId) {
            const startForm = new FormData();

            startForm.append("parent_id", parentId);
            startForm.append("name", file.name);
            startForm.append("type", file.type);
            startForm.append("full_size", String(file.size));

            const response = await fileService.startUpload(startForm);

            file.uploadSessionId = response.data.id;

            return response.data;
        },

        async finishMultiUpload(file, uploadSessionId, parentId) {
            const filesStore = useFilesStore();

            const finishForm = new FormData();

            finishForm.append("upload_session_id", uploadSessionId);

            try {
                const response = await fileService.finishUpload(finishForm, file.controller.signal);

                this.markCompleted(file.id);

                if (filesStore.view.currentDirectory === parentId) {
                    filesStore.addFile(response.data);
                }
            } catch (err) {
                if (axios.isCancel(err) || err.code === "ERR_CANCELED") {
                    this.markCancelled(file.id);

                    return;
                }

                this.markError(file.id, err);
                throw err;
            }
        },

        async createFoldersAndUploadFiles(node, parentId, rootFolder) {
            if (rootFolder.status == "cancelled") {
                return;
            }

            const filesStore = useFilesStore();

            // node is ALWAYS a folder here
            const response = await fileService.createFolder({
                id: parentId,
                name: node.name,
            });

            if (node.root && filesStore.view.currentDirectory === parentId) {
                filesStore.addFile(response.data);
            }

            const children = Array.isArray(node.children) ? node.children : [];

            for (let i = 0; i < children.length; i++) {
                const child = children[i];

                child.targetId = response.data.id;

                if (rootFolder.status === "cancelled") {
                    return;
                }

                if (child.status === "cancelled") {
                    continue;
                }

                if (child.type === "inode/directory") {
                    await this.createFoldersAndUploadFiles(child, child.targetId, rootFolder);
                } else {
                    if (child.size < this.chunkSize) {
                        await this.uploadFileNotInChunks(child, child.targetId).then(() => {
                            this.setFileProgress(
                                rootFolder.id,
                                this.computeRootProgress(rootFolder),
                            );
                        });
                    } else {
                        await this.uploadFileInChunks(child, child.targetId).then(() => {
                            this.setFileProgress(
                                rootFolder.id,
                                this.computeRootProgress(rootFolder),
                            );
                        });
                    }
                }
            }
        },
    },
});
