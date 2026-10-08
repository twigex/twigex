// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

/**
 * Extracts a user-friendly error message from an Axios error object.
 * Handles JSON responses ({"error": "msg"}), plain text responses, and network errors.
 *
 * @param {Error} error - The Axios error object
 * @param {string} [fallback] - Optional fallback message if no error can be extracted
 * @returns {string}
 */
export function extractErrorMessage(error, fallback = "An unexpected error occurred") {
    if (error?.response?.data?.error) {
        return error.response.data.error;
    }

    if (typeof error?.response?.data === "string" && error.response.data.trim()) {
        return error.response.data;
    }

    if (error?.message) {
        return error.message;
    }

    return fallback;
}
