<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="relative inline-block"
        @mouseenter="startShowTimer"
        @mouseleave="startHideTimer"
        ref="triggerEl"
    >
        <slot />

        <Teleport to="body">
            <TransitionRoot :show="visible" as="template">
                <TransitionChild
                    as="div"
                    class="fixed bg-gray-800 text-white text-sm rounded py-2 px-4 z-[9999]"
                    :style="fixedStyle"
                    @mouseenter="cancelHideTimer"
                    @mouseleave="startHideTimer"
                    enter="transition ease-out duration-200"
                    enter-from="opacity-0 scale-95"
                    enter-to="opacity-100 scale-100"
                    leave="transition ease-in duration-150"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <slot name="text" />
                </TransitionChild>
            </TransitionRoot>
        </Teleport>
    </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from "vue";
import { TransitionRoot, TransitionChild } from "@headlessui/vue";

const props = defineProps({
    position: { type: String, default: "top" },
    offset: { type: Number, default: 8 },
    showDelay: { type: Number, default: 200 },
    hideDelay: { type: Number, default: 300 },
});

const visible = ref(false);
const triggerEl = ref(null);
const coords = ref({ top: 0, left: 0 });
let showTimeout = null;
let hideTimeout = null;

function updatePosition() {
    const r = triggerEl.value?.getBoundingClientRect();

    if (!r) return;
    let top, left, transform;

    switch (props.position) {
        case "top":
            top = r.top - props.offset;
            left = r.left + r.width / 2;
            transform = "translate(-50%, -100%)";
            break;
        case "bottom":
            top = r.bottom + props.offset;
            left = r.left + r.width / 2;
            transform = "translate(-50%, 0)";
            break;
        case "left":
            top = r.top + r.height / 2;
            left = r.left - props.offset;
            transform = "translate(-100%, -50%)";
            break;
        case "right":
            top = r.top + r.height / 2;
            left = r.right + props.offset;
            transform = "translate(0, -50%)";
            break;
    }

    coords.value = { top, left, transform };
}

const fixedStyle = computed(() => ({
    top: `${coords.value.top ?? 0}px`,
    left: `${coords.value.left ?? 0}px`,
    transform: coords.value.transform,
    whiteSpace: "nowrap",
}));

function startShowTimer() {
    clearTimeout(hideTimeout);
    showTimeout = setTimeout(() => {
        visible.value = true;
        updatePosition();
        window.addEventListener("scroll", updatePosition, true);
        window.addEventListener("resize", updatePosition);
    }, props.showDelay);
}

function startHideTimer() {
    clearTimeout(showTimeout);
    hideTimeout = setTimeout(() => {
        visible.value = false;
        window.removeEventListener("scroll", updatePosition, true);
        window.removeEventListener("resize", updatePosition);
    }, props.hideDelay);
}

function cancelHideTimer() {
    clearTimeout(hideTimeout);
}

onBeforeUnmount(() => {
    clearTimeout(showTimeout);
    clearTimeout(hideTimeout);
    window.removeEventListener("scroll", updatePosition, true);
    window.removeEventListener("resize", updatePosition);
});
</script>
