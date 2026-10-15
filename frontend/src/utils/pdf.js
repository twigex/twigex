// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

function hasIteratorHelpers() {
    if (typeof Iterator === "undefined") {
        return false;
    }

    return typeof Iterator.prototype.toArray === "function";
}

// The modern pdf.js build calls iterator helpers, which Safari only has from
// 18.4. The legacy build carries polyfills for older browsers.
export async function loadPdfLibrary() {
    if (hasIteratorHelpers()) {
        const library = await import("pdfjs-dist");
        const worker = await import("pdfjs-dist/build/pdf.worker.mjs?url");

        library.GlobalWorkerOptions.workerSrc = worker.default;

        return library;
    }

    const library = await import("pdfjs-dist/legacy/build/pdf.mjs");
    const worker = await import("pdfjs-dist/legacy/build/pdf.worker.mjs?url");

    library.GlobalWorkerOptions.workerSrc = worker.default;

    return library;
}
