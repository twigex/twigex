// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRouter } from "vue-router";
import { useCollimatoStore } from "@/store/collimato";

const SECTIONS = ["dashboards", "charts", "data", "connections", "collimato-users"];

export function useWorkspaceAccess() {
    const router = useRouter();
    const collimatoStore = useCollimatoStore();

    function canEnter(to) {
        const permissions = [to.meta?.collimatoPermission ?? []].flat();

        return permissions.every((permission) => collimatoStore.can(permission));
    }

    function canOpen(name, workspaceId) {
        return canEnter(router.resolve({ name, params: { workspaceId } }));
    }

    function firstSection(workspaceId) {
        const name = SECTIONS.find((section) => canOpen(section, workspaceId));

        return { name, params: { workspaceId } };
    }

    return { canEnter, canOpen, firstSection };
}
