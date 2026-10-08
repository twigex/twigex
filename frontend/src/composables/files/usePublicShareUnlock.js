// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { extractErrorMessage } from "@/utils/errors";
import publicShareService from "@/services/publicShareService";

export function usePublicShareUnlock(token, load) {
    const password = ref("");
    const passwordError = ref("");
    const unlocking = ref(false);

    async function unlock() {
        if (unlocking.value) return;
        unlocking.value = true;
        passwordError.value = "";
        try {
            await publicShareService.authenticate(token, password.value);
            await load();
        } catch (error) {
            passwordError.value = extractErrorMessage(error);
        } finally {
            unlocking.value = false;
        }
    }

    return {
        password,
        passwordError,
        unlocking,
        unlock,
    };
}
