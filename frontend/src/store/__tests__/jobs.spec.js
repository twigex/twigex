// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useJobsStore } from "@/store/jobs";
import { moveJobName } from "@/composables/files/useFileProgress";
import fileService from "@/services/fileService";

vi.mock("@/services/fileService", () => ({
    default: {
        getRunningJobs: vi.fn(),
        acknowledgeJob: vi.fn(),
        cancelJob: vi.fn(),
    },
}));

vi.mock("@/store/files", () => ({
    useFilesStore: () => ({ completeMoveJob: vi.fn() }),
}));

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
});

describe("jobs store", () => {
    it("drops finished jobs on dismiss and hides the running ones", () => {
        const jobs = useJobsStore();

        jobs.jobs = [
            { id: "a", status: "completed" },
            { id: "b", status: "running" },
            { id: "c", status: "failed" },
        ];

        jobs.dismiss();

        expect(jobs.jobs.map((job) => job.id)).toEqual(["b"]);
        expect(jobs.hidden).toBe(true);
    });

    it("does not hide anything when every dismissed job had finished", () => {
        const jobs = useJobsStore();

        jobs.jobs = [{ id: "a", status: "completed" }];

        jobs.dismiss();

        expect(jobs.jobs).toEqual([]);
        expect(jobs.hidden).toBe(false);
    });

    it("keeps finished jobs out while hidden", async () => {
        const jobs = useJobsStore();

        jobs.hidden = true;
        fileService.getRunningJobs.mockResolvedValue({
            data: [
                { id: "a", status: "completed" },
                { id: "b", status: "running" },
            ],
        });

        await jobs.fetchJobs();

        expect(jobs.jobs.map((job) => job.id)).toEqual(["b"]);
    });

    it("shows the card again when a new move starts", () => {
        const jobs = useJobsStore();

        jobs.hidden = true;
        fileService.getRunningJobs.mockResolvedValue({ data: [] });

        jobs.startPolling();

        expect(jobs.hidden).toBe(false);
    });
});

describe("moveJobName", () => {
    function job(payload) {
        return { type: "move_files", payload: btoa(JSON.stringify(payload)) };
    }

    it("counts the moved items", () => {
        expect(moveJobName(job({ root_ids: [1, 2], subfile_ids: [3] }))).toBe("Moving 3 items");
        expect(moveJobName(job({ root_ids: [1] }))).toBe("Moving 1 item");
    });

    it("falls back when the payload can't be read", () => {
        expect(moveJobName({ type: "move_files", payload: "not base64!" })).toBe("Moving files");
    });
});
