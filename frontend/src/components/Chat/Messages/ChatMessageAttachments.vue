<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="relative inline-block">
        <div v-if="message.metadata.files.length == 1">
            <div
                v-for="(attachment, index) in message.metadata.files"
                :key="index"
                @click="openFile(attachment)"
                class="relative inline-block mr-2"
            >
                <div v-if="attachment.kind == 'gifv'" class="inline-block max-w-full">
                    <video
                        :src="attachment.url"
                        :poster="attachment.poster || ''"
                        autoplay
                        loop
                        muted
                        playsinline
                        preload="metadata"
                        class="rounded-md object-contain"
                        :style="gifSize(attachment)"
                    ></video>
                </div>

                <div v-if="attachment.kind == 'gif'" class="inline-block max-w-full">
                    <img
                        :src="attachment.url"
                        alt="GIF"
                        class="rounded-md object-contain"
                        :style="gifSize(attachment)"
                    />
                </div>

                <div
                    v-if="
                        attachment.kind == 'image' &&
                        attachment.kind != 'gifv' &&
                        attachment.kind != 'gif'
                    "
                >
                    <div class="relative">
                        <!-- 8rem = 128px = h-32, same as max-h-96 -->
                        <img
                            :src="fileUrl(attachment)"
                            alt="Attachment"
                            @load="emit('update:imageLoaded', true)"
                            class="rounded-lg max-w-full max-h-96 w-auto shadow-md transition-opacity duration-300"
                            :class="{
                                'opacity-100 h-auto': imageLoaded,
                                'opacity-0 h-96': !imageLoaded,
                            }"
                        />

                        <div
                            v-if="!imageLoaded"
                            class="absolute inset-0 bg-gray-200 animate-pulse rounded-lg"
                        ></div>
                    </div>

                    <button
                        v-if="hovered"
                        @click.stop="downloadFile(attachment)"
                        class="absolute top-5 right-5 transform translate-x-1/2 -translate-y-1/2 text-gray-500 bg-white hover:bg-gray-100 bg-opacity-80 p-2 rounded-md"
                    >
                        <ArrowDownTrayIcon class="h-4 w-4" aria-hidden="true" />
                    </button>
                </div>

                <div v-else-if="attachment.kind != 'gifv' && attachment.kind != 'gif'">
                    <div
                        @click.stop="openFile(attachment)"
                        class="rounded group border p-2 flex flex-row items-center w-96 h-16 hover:bg-gray-50 hover:shadow-md gap-x-2 m-1 justify-between"
                    >
                        <div class="flex flex-row gap-x-2 items-center w-72">
                            <component
                                :is="
                                    fileTypes[attachment.mime_type]?.icon ??
                                    fileTypes['not-found'].icon
                                "
                                :class="
                                    'h-10 w-10 flex-none ' +
                                    (fileTypes[attachment.mime_type]?.color ??
                                        fileTypes['not-found'].color)
                                "
                            />

                            <div class="truncate">
                                {{ attachment.name }}
                            </div>
                        </div>

                        <button
                            @click.stop="downloadFile(attachment)"
                            class="relative text-gray-500 bg-gray-50 hover:bg-gray-200 bg-opacity-80 p-2 rounded-md opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                        >
                            <ArrowDownTrayIcon
                                @click.stop="downloadFile(attachment)"
                                class="h-4 w-4 text-gray-500 hover:text-gray-700 cursor-pointer"
                                aria-hidden="true"
                            />
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-else>
            <div
                v-for="(attachment, index) in message.metadata.files.filter(
                    (file) => file.mime_type != 'image/gif',
                )"
                :key="index"
                class="relative inline-block mr-2"
            >
                <div
                    @click="openFile(attachment)"
                    class="rounded group border p-2 flex flex-row items-center w-96 h-16 hover:bg-gray-50 hover:shadow-md gap-x-2 m-1 justify-between"
                >
                    <div class="flex flex-row gap-x-2 items-center w-72">
                        <img
                            v-if="isImageFile(attachment)"
                            :src="fileUrl(attachment)"
                            alt="Attachment"
                            class="w-10 h-10 object-cover rounded shadow-md"
                        />

                        <div v-else>
                            <component
                                :is="
                                    fileTypes[attachment.mime_type]?.icon ??
                                    fileTypes['not-found'].icon
                                "
                                :class="
                                    'h-10 w-10 flex-none ' +
                                    (fileTypes[attachment.mime_type]?.color ??
                                        fileTypes['not-found'].color)
                                "
                            />
                        </div>

                        <div class="truncate">
                            {{ attachment.name }}
                        </div>
                    </div>

                    <button
                        @click.stop="downloadFile(attachment)"
                        class="relative text-gray-500 bg-gray-50 hover:bg-gray-200 bg-opacity-80 p-2 rounded-md opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                    >
                        <ArrowDownTrayIcon
                            @click.stop="downloadFile(attachment)"
                            class="h-4 w-4 text-gray-500 hover:text-gray-700 cursor-pointer"
                            aria-hidden="true"
                        />
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ArrowDownTrayIcon } from "@heroicons/vue/20/solid";
import fileTypes from "@/constants/fileTypes";
import officeService from "@/services/officeService";
import { useMediaStore } from "@/store/media";
import { useOfficeStore } from "@/store/office";

