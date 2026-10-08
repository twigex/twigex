// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import fileTypes from "@/constants/fileTypes";

const FAMILIES = {
    image: "image/png",
    video: "video/mp4",
    audio: "audio/mpeg",
};

// The icon and colour Files shows for a file's type. A type Files does not
// list takes its family's, an image as an image.
export function fileTypeOf(file) {
    const type = file?.type || "";

    return fileTypes[type] ?? fileTypes[FAMILIES[type.split("/")[0]]] ?? fileTypes["not-found"];
}

// As in Files, the media viewer shows images, videos and PDFs only; any other
// file opens in the office editor or is downloaded.
export function isViewable(file) {
    const type = file?.type || "";

    return type.includes("image") || type.includes("video") || type.includes("pdf");
}
