// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import useFileOperations from "@/composables/files/useFileOperations";
import fileService from "@/services/fileService";

const { findFileInTree } = useFileOperations();

export const useFilesStore = defineStore("files", {
    state: () => {
        return {
            treeData: [],
            fileOrder: null,
            files: [],
            currentDirectory: null,
            view: {
                currentDirectory: "",
                currentName: "",
            },
        };
    },

    getters: {
        getFiles() {
            return this.files;
        },
    },

    actions: {
        async completeMoveJob(data) {
            const payload = data;

            const { root_ids } = payload;
            const currentDir = this.view.currentDirectory;

            for (const item of root_ids) {
                const { id, source, target } = item;

                // REMOVE from current view if we are in source dir
                if (currentDir === source) {
                    this.files = this.files.filter((f) => f.id !== id);
                }

                // ADD to current view if we are in target dir
                if (currentDir === target) {
                    try {
                        const res = await fileService.getFile(id);

                        this.files.unshift(res.data);
                    } catch (err) {
                        console.error("Failed to fetch moved file:", id, err);
                    }
                }
            }

            const treeData = JSON.parse(JSON.stringify(this.treeData));

            for (const item of root_ids) {
                const { id, source, target, is_folder } = item;

                if (!is_folder) {
                    continue;
                }

                const sourceNode = findFileInTree(treeData, source);
                const targetNode = findFileInTree(treeData, target);

                let movedNode = null;

                // Remove from source if loaded
                if (sourceNode?.children) {
                    const idx = sourceNode.children.findIndex((c) => c.id === id);

                    if (idx !== -1) {
                        movedNode = sourceNode.children.splice(idx, 1)[0];
                    }
                }

                // Add to target
                if (targetNode?.children) {
                    if (movedNode) {
                        targetNode.children.push(movedNode);
                    } else {
                        // If not in memory, fetch it
                        try {
                            const res = await fileService.getFile(id);

                            targetNode.children.push(res.data);
                        } catch (err) {
                            console.error("Failed to fetch tree node:", id, err);
                        }
                    }
                }
            }

            this.treeData = treeData;
        },

        setTreeData(treeData) {
            this.treeData = treeData;
        },

        setFileOrder(fileOrder) {
            this.fileOrder = fileOrder;
        },

        setFiles(files) {
            this.files = files;
        },

        setCurrentView(view) {
            this.view = view;
        },

        addFile(file) {
            this.files.unshift(file);
        },
    },
});
