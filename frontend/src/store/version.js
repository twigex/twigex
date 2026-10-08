// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

function parseVersion(v) {
    const [core, pre = ""] = String(v).split("-");
    const nums = core.split(".").map((n) => parseInt(n, 10) || 0);

    return { nums, pre };
}

// Guards against non-release builds ("dev", "unknown") so they never trip the version gate.
function isRealVersion(v) {
    return /^\d/.test(String(v));
}

// A release outranks the same core version carrying a pre-release suffix (0.13.0 > 0.13.0-rc2).
function compareVersions(a, b) {
    const pa = parseVersion(a);
    const pb = parseVersion(b);

    const len = Math.max(pa.nums.length, pb.nums.length);

    for (let i = 0; i < len; i++) {
        const diff = (pa.nums[i] || 0) - (pb.nums[i] || 0);

        if (diff !== 0) return diff < 0 ? -1 : 1;
    }

    if (pa.pre === pb.pre) return 0;
    if (!pa.pre) return 1;
    if (!pb.pre) return -1;

    return pa.pre < pb.pre ? -1 : 1;
}

export const useVersionStore = defineStore("version", {
    state: () => ({
        bootDeploy: null,
        bootVersion: null,
        latestDeploy: null,
        dismissedDeploy: null,
        updateRequired: false,
    }),

    getters: {
        updateAvailable: (state) =>
            state.bootDeploy !== null &&
            state.latestDeploy !== null &&
            state.latestDeploy > state.bootDeploy,

        // A dismissal hides only the deploy it was made on; a strictly newer deploy brings the banner back.
        showBanner() {
            if (this.updateRequired) return true;
            if (!this.updateAvailable) return false;

            return this.dismissedDeploy === null || this.latestDeploy > this.dismissedDeploy;
        },
    },

    actions: {
        // Fed by the response interceptor's headers; the first call sets the baseline the rest compare against.
        check({ deploy, version, minVersion }) {
            const parsedDeploy = parseInt(deploy, 10);

            if (!Number.isNaN(parsedDeploy)) {
                if (this.bootDeploy === null) {
                    this.bootDeploy = parsedDeploy;
                    this.latestDeploy = parsedDeploy;
                } else if (parsedDeploy > this.latestDeploy) {
                    this.latestDeploy = parsedDeploy;
                }
            }

            if (this.bootVersion === null && version) {
                this.bootVersion = version;
            }

            if (
                isRealVersion(this.bootVersion) &&
                isRealVersion(minVersion) &&
                compareVersions(this.bootVersion, minVersion) < 0
            ) {
                this.updateRequired = true;
            }
        },

        dismiss() {
            this.dismissedDeploy = this.latestDeploy;
        },
    },
});
