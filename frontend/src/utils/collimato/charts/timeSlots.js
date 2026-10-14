// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

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
        case "week": {
            const dayOfWeek = date.getUTCDay(); // 0 (Sunday) to 6 (Saturday)
            const daysToMonday = (dayOfWeek === 0 ? -6 : 1) - dayOfWeek; // Adjust to Monday

            if (daysToMonday !== 0) {
                date.setUTCDate(date.getUTCDate() + daysToMonday); // Align to Monday
            }

            date.setUTCDate(date.getUTCDate() + 7); // Increment by 7 days
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        }

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

export function generateTimeArray(startTime, endTime, granularity, withTime = true) {
    const startDate = new Date(startTime + "Z"); // Ensure UTC by adding 'Z'
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
