// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Application dialog state store
import { defineStore } from "pinia";

export const useDialogStore = defineStore("dialog", {
    state: () => {
        return {
            fileDialogOpen: false,
            fileType: "",
            renameDialogOpen: false,
            deleteDialogOpen: false,
            moveDialogOpen: false,
            shareDialogOpen: false,
            file: null,
            shareDialog: {
                file: null,
                editItem: null,
                editLink: null,
                editPermission: "",
                editShareType: 1,
                accessLevel: "viewer",
                expiry: null,
            },
        };
    },
    actions: {
        openFileDialog(fileType) {
            this.fileType = fileType;
            this.fileDialogOpen = true;
        },
        closeFileDialog() {
            this.fileDialogOpen = false;
        },
        openRenameDialog() {
            this.renameDialogOpen = true;
        },
        closeRenameDialog() {
            this.renameDialogOpen = false;
        },
        openDeleteDialog() {
            this.deleteDialogOpen = true;
        },
        closeDeleteDialog() {
            this.deleteDialogOpen = false;
        },
        openMoveDialog() {
            this.moveDialogOpen = true;
        },
        closeMoveDialog() {
            this.moveDialogOpen = false;
        },
        openShareDialog(file) {
            this.shareDialog.file = file;
            this.shareDialog.editLink = null;
            this.shareDialogOpen = true;
        },
        openShareDialogEditLink(file, link) {
            this.shareDialog.file = file;
            this.shareDialog.editLink = link;
            this.shareDialogOpen = true;
        },
        openEditShareDialog(file, share, shareType = 1) {
            this.shareDialog.file = file;
            this.shareDialog.editItem = share;
            this.shareDialog.editPermission = "share";
            this.shareDialog.editShareType = shareType;
            this.shareDialog.accessLevel = share.accessLevel ?? share.access_level ?? "viewer";
            this.shareDialog.expiry = share.expiration;
            this.shareDialogOpen = true;
        },
        closeShareDialog() {
            this.shareDialogOpen = false;
        },
    },
});
