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
