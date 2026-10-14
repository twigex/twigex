// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { currentLocale } from "@/i18n/index";

const DATE_LOCALES = { en: "en-GB", lv: "lv", kk: "kk", pl: "pl" };

const DAY_FORMAT = { day: "numeric", month: "short", year: "numeric", timeZone: "UTC" };
const MONTH_FORMAT = { month: "short", year: "numeric", timeZone: "UTC" };
const YEAR_FORMAT = { year: "numeric", timeZone: "UTC" };

function incrementDate(date, granularity) {
    switch (granularity) {
        case "second":
            date.setUTCSeconds(date.getUTCSeconds() + 1);
            break;
        case "minute":
            date.setUTCMinutes(date.getUTCMinutes() + 1);
            break;
        case "hour":
            date.setUTCHours(date.getUTCHours() + 1);
            break;
        case "day":
            date.setUTCDate(date.getUTCDate() + 1);
            date.setUTCHours(0, 0, 0, 0); // Reset hours, minutes, seconds, and milliseconds
            break;
        case "week":
            date.setUTCDate(date.getUTCDate() + 7); // Increment by 7 days
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;

        case "month":
            date.setUTCMonth(date.getUTCMonth() + 1);
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        case "quarter":
            date.setUTCMonth(date.getUTCMonth() + 3);
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        case "year":
            date.setUTCFullYear(date.getUTCFullYear() + 1);
            date.setUTCMonth(0, 1); // Reset to January 1
            date.setUTCHours(0, 0, 0, 0);
            break;
        default:
            throw new Error("Unsupported granularity");
    }
}

// Cube starts each bucket at the beginning of its period (Monday, the 1st of
// the month), so the first slot must start there too, not on the range's
// first day.
function startOfPeriod(date, granularity) {
    const start = new Date(date.getTime());

    switch (granularity) {
        case "second":
            start.setUTCMilliseconds(0);
            break;
        case "minute":
            start.setUTCSeconds(0, 0);
            break;
        case "hour":
            start.setUTCMinutes(0, 0, 0);
            break;
        case "day":
            start.setUTCHours(0, 0, 0, 0);
            break;
        case "week": {
            const dayOfWeek = start.getUTCDay();
            let daysSinceMonday = dayOfWeek - 1;

            if (dayOfWeek === 0) {
                daysSinceMonday = 6;
            }

            start.setUTCDate(start.getUTCDate() - daysSinceMonday);
            start.setUTCHours(0, 0, 0, 0);
            break;
        }

        case "month":
            start.setUTCDate(1);
            start.setUTCHours(0, 0, 0, 0);
            break;
        case "quarter": {
            const month = start.getUTCMonth();
            const monthsIntoQuarter = month % 3;
            const firstMonthOfQuarter = month - monthsIntoQuarter;

            start.setUTCMonth(firstMonthOfQuarter, 1);
            start.setUTCHours(0, 0, 0, 0);
            break;
        }

        case "year":
            start.setUTCMonth(0, 1);
            start.setUTCHours(0, 0, 0, 0);
            break;
    }

    return start;
}

export function generateTimeArray(startTime, endTime, granularity, withTime = true) {
    const rangeStart = new Date(startTime + "Z"); // Ensure UTC by adding 'Z'
    const startDate = startOfPeriod(rangeStart, granularity);
    const endDate = new Date(endTime + "Z"); // Ensure UTC by adding 'Z'
    const times = [];

    for (
        let currentDate = new Date(startDate.getTime());
        currentDate <= endDate;
        incrementDate(currentDate, granularity)
    ) {
        if (withTime) {
            // Include full timestamp without the trailing 'Z'
            times.push(currentDate.toISOString().slice(0, -1));
        } else {
            // Include only the date part
            times.push(currentDate.toISOString().split("T")[0]);
        }
    }

    return times;
}

function slotDate(slot) {
    if (slot.includes("T")) {
        return new Date(slot + "Z");
    }

    return new Date(slot + "T00:00:00.000Z");
}

function endOfPeriod(start, granularity) {
    const next = new Date(start.getTime());

    incrementDate(next, granularity);

    return new Date(next.getTime() - 1);
}

function dateLocale() {
    const locale = DATE_LOCALES[currentLocale.value];

    if (!locale) {
        return "en-GB";
    }

    return locale;
}

function timeFormat(granularity, hour12) {
    const format = {
        ...DAY_FORMAT,
        hour: "2-digit",
        minute: "2-digit",
        hour12: hour12,
    };

    if (granularity === "second") {
        format.second = "2-digit";
    }

    return format;
}

function wholePeriodLabel(start, granularity, locale) {
    if (granularity === "month") {
        return new Intl.DateTimeFormat(locale, MONTH_FORMAT).format(start);
    }

    if (granularity === "quarter") {
        const quarter = Math.floor(start.getUTCMonth() / 3) + 1;
        const year = start.getUTCFullYear();

        return `Q${quarter} ${year}`;
    }

    if (granularity === "year") {
        return new Intl.DateTimeFormat(locale, YEAR_FORMAT).format(start);
    }

    return null;
}

// The first and last slot can reach outside the chosen range, but Cube only
// counts the days inside it, so the label names only those days.
function timeSlotLabel(slot, granularity, rangeStart, rangeEnd, display) {
    const locale = dateLocale();
    const start = slotDate(slot);

    if (granularity === "second" || granularity === "minute" || granularity === "hour") {
        const format = timeFormat(granularity, display.hour12);

        return new Intl.DateTimeFormat(locale, format).format(start);
    }

    const days = new Intl.DateTimeFormat(locale, DAY_FORMAT);

    if (granularity === "day") {
        return days.format(start);
    }

    const end = endOfPeriod(start, granularity);
    const startsInside = start >= rangeStart;
    const endsInside = end <= rangeEnd;

    if (startsInside && endsInside) {
        const whole = wholePeriodLabel(start, granularity, locale);

        if (whole) {
            return whole;
        }
    }

    let shownStart = start;
    let shownEnd = end;

    if (!startsInside) {
        shownStart = rangeStart;
    }

    if (!endsInside) {
        shownEnd = rangeEnd;
    }

    return days.formatRange(shownStart, shownEnd);
}

export function timeSlotLabels(slots, granularity, dateRange, display = {}) {
    const rangeStart = new Date(dateRange[0] + "Z");
    const rangeEnd = new Date(dateRange[1] + "Z");

    return slots.map((slot) => timeSlotLabel(slot, granularity, rangeStart, rangeEnd, display));
}
