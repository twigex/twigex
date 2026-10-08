<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Teleport to="body">
        <div
            v-if="visible"
            ref="popoverContent"
            :style="menuStyles"
            class="popover z-50 rounded-md bg-white shadow-lg ring-1 ring-black/5"
            @pointerdown.stop="closeParentMenu?.()"
        >
            <Menu as="div" class="relative inline-block text-left">
                <div class="py-1">
                    <MenuItem v-for="item in items" :key="item.label" v-slot="{ active }">
                        <a
                            href="#"
                            :class="[
                                item.disabled
                                    ? 'opacity-40 cursor-not-allowed text-gray-400'
                                    : active
                                      ? 'bg-gray-100 text-gray-900'
                                      : 'text-gray-700',
                                'group flex items-center px-4 py-2 text-sm',
                            ]"
                            @click.prevent="!item.disabled && item.action()"
                        >
                            <component
                                :is="item.icon"
                                :class="[
                                    active ? 'text-gray-500' : '',
                                    'mr-3 h-4 w-4 text-gray-400 pointer-events-none',
                                ]"
                                aria-hidden="true"
                            />
                            {{ item.label }}
                        </a>
                    </MenuItem>
                </div>
            </Menu>
        </div>
    </Teleport>
</template>

<script setup>
import { ref, toRef, onMounted, onBeforeUnmount } from "vue";
import { Menu, MenuItem } from "@headlessui/vue";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

// The menu a row, field, view or project item opens: its actions, anchored
// to what opened it, closed by a click anywhere else.
const props = defineProps({
    visible: { type: Boolean, default: false },
    // What the menu opens from: an element, or a point for a right click.
    anchor: { type: Object, default: null },
    items: { type: Array, default: () => [] },
    closeParentMenu: { type: Function, default: null },
});

const emit = defineEmits(["close"]);

const popoverContent = ref(null);
const { floatingStyles: menuStyles } = useAnchoredPopup({
    floating: popoverContent,
    anchor: toRef(props, "anchor"),
});

function handleClickOutside(event) {
    if (!props.visible) return;
    if (popoverContent.value?.contains(event.target)) return;
    emit("close");
}

onMounted(() => document.addEventListener("mousedown", handleClickOutside));
onBeforeUnmount(() => document.removeEventListener("mousedown", handleClickOutside));
</script>
