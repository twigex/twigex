// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index.js";

// The classes and style of an option's colour: a Tailwind background class
// as it is, any other colour as an inline background.
export const getStatusStyle = (status) => {
    if (!status) return { class: "bg-blue-600 hover:bg-blue-700 focus:ring-blue-500" };

    const val = status.color || "";

    if (typeof val === "string" && val.startsWith("bg-")) {
        return { class: val };
    }

    if (val) {
        return {
            class: "focus:outline-none focus:ring-1 focus:ring-offset-0",
            style: {
                backgroundColor: val,
                "--tw-ring-color": val,
            },
        };
    }

    return { class: "bg-blue-600 hover:bg-blue-700 focus:ring-blue-500" };
};

const STATUS_TRANSLATION_KEYS = {
    Completed: "projects.grid_view.status_completed",
    Cancelled: "projects.grid_view.status_cancelled",
};

// A status's name, the two built in translated.
export const getStatusDisplayName = (statusValue) => {
    if (!statusValue || !statusValue.name) return "";

    const statusName = statusValue.name;

    if (STATUS_TRANSLATION_KEYS[statusName]) {
        return t.value(STATUS_TRANSLATION_KEYS[statusName]);
    }

    return statusValue.name;
};
