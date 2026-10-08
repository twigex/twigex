<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full w-full bg-white shadow-lg flex flex-col">
        <!-- Header (Fixed Height) -->
        <div
            class="flex items-center justify-between border-b border-gray-200"
            style="height: 55px"
        >
            <h2 class="text-sm font-medium text-gray-600 px-6">
                {{ t("projects.navigation.custom_cards_navigation.title") }}
            </h2>
            <button
                class="text-gray-500 hover:text-gray-700 px-4"
                @click="rightSideNavigation = false"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                    />
                </svg>
            </button>
        </div>

        <!-- Scrollable Content -->
        <div class="flex-1 overflow-y-auto py-4 px-6">
            <ul class="space-y-4">
                <li v-for="(header, index) in headers" :key="index">
                    <ToggleSwitch
                        :model-value="header.name === 'name' || header.visible !== false"
                        :label="header.display_name || header.name"
                        :disabled="header.name === 'name'"
                        @update:model-value="toggleHeader(index)"
                    />
                </li>
            </ul>
        </div>
    </div>
</template>
<script setup>
import { t } from "@/i18n/index.js";
import { computed } from "vue";
import { useWorkspaceStore } from "@/store/workspaces";
import ToggleSwitch from "@/components/ToggleSwitch.vue";
import workspaceService from "@/services/workspaceService";
import { useRoute } from "vue-router";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const route = useRoute();
const workspaceStore = useWorkspaceStore();

// Computed headers list
const headers = computed(() => workspaceStore.getTableHeaders);

// Function to toggle header visibility
const toggleHeader = (index) => {
    const header = headers.value[index];

    // Ensure 'name' header is always true
    if (header.name === "name") {
        return;
    }

    // A field with no visible value is shown, as the grid shows it.
    header.visible = header.visible === false;

    // Save the new value after toggling
    saveHeaderValue(index);
};

const rightSideNavigation = computed({
    get() {
        return workspaceStore.getRightNavigation;
    },
    set(value) {
        workspaceStore.setRightNavigation(value);
    },
});

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

// Function to save display name & visibility to backend
const saveHeaderValue = (index) => {
    if (route.name === "grid-view") {
        const header = headers.value[index];

        // Prevent saving if visible is undefined
        if (header.visible === undefined) {
            return;
        }

        workspaceService
            .updateDisplayName({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                view_id: route.params.fid,
                header_name: header.name,
                display_name: header.display_name,
                visible: header.visible, // Ensures only true/false values are sent
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    } else if (route.name === "kanban-view") {
        const header = headers.value[index];

        // Ensure `order` is an object before modifying it
        if (typeof selectedView.value.order === "string") {
            selectedView.value.order = JSON.parse(selectedView.value.order);
        }

        // Ensure `fieldsVisible` exists in `selectedView.order`
        if (!Array.isArray(selectedView.value.order.fieldsVisible)) {
            selectedView.value.order.fieldsVisible = [];
        }

        // Find if the header already exists in fieldsVisible
        const existingFieldIndex = selectedView.value.order.fieldsVisible.findIndex(
            (field) => field.name === header.name,
        );

        if (existingFieldIndex !== -1) {
            // Update existing header visibility
            selectedView.value.order.fieldsVisible[existingFieldIndex].visible = header.visible;
        } else {
            // Add new header visibility entry
            selectedView.value.order.fieldsVisible.push({
                name: header.name,
                visible: header.visible,
            });
        }

        // Ensure all headers are stored in `fieldsVisible`
        headers.value.forEach((hdr) => {
            const fieldExists = selectedView.value.order.fieldsVisible.some(
                (field) => field.name === hdr.name,
            );

            if (!fieldExists) {
                selectedView.value.order.fieldsVisible.push({
                    name: hdr.name,
                    visible: hdr.visible ?? false, // Default to false if undefined
                });
            }
        });

        // Convert `order` back to JSON string
        selectedView.value.order = JSON.stringify(selectedView.value.order);

        workspaceService
            .updateView({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                view_id: selectedView.value.id,
                order: selectedView.value.order,
                name: selectedView.value.name,
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }
};
</script>
