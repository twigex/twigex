<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('main.sections.files')">
        <template #sidebar="{ mini }">
            <FilesNavigation :mini="mini" />
        </template>

        <div class="flex h-full w-full overflow-hidden">
            <div
                v-if="!can('view_files')"
                class="flex flex-1 flex-col items-center justify-center h-full text-center px-6"
            >
                <LockClosedIcon class="h-12 w-12 text-gray-300" aria-hidden="true" />
                <h3 class="mt-4 text-base font-semibold text-gray-900">Access restricted</h3>
                <p class="mt-1 text-sm text-gray-500 max-w-sm">
                    You don't have permission to view files. Contact your administrator to request
                    access.
                </p>
            </div>

            <template v-else>
                <!-- min-w-0 lets the list shrink so the details panel fits beside it instead of overflowing -->
                <div class="flex-1 min-w-0 h-full overflow-hidden">
                    <RouterView />
                </div>
                <FileDetails />
                <MobileDetails />
            </template>
        </div>
    </SidebarLayout>
</template>

<script setup>
import { RouterView } from "vue-router";
import { t } from "@/i18n/index.js";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import FilesNavigation from "@/components/Files/FilesNavigation.vue";
import { usePermissions } from "@/composables/usePermissions";
import FileDetails from "@/components/Files/Details/FileDetails.vue";
import MobileDetails from "@/components/Files/Details/MobileDetails.vue";
import { LockClosedIcon } from "@heroicons/vue/24/outline";

const { can } = usePermissions();
</script>
