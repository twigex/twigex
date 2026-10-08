// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { extractErrorMessage } from "@/utils/errors";
import { isOfficeType, previewable } from "@/utils/files/publicShareTypes";
import publicShareService from "@/services/publicShareService";
import { useAlertStore } from "@/store/alerts";
import { useOfficeStore } from "@/store/office";
import { useMediaStore } from "@/store/media";

export function usePublicShareOpen({ token, share, sortedChildren, openFolder }) {
    const alertStore = useAlertStore();
    const officeStore = useOfficeStore();
    const mediaStore = useMediaStore();

    function onChildClick(child) {
        // Touch has no dblclick, so below lg a tap opens; pointer devices open on dblclick.
        if (window.matchMedia("(max-width: 1023px)").matches) {
            openChild(child);
        }
    }

    function openChild(child) {
        if (child.IsFolder) {
            openFolder(child);

            return;
        }

        if (previewable(child.Type)) {
            openPreview(child);

            return;
        }

        if (isOfficeType(child.Type)) {
            openInOffice(child.ID);

            return;
        }

        if (child.AllowDownload) {
            downloadChild(child);
        }
    }

    function downloadChild(child) {
        const a = document.createElement("a");

        a.href = publicShareService.downloadUrl(token, child.ID);
        a.download = child.Name || "";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    }

    function openPreview(child) {
        const items = sortedChildren.value
            .filter((c) => !c.IsFolder && previewable(c.Type))
            .map((c) => ({
                id: c.ID,
                name: c.Name,
                type: c.Type,
                url: window.location.origin + publicShareService.viewUrl(token, c.ID),
            }));
        const index = items.findIndex((i) => i.id === child.ID);

        mediaStore.openViewer(items, index < 0 ? 0 : index);
    }

    function previewRoot() {
        mediaStore.openViewer(
            [
                {
                    id: share.value.ID,
                    name: share.value.Name,
                    type: share.value.Type,
                    url: window.location.origin + publicShareService.viewUrl(token),
                },
            ],
            0,
        );
    }

    async function openInOffice(childId = "") {
        try {
            const { data } = await publicShareService.getOffice(token, childId);

            officeStore.openOffice(data);
        } catch (error) {
            alertStore.showError(extractErrorMessage(error));
        }
    }

    return {
        onChildClick,
        openChild,
        previewRoot,
        openInOffice,
    };
}
