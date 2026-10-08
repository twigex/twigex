<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <Menu
            v-if="can('upload_files') || can('create_files')"
            as="div"
            class="z-50 relative inline-block text-left"
        >
            <div>
                <MenuButton
                    :disabled="route.name != 'file'"
                    class="z-10 rounded-full bg-indigo-600 p-2 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    :class="route.name != 'file' ? 'opacity-50 cursor-default' : ''"
                >
                    <PlusIcon class="h-5 w-5" aria-hidden="true" />
                </MenuButton>
            </div>

            <transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
            >
                <MenuItems
                    class="absolute left-0 z-10 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                >
                    <div class="py-1">
                        <template v-if="can('upload_files')">
                            <MenuItem v-slot="{ active }">
                                <div>
                                    <a
                                        href="#"
                                        :class="[
                                            active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                            'group flex items-center px-4 py-2 text-sm',
                                        ]"
                                        @click="$refs.file.click()"
                                    >
                                        <ArrowUpOnSquareStackIcon
                                            class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500"
                                            aria-hidden="true"
                                        />
                                        {{ t("files.menu.upload_file") }}
                                    </a>
                                    <input
                                        type="file"
                                        ref="file"
                                        style="display: none"
                                        @change="onFilesSelected"
                                        multiple
                                    />
                                </div>
                            </MenuItem>
                            <MenuItem v-slot="{ active }">
                                <div>
                                    <a
                                        href="#"
                                        :class="[
                                            active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                            'group flex items-center px-4 py-2 text-sm',
                                        ]"
                                        @click="$refs.directory.click()"
                                    >
                                        <ArrowUpOnSquareStackIcon
                                            class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500"
                                            aria-hidden="true"
                                        />
                                        {{ t("files.menu.upload_folder") }}
                                    </a>
                                    <input
                                        type="file"
                                        ref="directory"
                                        style="display: none"
                                        @change="onDirectorySelected"
                                        webkitdirectory
                                        directory
                                        multiple
                                    />
                                </div>
                            </MenuItem>
                        </template>
                        <template v-if="can('create_files')">
                            <MenuItem v-slot="{ active }" @click="openCreateFileDialog('document')">
                                <a
                                    href="#"
                                    :class="[
                                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                        'group flex items-center px-4 py-2 text-sm',
                                    ]"
                                >
                                    <DocumentTextIcon
                                        class="mr-3 h-5 w-5 text-blue-500 group-hover:text-blue-500"
                                        aria-hidden="true"
                                    />
                                    {{ t("files.menu.document") }}
                                </a>
                            </MenuItem>
                            <MenuItem
                                v-slot="{ active }"
                                @click="openCreateFileDialog('spreadsheet')"
                            >
                                <a
                                    href="#"
                                    :class="[
                                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                        'group flex items-center px-4 py-2 text-sm',
                                    ]"
                                >
                                    <DocumentTextIcon
                                        class="mr-3 h-5 w-5 text-green-500 group-hover:text-green-500"
                                        aria-hidden="true"
                                    />
                                    {{ t("files.menu.spreadsheet") }}
                                </a>
                            </MenuItem>
                            <MenuItem
                                v-slot="{ active }"
                                @click="openCreateFileDialog('presentation')"
                            >
                                <a
                                    href="#"
                                    :class="[
                                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                        'group flex items-center px-4 py-2 text-sm',
                                    ]"
                                >
                                    <DocumentTextIcon
                                        class="mr-3 h-5 w-5 text-orange-500 group-hover:text-orange-500"
                                        aria-hidden="true"
                                    />
                                    {{ t("files.menu.presentation") }}
                                </a>
                            </MenuItem>
                            <!--     v-slot="{ active }" -->
                            <!--     @click="openCreateFileDialog('text')" -->
                            <!-- > -->
                            <!--         href="#" -->
                            <!--         :class="[ -->
                            <!--             active -->
                            <!--                 ? 'bg-gray-100 text-gray-900' -->
                            <!--                 : 'text-gray-700', -->
                            <!--             'group flex items-center px-4 py-2 text-sm', -->
                            <!--         ]" -->
                            <!--     > -->
                            <!--             class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500" -->
                            <!--             aria-hidden="true" -->
                            <!--         /> -->
                            <!--         {{ t("files.menu.text_file") }} -->
                            <MenuItem v-slot="{ active }" @click="openCreateFileDialog('folder')">
                                <a
                                    href="#"
                                    :class="[
                                        active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                        'group flex items-center px-4 py-2 text-sm',
                                    ]"
                                >
                                    <FolderIcon
                                        class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500"
                                        aria-hidden="true"
                                    />
                                    {{ t("files.menu.new_folder") }}
                                </a>
                            </MenuItem>
                        </template>
                    </div>
                </MenuItems>
            </transition>
        </Menu>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useRoute } from "vue-router";
import { useDialogStore } from "@/store/dialogs";
import { usePermissions } from "@/composables/usePermissions";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import {
    PlusIcon,
    DocumentTextIcon,
    FolderIcon,
    ArrowUpOnSquareStackIcon,
} from "@heroicons/vue/20/solid";

const emits = defineEmits(["filesSelected"]);
const route = useRoute();
const dialogStore = useDialogStore();
const { can } = usePermissions();

function openCreateFileDialog(fileType) {
    dialogStore.openFileDialog(fileType);
}

function onFilesSelected(event) {
    const selectedFiles = event.target.files;

    if (selectedFiles && selectedFiles.length) {
        emits("filesSelected", { files: selectedFiles, type: "file" });
    }
}

function onDirectorySelected(event) {
    const selectedFiles = event.target.files;

    if (selectedFiles && selectedFiles.length) {
        emits("filesSelected", { files: selectedFiles, type: "directory" });
    }
}
</script>
