<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Menu as="div" class="relative inline-block text-left">
        <MenuButton
            ref="reference"
            :disabled="disabled"
            :class="['focus:outline-none', disabled ? 'cursor-not-allowed opacity-50' : '']"
        >
            <slot name="button">
                <div
                    class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-xs ring-1 ring-gray-300 hover:bg-gray-50"
                >
                    Options
                    <ChevronDownIcon class="-mr-1 h-5 w-5 text-gray-400" aria-hidden="true" />
                </div>
            </slot>
        </MenuButton>

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
                :class="[
                    'absolute right-0 z-[60] mt-2 w-max origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none',
                    sizeClasses[size],
                ]"
            >
                <slot name="items">
                    <div v-if="items?.length" class="py-1">
                        <MenuItem v-for="(item, i) in items" :key="i" v-slot="{ active }">
                            <button
                                type="button"
                                @click="onItemClick(item)"
                                :class="[
                                    active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                    'flex w-full items-center gap-x-2 px-4 py-2 text-left text-sm',
                                ]"
                            >
                                <component
                                    v-if="item.icon"
                                    :is="item.icon"
                                    class="h-4 w-4 text-gray-400"
                                />
                                <span>{{ item.label }}</span>
                            </button>
                        </MenuItem>
                    </div>
                </slot>
            </MenuItems>
        </transition>
    </Menu>
</template>

<script setup>
import { ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { ChevronDownIcon } from "@heroicons/vue/20/solid";
import { useFloating, autoPlacement } from "@floating-ui/vue";
import { autoUpdate } from "@floating-ui/dom";

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    transform: false,
    middleware: [
        autoPlacement({
            allowedPlacements: ["bottom-end", "bottom-start", "top-end", "top-start"],
        }),
    ],
    whileElementsMounted: autoUpdate,
});

defineProps({
    items: {
        type: Array,
        default: () => [],
    },
    size: {
        type: String,
        default: "auto",
        validator: (v) => ["auto", "sm", "md", "lg"].includes(v),
    },
    disabled: {
        type: Boolean,
        default: false,
    },
});

const sizeClasses = {
    auto: "",
    sm: "min-w-40",
    md: "min-w-56",
    lg: "min-w-72",
};

const emit = defineEmits(["select"]);

function onItemClick(item) {
    emit("select", item);
}
</script>
