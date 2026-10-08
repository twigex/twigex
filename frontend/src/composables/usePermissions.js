// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { usePermissionsStore } from "@/store/permissions";

export function usePermissions() {
    const store = usePermissionsStore();

    function can(permissionId) {
        return store.permissions.includes(permissionId);
    }

    return { can };
}
