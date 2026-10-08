<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Menu as="div" class="relative ml-1 flex" @click.stop>
        <MenuButton
            ref="button"
            class="flex items-center rounded p-0.5 text-gray-400 hover:bg-gray-200 hover:text-gray-600 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600"
        >
            <span class="sr-only">{{ t("projects.folder_view.actions") }}</span>
            <EllipsisHorizontalIcon class="h-3.5 w-3.5" aria-hidden="true" />
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
                    class="z-[60] w-48 rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                >
                    <MenuItem
                        v-for="action in actions"
                        :key="action.key"
                        v-slot="{ active, disabled }"
                        :disabled="action.disabled"
                    >
                        <button
                            type="button"
                            :class="[
                                active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                disabled ? 'cursor-not-allowed opacity-50' : '',
                                'group flex w-full items-center px-4 py-2 text-sm',
                            ]"
                            @click="emit(action.key, view)"
                        >
                            <component
                                :is="action.icon"
                                class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500"
                                aria-hidden="true"
                            />
                            {{ action.label }}
                        </button>
                    </MenuItem>
                </MenuItems>
            </transition>
        </Teleport>
    </Menu>
</template>

<script setup>
import { computed, ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import {
    EllipsisHorizontalIcon,
    LockClosedIcon,
    LockOpenIcon,
    PencilSquareIcon,
    TrashIcon,
} from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

const props = defineProps({
    view: { type: Object, required: true },
    canEdit: { type: Boolean, default: false },
    canDelete: { type: Boolean, default: false },
});

const emit = defineEmits(["toggle-visibility", "edit", "delete"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({
    placement: "bottom-start",
    anchor: button,
});

const actions = computed(() => {
    const isPublic = props.view.is_public !== false;

    return [
        {
            key: "toggle-visibility",
            label: isPublic
                ? t.value("projects.view.make_private")
                : t.value("projects.view.make_public"),
            icon: isPublic ? LockClosedIcon : LockOpenIcon,
            disabled: !props.canEdit,
        },
        {
            key: "edit",
            label: t.value("common.button.edit"),
            icon: PencilSquareIcon,
            disabled: !props.canEdit,
        },
        {
            key: "delete",
            label: t.value("common.button.delete"),
            icon: TrashIcon,
            disabled: !props.canDelete,
        },
    ];
});
</script>
