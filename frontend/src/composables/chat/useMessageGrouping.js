// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Each function takes the message before it rather than searching the list:
// looking it up again turned drawing a channel into a scan per message.

// Posts carry seconds or milliseconds depending on their age.
const asMilliseconds = (timestamp) => {
    const ts = Number(timestamp);

    return Number.isFinite(ts) && ts > 1e12 ? ts : ts * 1000;
};

const utcDay = (timestamp) => {
    const d = new Date(asMilliseconds(timestamp));

    return Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate());
};

export function isFirstMessageOfDay(message, previousMessage) {
    if (!previousMessage) return true;

    return utcDay(message.created) !== utcDay(previousMessage.created);
}

export function shouldCollapse(message, previousMessage, windowMs) {
    if (!previousMessage) return false;
    if (previousMessage.user_id !== message.user_id) return false;

    return asMilliseconds(message.created) - asMilliseconds(previousMessage.created) < windowMs;
}

// `now` is a parameter so a test can pin it.
export function dateSeparatorLabel(timestamp, { locale, translate, now = new Date() }) {
    const date = new Date(asMilliseconds(timestamp));

    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);

    yesterday.setDate(today.getDate() - 1);
    const oneWeekAgo = new Date(today);

    oneWeekAgo.setDate(today.getDate() - 7);

    if (date >= today) return translate("dates.today");
    if (date >= yesterday) return translate("dates.yesterday");
    if (date >= oneWeekAgo) {
        return date.toLocaleDateString(locale, { weekday: "long" });
    }

    return date.toLocaleDateString(locale, {
        year: "numeric",
        month: "long",
        day: "numeric",
    });
}
