<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        ref="menuElement"
        v-if="show"
        class="absolute z-50 select-none"
        :style="{
            top: `${menuPosition.y}px`,
            left: `${menuPosition.x}px`,
        }"
    >
        <div class="min-w-56 rounded-md bg-white shadow-lg ring-1 ring-black/5">
            <div class="py-1 text-sm">
                <!-- file -->
                <div v-if="caller == 'file'">
                    <div
                        v-for="item in files"
                        :key="item.name"
                        class="flex items-center gap-x-2 px-3 py-2 mx-1 rounded-md cursor-pointer transition hover:bg-gray-100"
                        @click="emitEvent(item.event)"
                    >
                        <component
                            :is="item.icon"
                            :class="'h-4 w-4 flex-none ' + item.color"
                            aria-hidden="true"
                        />

                        <a :href="item.href" class="flex-1 truncate text-gray-700">
                            {{ item.name }}
                        </a>
                    </div>
                </div>

                <!-- empty -->
                <div v-if="caller == 'empty'">
                    <div
                        v-for="item in empty"
                        :key="item.name"
                        class="flex items-center gap-x-2 px-3 py-2 mx-1 rounded-md cursor-pointer transition hover:bg-gray-100"
                        @click="emitEvent(item.event)"
                    >
                        <component
                            :is="item.icon"
                            :class="'h-4 w-4 flex-none ' + item.color"
                            aria-hidden="true"
                        />

                        <a :href="item.href" class="flex-1 truncate text-gray-700">
                            {{ item.name }}
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import {
    StarIcon,
    ArrowDownOnSquareIcon,
    ArrowsPointingOutIcon,
    InformationCircleIcon,
    PencilSquareIcon,
    ArrowRightStartOnRectangleIcon,
    ShareIcon,
    TrashIcon,
    FolderPlusIcon,
    DocumentTextIcon,
} from "@heroicons/vue/20/solid";

import { onMounted, onUnmounted, computed, ref } from "vue";
import { usePermissions } from "@/composables/usePermissions";

const props = defineProps({
    show: {
        type: Boolean,
        default: false,
    },
    coordinates: {
        type: Object,
        default: () => ({
            x: 0,
            y: 0,
        }),
    },
    caller: {
        type: String,
        default: "file",
    },
    file: {
        type: Object,
        default: () => ({}),
    },
});

const emit = defineEmits([
    "close",
    "add-to-favorites",
    "download",
    "open",
    "details",
    "rename",
    "move",
    "share",
    "delete",
    "new-folder",
    "file-upload",
    "folder-upload",
    "document",
    "spreadsheet",
    "presentation",
    "text",
]);
const menuElement = ref(null);
const { can } = usePermissions();

const handleClickOutside = () => {
    emit("close");
};

const favoriteIcon = computed(() => StarIcon);

const favoriteIconColor = computed(() =>
    props.file.favourite ? "text-yellow-500" : "text-gray-600 group-hover:text-indigo-600",
);
const favoriteName = computed(() =>
    props.file.favourite
        ? t.value("files.contextmenu.remove_from_favorites")
        : t.value("files.contextmenu.add_to_favorites"),
);

const menuPosition = computed(() => {
    const menuWidth = menuElement.value?.offsetWidth + 50 || 0;
    const menuHeight = menuElement.value?.offsetHeight + 20 || 0;
    let x = props.coordinates.x;
    let y = props.coordinates.y;

    if (x + menuWidth > window.innerWidth) {
        x = window.innerWidth - menuWidth;
    }

    if (y + menuHeight > window.innerHeight) {
        y = window.innerHeight - menuHeight;
    }

    return { x, y };
});

onMounted(() => {
    window.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
    window.removeEventListener("click", handleClickOutside);
});

function emitEvent(event) {
    emit(event);
}

const files = computed(() => {
    const items = [
        {
            get name() {
                return favoriteName.value;
            },
            get icon() {
                return favoriteIcon.value;
            },
            event: "add-to-favorites",
            get color() {
                return favoriteIconColor.value;
            },
        },
        ...(can("download_files")
            ? [
                  {
                      name: t.value("files.contextmenu.download"),
                      icon: ArrowDownOnSquareIcon,
                      event: "download",
                      color: "text-gray-600 group-hover:text-indigo-600",
                  },
              ]
            : []),
        {
            name: t.value("files.contextmenu.open"),
            icon: ArrowsPointingOutIcon,
            event: "open",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
        {
            name: t.value("files.contextmenu.details"),
            icon: InformationCircleIcon,
            event: "details",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
        {
            name: t.value("files.contextmenu.rename"),
            icon: PencilSquareIcon,
            event: "rename",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
        {
            name: t.value("files.contextmenu.move"),
            icon: ArrowRightStartOnRectangleIcon,
            event: "move",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
        ...(can("share_files")
            ? [
                  {
                      name: t.value("files.contextmenu.share"),
                      icon: ShareIcon,
                      event: "share",
                      color: "text-gray-600 group-hover:text-indigo-600",
                  },
              ]
            : []),
        {
            name: t.value("files.contextmenu.delete"),
            icon: TrashIcon,
            event: "delete",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
    ];

    return items;
});

const empty = computed(() => {
    if (!can("create_files")) return [];

    return [
        {
            name: t.value("files.contextmenu.new_folder"),
            icon: FolderPlusIcon,
            event: "new-folder",
            color: "text-gray-600 group-hover:text-indigo-600",
        },
        {
            name: t.value("files.contextmenu.new_document"),
            icon: DocumentTextIcon,
            event: "document",
            color: "text-blue-600 group-hover:text-blue-600",
        },
        {
            name: t.value("files.contextmenu.new_spreadsheet"),
            icon: DocumentTextIcon,
            event: "spreadsheet",
            color: "text-green-600 group-hover:text-green-600",
        },
        {
            name: t.value("files.contextmenu.new_presentation"),
            icon: DocumentTextIcon,
            event: "presentation",
            color: "text-orange-600 group-hover:text-orange-600",
        },
    ];
});
</script>
