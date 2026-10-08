// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function findItem(tosearch, attr, array) {
    return array.reduce((a, item) => {
        if (a) return a;
        switch (tosearch) {
            case "id":
                if (attr == item.id) return item;
                break;
            case "path":
                if (attr == item.path) return item;
                break;
            case "array":
                if (attr == item.id) return array;
                break;
            case "name":
                if (attr == item.name) return item;
                break;
            case "title":
                if (attr == item.title) return item;
                break;
            default:
                break;
        }

        if (item.children) return findItem(tosearch, attr, item.children);
    }, null);
}

// Binary, not decimal: a standard would call this a GiB. The storage limit is
// counted the same way, so a limit and the files filling it agree on screen.
export const BYTES_PER_GB = 1024 * 1024 * 1024;

export function convertSize(size) {
    var sizeType = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

    if (size == 0) {
        return "0 " + sizeType[0];
    }

    var i = parseInt(Math.floor(Math.log(size) / Math.log(1024)));

    return (size / Math.pow(1024, i)).toFixed(1) + " " + sizeType[i];
}

function readCookie(name) {
    const cookies = document.cookie.split(";");

    for (let i = 0; i < cookies.length; i++) {
        const cookie = cookies[i].trim();

        if (cookie.startsWith(name + "=")) {
            return cookie.slice(name.length + 1);
        }
    }

    return "";
}

export function getCsrfFromCookie() {
    return readCookie("tw_csrf");
}

export function hasSessionCookie() {
    return readCookie("tw_logged_in") !== "";
}

export function hexToRGB(hex, alpha) {
    var r = parseInt(hex.slice(1, 3), 16),
        g = parseInt(hex.slice(3, 5), 16),
        b = parseInt(hex.slice(5, 7), 16);

    if (alpha) {
        return [r, g, b, alpha];
    } else {
        return [r, g, b];
    }
}
