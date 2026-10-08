<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="visible"
        :style="{ top: `${menuPosition.y}px`, left: `${menuPosition.x}px` }"
        class="absolute bg-white shadow-md border rounded z-50"
        ref="menu"
    >
        <ul class="list-none p-2">
            <li
                class="py-1 px-4 hover:bg-gray-100 cursor-pointer relative"
                @mouseenter="showSubMenu"
                @mouseleave="hideSubMenu"
            >
                {{ t("collimato.dashboard.widget.drill_by") }}
                <!-- Submenu for dimensions -->
                <ul
                    v-if="subMenuVisible"
                    :style="submenuStyle"
                    class="absolute bg-white shadow-md border rounded z-50 list-none p-2 max-h-96 overflow-y-auto"
                    ref="submenu"
                >
                    <li
                        v-for="dimension in dimensions"
                        :key="dimension.name"
                        class="py-1 px-4 hover:bg-gray-100 cursor-pointer"
                        @click="handleDimensionClick(dimension)"
                    >
                        {{ dimension.name }}
                    </li>
                </ul>
            </li>
        </ul>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, onBeforeUnmount, nextTick } from "vue";

defineProps({
    dimensions: {
        type: Array,
        required: true,
    },
});

const emits = defineEmits(["drill-by"]);

const visible = ref(false);
const subMenuVisible = ref(false); // Toggle for submenu visibility
const menu = ref(null);
const submenu = ref(null);
const x = ref(0);
const y = ref(0);

const menuPosition = ref({ x: 0, y: 0 });
const submenuStyle = ref({ left: "100%", top: "0" });

// Function to show the context menu
function showContextMenu(position) {
    x.value = position.x;
    y.value = position.y;
    visible.value = true;

    // Adjust position after menu is visible
    nextTick(() => {
        adjustMenuPosition();
    });

    document.addEventListener("click", handleOutsideClick);
}

// Function to hide the context menu
function hideContextMenu() {
    visible.value = false;
    subMenuVisible.value = false;
    document.removeEventListener("click", handleOutsideClick);
}

// Function to handle dimension clicks
function handleDimensionClick(dimension) {
    hideContextMenu();
    emits("drill-by", dimension.name);
}

// Function to detect outside clicks
function handleOutsideClick(event) {
    if (menu.value && !menu.value.contains(event.target)) {
        hideContextMenu();
    }
}

// Adjust the position of the context menu to prevent overflow
function adjustMenuPosition() {
    const menuRect = menu.value.getBoundingClientRect();
    const menuWidth = menuRect.width;
    const menuHeight = menuRect.height;

    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    let newX = x.value;
    let newY = y.value;

    // Adjust if the menu overflows off the right side of the screen
    if (newX + menuWidth > viewportWidth) {
        newX = viewportWidth - menuWidth - 10; // Add some padding from the edge
    }

    // Adjust if the menu overflows off the bottom of the screen
    if (newY + menuHeight > viewportHeight) {
        newY = viewportHeight - menuHeight - 10; // Add some padding from the edge
    }

    menuPosition.value = { x: newX, y: newY };
}

// Adjust submenu positioning to avoid overflow
function adjustSubMenuPosition() {
    if (submenu.value) {
        const submenuRect = submenu.value.getBoundingClientRect();
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        let leftPosition = "100%"; // Default position
        let topPosition = "0"; // Default position

        // Check if the submenu overflows on the right
        if (submenuRect.right > viewportWidth) {
            leftPosition = `-${submenuRect.width}px`; // Shift the submenu to the left
        }

        // Check if the submenu overflows on the bottom
        if (submenuRect.bottom > viewportHeight) {
            const overflow = submenuRect.bottom - viewportHeight;

            topPosition = `-${overflow}px`; // Adjust submenu upwards
        }

        submenuStyle.value = { left: leftPosition, top: topPosition };
    }
}

// Function called to show submenu and adjust position
function showSubMenu() {
    // Reset the submenu position before showing
    submenuStyle.value = { left: "100%", top: "0" }; // Reset position to avoid retaining old values

    subMenuVisible.value = true;
    nextTick(() => {
        adjustSubMenuPosition(); // Adjust position after submenu becomes visible
    });
}

// Function to hide the submenu
function hideSubMenu() {
    subMenuVisible.value = false;
}

// Remove the listener when the component is destroyed to avoid memory leaks
onBeforeUnmount(() => {
    document.removeEventListener("click", handleOutsideClick);
});

// Expose the showContextMenu method to the parent component
defineExpose({
    showContextMenu,
});
</script>

<style scoped>
.absolute {
    position: absolute;
}
.relative {
    position: relative;
}
</style>
