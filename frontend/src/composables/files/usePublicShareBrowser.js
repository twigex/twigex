// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed, watch } from "vue";
import { t } from "@/i18n";
import { extractErrorMessage } from "@/utils/errors";
import publicShareService from "@/services/publicShareService";

export function usePublicShareBrowser(token) {
    const state = ref("loading");
    const errorMessage = ref("");
    const share = ref(null);
    const rootName = ref("");
    const folderPath = ref([]);
    const failedThumbs = ref(new Set());
    const viewMode = ref(localStorage.getItem("public_share_view") || "grid");
    const sort = ref("name-asc");
    const page = ref(1);
    const navigating = ref(false);

    const isFolder = computed(() => share.value?.IsFolder ?? false);
    const children = computed(() => share.value?.Children ?? []);
    const showBrowser = computed(
        () => state.value === "ready" && isFolder.value && !!share.value?.AllowView,
    );
    const currentChildId = computed(() =>
        folderPath.value.length ? folderPath.value[folderPath.value.length - 1].id : "",
    );
    const sortedChildren = computed(() => {
        const [key, dir] = sort.value.split("-");
        const arr = [...children.value];

        arr.sort((a, b) => {
            if (a.IsFolder !== b.IsFolder) return a.IsFolder ? -1 : 1;
            let cmp = key === "size" ? (a.Size || 0) - (b.Size || 0) : a.Name.localeCompare(b.Name);

            return dir === "desc" ? -cmp : cmp;
        });

        return arr;
    });
    const perPage = computed(() => (viewMode.value === "grid" ? 24 : 50));
    const pageCount = computed(() =>
        Math.max(1, Math.ceil(sortedChildren.value.length / perPage.value)),
    );
    const pagedChildren = computed(() => {
        const start = (page.value - 1) * perPage.value;

        return sortedChildren.value.slice(start, start + perPage.value);
    });

    watch(viewMode, (mode) => {
        localStorage.setItem("public_share_view", mode);
        page.value = 1;
    });
    watch(sort, () => {
        page.value = 1;
    });

    function thumbFailed(id) {
        failedThumbs.value.add(id);
        failedThumbs.value = new Set(failedThumbs.value);
    }

    async function load() {
        const isNav = state.value === "ready";

        if (isNav) navigating.value = true;
        page.value = 1;
        try {
            const { data } = await publicShareService.getShare(token, currentChildId.value);

            share.value = data;
            if (!currentChildId.value) {
                rootName.value = data.Name;
            }

            document.title = data.Name || t.value("public_share.document_title");
            state.value = "ready";
        } catch (error) {
            if (error.response?.status === 401) {
                state.value = "password";

                return;
            }

            errorMessage.value = extractErrorMessage(error);
            state.value = "error";
        } finally {
            navigating.value = false;
        }
    }

    function openFolder(child) {
        folderPath.value.push({ id: child.ID, name: child.Name });
        load();
    }

    function navigateTo(index) {
        folderPath.value = folderPath.value.slice(0, index + 1);
        load();
    }

    return {
        state,
        errorMessage,
        share,
        rootName,
        folderPath,
        failedThumbs,
        viewMode,
        sort,
        page,
        navigating,
        isFolder,
        children,
        showBrowser,
        currentChildId,
        sortedChildren,
        pageCount,
        pagedChildren,
        thumbFailed,
        load,
        openFolder,
        navigateTo,
    };
}
