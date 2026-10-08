// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useUploadStore } from "@/store/upload";
import { useJobsStore } from "@/store/jobs";
import { useFileProgress } from "@/composables/files/useFileProgress";

vi.mock("@/services/fileService", () => ({
    default: {},
}));

beforeEach(() => {
    setActivePinia(createPinia());
});

describe("useFileProgress errors", () => {
    it("keeps the server's message on a failed upload", () => {
        const uploads = useUploadStore();

        uploads.files = [{ id: "f1", name: "video.mov", status: "uploading" }];

        uploads.markError("f1", { response: { data: { error: "Storage quota exceeded" } } });

        const { uploads: items } = useFileProgress();

        expect(items.value[0].error).toBe("Storage quota exceeded");
    });

    it("falls back to a translated message when the server gave none", () => {
        const uploads = useUploadStore();

        uploads.files = [{ id: "f1", name: "video.mov", status: "uploading" }];

        uploads.markError("f1", {});

        expect(uploads.files[0].error).toBe("Upload failed");
    });

    it("counts failed uploads in the title", () => {
        const uploads = useUploadStore();

        uploads.files = [
            { id: "a", status: "completed" },
            { id: "b", status: "error" },
            { id: "c", status: "uploading" },
        ];

        const { uploadTitle } = useFileProgress();

        expect(uploadTitle.value).toBe("Uploaded 1 of 3, 1 failed");
    });

    it("explains a failed move without showing the raw server error", () => {
        const jobs = useJobsStore();

        jobs.jobs = [{ id: "j1", type: "move_files", status: "failed", error: "sql: no rows" }];

        const { moves, moveTitle } = useFileProgress();

        expect(moves.value[0].error).toBe("Couldn't move these items");
        expect(moveTitle.value).toBe("Move failed");
    });
});
