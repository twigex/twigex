<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <button
        v-if="disabled"
        type="button"
        class="flex h-full w-full cursor-pointer items-center truncate text-left text-sm font-medium text-gray-800 focus:outline-none"
        style="padding: 9px 6px"
        :title="label || t('projects.grid_view.not_set')"
        @click.stop="emit('denied')"
    >
        {{ label || "—" }}
    </button>
    <Menu v-else as="div" class="h-full w-full" @click.stop>
        <MenuButton
            ref="button"
            class="flex h-full w-full cursor-pointer items-center truncate text-left text-sm font-medium text-gray-800 hover:bg-gray-50 focus:outline-none"
            style="padding: 9px 6px"
            :title="label || t('projects.grid_view.not_set')"
        >
            {{ label || "—" }}
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
                    class="z-[60] min-w-[10rem] rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                >
                    <MenuItem v-for="option in options" :key="option.value" v-slot="{ active }">
                        <button
                            type="button"
                            :class="[
                                active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                'flex w-full items-center justify-between px-4 py-2 text-left text-sm',
                            ]"
                            @click="emit('select', option.value)"
                        >
                            {{ option.label }}
                            <CheckIcon
                                v-if="option.value === value"
                                class="h-4 w-4 text-indigo-600"
                                aria-hidden="true"
                            />
                        </button>
                    </MenuItem>
                </MenuItems>
            </transition>
        </Teleport>
    </Menu>
</template>

<script setup>
import { ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { CheckIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

defineProps({
    value: { type: String, default: "" },
    label: { type: String, default: "" },
    options: { type: Array, default: () => [] },
    // Shown but not opened; a click is reported as denied instead.
    disabled: { type: Boolean, default: false },
});

const emit = defineEmits(["select", "denied"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({
    placement: "bottom-start",
    anchor: button,
});
</script>
