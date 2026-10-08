// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index";
import { BYTES_PER_GB, convertSize } from "@/utils/utils";

// 0 is unlimited, matching the server.
const STORAGE_UNLIMITED = 0;

const storageLimitOptions = [
    { value: 5 * BYTES_PER_GB, label: "5 GB" },
    { value: 10 * BYTES_PER_GB, label: "10 GB" },
    { value: 20 * BYTES_PER_GB, label: "20 GB" },
    { value: 50 * BYTES_PER_GB, label: "50 GB" },
    { value: 100 * BYTES_PER_GB, label: "100 GB" },
    { value: STORAGE_UNLIMITED, label: null },
];

// A limit set outside the picker still has to read as something.
function storageLabel(value) {
    if (value === STORAGE_UNLIMITED) {
        return t.value("settings.edit_user.storage_unlimited");
    }

    const known = storageLimitOptions.find((o) => o.value === value);

    if (known) {
        return known.label;
    }

    return convertSize(value);
}

export { STORAGE_UNLIMITED, storageLimitOptions, storageLabel };
