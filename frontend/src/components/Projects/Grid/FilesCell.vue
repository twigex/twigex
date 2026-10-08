<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex min-w-0 flex-1 items-center gap-x-1">
        <button
            v-if="files.length"
            type="button"
            class="inline-flex min-w-0 items-center gap-x-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 hover:bg-gray-200"
            :title="nameOf(files[0])"
            @click.stop="emit('open', files[0])"
        >
            <component
                :is="fileTypeOf(files[0]).icon"
                :class="[fileTypeOf(files[0]).color, 'h-3.5 w-3.5 flex-none']"
                aria-hidden="true"
            />
            <span class="truncate">{{ nameOf(files[0]) }}</span>
        </button>

        <Menu v-if="files.length > 1" as="div" class="flex flex-none" @click.stop>
            <MenuButton
                ref="button"
                class="rounded-full bg-white px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600"
                :title="t('projects.grid_view.show_files')"
            >
                +{{ files.length - 1 }}
            </MenuButton>
            <Teleport to="body">
                <transition
                    enter-active-class="transition ease-out duration-100"
                    enter-from-class="transform opacity-0 scale-95"
                    enter-to-class="transform opacity-100 scale-100"
                    leave-active-class="transition ease-in duration-75"
                    leave-from-class="transform opacity-100 scale-100"
                    leave-to-class="transform opacity-0 scale-95"
                >
                    <MenuItems
                        ref="floating"
                        :style="floatingStyles"
                        class="z-[60] max-h-60 w-64 overflow-y-auto rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                    >
                        <p
                            class="px-4 pb-1 pt-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500"
                        >
                            {{ t("projects.grid_view.files_count", { count: files.length }) }}
                        </p>
                        <MenuItem
                            v-for="(file, index) in files"
                            :key="file.id ?? index"
                            v-slot="{ active }"
                        >
                            <button
                                type="button"
                                :class="[
                                    active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                    'flex w-full min-w-0 items-center gap-x-2 px-4 py-2 text-left text-sm',
                                ]"
                                :title="nameOf(file)"
                                @click="emit('open', file)"
                            >
                                <component
                                    :is="fileTypeOf(file).icon"
                                    :class="[fileTypeOf(file).color, 'h-4 w-4 flex-none']"
                                    aria-hidden="true"
                                />
                                <span class="truncate">{{ nameOf(file) }}</span>
                            </button>
                        </MenuItem>
                    </MenuItems>
                </transition>
            </Teleport>
        </Menu>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { t } from "@/i18n/index.js";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { fileTypeOf } from "@/utils/projects/files";

defineProps({
    files: { type: Array, default: () => [] },
});

const emit = defineEmits(["open"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({
    placement: "bottom-start",
    anchor: button,
});

const nameOf = (file) =>
    file?.name || file?.filename || t.value("projects.full_task_navigation.unknow_file");
</script>
