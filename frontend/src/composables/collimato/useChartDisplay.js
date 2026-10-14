// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useUserStore } from "@/store/user";

export function useChartDisplay() {
    const userStore = useUserStore();

    function chartDisplay() {
        const clock = userStore.getClockDisplay;

        return { hour12: clock === "12h" };
    }

    return { chartDisplay };
}