const props = defineProps({
    message: {
        type: Object,
        required: true,
    },
    hovered: {
        type: Boolean,
        default: false,
    },
    imageLoaded: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:imageLoaded"]);

const mediaStore = useMediaStore();
const officeStore = useOfficeStore();

const fileUrl = (attachment) =>
    `/api/channels/${props.message.channel_id}/posts/${props.message.id}/file/${attachment.id}`;

const isImageFile = (file) =>
    (file && typeof file === "string") ||
    (file && file.mime_type && file.mime_type.startsWith("image/"));

const downloadFile = (attachment) => {
    const filePath = fileUrl(attachment);

    // Create a temporary anchor tag to trigger the download
    const a = document.createElement("a");

    a.href = filePath;
    a.download = attachment.name || ""; // Optional: specify a default filename
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
};

function openFile(file) {
    if (
        file.mime_type.includes("image") ||
        file.mime_type.includes("video") ||
        file.mime_type.includes("pdf")
    ) {
        let fileTypes = [];

        props.message.metadata.files.forEach((file) => {
            if (
                file.mime_type.includes("image") ||
                file.mime_type.includes("video") ||
                file.mime_type.includes("pdf")
            ) {
                let url = window.location.origin + fileUrl(file);

                if (file.kind == "gifv" || file.kind == "gif") {
                    url = file.url;
                }

                fileTypes.push({
                    id: file.id,
                    name: file.name,
                    type: file.mime_type,
                    kind: file.kind,
                    url: url,
                });
            }
        });

        let currentIndex = fileTypes.findIndex((f) => f.id === file.id);

        setTimeout(() => {
            mediaStore.openViewer(fileTypes, currentIndex);
        }, 100);

        return;
    }

    officeService
        .supportFileType(file.mime_type)
        .then((res) => {
            if (res.data) {
                officeService.openFile(file.id, "chat").then((res) => {
                    officeStore.openOffice(res.data);
                });
            } else {
                downloadFile(file);
            }
        })
        .catch(() => {
            downloadFile(file);
        });
}

function gifSize(attachment) {
    if (attachment.width && attachment.height) {
        //if vertical allow height up to 300px else 200px
        let heightLimit = 200;

        if (attachment.height > attachment.width) {
            heightLimit = 300;
        }

        return {
            width: "auto",
            height: heightLimit + "px",
        };
    }

    // Fallback when width/height unknown
    return {
        width: "auto",
        height: "auto",
    };
}
</script>
