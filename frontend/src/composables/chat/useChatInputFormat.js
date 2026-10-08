// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { nextTick } from "vue";

export function useChatInputFormat({ text, input }) {
    function applyFormat(type) {
        const el = input();

        if (!el) return;

        const start = el.selectionStart;
        const end = el.selectionEnd;
        const hasSelection = start !== end;
        const full = text.value;
        const selected = full.slice(start, end);

        if (type === "ul" || type === "quote" || type === "ol") {
            const linePrefix = type === "ul" ? "- " : type === "quote" ? "> " : null;
            const isOlLine = (l) => /^\d+\. /.test(l);

            if (hasSelection) {
                const lines = selected.split("\n");

                if (type === "ol") {
                    const allNumbered = lines.every(isOlLine);
                    const replaced = allNumbered
                        ? lines.map((l) => l.replace(/^\d+\. /, "")).join("\n")
                        : lines.map((l, i) => `${i + 1}. ${l}`).join("\n");

                    text.value = full.slice(0, start) + replaced + full.slice(end);
                    nextTick(() => {
                        el.selectionStart = start;
                        el.selectionEnd = start + replaced.length;
                        el.focus();
                    });
                } else {
                    const allPrefixed = lines.every((l) => l.startsWith(linePrefix));
                    const replaced = allPrefixed
                        ? lines.map((l) => l.slice(linePrefix.length)).join("\n")
                        : lines.map((l) => linePrefix + l).join("\n");

                    text.value = full.slice(0, start) + replaced + full.slice(end);
                    nextTick(() => {
                        el.selectionStart = start;
                        el.selectionEnd = start + replaced.length;
                        el.focus();
                    });
                }
            } else {
                // No selection, toggle on the current line
                const lineStart = full.lastIndexOf("\n", start - 1) + 1;
                const lineEnd =
                    full.indexOf("\n", start) === -1 ? full.length : full.indexOf("\n", start);
                const line = full.slice(lineStart, lineEnd);

                let newLine, delta;

                if (type === "ol") {
                    if (isOlLine(line)) {
                        const stripped = line.replace(/^\d+\. /, "");

                        delta = stripped.length - line.length; // negative
                        newLine = stripped;
                    } else {
                        newLine = "1. " + line;
                        delta = 3;
                    }
                } else {
                    if (line.startsWith(linePrefix)) {
                        newLine = line.slice(linePrefix.length);
                        delta = -linePrefix.length;
                    } else {
                        newLine = linePrefix + line;
                        delta = linePrefix.length;
                    }
                }

                text.value = full.slice(0, lineStart) + newLine + full.slice(lineEnd);
                nextTick(() => {
                    const cursor = Math.max(lineStart, start + delta);

                    el.selectionStart = cursor;
                    el.selectionEnd = cursor;
                    el.focus();
                });
            }

            return;
        }

        const wrapMap = {
            bold: ["**", "**"],
            italic: ["*", "*"],
            strikethrough: ["~~", "~~"],
            code: ["`", "`"],
            codeblock: ["```\n", "\n```"],
        };
        const [prefix, suffix] = wrapMap[type] ?? ["", ""];

        // Toggle off: markers are just outside the current selection
        if (
            hasSelection &&
            full.slice(start - prefix.length, start) === prefix &&
            full.slice(end, end + suffix.length) === suffix
        ) {
            text.value =
                full.slice(0, start - prefix.length) + selected + full.slice(end + suffix.length);
            nextTick(() => {
                el.selectionStart = start - prefix.length;
                el.selectionEnd = end - prefix.length;
                el.focus();
            });

            return;
        }

        // Toggle off: markers are inside the selection
        if (
            hasSelection &&
            selected.startsWith(prefix) &&
            selected.endsWith(suffix) &&
            selected.length > prefix.length + suffix.length
        ) {
            const inner = selected.slice(prefix.length, selected.length - suffix.length);

            text.value = full.slice(0, start) + inner + full.slice(end);
            nextTick(() => {
                el.selectionStart = start;
                el.selectionEnd = start + inner.length;
                el.focus();
            });

            return;
        }

        // Apply
        if (hasSelection) {
            text.value = full.slice(0, start) + prefix + selected + suffix + full.slice(end);
            nextTick(() => {
                el.selectionStart = start + prefix.length;
                el.selectionEnd = end + prefix.length;
                el.focus();
            });
        } else {
            text.value = full.slice(0, start) + prefix + suffix + full.slice(start);
            nextTick(() => {
                const cursor = start + prefix.length;

                el.selectionStart = cursor;
                el.selectionEnd = cursor;
                el.focus();
            });
        }
    }

    return { applyFormat };
}
