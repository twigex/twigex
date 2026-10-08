// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export default function useFileOperations() {
    const findFileInTree = (nodes, id, remove = false) => {
        for (let i = 0; i < nodes.length; i++) {
            if (nodes[i].id === id) {
                return nodes[i];
            }

            if (nodes[i].children) {
                const found = findFileInTree(nodes[i].children, id, remove);

                if (found) {
                    if (remove) {
                        nodes[i].children = nodes[i].children.filter((item) => item.id !== id);
                    }

                    return found;
                }
            }
        }

        return null;
    };

    const renameFile = async (fileService, event) => {
        return fileService.renameFile({ id: event.file.id, name: event.name });
    };

    const addToFavorite = async (fileService, contextMenuFile) => {
        return fileService.addToFavorites(contextMenuFile.id);
    };

    const deleteFile = async (fileService, files) => {
        await fileService.deleteFile({ id: files });
    };

    const downloadFile = (fileService, file) => {
        const filePath = `/api/files/download/${file.id}`;

        const a = document.createElement("a");

        a.href = filePath;
        a.download = file.name || "";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    };

    const shareFile = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService
                .shareFile({
                    fileID: event.file,
                    users: event.users ?? [],
                    groups: event.groups ?? [],
                    expiration: event.expiry,
                    accessLevel: event.accessLevel,
                })
                .then(resolve)
                .catch(reject);
        });
    };

    const unshareFile = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService.unshareFile(event.fileId, event.userId).then(resolve).catch(reject);
        });
    };

    const unshareGroup = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService.unshareGroup(event.fileId, event.groupId).then(resolve).catch(reject);
        });
    };

    const updateShare = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService
                .updateShare({
                    id: event.id,
                    fileID: event.fileID,
                    shareType: event.shareType,
                    accessLevel: event.accessLevel,
                    expiration: event.expiry,
                })
                .then(resolve)
                .catch(reject);
        });
    };

    const shareLink = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService
                .shareLink({
                    item: event.file,
                    expiration: event.expiry,
                    passwordProtected: event.passwordProtected,
                    password: event.password,
                    allowDownload: event.download,
                    allowUpload: event.upload,
                    message: event.message,
                })
                .then(resolve)
                .catch(reject);
        });
    };

    const updateLinkShare = (shareService, event) => {
        return new Promise((resolve, reject) => {
            shareService
                .updateLinkShare({
                    token: event.token,
                    expiration: event.expiry,
                    passwordProtected: event.passwordProtected,
                    password: event.password,
                    allowDownload: event.download,
                    allowUpload: event.upload,
                    message: event.message,
                })
                .then(resolve)
                .catch(reject);
        });
    };

    return {
        findFileInTree,
        renameFile,
        addToFavorite,
        deleteFile,
        downloadFile,
        shareFile,
        unshareFile,
        unshareGroup,
        updateShare,
        shareLink,
        updateLinkShare,
    };
}
