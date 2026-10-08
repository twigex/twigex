// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { getActivePinia } from "pinia";

// Called on logout so a second user signing in on the same SPA session never
// sees the previous user's cached data.
export function resetAllStores() {
    const pinia = getActivePinia();

    if (!pinia) return;

    pinia._s.forEach((store) => store.$reset());
}
