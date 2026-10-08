// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import {
    DocumentIcon,
    FolderIcon,
    FilmIcon,
    CalendarIcon,
    CodeBracketIcon,
    ArchiveBoxIcon,
    PhotoIcon,
    MusicalNoteIcon,
} from "@heroicons/vue/24/solid";

const fileTypes = {
    "inode/directory": {
        icon: FolderIcon,
        fileType: "folder",
        color: "text-indigo-500",
    },
    "cloud#drive": {
        icon: FolderIcon,
        fileType: "folder",
        color: "text-gray-500",
    },
    "application/pdf": {
        icon: DocumentIcon,
        fileType: "pdf",
        color: "text-red-500",
    },
    "image/png": {
        icon: PhotoIcon,
        fileType: "image",
        color: "text-yellow-500",
    },
    "image/jpeg": {
        icon: PhotoIcon,
        fileType: "image",
        color: "text-yellow-500",
    },
    "image/gif": {
        icon: PhotoIcon,
        fileType: "image",
        color: "text-yellow-500",
    },
    "image/bmp": {
        icon: PhotoIcon,
        fileType: "image",
        color: "text-yellow-500",
    },
    "image/tiff": {
        icon: PhotoIcon,
        fileType: "image",
        color: "text-yellow-500",
    },
    "image/svg+xml": {
        icon: PhotoIcon,
        fileType: "svg file",
        color: "text-yellow-500",
    },
    "image/x-canon-cr2": {
        icon: PhotoIcon,
        fileType: "raw file",
        color: "text-yellow-500",
    },
    "text/plain": {
        icon: DocumentIcon,
        fileType: "text",
        color: "text-gray-500",
    },
    "text/x-go": {
        icon: CodeBracketIcon,
        fileType: "text",
        color: "text-blue-500",
    },
    "text/x-csrc": {
        icon: CodeBracketIcon,
        fileType: "text",
        color: "text-blue-500",
    },
    "text/css": {
        icon: CodeBracketIcon,
        fileType: "text",
        color: "text-blue-500",
    },
    "text/html": {
        icon: CodeBracketIcon,
        fileType: "text",
        color: "text-orange-500",
    },
    "text/csv": {
        icon: DocumentIcon,
        fileType: "text",
        color: "text-green-500",
    },
    "text/calendar": {
        icon: CalendarIcon,
        fileType: "text",
        color: "text-blue-500",
    },
    "text/javascript": {
        icon: CodeBracketIcon,
        fileType: "text",
        color: "text-yellow-500",
    },
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {
        icon: DocumentIcon,
        fileType: "excel",
        color: "text-green-500",
    },
    "application/vnd.ms-excel": {
        icon: DocumentIcon,
        fileType: "excel",
        color: "text-green-500",
    },
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": {
        icon: DocumentIcon,
        fileType: "word",
        color: "text-blue-500",
    },
    "application/msword": {
        icon: DocumentIcon,
        fileType: "word",
        color: "text-blue-500",
    },
    "application/vnd.openxmlformats-officedocument.presentationml.presentation": {
        icon: DocumentIcon,
        fileType: "powerpoint",
        color: "text-orange-500",
    },
    "application/vnd.ms-powerpoint": {
        icon: DocumentIcon,
        fileType: "powerpoint",
        color: "text-orange-500",
    },
    "application/x-cd-image": {
        icon: DocumentIcon,
        fileType: "x-cd-image",
        color: "text-gray-500",
    },
    "application/x-raw-disk-image": {
        icon: DocumentIcon,
        fileType: "x-cd-image",
        color: "text-gray-500",
    },
    "application/vnd.debian.binary-package": {
        icon: ArchiveBoxIcon,
        fileType: "debian package",
        color: "text-gray-500",
    },
    "application/zip": {
        icon: ArchiveBoxIcon,
        fileType: "zip archive",
        color: "text-gray-500",
    },
    "application/gzip": {
        icon: ArchiveBoxIcon,
        fileType: "gzip archive",
        color: "text-gray-500",
    },
    "application/x-xz-compressed-tar": {
        icon: ArchiveBoxIcon,
        fileType: "tar archive",
        color: "text-gray-500",
    },
    "audio/aac": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "audio/mpeg": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "audio/ogg": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "audio/wav": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "audio/webm": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "audio/x-wav": {
        icon: MusicalNoteIcon,
        fileType: "audio",
        color: "text-purple-500",
    },
    "video/mp4": {
        icon: FilmIcon,
        fileType: "video",
        color: "text-red-500",
    },
    "video/x-msvideo": {
        icon: FilmIcon,
        fileType: "video",
        color: "text-red-500",
    },
    "video/mpeg": {
        icon: FilmIcon,
        fileType: "video",
        color: "text-red-500",
    },
    "video/ogg": {
        icon: FilmIcon,
        fileType: "video",
        color: "text-red-500",
    },
    "video/webm": {
        icon: FilmIcon,
        fileType: "video",
        color: "text-red-500",
    },
    "not-found": {
        icon: DocumentIcon,
        fileType: "unknown",
        color: "text-gray-500",
    },
};

export default fileTypes;
