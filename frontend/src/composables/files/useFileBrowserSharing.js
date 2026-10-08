// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useDetailsStore } from "@/store/details";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import shareService from "@/services/shareService";
import useFileOperations from "@/composables/files/useFileOperations";

export function useFileBrowserSharing() {
    const detailsStore = useDetailsStore();
    const { shareFile, unshareFile, unshareGroup, updateShare, shareLink, updateLinkShare } =
        useFileOperations();

    function mergeById(existing, incoming) {
        const byId = new Map((existing ?? []).map((row) => [row.id, row]));

        for (const row of incoming ?? []) {
            byId.set(row.id, row);
        }

        return [...byId.values()];
    }

    function handleShareFile(event) {
        shareFile(shareService, event).then((response) => {
            if (detailsStore.details != null && detailsStore.file.id == event.file) {
                detailsStore.details.sharedUsers = mergeById(
                    detailsStore.details.sharedUsers,
                    response.data.shared_users,
                );
                detailsStore.details.sharedGroups = mergeById(
                    detailsStore.details.sharedGroups,
                    response.data.shared_groups,
                );
            }
        });
    }

    function handleUnshareFile(event) {
        unshareFile(shareService, event)
            .then(() => {
                if (detailsStore.details?.sharedUsers) {
                    detailsStore.details.sharedUsers = detailsStore.details.sharedUsers.filter(
                        (user) => user.id !== event.userId,
                    );
                }
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    function handleUnshareGroup(event) {
        unshareGroup(shareService, event)
            .then(() => {
                if (detailsStore.details?.sharedGroups) {
                    detailsStore.details.sharedGroups = detailsStore.details.sharedGroups.filter(
                        (g) => g.id !== event.groupId,
                    );
                }
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }

    function handleShareFileUpdate(event) {
        updateShare(shareService, event);
    }

    function handleLinkShare(event) {
        shareLink(shareService, event);
    }

    function handleLinkShareUpdate(event) {
        updateLinkShare(shareService, event);
    }

    return {
        handleShareFile,
        handleUnshareFile,
        handleUnshareGroup,
        handleShareFileUpdate,
        handleLinkShare,
        handleLinkShareUpdate,
    };
}
