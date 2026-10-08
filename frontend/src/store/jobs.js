// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import fileService from "@/services/fileService";
import { useFilesStore } from "@/store/files";

function isActive(job) {
    return job.status === "pending" || job.status === "running";
}

export const useJobsStore = defineStore("jobs", {
    state: () => ({
        jobs: [], // raw jobs from backend
        loading: false,
        poller: null,
        hidden: false,
    }),

    getters: {},

    actions: {
        async fetchJobs() {
            this.loading = true;
            try {
                const res = await fileService.getRunningJobs();
                const incoming = res.data;

                const byId = new Map(this.jobs.map((j) => [j.id, j]));

                for (const job of incoming) {
                    if (job.status == "completed") {
                        //pass down to file store to resolve this
                        useFilesStore().completeMoveJob(job.payload);
                        await fileService.acknowledgeJob(job.id);
                    }

                    byId.set(job.id, job);
                }

                this.jobs = Array.from(byId.values());

                if (this.hidden) {
                    this.jobs = this.jobs.filter(isActive);
                }
            } finally {
                this.loading = false;
            }
        },

        startPolling(interval = 3000) {
            this.hidden = false;

            if (this.poller) return;

            const poll = async () => {
                await this.fetchJobs();

                if (this.jobs.length === 0) {
                    this.stopPolling();

                    return;
                }

                if (!this.jobs.some(isActive)) {
                    this.stopPolling();

                    return;
                }

                this.poller = setTimeout(poll, interval);
            };

            poll();
        },

        stopPolling() {
            if (this.poller) {
                clearTimeout(this.poller);
                this.poller = null;
            }
        },

        cancelJob(id) {
            fileService.cancelJob(id);
        },

        dismiss() {
            this.jobs = this.jobs.filter(isActive);
            this.hidden = this.jobs.length > 0;
        },
    },
});
