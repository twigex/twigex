// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useUserStore } from "@/store/user";
import moment from "moment";

export default function useDateOperations() {
    const getDate = (tstamp) => {
        if (!tstamp) {
            return "";
        }

        tstamp = isUnixMilliseconds(tstamp) ? tstamp : tstamp * 1000;

        if (useUserStore().getClockDisplay == "12h") {
            return moment(tstamp).tz(useUserStore().getTimezone).format("MM/DD/YYYY");
        }

        return moment(tstamp).tz(useUserStore().getTimezone).format("DD.MM.YYYY");
    };

    const getDateAndTime = (tstamp) => {
        if (!tstamp) return "";

        // Normalize seconds → milliseconds
        const ts = isUnixMilliseconds(tstamp) ? tstamp : tstamp * 1000;

        const userStore = useUserStore();
        const tz = userStore.getTimezone;

        return moment(ts)
            .tz(tz)
            .format(userStore.getClockDisplay === "12h" ? "MM/DD/YYYY LT" : "DD.MM.YYYY HH:mm");
    };

    const getTime = (tstamp) => {
        if (!tstamp) {
            return "";
        }

        tstamp = isUnixMilliseconds(tstamp) ? tstamp : tstamp * 1000;

        if (useUserStore().getClockDisplay == "12h") {
            return moment(tstamp).tz(useUserStore().getTimezone).format("LT");
        }

        return moment(tstamp).tz(useUserStore().getTimezone).format("HH:mm");
    };

    //Timestamp to last activity
    const getLastActivity = (tstamp) => {
        if (!tstamp) {
            return "No activity";
        }

        return moment(tstamp * 1000)
            .tz(useUserStore().getTimezone)
            .fromNow();
    };

    const getTimestampFromDateString = (dateString) => {
        if (!dateString) return null;

        // A date field comes as its day at midnight UTC. Read as a moment, it
        // would be the day before for anyone west of UTC.
        const day =
            typeof dateString === "string" &&
            dateString.match(/^(\d{4}-\d{2}-\d{2})(?:T00:00:00(?:\.0+)?Z)?$/);
        const date = day
            ? moment.tz(day[1], "YYYY-MM-DD", useUserStore().getTimezone)
            : moment.tz(dateString, useUserStore().getTimezone);

        if (!date.isValid()) return null;

        return Math.floor(date.startOf("day").valueOf() / 1000);
    };

    // A date as a date picker takes it, YYYY-MM-DD: a date field's day as it
    // is, a timestamp as its day in the user's timezone.
    const getDateInputValue = (value, isDay) => {
        if (value === null || value === undefined || value === "" || value === "0") return "";
        if (isDay) return String(value).slice(0, 10);

        const n = Number(value);

        if (!Number.isFinite(n) || n <= 0) return "";

        return moment(n > 1e12 ? n : n * 1000)
            .tz(useUserStore().getTimezone)
            .format("YYYY-MM-DD");
    };

    // The day a task's date falls on, YYYY-MM-DD, whether it is stored as a day
    // or as a timestamp, in seconds or milliseconds; null without one.
    const getDayOf = (value) => {
        if (!value) return null;
        const s = String(value).trim();

        if (/^\d{4}-\d{2}-\d{2}/.test(s)) return s.slice(0, 10);
        const n = Number(s);

        if (!isNaN(n) && n > 0) {
            const ms = n < 1e12 ? n * 1000 : n;

            return moment(ms).tz(useUserStore().getTimezone).format("YYYY-MM-DD");
        }

        return null;
    };

    // A field holding a day rather than a moment, stored as YYYY-MM-DD.
    const isDayField = (header) =>
        (header?.header_type || "").toUpperCase() === "DATE" || header?.header_usage === "date";

    const isUnixMilliseconds = (timestamp) => {
        const ts = Number(timestamp);

        return Number.isFinite(ts) && ts > 1e12;
    };

    return {
        getDate,
        getDateAndTime,
        getTime,
        getLastActivity,
        getTimestampFromDateString,
        getDateInputValue,
        getDayOf,
        isDayField,
        isUnixMilliseconds,
    };
}
