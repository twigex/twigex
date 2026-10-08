// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { reactive } from "vue";
import { setActivePinia, createPinia } from "pinia";
import { useFilesStore } from "@/store/files";
import { useFileBrowserFileList } from "@/composables/files/useFileBrowserFileList";

vi.mock("@/services/fileService", () => ({
    default: {},
}));

const files = [
    { id: "b", name: "Beta", created: "2026-01-02" },
    { id: "a", name: "Alpha", created: "2026-01-03" },
    { id: "c", name: "Gamma", created: "2026-01-01" },
];

beforeEach(() => {
    setActivePinia(createPinia());
});

describe("fileList", () => {
    it("sorts without reordering the store", () => {
        const filesStore = useFilesStore();

        filesStore.files = [...files];
        filesStore.fileOrder = { order: "asc", type: "name" };

        const state = reactive({
            currentPage: 1,
            itemsPerPage: 20,
        });

        const { fileList } = useFileBrowserFileList({ state });

        expect(fileList.value.map((f) => f.id)).toEqual(["a", "b", "c"]);
        expect(filesStore.files.map((f) => f.id)).toEqual(["b", "a", "c"]);
    });

    it("defaults to newest first and pages the sorted list", () => {
        const filesStore = useFilesStore();

        filesStore.files = [...files];

        const state = reactive({
            currentPage: 2,
            itemsPerPage: 2,
        });

        const { fileList } = useFileBrowserFileList({ state });

        expect(fileList.value.map((f) => f.id)).toEqual(["c"]);
    });
});
