// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRoute } from "vue-router";
import { useUploadStore } from "@/store/upload";

export function useFileBrowserUpload() {
    const route = useRoute();
    const uploadStore = useUploadStore();

    async function handleDroppedFiles(event, id) {
        let preparedFiles = [];
        let preparedDirectories = [];

        for (let i = 0; i < event.dataTransfer.items.length; i++) {
            if (event.dataTransfer.items[i].type !== "") {
                let file = event.dataTransfer.items[i].getAsFile();

                if (file) {
                    preparedFiles.push(...prepareFiles([file], id));
                }
            } else {
                let dirFiles = await collectFilesFromDataTransfer(
                    event.dataTransfer,
                    event.dataTransfer.items[i],
                );

                preparedDirectories.push(prepareDirectory(dirFiles, id));
            }
        }

        if (uploadStore.files.length > 0) {
            uploadStore.addFiles(preparedFiles);
            uploadStore.addFiles(preparedDirectories);
        } else {
            uploadStore.setFiles([...preparedFiles, ...preparedDirectories]);
            uploadStore.startUpload(id);
        }
    }

    function handleFilesSelected(event) {
        if (event.type == "file") {
            let preparedFiles = prepareFiles(event.files, route.params.id);

            addFilesToUploadQueue(preparedFiles, route.params.id);
        } else {
            let directory = prepareDirectory(event.files, route.params.id);

            addFilesToUploadQueue([directory], route.params.id);
        }
    }

    function addFilesToUploadQueue(files, targetId) {
        if (uploadStore.files.length > 0) {
            uploadStore.addFiles(files);
        } else {
            uploadStore.setFiles(files);
            uploadStore.startUpload(targetId);
        }

        uploadStore.open();
    }

    function prepareFiles(files, targetId) {
        let filesArray = [];

        for (let i = 0; i < files.length; i++) {
            filesArray.push(
                createNode({
                    targetId: targetId,
                    name: files[i].name,
                    isFile: true,
                    file: files[i],
                    size: files[i].size,
                    type: files[i].type,
                }),
            );
        }

        return filesArray;
    }

    function prepareDirectory(files, targetId) {
        const rootFolder = buildFileTree(files, targetId);

        return rootFolder;
    }

    function buildFileTree(files, targetId) {
        if (!files.length) return null;

        const rootName = files[0].webkitRelativePath.split("/")[0];

        const root = createNode({
            targetId: targetId,
            name: rootName,
            isFile: false,
            type: "inode/directory",
            root: true,
        });

        root.root = true;

        for (const file of files) {
            const path = file.webkitRelativePath.split("/");
            let current = root;

            for (let i = 1; i < path.length; i++) {
                const part = path[i];

                if (i === path.length - 1) {
                    current.children.push(
                        createNode({
                            name: part,
                            isFile: true,
                            file,
                            size: file.size,
                            type: file.type,
                            rootId: root.id,
                        }),
                    );
                } else {
                    let dir = current.children.find((n) => !n.isFile && n.name === part);

                    if (!dir) {
                        dir = createNode({
                            name: part,
                            isFile: false,
                            type: "inode/directory",
                            rootId: root.id,
                        });
                        current.children.push(dir);
                    }

                    current = dir;
                }
            }
        }

        return root;
    }

    function createNode({
        targetId,
        name,
        isFile,
        file = null,
        size = 0,
        type,
        root = false,
        rootId = null,
    }) {
        return {
            id: crypto.randomUUID(),
            targetId: targetId || null,
            name,
            isFile,
            file,
            size,
            type,
            rootFolder: root,
            rootId: rootId,
            status: "pending",
            progress: 0,
            totalFiles: 0,
            children: isFile ? null : [],
        };
    }

    async function collectFilesFromDataTransfer(dt, item) {
        const out = [];

        const items = [item];

        if (items.length && "getAsFileSystemHandle" in items[0]) {
            for (const item of items) {
                const handle = await item.getAsFileSystemHandle();

                if (handle) await traverseHandle(handle, out, "");
            }

            return out;
        }

        if (items.length && items[0].webkitGetAsEntry) {
            for (const item of items) {
                const entry = item.webkitGetAsEntry();

                if (entry) await traverseEntry(entry, out, "");
            }

            return out;
        }

        for (const f of Array.from(dt.files || [])) out.push(f);

        return out;
    }

    async function traverseHandle(handle, out, prefix) {
        if (handle.kind === "file") {
            const file = await handle.getFile();

            file.relativePath = prefix + file.name;
            out.push(file);
        } else if (handle.kind === "directory") {
            for await (const [, child] of handle.entries()) {
                await traverseHandle(child, out, `${prefix}${handle.name}/`);
            }
        }
    }

    function readAllEntries(dirReader) {
        return new Promise((resolve) => {
            const entries = [];
            const read = () =>
                dirReader.readEntries((chunk) => {
                    if (chunk.length) {
                        entries.push(...chunk);
                        read();
                    } else {
                        resolve(entries);
                    }
                });

            read();
        });
    }

    async function traverseEntry(entry, out, prefix) {
        if (entry.isFile) {
            const file = await new Promise((res) => entry.file(res));

            file.relativePath = prefix + file.name;
            out.push(file);
        } else if (entry.isDirectory) {
            const reader = entry.createReader();
            const children = await readAllEntries(reader);

            for (const child of children) {
                await traverseEntry(child, out, `${prefix}${entry.name}/`);
            }
        }
    }

    return {
        handleDroppedFiles,
        handleFilesSelected,
    };
}
